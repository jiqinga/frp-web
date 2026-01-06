package service

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"frp-web-panel/internal/events"
	"frp-web-panel/internal/model"
	"frp-web-panel/internal/repository"
	"frp-web-panel/internal/websocket"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
)

// ACMEUser 实现 lego 的 User 接口
type ACMEUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *ACMEUser) GetEmail() string                        { return u.Email }
func (u *ACMEUser) GetRegistration() *registration.Resource { return u.Registration }
func (u *ACMEUser) GetPrivateKey() crypto.PrivateKey        { return u.key }

// CustomDNSProvider 自定义 DNS 提供商，实现 lego 的 challenge.Provider 接口
type CustomDNSProvider struct {
	operator DNSOperator
	records  map[string]string // domain -> recordID
}

func NewCustomDNSProvider(operator DNSOperator) *CustomDNSProvider {
	return &CustomDNSProvider{
		operator: operator,
		records:  make(map[string]string),
	}
}

func (p *CustomDNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	log.Printf("[ACME] 添加 DNS TXT 记录: %s -> %s", fqdn, value)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	recordID, err := p.operator.AddTXTRecord(ctx, fqdn, value)
	if err != nil {
		return fmt.Errorf("添加 TXT 记录失败: %w", err)
	}
	p.records[fqdn] = recordID
	return nil
}

func (p *CustomDNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)
	log.Printf("[ACME] 清理 DNS TXT 记录: %s", fqdn)

	recordID, ok := p.records[fqdn]
	if !ok {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := p.operator.DeleteTXTRecord(ctx, fqdn, recordID); err != nil {
		log.Printf("[ACME] 清理 TXT 记录失败: %v", err)
	}
	delete(p.records, fqdn)
	return nil
}

// ACMEService ACME 证书服务
type ACMEService struct {
	certRepo       *repository.CertificateRepository
	providerRepo   *repository.DNSProviderRepository
	proxyRepo      *repository.ProxyRepository
	dnsService     *DNSService
	settingService *SettingService
	staging        bool // 是否使用 Let's Encrypt 测试环境
	eventBus       *events.EventBus
	daemonHub      *websocket.ClientDaemonHub
	eventNotifier  *SystemEventNotifier
}

func NewACMEService(staging bool) *ACMEService {
	return &ACMEService{
		certRepo:       repository.NewCertificateRepository(),
		providerRepo:   repository.NewDNSProviderRepository(),
		proxyRepo:      repository.NewProxyRepository(),
		dnsService:     NewDNSService(),
		settingService: NewSettingService(),
		staging:        staging,
		eventBus:       events.GetEventBus(),
	}
}

// SetEventNotifier 设置系统事件通知器
func (s *ACMEService) SetEventNotifier(notifier *SystemEventNotifier) {
	s.eventNotifier = notifier
}

// getEmail 动态获取 ACME 邮箱
func (s *ACMEService) getEmail() string {
	return s.settingService.GetAcmeEmail()
}

// SetDaemonHub 设置 ClientDaemonHub
func (s *ACMEService) SetDaemonHub(daemonHub *websocket.ClientDaemonHub) {
	s.daemonHub = daemonHub
}

// notifyProgress 发送进度通知
func (s *ACMEService) notifyProgress(taskID, domain, step, message, errMsg string) {
	s.eventBus.Publish(events.CertProgressEvent{
		TaskID:    taskID,
		Domain:    domain,
		Step:      step,
		Message:   message,
		Error:     errMsg,
		Timestamp: time.Now().Format("15:04:05"),
	})
}

// RequestCertificate 申请证书
func (s *ACMEService) RequestCertificate(proxyID uint, domain string, providerID uint) (*model.Certificate, error) {
	return s.RequestCertificateWithTaskAndAutoRenew("", proxyID, domain, providerID, true)
}

// RequestCertificateWithTask 申请证书（带任务ID）
func (s *ACMEService) RequestCertificateWithTask(taskID string, proxyID uint, domain string, providerID uint) (*model.Certificate, error) {
	return s.RequestCertificateWithTaskAndAutoRenew(taskID, proxyID, domain, providerID, true)
}

// createHTTPClientWithTimeout 创建带超时的 HTTP 客户端
func createHTTPClientWithTimeout(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: timeout,
			IdleConnTimeout:       90 * time.Second,
		},
	}
}

