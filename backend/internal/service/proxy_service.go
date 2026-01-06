/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-14 15:32:13
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-24 17:21:23
 * @FilePath            : frp-web-testbackendinternalserviceproxy_service.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"frp-web-panel/internal/model"
	"frp-web-panel/internal/repository"
	"frp-web-panel/pkg/database"
	"log"
	"math/rand"
	"time"
)

const (
	// 随机端口范围
	minRandomPort = 10000
	maxRandomPort = 65535
)

type ProxyService struct {
	proxyRepo      *repository.ProxyRepository
	clientRepo     *repository.ClientRepository
	dnsService     *DNSService
	frpServerRepo  *repository.FrpServerRepository
	certRepo       *repository.CertificateRepository
	settingService *SettingService
}

func NewProxyService() *ProxyService {
	return &ProxyService{
		proxyRepo:      repository.NewProxyRepository(),
		clientRepo:     repository.NewClientRepository(),
		dnsService:     NewDNSService(),
		frpServerRepo:  repository.NewFrpServerRepository(database.DB),
		certRepo:       repository.NewCertificateRepository(),
		settingService: NewSettingService(),
	}
}

func (s *ProxyService) GetProxiesByClient(clientID uint) ([]model.Proxy, error) {
	return s.proxyRepo.FindByClientID(clientID)
}

// GetAllProxies 获取所有代理列表
func (s *ProxyService) GetAllProxies() ([]model.Proxy, error) {
	return s.proxyRepo.FindAll()
}

func (s *ProxyService) GetProxy(id uint) (*model.Proxy, error) {
	return s.proxyRepo.FindByID(id)
}

func (s *ProxyService) CreateProxy(proxy *model.Proxy) error {
	// 对于 TCP 和 UDP 类型的代理，如果远程端口为 0，则自动分配一个随机可用端口
	if (proxy.Type == "tcp" || proxy.Type == "udp") && proxy.RemotePort == 0 {
		randomPort, err := s.generateRandomPort()
		if err != nil {
			log.Printf("[代理创建] ⚠️ 自动分配端口失败: %v，将使用默认值 0", err)
		} else {
			proxy.RemotePort = randomPort
			log.Printf("[代理创建] ✅ 自动分配远程端口: %d", randomPort)
		}
	}

	// HTTPS 类型代理必须选择证书
	if proxy.Type == "https" && proxy.CertID == nil {
		return fmt.Errorf("HTTPS 代理必须选择证书")
	}

	// 如果选择了证书，验证证书是否存在且有效
	if proxy.CertID != nil {
		cert, err := s.certRepo.FindByID(*proxy.CertID)
		if err != nil {
			return fmt.Errorf("证书不存在")
		}
		if cert.Status != model.CertStatusActive {
			return fmt.Errorf("所选证书无效，请选择有效的证书")
		}
	}

	// 创建代理
	if err := s.proxyRepo.Create(proxy); err != nil {
		return err
	}

	// 如果启用了 DNS 同步且有自定义域名，则同步 DNS 记录
	if proxy.EnableDNSSync && proxy.CustomDomains != "" {
		go s.syncDNSRecordAsync(proxy)
	}

	return nil
}

// syncDNSRecordAsync 异步同步 DNS 记录
func (s *ProxyService) syncDNSRecordAsync(proxy *model.Proxy) {
	log.Printf("[DNS同步] 开始为代理 %s (ID=%d) 同步 DNS 记录, 域名: %s", proxy.Name, proxy.ID, proxy.CustomDomains)

	// 获取客户端信息以获取 FRP 服务器 ID
	client, err := s.clientRepo.FindByID(proxy.ClientID)
	if err != nil {
		log.Printf("[DNS同步] ❌ 获取客户端信息失败: %v", err)
		return
	}

	// 同步 DNS 记录
	if err := s.dnsService.SyncDNSRecord(proxy, client.FrpServerID); err != nil {
		log.Printf("[DNS同步] ❌ 同步 DNS 记录失败: %v", err)
	} else {
		log.Printf("[DNS同步] ✅ DNS 记录同步成功")
	}
}

