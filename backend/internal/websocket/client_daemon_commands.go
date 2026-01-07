package websocket

import (
	"encoding/json"
	"fmt"
	"frp-web-panel/internal/logger"
	"time"
)

// PushConfigUpdate 推送配置更新到指定客户端
func (h *ClientDaemonHub) PushConfigUpdate(clientID uint, config string, version int) error {
	h.mu.RLock()
	conn, exists := h.clients[clientID]
	h.mu.RUnlock()

	if !exists {
		return fmt.Errorf("客户端 %d 未连接", clientID)
	}

	msg := Message{
		Type: "config_update",
		Data: map[string]interface{}{
			"config":  config,
			"version": version,
		},
		Timestamp: time.Now().Unix(),
	}

	data, _ := json.Marshal(msg)
	select {
	case conn.Send <- data:
		return nil
	default:
		return fmt.Errorf("发送队列已满")
	}
}

// SendUpdateCommand 向指定客户端发送更新命令
func (h *ClientDaemonHub) SendUpdateCommand(clientID uint, updateType string, version string, downloadURL string, mirrorID uint) error {
	h.mu.RLock()
	conn, exists := h.clients[clientID]
	h.mu.RUnlock()

	if !exists {
		logger.Warnf("[ClientDaemonHub] 客户端 %d 未连接，无法发送更新命令", clientID)
		return fmt.Errorf("客户端 %d 未连接", clientID)
	}

	msg := Message{
		Type:     "update",
		ClientID: clientID,
		Data: map[string]interface{}{
			"update_type":  updateType,
			"version":      version,
			"download_url": downloadURL,
			"mirror_id":    mirrorID,
		},
		Timestamp: time.Now().Unix(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化更新命令失败: %v", err)
	}

	select {
	case conn.Send <- data:
		logger.Infof("[ClientDaemonHub] 已向客户端 %d 发送更新命令: type=%s, version=%s", clientID, updateType, version)
		return nil
	default:
		return fmt.Errorf("发送队列已满，无法发送更新命令")
	}
}

// PushCertSync 推送证书同步到指定客户端
func (h *ClientDaemonHub) PushCertSync(clientID uint, domain string, certPEM string, keyPEM string) error {
	h.mu.RLock()
	conn, exists := h.clients[clientID]
	h.mu.RUnlock()

	if !exists {
		logger.Warnf("[ClientDaemonHub] 客户端 %d 未连接，无法推送证书", clientID)
		return fmt.Errorf("客户端 %d 未连接", clientID)
	}

	msg := Message{
		Type:     "cert_sync",
		ClientID: clientID,
		Data: map[string]interface{}{
			"domain":   domain,
			"cert_pem": certPEM,
			"key_pem":  keyPEM,
		},
		Timestamp: time.Now().Unix(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化证书同步消息失败: %v", err)
	}

	select {
	case conn.Send <- data:
		logger.Infof("[ClientDaemonHub] 已向客户端 %d 推送证书: domain=%s", clientID, domain)
		return nil
	default:
		return fmt.Errorf("发送队列已满，无法推送证书")
	}
}

// PushCertDelete 推送证书删除命令到指定客户端
func (h *ClientDaemonHub) PushCertDelete(clientID uint, domain string) error {
	h.mu.RLock()
	conn, exists := h.clients[clientID]
	h.mu.RUnlock()

	if !exists {
		logger.Warnf("[ClientDaemonHub] 客户端 %d 未连接，无法推送证书删除", clientID)
		return fmt.Errorf("客户端 %d 未连接", clientID)
	}

	msg := Message{
		Type:     "cert_delete",
		ClientID: clientID,
		Data: map[string]interface{}{
			"domain": domain,
		},
		Timestamp: time.Now().Unix(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化证书删除消息失败: %v", err)
	}

	select {
	case conn.Send <- data:
		logger.Infof("[ClientDaemonHub] 已向客户端 %d 推送证书删除: domain=%s", clientID, domain)
		return nil
	default:
		return fmt.Errorf("发送队列已满，无法推送证书删除")
	}
}

// SendLogStreamCommand 向指定客户端发送日志流命令
func (h *ClientDaemonHub) SendLogStreamCommand(clientID uint, logType string, action string, lines int) error {
	h.mu.RLock()
	conn, exists := h.clients[clientID]
	h.mu.RUnlock()

	if !exists {
		logger.Warnf("[ClientDaemonHub] 客户端 %d 未连接，无法发送日志流命令", clientID)
		return fmt.Errorf("客户端 %d 未连接", clientID)
	}

	msg := Message{
		Type:     "log_stream",
		ClientID: clientID,
		Data: map[string]interface{}{
			"log_type": logType,
			"action":   action,
			"lines":    lines,
		},
		Timestamp: time.Now().Unix(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化日志流命令失败: %v", err)
	}

	select {
	case conn.Send <- data:
		logger.Infof("[ClientDaemonHub] 已向客户端 %d 发送日志流命令: type=%s, action=%s", clientID, logType, action)
		return nil
	default:
		return fmt.Errorf("发送队列已满，无法发送日志流命令")
	}
}

// FrpcControlResult frpc控制结果
type FrpcControlResult struct {
	Success bool
	Message string
}

// SendFrpcControlCommand 向指定客户端发送frpc控制命令
func (h *ClientDaemonHub) SendFrpcControlCommand(clientID uint, action string) error {
	h.mu.RLock()
	conn, exists := h.clients[clientID]
	h.mu.RUnlock()

	if !exists {
		logger.Warnf("[ClientDaemonHub] 客户端 %d 未连接，无法发送frpc控制命令", clientID)
		return fmt.Errorf("客户端 %d 未连接", clientID)
	}

	msg := Message{
		Type:     "frpc_control",
		ClientID: clientID,
		Data: map[string]interface{}{
			"action": action,
		},
		Timestamp: time.Now().Unix(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化frpc控制命令失败: %v", err)
	}

	select {
	case conn.Send <- data:
		logger.Infof("[ClientDaemonHub] 已向客户端 %d 发送frpc控制命令: action=%s", clientID, action)
		return nil
	default:
		return fmt.Errorf("发送队列已满，无法发送frpc控制命令")
	}
}

// SendFrpcControlCommandAndWait 发送frpc控制命令并等待结果
func (h *ClientDaemonHub) SendFrpcControlCommandAndWait(clientID uint, action string, timeout time.Duration) (*FrpcControlResult, error) {
	// 创建结果通道
	resultChan := make(chan *FrpcControlResult, 1)
	key := fmt.Sprintf("%d_%s_%d", clientID, action, time.Now().UnixNano())

	// 注册等待
	h.mu.Lock()
	if h.frpcControlWaiters == nil {
		h.frpcControlWaiters = make(map[uint]chan *FrpcControlResult)
	}
	h.frpcControlWaiters[clientID] = resultChan
	h.mu.Unlock()

	// 清理函数
	defer func() {
		h.mu.Lock()
		delete(h.frpcControlWaiters, clientID)
		h.mu.Unlock()
	}()

	// 发送命令
	if err := h.SendFrpcControlCommand(clientID, action); err != nil {
		return nil, err
	}

	logger.Debugf("[ClientDaemonHub] 等待客户端 %d 的frpc控制结果: key=%s", clientID, key)

	// 等待结果或超时
	select {
	case result := <-resultChan:
		logger.Debugf("[ClientDaemonHub] 收到客户端 %d 的frpc控制结果: success=%v", clientID, result.Success)
		return result, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("等待frpc控制结果超时")
	}
}

// NotifyFrpcControlResult 通知frpc控制结果
func (h *ClientDaemonHub) NotifyFrpcControlResult(clientID uint, success bool, message string) {
	h.mu.RLock()
	resultChan, exists := h.frpcControlWaiters[clientID]
	h.mu.RUnlock()

	if exists {
		select {
		case resultChan <- &FrpcControlResult{Success: success, Message: message}:
			logger.Debugf("[ClientDaemonHub] 已通知frpc控制结果: clientID=%d, success=%v", clientID, success)
		default:
			logger.Warnf("[ClientDaemonHub] 通知frpc控制结果失败，通道已满: clientID=%d", clientID)
		}
	}
}

// SendShutdownCommand 向指定客户端发送停止命令
func (h *ClientDaemonHub) SendShutdownCommand(clientID uint) error {
	h.mu.RLock()
	conn, exists := h.clients[clientID]
	h.mu.RUnlock()

	if !exists {
		logger.Warnf("[ClientDaemonHub] 客户端 %d 未连接，无法发送停止命令", clientID)
		return fmt.Errorf("客户端 %d 未连接", clientID)
	}

	msg := Message{
		Type:      "shutdown",
		ClientID:  clientID,
		Timestamp: time.Now().Unix(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化停止命令失败: %v", err)
	}

	select {
	case conn.Send <- data:
		logger.Infof("[ClientDaemonHub] 已向客户端 %d 发送停止命令", clientID)
		return nil
	default:
		return fmt.Errorf("发送队列已满，无法发送停止命令")
	}
}
