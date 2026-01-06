/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-19 17:07:04
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-05 11:04:27
 * @FilePath            : frp-web-testbackendinternalfrptypes.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package frp

import "time"

// ServerInfo frps 服务器信息
type ServerInfo struct {
	Version         string `json:"version"`
	BindPort        int    `json:"bindPort"`
	TotalTrafficIn  int64  `json:"totalTrafficIn"`
	TotalTrafficOut int64  `json:"totalTrafficOut"`
	CurConns        int    `json:"curConns"`
	ClientCounts    int    `json:"clientCounts"`
}

// ProxyConf 代理配置
type ProxyConf struct {
	Type       string `json:"type"`
	RemotePort int    `json:"remotePort,omitempty"`
}

// ProxyInfo 代理信息
type ProxyInfo struct {
	Name            string    `json:"name"`
	Conf            ProxyConf `json:"conf"`
	TodayTrafficIn  int64     `json:"todayTrafficIn"`
	TodayTrafficOut int64     `json:"todayTrafficOut"`
	CurConns        int       `json:"curConns"`
	Status          string    `json:"status"`
}

// ProxyList 代理列表
type ProxyList struct {
	Proxies []ProxyInfo `json:"proxies"`
}

// ClientStatus frpc 客户端状态
type ClientStatus struct {
	Proxies []ClientProxyStatus `json:"proxies"`
}

// ClientProxyStatus 客户端代理状态
type ClientProxyStatus struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Status string `json:"status"`
	Err    string `json:"err,omitempty"`
}

// TrafficData 流量数据
type TrafficData struct {
	Name       string      `json:"name"`
	TrafficIn  []int64     `json:"trafficIn"`
	TrafficOut []int64     `json:"trafficOut"`
	Timestamps []time.Time `json:"timestamps"`
}

// ClientInfo 客户端信息
type ClientInfo struct {
	User         string    `json:"user"`
	RunID        string    `json:"runId"`
	Version      string    `json:"version"`
	ConnectTime  time.Time `json:"connectTime"`
	LastSeenTime time.Time `json:"lastSeenTime"`
}

// ClientListResponse 客户端列表响应
type ClientListResponse struct {
	Clients []ClientInfo `json:"clients"`
}

// ProxyTrafficData 单个隧道的流量数据
type ProxyTrafficData struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	TrafficIn  int64  `json:"traffic_in"`
	TrafficOut int64  `json:"traffic_out"`
}

// FrpsMetrics frps服务器指标数据
type FrpsMetrics struct {
	// FRP服务器指标
	ClientCounts  int64              `json:"client_counts"`  // 客户端数量
	ProxyCounts   map[string]int64   `json:"proxy_counts"`   // 按类型统计的代理数量
	TotalProxies  int64              `json:"total_proxies"`  // 代理总数
	TrafficIn     int64              `json:"traffic_in"`     // 总入站流量(字节)
	TrafficOut    int64              `json:"traffic_out"`    // 总出站流量(字节)
	ProxyTraffics []ProxyTrafficData `json:"proxy_traffics"` // 每个隧道的流量数据

	// 进程指标
	CpuSeconds  float64 `json:"cpu_seconds"`  // CPU使用时间(秒)
	MemoryBytes int64   `json:"memory_bytes"` // 内存占用(字节)
	StartTime   float64 `json:"start_time"`   // 启动时间(Unix时间戳)
	Uptime      int64   `json:"uptime"`       // 运行时长(秒)

	// Go运行时指标
	Goroutines int64 `json:"goroutines"` // goroutine数量
}
