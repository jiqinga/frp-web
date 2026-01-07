/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-14 15:31:40
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2026-01-07 11:07:10
 * @FilePath            : frp-web-testbackendinternalserviceclient_service.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package service

import (
	"crypto/rand"
	"encoding/hex"
	"frp-web-panel/internal/logger"
	"frp-web-panel/internal/model"
	"frp-web-panel/internal/repository"
	"frp-web-panel/internal/websocket"
	"frp-web-panel/pkg/database"
	"time"

	"gorm.io/gorm"
)

type ClientService struct {
	clientRepo    *repository.ClientRepository
	frpServerRepo *repository.FrpServerRepository
}

func NewClientService() *ClientService {
	return &ClientService{
		clientRepo:    repository.NewClientRepository(),
		frpServerRepo: repository.NewFrpServerRepository(database.DB),
	}
}

func (s *ClientService) GetClients(page, pageSize int, keyword string) ([]model.Client, int64, error) {
	return s.clientRepo.FindAll(page, pageSize, keyword)
}

func (s *ClientService) GetClient(id uint) (*model.Client, error) {
	return s.clientRepo.FindByID(id)
}

func (s *ClientService) CreateClient(client *model.Client) error {
	// 如果没有设置 Admin API 配置，设置默认值
	if client.FrpcAdminPort == 0 {
		client.FrpcAdminHost = "127.0.0.1"
		client.FrpcAdminPort = 7400
		client.FrpcAdminUser = "admin"
		// 生成随机密码
		pwdBytes := make([]byte, 12)
		if _, err := rand.Read(pwdBytes); err == nil {
			client.FrpcAdminPwd = hex.EncodeToString(pwdBytes)[:16]
		} else {
			client.FrpcAdminPwd = "admin" // 降级使用默认密码
		}
		logger.Infof("客户端创建 自动设置 Admin API 配置: addr=%s, port=%d, user=%s",
			client.FrpcAdminHost, client.FrpcAdminPort, client.FrpcAdminUser)
	}
	return s.clientRepo.Create(client)
}

func (s *ClientService) UpdateClient(client *model.Client) error {
	return s.clientRepo.Update(client)
}

func (s *ClientService) DeleteClient(id uint) error {
	// 在删除前，先尝试向 ws-daemon 发送停止命令
	// 即使发送失败也继续删除（客户端可能已离线）
	if websocket.ClientDaemonHubInstance.IsClientOnline(id) {
		logger.Infof("客户端删除 客户端 ID=%d 在线，发送停止命令...", id)
		if err := websocket.ClientDaemonHubInstance.SendShutdownCommand(id); err != nil {
			logger.Warnf("客户端删除 发送停止命令失败: %v，继续删除操作", err)
		} else {
			logger.Info("客户端删除 停止命令已发送，等待 daemon 关闭...")
			// 等待短暂时间让 daemon 有机会优雅关闭
			time.Sleep(500 * time.Millisecond)
		}
	} else {
		logger.Infof("客户端删除 客户端 ID=%d 不在线，跳过发送停止命令", id)
	}

	// 使用事务确保数据一致性：先删除关联的代理，再删除客户端
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 先删除该客户端关联的所有代理
		if err := tx.Where("client_id = ?", id).Delete(&model.Proxy{}).Error; err != nil {
			logger.Errorf("客户端删除 删除客户端 ID=%d 的关联代理失败: %v", id, err)
			return err
		}
		logger.Infof("客户端删除 已删除客户端 ID=%d 的所有关联代理", id)

		// 再删除客户端记录
		if err := tx.Delete(&model.Client{}, id).Error; err != nil {
			logger.Errorf("客户端删除 删除客户端 ID=%d 失败: %v", id, err)
			return err
		}
		logger.Infof("客户端删除 客户端 ID=%d 删除成功", id)

		return nil
	})
}