// generateRandomPort 生成一个随机可用的远程端口
// 端口范围: 10000-65535
// 会检查端口是否已被其他代理使用
func (s *ProxyService) generateRandomPort() (int, error) {
	// 获取已使用的端口列表
	usedPorts, err := s.proxyRepo.GetUsedRemotePorts()
	if err != nil {
		return 0, fmt.Errorf("获取已使用端口失败: %v", err)
	}

	// 将已使用端口转换为 map 以便快速查找
	usedPortMap := make(map[int]bool)
	for _, port := range usedPorts {
		usedPortMap[port] = true
	}

	// 初始化随机数生成器
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// 尝试生成随机端口，最多尝试 100 次
	for i := 0; i < 100; i++ {
		port := rng.Intn(maxRandomPort-minRandomPort+1) + minRandomPort
		if !usedPortMap[port] {
			return port, nil
		}
	}

	// 如果随机尝试失败，顺序查找可用端口
	for port := minRandomPort; port <= maxRandomPort; port++ {
		if !usedPortMap[port] {
			return port, nil
		}
	}

	return 0, fmt.Errorf("没有可用的端口")
}

func (s *ProxyService) UpdateProxy(proxy *model.Proxy) error {
	// 获取旧的代理信息
	oldProxy, err := s.proxyRepo.FindByID(proxy.ID)
	if err != nil {
		log.Printf("[代理更新] ⚠️ 获取旧代理信息失败: %v", err)
	}

	// 处理 DNS 同步逻辑 - 在更新代理之前先删除旧的 DNS 记录
	if oldProxy != nil {
		// 如果之前启用了 DNS 同步但现在禁用了，或者域名变了，同步删除旧的 DNS 记录
		if oldProxy.EnableDNSSync && (!proxy.EnableDNSSync || oldProxy.CustomDomains != proxy.CustomDomains) {
			log.Printf("[代理更新] 域名变更或DNS同步关闭，同步删除旧DNS记录: %s -> %s", oldProxy.CustomDomains, proxy.CustomDomains)
			if err := s.dnsService.DeleteDNSRecord(oldProxy.ID); err != nil {
				log.Printf("[代理更新] ⚠️ 删除旧DNS记录失败: %v", err)
				// 继续执行，不阻塞代理更新
			} else {
				log.Printf("[代理更新] ✅ 旧DNS记录删除成功")
			}
		}
	}

	// 更新代理
	if err := s.proxyRepo.Update(proxy); err != nil {
		return err
	}

	// 如果现在启用了 DNS 同步且有自定义域名，异步同步新的 DNS 记录
	if proxy.EnableDNSSync && proxy.CustomDomains != "" {
		go s.syncDNSRecordAsync(proxy)
	}

	return nil
}

// DeleteProxy 删除代理
// deleteDNS: 是否同时删除关联的 DNS 记录
func (s *ProxyService) DeleteProxy(id uint, deleteDNS bool) error {
	// 先获取代理信息，用于删除 DNS 记录
	proxy, err := s.proxyRepo.FindByID(id)
	if err != nil {
		log.Printf("[代理删除] ⚠️ 获取代理信息失败: %v", err)
	} else if proxy.EnableDNSSync && deleteDNS {
		// 同步删除 DNS 记录
		log.Printf("[代理删除] 开始删除代理 %s (ID=%d) 的 DNS 记录", proxy.Name, proxy.ID)
		if err := s.dnsService.DeleteDNSRecord(proxy.ID); err != nil {
			log.Printf("[代理删除] ❌ 删除 DNS 记录失败: %v", err)
		} else {
			log.Printf("[代理删除] ✅ DNS 记录删除成功")
		}
	} else if proxy.EnableDNSSync && !deleteDNS {
		log.Printf("[代理删除] 用户选择保留 DNS 记录，跳过删除")
	}

	return s.proxyRepo.Delete(id)
}

// ToggleProxy 切换代理的启用/禁用状态
func (s *ProxyService) ToggleProxy(id uint) (*model.Proxy, error) {
	return s.proxyRepo.ToggleEnabled(id)
}

