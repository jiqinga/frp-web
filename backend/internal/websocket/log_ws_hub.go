package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// LogWSHub 管理前端日志 WebSocket 连接
type LogWSHub struct {
	clients map[uint]map[*LogWSClient]bool // clientID -> connections
	mu      sync.RWMutex
}

// LogWSClient 前端日志 WebSocket 客户端
type LogWSClient struct {
	ClientID uint
	Conn     *websocket.Conn
	Send     chan []byte
	Hub      *LogWSHub
}

// LogWSHubInstance 全局实例
var LogWSHubInstance *LogWSHub

func init() {
	LogWSHubInstance = NewLogWSHub()
}

// NewLogWSHub 创建新的日志 WebSocket Hub
func NewLogWSHub() *LogWSHub {
	return &LogWSHub{
		clients: make(map[uint]map[*LogWSClient]bool),
	}
}

// Register 注册客户端连接
func (h *LogWSHub) Register(client *LogWSClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[client.ClientID] == nil {
		h.clients[client.ClientID] = make(map[*LogWSClient]bool)
	}
	h.clients[client.ClientID][client] = true
	log.Printf("[LogWSHub] 前端连接已注册，clientID=%d", client.ClientID)
}

// Unregister 注销客户端连接
func (h *LogWSHub) Unregister(client *LogWSClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conns, ok := h.clients[client.ClientID]; ok {
		if _, exists := conns[client]; exists {
			delete(conns, client)
			close(client.Send)
			if len(conns) == 0 {
				delete(h.clients, client.ClientID)
			}
			log.Printf("[LogWSHub] 前端连接已注销，clientID=%d", client.ClientID)
		}
	}
}

// BroadcastLog 向指定 clientID 的所有前端连接广播日志
func (h *LogWSHub) BroadcastLog(clientID uint, logType string, line string, timestamp int64) {
	h.mu.RLock()
	conns := h.clients[clientID]
	connCount := len(conns)
	h.mu.RUnlock()

	log.Printf("[LogWSHub] BroadcastLog: clientID=%d, logType=%s, 连接数=%d, line=%s", clientID, logType, connCount, line)

	if connCount == 0 {
		log.Printf("[LogWSHub] ⚠️ 没有找到 clientID=%d 的前端连接", clientID)
		return
	}

	msg := map[string]interface{}{
		"type":      "log_data",
		"log_type":  logType,
		"content":   line,
		"timestamp": timestamp,
	}
	data, _ := json.Marshal(msg)

	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range conns {
		select {
		case client.Send <- data:
		default:
			// 发送缓冲区满，跳过
		}
	}
}

// WritePump 处理发送消息
func (c *LogWSClient) WritePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
		c.Hub.Unregister(c)
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ReadPump 处理接收消息（保持连接活跃）
func (c *LogWSClient) ReadPump() {
	defer func() {
		c.Hub.Unregister(c)
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
