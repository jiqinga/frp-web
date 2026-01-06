/*
 * @Author              : 寂情啊
 * @Date                : 2025-12-04 17:26:50
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-04 17:27:01
 * @FilePath            : frp-web-testbackendinternalmodelserver_metrics.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package model

import "time"

// ServerMetricsHistory 服务器指标历史记录
type ServerMetricsHistory struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	ServerID    uint      `json:"server_id" gorm:"index:idx_server_time;not null"`
	CpuPercent  float64   `json:"cpu_percent"`
	MemoryBytes int64     `json:"memory_bytes"`
	TrafficIn   int64     `json:"traffic_in"`
	TrafficOut  int64     `json:"traffic_out"`
	RecordTime  time.Time `json:"record_time" gorm:"index:idx_server_time;not null"`
	CreatedAt   time.Time `json:"created_at"`
}

func (ServerMetricsHistory) TableName() string {
	return "server_metrics_history"
}
