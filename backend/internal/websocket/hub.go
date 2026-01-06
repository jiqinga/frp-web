/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-17 16:16:18
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2026-01-06 14:59:29
 * @FilePath            : frp-web-testbackendinternalwebsockethub.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	h.RunWithContext(context.Background())
}

func (h *Hub) RunWithContext(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			h.mu.Lock()
			for client := range h.clients {
				close(client.Send)
				delete(h.clients, client)
			}
			h.mu.Unlock()
			return
		case client := <-h.Register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) BroadcastMessage(message []byte) {
	h.broadcast <- message
}

type SSHLogMessage struct {
	Type      string `json:"type"`
	ServerID  uint   `json:"server_id"`
	Operation string `json:"operation"`
	Log       string `json:"log"`
	LogType   string `json:"log_type"`
	Progress  int    `json:"progress,omitempty"`
	Timestamp string `json:"timestamp"`
}

func (h *Hub) BroadcastSSHLog(serverID uint, operation string, log string) {
	logType := "info"
	progress := 0

	if len(log) > 5 && log[0:5] == "下载进度:" {
		logType = "progress"
		if len(log) > 7 {
			var p int
			if n, _ := fmt.Sscanf(log[5:], " %d%%", &p); n == 1 {
				progress = p
			}
		}
	} else if len(log) > 2 && (log[0:2] == "错误" || log[0:2] == "失败") {
		logType = "error"
	} else if len(log) > 2 && (log[0:2] == "完成" || log[0:2] == "成功") {
		logType = "success"
	}

	msg := SSHLogMessage{
		Type:      "ssh_log",
		ServerID:  serverID,
		Operation: operation,
		Log:       log,
		LogType:   logType,
		Progress:  progress,
		Timestamp: time.Now().Format("15:04:05"),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	h.BroadcastMessage(data)
}

func (h *Hub) BroadcastServerStatus(serverID uint, serverName string, status string) {
	message := map[string]interface{}{
		"type":      "server_status_update",
		"timestamp": time.Now().Format(time.RFC3339),
		"data": []map[string]interface{}{
			{
				"server_id":   serverID,
				"server_name": serverName,
				"status":      status,
			},
		},
	}

	data, err := json.Marshal(message)
	if err != nil {
		return
	}

	h.BroadcastMessage(data)
}

// CertProgressMessage 证书申请进度消息
type CertProgressMessage struct {
	Type      string `json:"type"`
	TaskID    string `json:"task_id"`
	Domain    string `json:"domain"`
	Step      string `json:"step"`
	Message   string `json:"message,omitempty"`
	Error     string `json:"error,omitempty"`
	Timestamp string `json:"timestamp"`
}

// BroadcastCertProgress 广播证书申请进度
func (h *Hub) BroadcastCertProgress(taskID, domain, step, message, errMsg string) {
	msg := CertProgressMessage{
		Type:      "cert_progress",
		TaskID:    taskID,
		Domain:    domain,
		Step:      step,
		Message:   message,
		Error:     errMsg,
		Timestamp: time.Now().Format("15:04:05"),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	h.BroadcastMessage(data)
}

// FrpcControlResultMessage frpc控制结果消息
type FrpcControlResultMessage struct {
	Type      string `json:"type"`
	ClientID  uint   `json:"client_id"`
	Action    string `json:"action"`
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// BroadcastFrpcControlResult 广播frpc控制结果
func (h *Hub) BroadcastFrpcControlResult(clientID uint, action string, success bool, message string) {
	msg := FrpcControlResultMessage{
		Type:      "frpc_control_result",
		ClientID:  clientID,
		Action:    action,
		Success:   success,
		Message:   message,
		Timestamp: time.Now().Format("15:04:05"),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	h.BroadcastMessage(data)
}
