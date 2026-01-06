package service

import (
	"fmt"
	"frp-web-panel/internal/model"
	"frp-web-panel/internal/repository"
	"log"
	"time"
)

// CertRenewalScheduler 证书续期调度器
type CertRenewalScheduler struct {
	certRepo      *repository.CertificateRepository
	acmeService   *ACMEService
	eventNotifier *SystemEventNotifier
	logService    *LogService
	interval      time.Duration
	stopChan      chan struct{}
	retryQueue    map[uint]int // certID -> retry count
}

func NewCertRenewalScheduler(acmeService *ACMEService) *CertRenewalScheduler {
	return &CertRenewalScheduler{
		certRepo:    repository.NewCertificateRepository(),
		acmeService: acmeService,
		logService:  NewLogService(),
		interval:    24 * time.Hour, // 每天检查一次
		stopChan:    make(chan struct{}),
		retryQueue:  make(map[uint]int),
	}
}

// SetEventNotifier 设置系统事件通知器
func (s *CertRenewalScheduler) SetEventNotifier(notifier *SystemEventNotifier) {
	s.eventNotifier = notifier
}

// Start 启动调度器
func (s *CertRenewalScheduler) Start() {
	log.Println("[证书续期] 调度器已启动，检查间隔:", s.interval)
	go s.run()
}

// Stop 停止调度器
func (s *CertRenewalScheduler) Stop() {
	close(s.stopChan)
	log.Println("[证书续期] 调度器已停止")
}

func (s *CertRenewalScheduler) run() {
	// 启动时立即检查一次
	s.checkAndRenew()

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.checkAndRenew()
		case <-s.stopChan:
			return
		}
	}
}

func (s *CertRenewalScheduler) checkAndRenew() {
	log.Println("[证书续期] 开始检查即将过期的证书...")

	// 更新证书状态
	s.updateCertificateStatuses()

	// 获取需要续期的证书
	certs, err := s.certRepo.FindExpiring()
	if err != nil {
		log.Printf("[证书续期] ❌ 获取即将过期证书失败: %v", err)
		return
	}

	if len(certs) == 0 {
		log.Println("[证书续期] ✅ 没有需要续期的证书")
		return
	}

	log.Printf("[证书续期] 发现 %d 个需要续期的证书", len(certs))

	for _, cert := range certs {
		if !cert.AutoRenew {
			continue
		}

		log.Printf("[证书续期] 开始续期证书: ID=%d, 域名=%s", cert.ID, cert.Domain)
		if err := s.acmeService.RenewCertificate(cert.ID); err != nil {
			log.Printf("[证书续期] ❌ 续期失败: %v", err)
			s.addToRetryQueue(cert.ID)
			// 记录续签失败日志
			s.logService.CreateLogAsync(0, "auto_renew", "certificate", cert.ID,
				fmt.Sprintf("系统自动续签证书失败: %s, 错误: %v", cert.Domain, err), "127.0.0.1")
		} else {
			log.Printf("[证书续期] ✅ 续期成功: ID=%d", cert.ID)
			delete(s.retryQueue, cert.ID)
			// 记录续签成功日志
			s.logService.CreateLogAsync(0, "auto_renew", "certificate", cert.ID,
				fmt.Sprintf("系统自动续签证书成功: %s", cert.Domain), "127.0.0.1")
		}
	}

	// 处理重试队列
	s.processRetryQueue()
}

// addToRetryQueue 添加到重试队列
func (s *CertRenewalScheduler) addToRetryQueue(certID uint) {
	count := s.retryQueue[certID]
	if count < 3 { // 最多重试3次
		s.retryQueue[certID] = count + 1
	}
}

// processRetryQueue 处理重试队列
func (s *CertRenewalScheduler) processRetryQueue() {
	if len(s.retryQueue) == 0 {
		return
	}

	log.Printf("[证书续期] 处理重试队列，共 %d 个证书", len(s.retryQueue))

	for certID, retryCount := range s.retryQueue {
		if retryCount > 3 {
			log.Printf("[证书续期] ⚠️ 证书 %d 已达最大重试次数，跳过", certID)
			delete(s.retryQueue, certID)
			continue
		}

		// 等待一段时间后重试（指数退避）
		waitTime := time.Duration(retryCount*30) * time.Minute
		log.Printf("[证书续期] 证书 %d 将在 %v 后重试（第 %d 次）", certID, waitTime, retryCount)

		go func(id uint, wait time.Duration) {
			time.Sleep(wait)
			if err := s.acmeService.RenewCertificate(id); err != nil {
				log.Printf("[证书续期] ❌ 重试失败: ID=%d, err=%v", id, err)
			} else {
				log.Printf("[证书续期] ✅ 重试成功: ID=%d", id)
				delete(s.retryQueue, id)
			}
		}(certID, waitTime)
	}
}

// updateCertificateStatuses 更新所有证书的状态
func (s *CertRenewalScheduler) updateCertificateStatuses() {
	certs, err := s.certRepo.FindAll()
	if err != nil {
		log.Printf("[证书续期] ❌ 获取证书列表失败: %v", err)
		return
	}

	now := time.Now()
	for _, cert := range certs {
		if cert.NotAfter == nil || cert.Status == model.CertStatusFailed {
			continue
		}

		var newStatus string
		daysUntilExpiry := cert.NotAfter.Sub(now).Hours() / 24

		if daysUntilExpiry <= 0 {
			newStatus = model.CertStatusExpired
		} else if daysUntilExpiry <= 30 {
			newStatus = model.CertStatusExpiring
		} else {
			newStatus = model.CertStatusActive
		}

		if cert.Status != newStatus {
			oldStatus := cert.Status
			cert.Status = newStatus
			if err := s.certRepo.Update(&cert); err != nil {
				log.Printf("[证书续期] ⚠️ 更新证书状态失败: ID=%d, %v", cert.ID, err)
			} else {
				// 状态变更时发送通知和记录日志
				if newStatus == model.CertStatusExpiring && oldStatus != model.CertStatusExpiring {
					if s.eventNotifier != nil {
						go s.eventNotifier.NotifyCertExpiring(cert.Domain, cert.ID, *cert.NotAfter)
					}
					s.logService.CreateLogAsync(0, "status_change", "certificate", cert.ID,
						fmt.Sprintf("证书即将过期: %s", cert.Domain), "127.0.0.1")
				} else if newStatus == model.CertStatusExpired && oldStatus != model.CertStatusExpired {
					if s.eventNotifier != nil {
						go s.eventNotifier.NotifyCertExpired(cert.Domain, cert.ID, *cert.NotAfter)
					}
					s.logService.CreateLogAsync(0, "status_change", "certificate", cert.ID,
						fmt.Sprintf("证书已过期: %s", cert.Domain), "127.0.0.1")
				}
			}
		}
	}
}
