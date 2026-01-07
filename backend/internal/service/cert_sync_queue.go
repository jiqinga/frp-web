package service

import (
	"frp-web-panel/internal/logger"
	"frp-web-panel/internal/repository"
	"frp-web-panel/internal/websocket"
	"sync"
	"time"
)

// PendingCertSync 待同步的证书信息
type PendingCertSync struct {
	ClientID  uint
	Domain    string
	CertPEM   string
	KeyPEM    string
	CreatedAt time.Time
}

// CertSyncQueue 证书同步队列
type CertSyncQueue struct {
	queue     map[uint][]PendingCertSync // clientID -> pending certs
	mu        sync.RWMutex
	daemonHub *websocket.ClientDaemonHub
	certRepo  *repository.CertificateRepository
	proxyRepo *repository.ProxyRepository
}

var certSyncQueueInstance *CertSyncQueue
var certSyncQueueOnce sync.Once

// GetCertSyncQueue 获取证书同步队列单例
func GetCertSyncQueue() *CertSyncQueue {
	certSyncQueueOnce.Do(func() {
		certSyncQueueInstance = &CertSyncQueue{
			queue:     make(map[uint][]PendingCertSync),
			certRepo:  repository.NewCertificateRepository(),
			proxyRepo: repository.NewProxyRepository(),
		}
	})
	return certSyncQueueInstance
}

// SetDaemonHub 设置 DaemonHub
func (q *CertSyncQueue) SetDaemonHub(hub *websocket.ClientDaemonHub) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.daemonHub = hub

	// 注册客户端上线回调
	if hub != nil {
		originalCallback := hub.GetStatusCallback()
		hub.SetStatusCallback(func(clientID uint, online bool) {
			if originalCallback != nil {
				originalCallback(clientID, online)
			}
			if online {
				go q.SyncPendingCerts(clientID)
			}
		})
		logger.Info("CertSyncQueue 已注册客户端上线回调")
	}
}

// AddPendingSync 添加待同步的证书
func (q *CertSyncQueue) AddPendingSync(clientID uint, domain, certPEM, keyPEM string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// 检查是否已存在相同域名的待同步记录，如果有则更新
	pending := q.queue[clientID]
	for i, p := range pending {
		if p.Domain == domain {
			pending[i] = PendingCertSync{
				ClientID:  clientID,
				Domain:    domain,
				CertPEM:   certPEM,
				KeyPEM:    keyPEM,
				CreatedAt: time.Now(),
			}
			q.queue[clientID] = pending
			logger.Infof("CertSyncQueue 更新待同步证书: clientID=%d, domain=%s", clientID, domain)
			return
		}
	}

	// 添加新记录
	q.queue[clientID] = append(pending, PendingCertSync{
		ClientID:  clientID,
		Domain:    domain,
		CertPEM:   certPEM,
		KeyPEM:    keyPEM,
		CreatedAt: time.Now(),
	})
	logger.Infof("CertSyncQueue 添加待同步证书: clientID=%d, domain=%s", clientID, domain)
}

// SyncPendingCerts 同步待处理的证书到客户端
func (q *CertSyncQueue) SyncPendingCerts(clientID uint) {
	q.mu.Lock()
	pending, exists := q.queue[clientID]
	if !exists || len(pending) == 0 {
		q.mu.Unlock()
		return
	}
	// 复制并清空队列
	toSync := make([]PendingCertSync, len(pending))
	copy(toSync, pending)
	delete(q.queue, clientID)
	q.mu.Unlock()

	if q.daemonHub == nil {
		logger.Warn("CertSyncQueue DaemonHub 未设置，无法同步证书")
		return
	}

	logger.Infof("CertSyncQueue 开始同步 %d 个待处理证书到客户端 %d", len(toSync), clientID)

	for _, cert := range toSync {
		if err := q.daemonHub.PushCertSync(clientID, cert.Domain, cert.CertPEM, cert.KeyPEM); err != nil {
			logger.Errorf("CertSyncQueue 同步证书失败: clientID=%d, domain=%s, err=%v", clientID, cert.Domain, err)
			// 重新加入队列
			q.AddPendingSync(clientID, cert.Domain, cert.CertPEM, cert.KeyPEM)
		} else {
			logger.Infof("CertSyncQueue 同步证书成功: clientID=%d, domain=%s", clientID, cert.Domain)
		}
		// 避免发送过快
		time.Sleep(100 * time.Millisecond)
	}
}

// CleanExpired 清理过期的待同步记录（超过7天）
func (q *CertSyncQueue) CleanExpired() {
	q.mu.Lock()
	defer q.mu.Unlock()

	expireTime := time.Now().Add(-7 * 24 * time.Hour)
	for clientID, pending := range q.queue {
		var valid []PendingCertSync
		for _, p := range pending {
			if p.CreatedAt.After(expireTime) {
				valid = append(valid, p)
			}
		}
		if len(valid) == 0 {
			delete(q.queue, clientID)
		} else {
			q.queue[clientID] = valid
		}
	}
}

// GetPendingCount 获取待同步数量
func (q *CertSyncQueue) GetPendingCount() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	count := 0
	for _, pending := range q.queue {
		count += len(pending)
	}
	return count
}