// RequestCertificateWithTaskAndAutoRenew 申请证书（带任务ID和自动续期参数）
func (s *ACMEService) RequestCertificateWithTaskAndAutoRenew(taskID string, proxyID uint, domain string, providerID uint, autoRenew bool) (*model.Certificate, error) {
	// 动态获取 ACME 邮箱
	email := s.getEmail()
	log.Printf("[ACME] 从设置中获取的邮箱: '%s'", email)

	// 校验 ACME 邮箱是否已配置
	if email == "" {
		s.notifyProgress(taskID, domain, "failed", "", "请先在系统设置中配置ACME证书申请邮箱")
		return nil, fmt.Errorf("请先在系统设置中配置ACME证书申请邮箱")
	}

	// 校验邮箱域名是否为保留域名
	if strings.Contains(email, "@example.com") || strings.Contains(email, "@example.org") || strings.Contains(email, "@example.net") {
		errMsg := fmt.Sprintf("邮箱域名无效: %s (example.com/org/net 是保留域名，请使用真实邮箱)", email)
		s.notifyProgress(taskID, domain, "failed", "", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	// 检查是否已存在相同域名的有效证书
	existingCert, err := s.certRepo.FindByDomain(domain)
	if err == nil && existingCert != nil {
		// 如果证书状态为 active、expiring 或 pending，则拒绝申请
		if existingCert.Status == model.CertStatusActive ||
			existingCert.Status == model.CertStatusExpiring ||
			existingCert.Status == model.CertStatusPending {
			errMsg := fmt.Sprintf("域名 %s 已存在有效证书，无需重复申请", domain)
			s.notifyProgress(taskID, domain, "failed", "", errMsg)
			return nil, fmt.Errorf("%s", errMsg)
		}
	}

	// 步骤1: 验证配置
	s.notifyProgress(taskID, domain, "validating", "正在验证DNS提供商配置...", "")
	log.Printf("[ACME] 开始为域名 %s 申请证书, 使用邮箱: %s", domain, email)

	// 获取 DNS 提供商
	log.Printf("[ACME] 正在获取DNS提供商 (ID: %d)...", providerID)
	provider, err := s.providerRepo.FindByID(providerID)
	if err != nil {
		s.notifyProgress(taskID, domain, "failed", "", "获取DNS提供商失败: "+err.Error())
		return nil, fmt.Errorf("获取 DNS 提供商失败: %w", err)
	}
	log.Printf("[ACME] DNS提供商获取成功: %s (类型: %s)", provider.Name, provider.Type)

	// 创建 DNS 操作实例
	log.Printf("[ACME] 正在创建DNS操作实例...")
	operator, err := s.dnsService.CreateDNSOperator(provider)
	if err != nil {
		s.notifyProgress(taskID, domain, "failed", "", "创建DNS操作实例失败: "+err.Error())
		return nil, fmt.Errorf("创建 DNS 操作实例失败: %w", err)
	}
	log.Printf("[ACME] DNS操作实例创建成功")

	// 生成私钥
	log.Printf("[ACME] 正在生成私钥...")
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		s.notifyProgress(taskID, domain, "failed", "", "生成私钥失败: "+err.Error())
		return nil, fmt.Errorf("生成私钥失败: %w", err)
	}
	log.Printf("[ACME] 私钥生成成功")

	user := &ACMEUser{
		Email: email,
		key:   privateKey,
	}

	// 使用重试机制创建 ACME 客户端
	s.notifyProgress(taskID, domain, "connecting", "正在连接Let's Encrypt服务器...", "")

	var client *lego.Client
	retryOperation := func() (*lego.Client, error) {
		log.Printf("[ACME] 尝试连接 Let's Encrypt 服务器...")

		// 配置 ACME 客户端
		config := lego.NewConfig(user)
		if s.staging {
			config.CADirURL = lego.LEDirectoryStaging
		} else {
			config.CADirURL = lego.LEDirectoryProduction
		}
		config.Certificate.KeyType = certcrypto.EC256
		// 设置带超时的 HTTP 客户端（30秒超时）
		config.HTTPClient = createHTTPClientWithTimeout(30 * time.Second)

		c, err := lego.NewClient(config)
		if err != nil {
			log.Printf("[ACME] 连接失败: %v", err)
			return nil, err
		}
		log.Printf("[ACME] 连接成功")
		return c, nil
	}

	// 使用简单重试，最多3次，指数退避
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		client, err = retryOperation()
		if err == nil {
			break
		}
		if attempt < maxRetries {
			waitTime := time.Duration(attempt*2) * time.Second
			log.Printf("[ACME] 第 %d 次尝试失败，%v 后重试...", attempt, waitTime)
			time.Sleep(waitTime)
		}
	}

	if err != nil {
		s.notifyProgress(taskID, domain, "failed", "", "连接Let's Encrypt服务器失败（已重试3次）: "+err.Error())
		return nil, fmt.Errorf("创建 ACME 客户端失败: %w", err)
	}

	// 步骤2: 添加DNS记录
	s.notifyProgress(taskID, domain, "adding_dns", "正在添加DNS TXT验证记录...", "")

	// 设置 DNS-01 验证
	// 使用国内可访问的DNS服务器，避免Google DNS在中国大陆不可达的问题
	dnsProvider := NewCustomDNSProvider(operator)
	if err := client.Challenge.SetDNS01Provider(dnsProvider,
		dns01.AddDNSTimeout(120*time.Second),
		dns01.AddRecursiveNameservers([]string{
			"223.5.5.5:53",       // 阿里DNS
			"223.6.6.6:53",       // 阿里DNS备用
			"119.29.29.29:53",    // 腾讯DNS
			"114.114.114.114:53", // 114 DNS
		}),
	); err != nil {
		s.notifyProgress(taskID, domain, "failed", "", "设置DNS验证失败: "+err.Error())
		return nil, fmt.Errorf("设置 DNS 验证失败: %w", err)
	}
	log.Printf("[ACME] DNS-01 验证提供商设置成功")

	// 注册账户（带重试）
	s.notifyProgress(taskID, domain, "registering", "正在注册ACME账户...", "")
	log.Printf("[ACME] 正在注册ACME账户...")

	var reg *registration.Resource
	regOperation := func() (*registration.Resource, error) {
		return client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	}

	// 使用简单重试，最多3次
	for attempt := 1; attempt <= maxRetries; attempt++ {
		reg, err = regOperation()
		if err == nil {
			break
		}
		if attempt < maxRetries {
			waitTime := time.Duration(attempt*2) * time.Second
			log.Printf("[ACME] 注册账户第 %d 次尝试失败，%v 后重试...", attempt, waitTime)
			time.Sleep(waitTime)
		}
	}

	if err != nil {
		s.notifyProgress(taskID, domain, "failed", "", "注册ACME账户失败: "+err.Error())
		return nil, fmt.Errorf("注册 ACME 账户失败: %w", err)
	}
	user.Registration = reg
	log.Printf("[ACME] ACME账户注册成功")

	// 步骤3: 等待DNS验证
	s.notifyProgress(taskID, domain, "waiting_dns", "等待DNS记录生效...", "")

	// 步骤4: 申请证书
	s.notifyProgress(taskID, domain, "requesting", "正在向Let's Encrypt申请证书...", "")
	log.Printf("[ACME] 正在申请证书...")

	request := certificate.ObtainRequest{
		Domains: []string{domain},
		Bundle:  true,
	}
	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		log.Printf("[ACME] 证书申请失败: %v", err)
		// 保存失败记录
		cert := &model.Certificate{
			ProxyID:    proxyID,
			Domain:     domain,
			ProviderID: providerID,
			Status:     model.CertStatusFailed,
			LastError:  err.Error(),
		}
		s.certRepo.Create(cert)
		s.notifyProgress(taskID, domain, "failed", "", "申请证书失败: "+err.Error())
		// 发送系统告警通知
		if s.eventNotifier != nil {
			go s.eventNotifier.NotifyCertApply(domain, 0, false, err.Error())
		}
		return nil, fmt.Errorf("申请证书失败: %w", err)
	}

	// 步骤5: 保存证书
	s.notifyProgress(taskID, domain, "saving", "正在保存证书...", "")
	log.Printf("[ACME] 正在保存证书...")

	// 解析证书有效期
	notBefore, notAfter, _ := parseCertificateDates(certificates.Certificate)

	// 保存证书
	cert := &model.Certificate{
		ProxyID:       proxyID,
		Domain:        domain,
		ProviderID:    providerID,
		Status:        model.CertStatusActive,
		CertPEM:       string(certificates.Certificate),
		KeyPEM:        string(certificates.PrivateKey),
		IssuerCertPEM: string(certificates.IssuerCertificate),
		NotBefore:     notBefore,
		NotAfter:      notAfter,
		AutoRenew:     autoRenew,
	}

	if err := s.certRepo.Create(cert); err != nil {
		s.notifyProgress(taskID, domain, "failed", "", "保存证书失败: "+err.Error())
		return nil, fmt.Errorf("保存证书失败: %w", err)
	}

	// 完成
	s.notifyProgress(taskID, domain, "completed", "证书申请成功！", "")
	log.Printf("[ACME] 证书申请成功: %s, 有效期至 %v", domain, notAfter)

	// 发送系统告警通知
	if s.eventNotifier != nil {
		go s.eventNotifier.NotifyCertApply(domain, cert.ID, true, "")
	}

	// 推送证书到客户端
	go s.pushCertToClient(proxyID, domain, cert.CertPEM, cert.KeyPEM)

	return cert, nil
}

