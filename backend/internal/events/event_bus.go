/*
 * 事件总线 - 解耦 Service 层与 WebSocket Hub
 */
package events

import (
	"sync"
	"time"
)

// EventType 事件类型
type EventType string

const (
	EventSSHLog         EventType = "ssh_log"
	EventServerStatus   EventType = "server_status"
	EventCertProgress   EventType = "cert_progress"
	EventTrafficUpdate  EventType = "traffic_update"
	EventUpdateProgress EventType = "update_progress"
	EventUpdateResult   EventType = "update_result"
)

// Event 事件接口
type Event interface {
	Type() EventType
}

// SSHLogEvent SSH日志事件
type SSHLogEvent struct {
	ServerID  uint
	Operation string
	Log       string
}

func (e SSHLogEvent) Type() EventType { return EventSSHLog }

// ServerStatusEvent 服务器状态事件
type ServerStatusEvent struct {
	ServerID   uint
	ServerName string
	Status     string
}

func (e ServerStatusEvent) Type() EventType { return EventServerStatus }

// CertProgressEvent 证书进度事件
type CertProgressEvent struct {
	TaskID    string
	Domain    string
	Step      string
	Message   string
	Error     string
	Timestamp string
}

func (e CertProgressEvent) Type() EventType { return EventCertProgress }

// TrafficUpdateEvent 流量更新事件
type TrafficUpdateEvent struct {
	Timestamp time.Time
	Data      []map[string]interface{}
}

func (e TrafficUpdateEvent) Type() EventType { return EventTrafficUpdate }

// UpdateProgressEvent 客户端更新进度事件
type UpdateProgressEvent struct {
	ClientID        uint
	UpdateType      string
	Stage           string
	Progress        int
	Message         string
	TotalBytes      int64
	DownloadedBytes int64
}

func (e UpdateProgressEvent) Type() EventType { return EventUpdateProgress }

// UpdateResultEvent 客户端更新结果事件
type UpdateResultEvent struct {
	ClientID   uint
	UpdateType string
	Success    bool
	Version    string
	Message    string
}

func (e UpdateResultEvent) Type() EventType { return EventUpdateResult }

// EventHandler 事件处理函数
type EventHandler func(Event)

// EventBus 事件总线
type EventBus struct {
	handlers map[EventType][]EventHandler
	mu       sync.RWMutex
}

// 全局事件总线实例
var (
	globalEventBus *EventBus
	eventBusOnce   sync.Once
)

// GetEventBus 获取全局事件总线实例
func GetEventBus() *EventBus {
	eventBusOnce.Do(func() {
		globalEventBus = &EventBus{
			handlers: make(map[EventType][]EventHandler),
		}
	})
	return globalEventBus
}

// Subscribe 订阅事件
func (eb *EventBus) Subscribe(eventType EventType, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.handlers[eventType] = append(eb.handlers[eventType], handler)
}

// Publish 发布事件
func (eb *EventBus) Publish(event Event) {
	eb.mu.RLock()
	handlers := eb.handlers[event.Type()]
	eb.mu.RUnlock()

	for _, handler := range handlers {
		go handler(event)
	}
}

// PublishSync 同步发布事件（阻塞直到所有处理完成）
func (eb *EventBus) PublishSync(event Event) {
	eb.mu.RLock()
	handlers := eb.handlers[event.Type()]
	eb.mu.RUnlock()

	for _, handler := range handlers {
		handler(event)
	}
}
