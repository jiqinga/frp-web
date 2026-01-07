package service

import (
	"context"
	"fmt"
	"frp-web-panel/internal/logger"
	"frp-web-panel/internal/model"
	"frp-web-panel/internal/repository"
	"frp-web-panel/pkg/database"
	"time"
)

// DNSOperator DNS 操作接口，所有 DNS 提供商都需要实现此接口
type DNSOperator interface {
	// AddRecord 添加 DNS A 记录，返回记录 ID
	AddRecord(ctx context.Context, domain, ip string) (string, error)
	// UpdateRecord 更新 DNS A 记录
	UpdateRecord(ctx context.Context, domain, ip, recordID string) error
	// DeleteRecord 删除 DNS A 记录
	DeleteRecord(ctx context.Context, domain, recordID string) error
	// TestConnection 测试 API 连接
	TestConnection(ctx context.Context) error
	// ListDomains 获取该提供商下托管的所有域名列表
	ListDomains(ctx context.Context) ([]string, error)
	// AddTXTRecord 添加 DNS TXT 记录（用于 ACME DNS-01 验证）
	AddTXTRecord(ctx context.Context, domain, value string) (string, error)
	// DeleteTXTRecord 删除 DNS TXT 记录
	DeleteTXTRecord(ctx context.Context, domain, recordID string) error
}

type DNSService struct {
	providerRepo  *repository.DNSProviderRepository
	recordRepo    *repository.DNSRecordRepository
	frpServerRepo *repository.FrpServerRepository
	eventNotifier *SystemEventNotifier
}

func NewDNSService() *DNSService {
	return &DNSService{
		providerRepo:  repository.NewDNSProviderRepository(),
		recordRepo:    repository.NewDNSRecordRepository(),
		frpServerRepo: repository.NewFrpServerRepository(database.DB),
	}
}

// SetEventNotifier 设置系统事件通知器
func (s *DNSService) SetEventNotifier(notifier *SystemEventNotifier) {
	s.eventNotifier = notifier
}

// CreateDNSOperator 根据提供商类型创建对应的 DNS 操作实例
func (s *DNSService) CreateDNSOperator(provider *model.DNSProvider) (DNSOperator, error) {
	switch provider.Type {
	case model.DNSProviderTypeAliyun:
		return NewAliyunDNSServiceV2(provider.AccessKey, provider.SecretKey)
	case model.DNSProviderTypeCloudflare:
		// Cloudflare 使用 API Token，存储在 AccessKey 字段
		return NewCloudflareDNSService(provider.AccessKey)
	case model.DNSProviderTypeTencent:
		return NewTencentDNSService(provider.AccessKey, provider.SecretKey)
	default:
		return nil, fmt.Errorf("不支持的 DNS 提供商类型: %s", provider.Type)
	}
}