// pushCertToClient 推送证书到关联的客户端
func (s *ACMEService) pushCertToClient(proxyID uint, domain string, certPEM string, keyPEM string) {
	// 获取代理信息以获取客户端ID
	proxy, err := s.proxyRepo.FindByID(proxyID)
	if err != nil {
		log.Printf("[ACME] ❌ 获取代理信息失败 (proxyID=%d): %v", proxyID, err)
		return
	}

	if s.daemonHub == nil {
		log.Printf("[ACME] ⚠️ DaemonHub 未设置，添加到同步队列")
		GetCertSyncQueue().AddPendingSync(proxy.ClientID, domain, certPEM, keyPEM)
		return
	}

	// 检查客户端是否在线
	if !s.daemonHub.IsClientOnline(proxy.ClientID) {
		log.Printf("[ACME] ⚠️ 客户端 %d 离线，添加到同步队列", proxy.ClientID)
		GetCertSyncQueue().AddPendingSync(proxy.ClientID, domain, certPEM, keyPEM)
		return
	}

	// 推送证书到客户端
	if err := s.daemonHub.PushCertSync(proxy.ClientID, domain, certPEM, keyPEM); err != nil {
		log.Printf("[ACME] ❌ 推送证书到客户端 %d 失败，添加到同步队列: %v", proxy.ClientID, err)
		GetCertSyncQueue().AddPendingSync(proxy.ClientID, domain, certPEM, keyPEM)
	} else {
		log.Printf("[ACME] ✅ 证书已推送到客户端 %d: domain=%s", proxy.ClientID, domain)
	}
}

