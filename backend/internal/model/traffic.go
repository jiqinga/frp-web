/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-14 16:00:15
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-09 16:03:53
 * @FilePath            : frp-web-testbackendinternalmodeltraffic.go
 * @Description         : 流量统计模型
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package model

// TrafficSummary 流量汇总统计
type TrafficSummary struct {
	TotalBytesIn   int64 `json:"total_bytes_in"`
	TotalBytesOut  int64 `json:"total_bytes_out"`
	CurrentRateIn  int64 `json:"current_rate_in"`
	CurrentRateOut int64 `json:"current_rate_out"`
	ActiveProxies  int   `json:"active_proxies"`
	TotalProxies   int   `json:"total_proxies"`
}
