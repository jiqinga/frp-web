/*
 * 事件订阅器 - 将 Service 层事件转发到 WebSocket Hub
 */
package websocket

import (
	"encoding/json"
	"fmt"
	"frp-web-panel/internal/events"
	"time"
)

// SetupEventSubscribers 设置事件订阅，将事件总线事件转发到 WebSocket Hub
func SetupEventSubscribers(hub *Hub) {
	eventBus := events.GetEventBus()

	// 订阅 SSH 日志事件
	eventBus.Subscribe(events.EventSSHLog, func(e events.Event) {
		event := e.(events.SSHLogEvent)
		logType := "info"
		progress := 0

		if len(event.Log) > 5 && event.Log[0:5] == "下载进度:" {
			logType = "progress"
			if len(event.Log) > 7 {
				var p int
				if n, _ := fmt.Sscanf(event.Log[5:], " %d%%", &p); n == 1 {
					progress = p
				}
			}
		} else if len(event.Log) > 2 && (event.Log[0:2] == "错误" || event.Log[0:2] == "失败") {
			logType = "error"
		} else if len(event.Log) > 2 && (event.Log[0:2] == "完成" || event.Log[0:2] == "成功") {
			logType = "success"
		}

		msg := map[string]interface{}{
			"type":      "ssh_log",
			"server_id": event.ServerID,
			"operation": event.Operation,
			"log":       event.Log,
			"log_type":  logType,
			"progress":  progress,
			"timestamp": time.Now().Format("15:04:05"),
		}
		data, _ := json.Marshal(msg)
		hub.BroadcastMessage(data)
	})

	// 订阅服务器状态事件
	eventBus.Subscribe(events.EventServerStatus, func(e events.Event) {
		event := e.(events.ServerStatusEvent)
		msg := map[string]interface{}{
			"type":      "server_status_update",
			"timestamp": time.Now().Format(time.RFC3339),
			"data": []map[string]interface{}{
				{
					"server_id":   event.ServerID,
					"server_name": event.ServerName,
					"status":      event.Status,
				},
			},
		}
		data, _ := json.Marshal(msg)
		hub.BroadcastMessage(data)
	})

	// 订阅证书进度事件
	eventBus.Subscribe(events.EventCertProgress, func(e events.Event) {
		event := e.(events.CertProgressEvent)
		msg := map[string]interface{}{
			"type":      "cert_progress",
			"task_id":   event.TaskID,
			"domain":    event.Domain,
			"step":      event.Step,
			"message":   event.Message,
			"error":     event.Error,
			"timestamp": event.Timestamp,
		}
		data, _ := json.Marshal(msg)
		hub.BroadcastMessage(data)
	})

	// 订阅流量更新事件
	eventBus.Subscribe(events.EventTrafficUpdate, func(e events.Event) {
		event := e.(events.TrafficUpdateEvent)
		msg := map[string]interface{}{
			"type":      "traffic_update",
			"timestamp": event.Timestamp.Format(time.RFC3339),
			"data":      event.Data,
		}
		data, _ := json.Marshal(msg)
		hub.BroadcastMessage(data)
	})

	// 订阅更新进度事件
	eventBus.Subscribe(events.EventUpdateProgress, func(e events.Event) {
		event := e.(events.UpdateProgressEvent)
		msg := map[string]interface{}{
			"type":      "client_update_progress",
			"timestamp": time.Now().Format(time.RFC3339),
			"data": map[string]interface{}{
				"client_id":        event.ClientID,
				"update_type":      event.UpdateType,
				"stage":            event.Stage,
				"progress":         event.Progress,
				"message":          event.Message,
				"total_bytes":      event.TotalBytes,
				"downloaded_bytes": event.DownloadedBytes,
			},
		}
		data, _ := json.Marshal(msg)
		hub.BroadcastMessage(data)
	})

	// 订阅更新结果事件
	eventBus.Subscribe(events.EventUpdateResult, func(e events.Event) {
		event := e.(events.UpdateResultEvent)
		msg := map[string]interface{}{
			"type":      "client_update_result",
			"timestamp": time.Now().Format(time.RFC3339),
			"data": map[string]interface{}{
				"client_id":   event.ClientID,
				"update_type": event.UpdateType,
				"success":     event.Success,
				"version":     event.Version,
				"message":     event.Message,
			},
		}
		data, _ := json.Marshal(msg)
		hub.BroadcastMessage(data)
	})
}