// RenewCertificate 续期证书
func (s *ACMEService) RenewCertificate(certID uint) error {
	cert, err := s.certRepo.FindByID(certID)
	if err != nil {
		return fmt.Errorf("获取证书失败: %w", err)
	}

	// 使用内部续签方法，直接更新现有证书，不创建新记录
	return s.renewCertificateInternal(cert)
}

// renewCertificateInternal 内部续签方法，直接更新现有证书记录
func (s *ACMEService) renewCertificateInternal(cert *model.Certificate) error {
	// 动态获取 ACME 邮箱
	email := s.getEmail()
	if email == "" {
		return fmt.Errorf("请先在系统设置中配置ACME证书申请邮箱")
	}

	// 校验邮箱域名是否为保留域名
	if strings.Contains(email, "@example.com") || strings.Contains(email, "@example.org") || strings.Contains(email, "@example.net") {
		return fmt.Errorf("邮箱域名无效: %s (example.com/org/net 是保留域名，请使用真实邮箱)", email)
	}

	log.Printf("[ACME] 开始续签证书: ID=%d, 域名=%s", cert.ID, cert.Domain)

	// 获取 DNS 提供商
	provider, err := s.providerRepo.FindByID(cert.ProviderID)
	if err != nil {
		return fmt.Errorf("获取 DNS 提供商失败: %w", err)
	}

	// 创建 DNS 操作实例
	operator, err := s.dnsService.CreateDNSOperator(provider)
	if err != nil {
		return fmt.Errorf("创建 DNS 操作实例失败: %w", err)
	}

	// 生成私钥
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("生成私钥失败: %w", err)
	}

	user := &ACMEUser{
		Email: email,
		key:   privateKey,
	}

	// 创建 ACME 客户端（带重试）
	var client *lego.Client
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		config := lego.NewConfig(user)
		if s.staging {
			config.CADirURL = lego.LEDirectoryStaging
		} else {
			config.CADirURL = lego.LEDirectoryProduction
		}
		config.Certificate.KeyType = certcrypto.EC256
		config.HTTPClient = createHTTPClientWithTimeout(30 * time.Second)

		client, err = lego.NewClient(config)
		if err == nil {
			break
		}
		if attempt < maxRetries {
			time.Sleep(time.Duration(attempt*2) * time.Second)
		}
	}
	if err != nil {
		return fmt.Errorf("创建 ACME 客户端失败: %w", err)
	}

	// 设置 DNS-01 验证
	dnsProvider := NewCustomDNSProvider(operator)
	if err := client.Challenge.SetDNS01Provider(dnsProvider,
		dns01.AddDNSTimeout(120*time.Second),
		dns01.AddRecursiveNameservers([]string{
			"223.5.5.5:53", "223.6.6.6:53", "119.29.29.29:53", "114.114.114.114:53",
		}),
	); err != nil {
		return fmt.Errorf("设置 DNS 验证失败: %w", err)
	}

	// 注册账户（带重试）
	var reg *registration.Resource
	for attempt := 1; attempt <= maxRetries; attempt++ {
		reg, err = client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
		if err == nil {
			break
		}
		if attempt < maxRetries {
			time.Sleep(time.Duration(attempt*2) * time.Second)
		}
	}
	if err != nil {
		return fmt.Errorf("注册 ACME 账户失败: %w", err)
	}
	user.Registration = reg

	// 申请证书
	request := certificate.ObtainRequest{
		Domains: []string{cert.Domain},
		Bundle:  true,
	}
	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		cert.LastError = err.Error()
		s.certRepo.Update(cert)
		if s.eventNotifier != nil {
			go s.eventNotifier.NotifyCertApply(cert.Domain, cert.ID, false, err.Error())
		}
		return fmt.Errorf("申请证书失败: %w", err)
	}

	// 解析证书有效期
	notBefore, notAfter, _ := parseCertificateDates(certificates.Certificate)

	// 更新原证书记录
	cert.CertPEM = string(certificates.Certificate)
	cert.KeyPEM = string(certificates.PrivateKey)
	cert.IssuerCertPEM = string(certificates.IssuerCertificate)
	cert.NotBefore = notBefore
	cert.NotAfter = notAfter
	cert.Status = model.CertStatusActive
	cert.LastError = ""

	if err := s.certRepo.Update(cert); err != nil {
		return fmt.Errorf("更新证书记录失败: %w", err)
	}

	log.Printf("[ACME] ✅ 证书续签成功: ID=%d, 域名=%s, 有效期至 %v", cert.ID, cert.Domain, notAfter)

	// 发送系统告警通知
	if s.eventNotifier != nil {
		go s.eventNotifier.NotifyCertApply(cert.Domain, cert.ID, true, "")
	}

	// 推送证书到客户端
	go s.pushCertToClient(cert.ProxyID, cert.Domain, cert.CertPEM, cert.KeyPEM)

	return nil
}