// SyncAllClientsWSStatus 同步所有客户端的 WebSocket 连接状态
// 注意：此方法只更新 WsConnected 字段，不更新 OnlineStatus
// OnlineStatus 由 frpc 健康检查消息单独更新
func (s *ClientService) SyncAllClientsWSStatus() error {
	clients, err := s.clientRepo.GetAllForStatusCheck()
	if err != nil {
		return err
	}

	for _, client := range clients {
		wsConnected := websocket.ClientDaemonHubInstance.IsClientOnline(client.ID)
		// 只在状态变化时更新
		if client.WsConnected != wsConnected {
			if err := s.clientRepo.UpdateWSStatus(client.ID, wsConnected); err != nil {
				logger.Errorf("WS状态同步 更新客户端 %s WS状态失败: %v", client.Name, err)
			}
		}
	}
	return nil
}

// UpdateHeartbeat 更新客户端心跳(保留原有方法供兼容)
func (s *ClientService) UpdateHeartbeat(clientID uint) error {
	now := time.Now()
	logger.Debugf("心跳上报 收到客户端 ID=%d 的心跳，时间: %v", clientID, now.Format("2006-01-02 15:04:05"))
	if err := s.clientRepo.UpdateOnlineStatus(clientID, "online", now); err != nil {
		logger.Errorf("心跳上报 更新失败: %v", err)
		return err
	}
	logger.Debugf("心跳上报 客户端 ID=%d 状态已更新为 online", clientID)
	return nil
}

// UpdateOnlineStatusDirectly 直接更新客户端在线状态（由 WebSocket 连接状态回调使用）
func (s *ClientService) UpdateOnlineStatusDirectly(clientID uint, status string) error {
	now := time.Now()
	logger.Infof("WS状态更新 客户端 ID=%d 状态更新为: %s", clientID, status)
	if err := s.clientRepo.UpdateOnlineStatus(clientID, status, now); err != nil {
		logger.Errorf("WS状态更新 更新失败: %v", err)
		return err
	}
	logger.Infof("WS状态更新 客户端 ID=%d 状态已更新为 %s", clientID, status)
	return nil
}

// UpdateWSStatus 更新客户端WebSocket连接状态
func (s *ClientService) UpdateWSStatus(clientID uint, connected bool) error {
	return s.clientRepo.UpdateWSStatus(clientID, connected)
}

// UpdateConfigSync 更新客户端配置同步信息
func (s *ClientService) UpdateConfigSync(clientID uint, version int, syncTime *time.Time) error {
	return s.clientRepo.UpdateConfigSync(clientID, version, syncTime)
}

// ResetAllClientStatus 重置所有客户端状态为离线（服务启动时调用）
func (s *ClientService) ResetAllClientStatus() error {
	logger.Info("客户端状态 重置所有客户端状态为离线...")
	if err := s.clientRepo.ResetAllClientStatus(); err != nil {
		logger.Errorf("客户端状态 重置失败: %v", err)
		return err
	}
	logger.Info("客户端状态 所有客户端状态已重置为离线")
	return nil
}

// UpdateConfigSyncStatus 更新客户端配置同步状态
func (s *ClientService) UpdateConfigSyncStatus(clientID uint, success bool, errorMsg string, rolledBack bool) error {
	var status string
	if success {
		status = "synced"
	} else if rolledBack {
		status = "rolled_back"
	} else {
		status = "failed"
	}

	syncTime := time.Now()
	logger.Infof("配置同步状态 客户端 ID=%d 同步结果: success=%v, status=%s, error=%s", clientID, success, status, errorMsg)

	if err := s.clientRepo.UpdateConfigSyncStatus(clientID, status, errorMsg, syncTime); err != nil {
		logger.Errorf("配置同步状态 更新失败: %v", err)
		return err
	}
	logger.Infof("配置同步状态 客户端 ID=%d 状态已更新为 %s", clientID, status)
	return nil
}

// SetConfigSyncPending 设置客户端配置同步状态为 pending
func (s *ClientService) SetConfigSyncPending(clientID uint) error {
	syncTime := time.Now()
	logger.Debugf("配置同步状态 客户端 ID=%d 设置为 pending", clientID)
	return s.clientRepo.UpdateConfigSyncStatus(clientID, "pending", "", syncTime)
}
