/*
 * @Author              : 寂情啊
 * @Date                : 2025-12-29 15:26:49
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-29 15:27:01
 * @FilePath            : frp-web-testbackendinternalwebsocketclient_daemon_types.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package websocket

// ClientStatusCallback 客户端状态变更回调函数类型
type ClientStatusCallback func(clientID uint, online bool)

// UpdateProgressCallback 更新进度回调函数类型
type UpdateProgressCallback func(clientID uint, updateType string, stage string, progress int, message string, totalBytes int64, downloadedBytes int64)

// UpdateResultCallback 更新结果回调函数类型
type UpdateResultCallback func(clientID uint, updateType string, success bool, version string, message string)

// VersionReportCallback 版本上报回调函数类型
type VersionReportCallback func(clientID uint, frpcVersion string, daemonVersion string, os string, arch string)

// LogDataCallback 日志数据回调函数类型
type LogDataCallback func(clientID uint, logType string, line string, timestamp int64)

// FrpcControlResultCallback frpc控制结果回调函数类型
type FrpcControlResultCallback func(clientID uint, action string, success bool, message string)

// ConfigSyncResultCallback 配置同步结果回调函数类型
type ConfigSyncResultCallback func(clientID uint, success bool, errorMsg string, rolledBack bool)

// Message WebSocket消息结构
type Message struct {
	Type      string                 `json:"type"`
	ClientID  uint                   `json:"client_id,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp int64                  `json:"timestamp"`
}

// MessageHandler 消息处理器接口
type MessageHandler interface {
	HandleMessage(clientID uint, msg *Message)
}