// ExportClientConfig 导出客户端配置（TOML格式）
// 生成完整的 frpc TOML 配置文件，包含：
// - 基础连接配置（serverAddr, serverPort, user, auth.token）
// - 日志配置（log.to, log.level, log.maxDays）
// - Web管理界面配置（webServer.*）
// - 所有启用的代理配置
func (s *ProxyService) ExportClientConfig(clientID uint) (string, error) {
	log.Printf("[配置导出] ========== 开始导出客户端 %d 的 TOML 配置 ==========", clientID)

	client, err := s.clientRepo.FindByID(clientID)
	if err != nil {
		log.Printf("[配置导出] ❌ 获取客户端信息失败: %v", err)
		return "", err
	}
	log.Printf("[配置导出] 客户端信息: Name=%s, ServerAddr=%s, ServerPort=%d",
		client.Name, client.ServerAddr, client.ServerPort)

	// 只获取启用的代理
	proxies, err := s.proxyRepo.FindEnabledByClientID(clientID)
	if err != nil {
		log.Printf("[配置导出] ❌ 获取代理列表失败: %v", err)
		return "", err
	}

	log.Printf("[配置导出] 找到 %d 个启用的代理:", len(proxies))
	for i, p := range proxies {
		log.Printf("[配置导出]   [%d] ID=%d, Name=%s, Type=%s, LocalPort=%d, RemotePort=%d, Enabled=%v",
			i+1, p.ID, p.Name, p.Type, p.LocalPort, p.RemotePort, p.Enabled)
	}

	var buf bytes.Buffer

	// ==================== 基础配置 ====================
	buf.WriteString("# FRP 客户端配置文件 (TOML格式)\n")
	buf.WriteString("# 由 FRP Web Panel 自动生成\n\n")

	// 服务器连接配置
	buf.WriteString(fmt.Sprintf("serverAddr = \"%s\"\n", client.ServerAddr))
	buf.WriteString(fmt.Sprintf("serverPort = %d\n", client.ServerPort))
	buf.WriteString(fmt.Sprintf("user = \"%s\"\n", client.Name))
	buf.WriteString("\n")

	// ==================== 认证配置 ====================
	if client.Token != "" {
		buf.WriteString(fmt.Sprintf("auth.token = \"%s\"\n", client.Token))
		buf.WriteString("\n")
	}

	// ==================== 日志配置 ====================
	// 重要：这些配置必须保留，否则更新代理后会丢失
	buf.WriteString("# 日志配置\n")
	buf.WriteString("log.to = \"/opt/frpc/frpc.log\"\n")
	buf.WriteString("log.level = \"info\"\n")
	buf.WriteString("log.maxDays = 7\n")
	buf.WriteString("\n")

	// ==================== Web管理界面配置 ====================
	if client.FrpcAdminPort > 0 {
		adminAddr := client.FrpcAdminHost
		if adminAddr == "" {
			adminAddr = "127.0.0.1" // 默认监听本地
		}
		buf.WriteString("# Web管理界面配置\n")
		buf.WriteString(fmt.Sprintf("webServer.addr = \"%s\"\n", adminAddr))
		buf.WriteString(fmt.Sprintf("webServer.port = %d\n", client.FrpcAdminPort))
		if client.FrpcAdminUser != "" {
			buf.WriteString(fmt.Sprintf("webServer.user = \"%s\"\n", client.FrpcAdminUser))
		}
		if client.FrpcAdminPwd != "" {
			buf.WriteString(fmt.Sprintf("webServer.password = \"%s\"\n", client.FrpcAdminPwd))
		}
		buf.WriteString("\n")
		log.Printf("[配置导出] WebServer 配置: addr=%s, port=%d, user=%s",
			adminAddr, client.FrpcAdminPort, client.FrpcAdminUser)
	} else {
		log.Printf("[配置导出] ⚠️ 客户端未配置 WebServer (FrpcAdminPort=0)")
	}

	// ==================== 代理配置 ====================
	for _, proxy := range proxies {
		buf.WriteString(fmt.Sprintf("# 代理: %s\n", proxy.Name))
		buf.WriteString(fmt.Sprintf("[[proxies]]\n"))
		buf.WriteString(fmt.Sprintf("name = \"%s\"\n", proxy.Name))
		buf.WriteString(fmt.Sprintf("type = \"%s\"\n", proxy.Type))
		buf.WriteString(fmt.Sprintf("localIP = \"%s\"\n", proxy.LocalIP))
		buf.WriteString(fmt.Sprintf("localPort = %d\n", proxy.LocalPort))

		// TCP/UDP 类型的远程端口
		if proxy.RemotePort > 0 {
			buf.WriteString(fmt.Sprintf("remotePort = %d\n", proxy.RemotePort))
		}

		// HTTP/HTTPS 类型的自定义域名
		if proxy.CustomDomains != "" {
			buf.WriteString(fmt.Sprintf("customDomains = [\"%s\"]\n", proxy.CustomDomains))
		}

		// 子域名
		if proxy.Subdomain != "" {
			buf.WriteString(fmt.Sprintf("subdomain = \"%s\"\n", proxy.Subdomain))
		}

		// HTTP 路由
		if proxy.Locations != "" {
			buf.WriteString(fmt.Sprintf("locations = [\"%s\"]\n", proxy.Locations))
		}

		// Host Header 重写
		if proxy.HostHeaderRewrite != "" {
			buf.WriteString(fmt.Sprintf("hostHeaderRewrite = \"%s\"\n", proxy.HostHeaderRewrite))
		}

		// HTTP 基本认证
		if proxy.HttpUser != "" {
			buf.WriteString(fmt.Sprintf("httpUser = \"%s\"\n", proxy.HttpUser))
		}
		if proxy.HttpPassword != "" {
			buf.WriteString(fmt.Sprintf("httpPassword = \"%s\"\n", proxy.HttpPassword))
		}

		// STCP/SUDP 密钥
		if proxy.SecretKey != "" {
			buf.WriteString(fmt.Sprintf("secretKey = \"%s\"\n", proxy.SecretKey))
		}

		// 允许的用户
		if proxy.AllowUsers != "" {
			buf.WriteString(fmt.Sprintf("allowUsers = [\"%s\"]\n", proxy.AllowUsers))
		}

		// 带宽限制
		if proxy.BandwidthLimit != "" {
			buf.WriteString(fmt.Sprintf("transport.bandwidthLimit = \"%s\"\n", proxy.BandwidthLimit))
			if proxy.BandwidthLimitMode != "" {
				buf.WriteString(fmt.Sprintf("transport.bandwidthLimitMode = \"%s\"\n", proxy.BandwidthLimitMode))
			}
		}

		// 加密和压缩
		if proxy.UseEncryption {
			buf.WriteString("transport.useEncryption = true\n")
		}
		if proxy.UseCompression {
			buf.WriteString("transport.useCompression = true\n")
		}

		// 健康检查
		if proxy.HealthCheckType != "" {
			buf.WriteString(fmt.Sprintf("healthCheck.type = \"%s\"\n", proxy.HealthCheckType))
			if proxy.HealthCheckTimeout > 0 {
				buf.WriteString(fmt.Sprintf("healthCheck.timeoutSeconds = %d\n", proxy.HealthCheckTimeout))
			}
			if proxy.HealthCheckInterval > 0 {
				buf.WriteString(fmt.Sprintf("healthCheck.intervalSeconds = %d\n", proxy.HealthCheckInterval))
			}
		}

		// 插件配置
		if proxy.PluginType != "" && proxy.PluginConfig != "" {
			buf.WriteString(fmt.Sprintf("[proxies.plugin]\n"))
			buf.WriteString(fmt.Sprintf("type = \"%s\"\n", proxy.PluginType))

			// 根据插件类型解析并输出配置
			switch proxy.PluginType {
			case model.PluginTypeHTTPProxy:
				var cfg model.HTTPProxyPluginConfig
				if err := json.Unmarshal([]byte(proxy.PluginConfig), &cfg); err == nil {
					if cfg.HttpUser != "" {
						buf.WriteString(fmt.Sprintf("httpUser = \"%s\"\n", cfg.HttpUser))
					}
					if cfg.HttpPassword != "" {
						buf.WriteString(fmt.Sprintf("httpPassword = \"%s\"\n", cfg.HttpPassword))
					}
				}
			case model.PluginTypeSocks5:
				var cfg model.Socks5PluginConfig
				if err := json.Unmarshal([]byte(proxy.PluginConfig), &cfg); err == nil {
					if cfg.Username != "" {
						buf.WriteString(fmt.Sprintf("username = \"%s\"\n", cfg.Username))
					}
					if cfg.Password != "" {
						buf.WriteString(fmt.Sprintf("password = \"%s\"\n", cfg.Password))
					}
				}
			case model.PluginTypeStaticFile:
				var cfg model.StaticFilePluginConfig
				if err := json.Unmarshal([]byte(proxy.PluginConfig), &cfg); err == nil {
					if cfg.LocalPath != "" {
						buf.WriteString(fmt.Sprintf("localPath = \"%s\"\n", cfg.LocalPath))
					}
					if cfg.StripPrefix != "" {
						buf.WriteString(fmt.Sprintf("stripPrefix = \"%s\"\n", cfg.StripPrefix))
					}
					if cfg.HttpUser != "" {
						buf.WriteString(fmt.Sprintf("httpUser = \"%s\"\n", cfg.HttpUser))
					}
					if cfg.HttpPassword != "" {
						buf.WriteString(fmt.Sprintf("httpPassword = \"%s\"\n", cfg.HttpPassword))
					}
				}
			case model.PluginTypeUnixDomainSocket:
				var cfg model.UnixDomainSocketPluginConfig
				if err := json.Unmarshal([]byte(proxy.PluginConfig), &cfg); err == nil {
					if cfg.UnixPath != "" {
						buf.WriteString(fmt.Sprintf("unixPath = \"%s\"\n", cfg.UnixPath))
					}
				}
			case model.PluginTypeHTTPS2HTTP:
				var cfg model.HTTPS2HTTPPluginConfig
				if err := json.Unmarshal([]byte(proxy.PluginConfig), &cfg); err == nil {
					if cfg.LocalAddr != "" {
						buf.WriteString(fmt.Sprintf("localAddr = \"%s\"\n", cfg.LocalAddr))
					}
					// 如果启用自动证书，使用证书表中的路径
					crtPath, keyPath := s.getCertPaths(&proxy, cfg.CrtPath, cfg.KeyPath)
					if crtPath != "" {
						buf.WriteString(fmt.Sprintf("crtPath = \"%s\"\n", crtPath))
					}
					if keyPath != "" {
						buf.WriteString(fmt.Sprintf("keyPath = \"%s\"\n", keyPath))
					}
					if cfg.HostHeaderRewrite != "" {
						buf.WriteString(fmt.Sprintf("hostHeaderRewrite = \"%s\"\n", cfg.HostHeaderRewrite))
					}
				}
			case model.PluginTypeHTTPS2HTTPS:
				var cfg model.HTTPS2HTTPSPluginConfig
				if err := json.Unmarshal([]byte(proxy.PluginConfig), &cfg); err == nil {
					if cfg.LocalAddr != "" {
						buf.WriteString(fmt.Sprintf("localAddr = \"%s\"\n", cfg.LocalAddr))
					}
					// 如果启用自动证书，使用证书表中的路径
					crtPath, keyPath := s.getCertPaths(&proxy, cfg.CrtPath, cfg.KeyPath)
					if crtPath != "" {
						buf.WriteString(fmt.Sprintf("crtPath = \"%s\"\n", crtPath))
					}
					if keyPath != "" {
						buf.WriteString(fmt.Sprintf("keyPath = \"%s\"\n", keyPath))
					}
					if cfg.HostHeaderRewrite != "" {
						buf.WriteString(fmt.Sprintf("hostHeaderRewrite = \"%s\"\n", cfg.HostHeaderRewrite))
					}
				}
			}
			log.Printf("[配置导出] 代理 %s 使用插件: %s", proxy.Name, proxy.PluginType)
		}

		buf.WriteString("\n")
	}

	configStr := buf.String()
	log.Printf("[配置导出] 生成的完整 TOML 配置:\n%s", configStr)
	log.Printf("[配置导出] ========== 配置导出完成 ==========")
	return configStr, nil
}

