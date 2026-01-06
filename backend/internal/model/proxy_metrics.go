/*
 * @Author              : 寂情啊
 * @Date                : 2025-12-05 11:05:15
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-05 11:05:23
 * @FilePath            : frp-web-testbackendinternalmodelproxy_metrics.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package model

import "time"

// ProxyMetricsHistory 隧道级别的流量历史记录
type ProxyMetricsHistory struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	ServerID   uint      `json:"server_id" gorm:"index:idx_proxy_metrics;not null"`
	ProxyName  string    `json:"proxy_name" gorm:"index:idx_proxy_metrics;size:100;not null"`
	ProxyType  string    `json:"proxy_type" gorm:"size:20"`
	TrafficIn  int64     `json:"traffic_in"`  // 累计入站流量
	TrafficOut int64     `json:"traffic_out"` // 累计出站流量
	RateIn     int64     `json:"rate_in"`     // 入站速率 bytes/s
	RateOut    int64     `json:"rate_out"`    // 出站速率 bytes/s
	RecordTime time.Time `json:"record_time" gorm:"index:idx_proxy_metrics;not null"`
	CreatedAt  time.Time `json:"created_at"`
}

func (ProxyMetricsHistory) TableName() string {
	return "proxy_metrics_history"
}