// SyncDNSRecord 同步DNS记录
// frpServerID: FRP服务器ID，用于获取服务器IP地址
func (s *DNSService) SyncDNSRecord(proxy *model.Proxy, frpServerID *uint) error {
	if !proxy.EnableDNSSync || proxy.CustomDomains == "" {
		return nil
	}

	// 获取服务器IP
	serverIP, err := s.GetServerIP(frpServerID)
	if err != nil {
		return fmt.Errorf("获取服务器IP失败: %v", err)
	}

	// 获取DNS提供商：优先使用代理指定的提供商，否则使用第一个启用的提供商
	var provider *model.DNSProvider
	if proxy.DNSProviderID != nil {
		provider, err = s.providerRepo.FindByID(*proxy.DNSProviderID)
		if err != nil {
			return fmt.Errorf("指定的DNS提供商不存在: %v", err)
		}
	} else {
		providers, err := s.providerRepo.FindEnabled()
		if err != nil || len(providers) == 0 {
			return fmt.Errorf("没有可用的DNS提供商")
		}
		provider = &providers[0]
	}

	// 创建 DNS 操作实例
	operator, err := s.CreateDNSOperator(provider)
	if err != nil {
		return fmt.Errorf("创建 DNS 操作实例失败: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 先检查是否已存在该代理的DNS记录，如果存在则先删除
	existingRecord, _ := s.recordRepo.FindByProxyID(proxy.ID)
	if existingRecord != nil && existingRecord.RecordID != "" {
		logger.Infof("DNS 发现已存在的DNS记录，先删除旧记录: %s", existingRecord.RecordID)
		if err := operator.DeleteRecord(ctx, existingRecord.Domain, existingRecord.RecordID); err != nil {
			logger.Warnf("DNS 删除旧记录失败: %v", err)
		}
		s.recordRepo.DeleteByProxyID(proxy.ID)
	}

	// 使用代理指定的根域名，如果没有则从 CustomDomains 解析
	rootDomain := proxy.DNSRootDomain
	if rootDomain == "" {
		rootDomain, _ = ParseDomain(proxy.CustomDomains)
	}

	// 添加新记录
	recordID, err := operator.AddRecord(ctx, proxy.CustomDomains, serverIP)
	if err != nil {
		record := &model.DNSRecord{
			ProxyID:     proxy.ID,
			ProviderID:  provider.ID,
			Domain:      proxy.CustomDomains,
			RootDomain:  rootDomain,
			RecordType:  "A",
			RecordValue: serverIP,
			Status:      model.DNSRecordStatusFailed,
			LastError:   err.Error(),
		}
		s.recordRepo.Create(record)
		if s.eventNotifier != nil {
			go s.eventNotifier.NotifyDNSSync(proxy.CustomDomains, "A", false, err.Error())
		}
		return err
	}

	record := &model.DNSRecord{
		ProxyID:     proxy.ID,
		ProviderID:  provider.ID,
		Domain:      proxy.CustomDomains,
		RootDomain:  rootDomain,
		RecordType:  "A",
		RecordValue: serverIP,
		RecordID:    recordID,
		Status:      model.DNSRecordStatusSynced,
	}
	if err := s.recordRepo.Create(record); err != nil {
		return err
	}
	if s.eventNotifier != nil {
		go s.eventNotifier.NotifyDNSSync(proxy.CustomDomains, "A", true, "")
	}
	return nil
}

// DeleteDNSRecord 删除DNS记录
func (s *DNSService) DeleteDNSRecord(proxyID uint) error {
	record, err := s.recordRepo.FindByProxyID(proxyID)
	if err != nil {
		return nil
	}

	if record.RecordID != "" {
		provider, err := s.providerRepo.FindByID(record.ProviderID)
		if err == nil {
			operator, err := s.CreateDNSOperator(provider)
			if err == nil {
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				if err := operator.DeleteRecord(ctx, record.Domain, record.RecordID); err != nil {
					logger.Warnf("DNS 删除DNS记录失败: %v", err)
				}
			}
		}
	}

	return s.recordRepo.DeleteByProxyID(proxyID)
}

// GetServerIP 获取FRP服务器IP
func (s *DNSService) GetServerIP(frpServerID *uint) (string, error) {
	if frpServerID == nil {
		return "", fmt.Errorf("未关联FRP服务器")
	}
	server, err := s.frpServerRepo.GetByID(*frpServerID)
	if err != nil {
		return "", err
	}

	// 调试日志：打印服务器信息
	logger.Debugf("DNS GetServerIP 服务器ID=%d, Name=%s, ServerType=%s, Host=%s, SSHHost=%s",
		server.ID, server.Name, server.ServerType, server.Host, server.SSHHost)

	// 如果 Host 为空或为 0.0.0.0，尝试使用 SSHHost
	host := server.Host
	if host == "" || host == "0.0.0.0" {
		if server.SSHHost != "" {
			logger.Debugf("DNS GetServerIP Host 为空或 0.0.0.0，回退使用 SSHHost: %s", server.SSHHost)
			host = server.SSHHost
		} else {
			logger.Warn("DNS GetServerIP Host 和 SSHHost 都为空或无效！")
		}
	}

	logger.Debugf("DNS GetServerIP 最终返回的IP: %s", host)
	return host, nil
}

// TestProvider 测试 DNS 提供商连接
func (s *DNSService) TestProvider(provider *model.DNSProvider) error {
	operator, err := s.CreateDNSOperator(provider)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return operator.TestConnection(ctx)
}