// getCertPaths 获取证书路径
func (s *ProxyService) getCertPaths(proxy *model.Proxy, defaultCrtPath, defaultKeyPath string) (string, string) {
	log.Printf("[证书路径诊断] 代理 %s: CertID=%v, defaultCrtPath=%s, defaultKeyPath=%s",
		proxy.Name, proxy.CertID, defaultCrtPath, defaultKeyPath)

	// 如果没有关联证书，使用默认路径
	if proxy.CertID == nil {
		log.Printf("[证书路径诊断] 代理 %s: CertID=nil，使用默认路径", proxy.Name)
		return defaultCrtPath, defaultKeyPath
	}

	// 获取证书信息
	cert, err := s.certRepo.FindByID(*proxy.CertID)
	if err != nil {
		log.Printf("[证书路径诊断] 代理 %s: 获取证书失败 (CertID=%d): %v，使用默认路径", proxy.Name, *proxy.CertID, err)
		return defaultCrtPath, defaultKeyPath
	}
	if cert == nil {
		log.Printf("[证书路径诊断] 代理 %s: 证书记录不存在 (CertID=%d)，使用默认路径", proxy.Name, *proxy.CertID)
		return defaultCrtPath, defaultKeyPath
	}

	log.Printf("[证书路径诊断] 代理 %s: 证书状态=%s, 域名=%s", proxy.Name, cert.Status, cert.Domain)

	if cert.Status != model.CertStatusActive {
		log.Printf("[证书路径诊断] 代理 %s: 证书状态不是 active (当前=%s)，使用默认路径", proxy.Name, cert.Status)
		return defaultCrtPath, defaultKeyPath
	}

	// 使用约定的证书路径（证书文件由 daemon 同步）
	domain := cert.Domain
	crtPath := fmt.Sprintf("/opt/frpc/certs/%s.crt", domain)
	keyPath := fmt.Sprintf("/opt/frpc/certs/%s.key", domain)

	log.Printf("[证书路径诊断] 代理 %s 使用证书路径: crt=%s, key=%s", proxy.Name, crtPath, keyPath)
	return crtPath, keyPath
}