// GetCertificateByProxyID 根据代理ID获取证书
func (s *ACMEService) GetCertificateByProxyID(proxyID uint) (*model.Certificate, error) {
	return s.certRepo.FindByProxyID(proxyID)
}

// ReapplyCertificate 重新申请失败的证书
func (s *ACMEService) ReapplyCertificate(certID uint) error {
	cert, err := s.certRepo.FindByID(certID)
	if err != nil {
		return fmt.Errorf("获取证书失败: %w", err)
	}

	if cert.Status != model.CertStatusFailed {
		return fmt.Errorf("只能重新申请失败状态的证书")
	}

	// 重置状态为pending
	cert.Status = model.CertStatusPending
	cert.LastError = ""
	if err := s.certRepo.Update(cert); err != nil {
		return fmt.Errorf("更新证书状态失败: %w", err)
	}

	// 异步执行证书申请
	go func() {
		newCert, err := s.RequestCertificateWithTask("", cert.ProxyID, cert.Domain, cert.ProviderID)
		if err != nil {
			log.Printf("[ACME] 重新申请证书失败: %v", err)
			return
		}
		// 更新原证书记录
		cert.CertPEM = newCert.CertPEM
		cert.KeyPEM = newCert.KeyPEM
		cert.IssuerCertPEM = newCert.IssuerCertPEM
		cert.NotBefore = newCert.NotBefore
		cert.NotAfter = newCert.NotAfter
		cert.Status = model.CertStatusActive
		cert.LastError = ""
		s.certRepo.Update(cert)
		// 删除新创建的重复记录
		s.certRepo.Delete(newCert.ID)
	}()

	return nil
}

// parseCertificateDates 解析证书有效期
func parseCertificateDates(certPEM []byte) (*time.Time, *time.Time, error) {
	block, _ := pem.Decode(certPEM)
	if block == nil {
		// 解析失败时使用默认值
		now := time.Now()
		notBefore := now
		notAfter := now.Add(90 * 24 * time.Hour)
		return &notBefore, &notAfter, fmt.Errorf("无法解析 PEM 证书")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		now := time.Now()
		notBefore := now
		notAfter := now.Add(90 * 24 * time.Hour)
		return &notBefore, &notAfter, fmt.Errorf("解析证书失败: %w", err)
	}

	return &cert.NotBefore, &cert.NotAfter, nil
}
