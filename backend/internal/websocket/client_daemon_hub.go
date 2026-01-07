package websocket

import (
	"frp-web-panel/internal/logger"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ClientDaemonHub 管理客户端守护程序的WebSocket连接
type ClientDaemonHub struct {
	clients                   map[uint]*DaemonConnection
	Register                  chan *DaemonConnection
	Unregister                chan *DaemonConnection
	mu                        sync.RWMutex
	statusCallback            ClientStatusCallback
	updateProgressCallback    UpdateProgressCallback
	updateResultCallback      UpdateResultCallback
	versionReportCallback     VersionReportCallback
	logDataCallback           LogDataCallback
	frpcControlResultCallback FrpcControlResultCallback
	configSyncResultCallback  ConfigSyncResultCallback
	frpcControlWaiters        map[uint]chan *FrpcControlResult
}

// DaemonConnection 客户端守护程序连接
type DaemonConnection struct {
	ClientID uint
	Conn     *websocket.Conn
	Send     chan []byte
	Hub      *ClientDaemonHub
}

var ClientDaemonHubInstance *ClientDaemonHub

func init() {
	ClientDaemonHubInstance = NewClientDaemonHub()
	go ClientDaemonHubInstance.Run()
}

func NewClientDaemonHub() *ClientDaemonHub {
	return &ClientDaemonHub{
		clients:    make(map[uint]*DaemonConnection),
		Register:   make(chan *DaemonConnection),
		Unregister: make(chan *DaemonConnection),
	}
}

// SetStatusCallback 设置客户端状态变更回调函数
func (h *ClientDaemonHub) SetStatusCallback(callback ClientStatusCallback) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.statusCallback = callback
	logger.Debugf("[ClientDaemonHub] 状态回调函数已设置")
}

// GetStatusCallback 获取当前状态回调函数
func (h *ClientDaemonHub) GetStatusCallback() ClientStatusCallback {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.statusCallback
}

// SetUpdateProgressCallback 设置更新进度回调函数
func (h *ClientDaemonHub) SetUpdateProgressCallback(callback UpdateProgressCallback) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.updateProgressCallback = callback
	logger.Debugf("[ClientDaemonHub] 更新进度回调函数已设置")
}

// SetUpdateResultCallback 设置更新结果回调函数
func (h *ClientDaemonHub) SetUpdateResultCallback(callback UpdateResultCallback) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.updateResultCallback = callback
	logger.Debugf("[ClientDaemonHub] 更新结果回调函数已设置")
}

// SetVersionReportCallback 设置版本上报回调函数
func (h *ClientDaemonHub) SetVersionReportCallback(callback VersionReportCallback) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.versionReportCallback = callback
	logger.Debugf("[ClientDaemonHub] 版本上报回调函数已设置")
}

// SetLogDataCallback 设置日志数据回调函数
func (h *ClientDaemonHub) SetLogDataCallback(callback LogDataCallback) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.logDataCallback = callback
	logger.Debugf("[ClientDaemonHub] 日志数据回调函数已设置")
}

// SetFrpcControlResultCallback 设置frpc控制结果回调函数
func (h *ClientDaemonHub) SetFrpcControlResultCallback(callback FrpcControlResultCallback) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.frpcControlResultCallback = callback
	logger.Debugf("[ClientDaemonHub] frpc控制结果回调函数已设置")
}

// SetConfigSyncResultCallback 设置配置同步结果回调函数
func (h *ClientDaemonHub) SetConfigSyncResultCallback(callback ConfigSyncResultCallback) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.configSyncResultCallback = callback
	logger.Debugf("[ClientDaemonHub] 配置同步结果回调函数已设置")
}

func (h *ClientDaemonHub) Run() {
	for {
		select {
		case conn := <-h.Register:
			h.mu.Lock()
			h.clients[conn.ClientID] = conn
			callback := h.statusCallback
			h.mu.Unlock()
			logger.Infof("[ClientDaemonHub] 客户端 %d 已连接", conn.ClientID)
			if callback != nil {
				go callback(conn.ClientID, true)
			}

		case conn := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.clients[conn.ClientID]; ok {
				delete(h.clients, conn.ClientID)
				close(conn.Send)
				logger.Infof("[ClientDaemonHub] 客户端 %d 已断开", conn.ClientID)
				callback := h.statusCallback
				h.mu.Unlock()
				if callback != nil {
					go callback(conn.ClientID, false)
				}
			} else {
				h.mu.Unlock()
			}
		}
	}
}

// IsClientOnline 检查客户端是否在线
func (h *ClientDaemonHub) IsClientOnline(clientID uint) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, exists := h.clients[clientID]
	return exists
}

// GetOnlineCount 获取在线客户端数量
func (h *ClientDaemonHub) GetOnlineCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// GetOnlineClientIDs 获取所有在线客户端ID列表
func (h *ClientDaemonHub) GetOnlineClientIDs() []uint {
	h.mu.RLock()
	defer h.mu.RUnlock()
	ids := make([]uint, 0, len(h.clients))
	for id := range h.clients {
		ids = append(ids, id)
	}
	return ids
}

// WritePump 处理发送消息
func (dc *DaemonConnection) WritePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		dc.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-dc.Send:
			dc.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				dc.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := dc.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			dc.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := dc.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ReadPump 处理接收消息
func (dc *DaemonConnection) ReadPump(handler MessageHandler) {
	defer func() {
		dc.Hub.Unregister <- dc
		dc.Conn.Close()
	}()

	dc.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	dc.Conn.SetPongHandler(func(string) error {
		dc.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var msg Message
		err := dc.Conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Errorf("WebSocket错误: %v", err)
			}
			break
		}
		handler.HandleMessage(dc.ClientID, &msg)
	}
}
