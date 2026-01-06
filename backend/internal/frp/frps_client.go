/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-19 17:08:26
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-05 11:04:56
 * @FilePath            : frp-web-testbackendinternalfrpfrps_client.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package frp

import (
	"fmt"
	"log"
	"strings"
	"time"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

type FrpsClient struct {
	*Client
}

func NewFrpsClient(host string, port int, username, password string) *FrpsClient {
	return &FrpsClient{
		Client: NewClient(host, port, username, password, 10*time.Second),
	}
}

func (c *FrpsClient) GetServerInfo() (*ServerInfo, error) {
	var info ServerInfo
	if err := c.doRequest("GET", "/api/serverinfo", &info); err != nil {
		return nil, err
	}
	return &info, nil
}

func (c *FrpsClient) GetProxies(proxyType string) (*ProxyList, error) {
	var list ProxyList
	path := fmt.Sprintf("/api/proxy/%s", proxyType)
	if err := c.doRequest("GET", path, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

// GetAllProxies 获取所有类型的代理列表
func (c *FrpsClient) GetAllProxies() (map[string][]ProxyInfo, error) {
	result := make(map[string][]ProxyInfo)
	proxyTypes := []string{"tcp", "udp", "http", "https", "stcp", "xtcp"}

	for _, pType := range proxyTypes {
		list, err := c.GetProxies(pType)
		if err != nil {
			// 忽略单个类型的错误,继续查询其他类型
			continue
		}
		if len(list.Proxies) > 0 {
			result[pType] = list.Proxies
		}
	}

	return result, nil
}

func (c *FrpsClient) HealthCheck() error {
	return c.doRequest("GET", "/healthz", nil)
}

// GetProxyTraffic 获取代理的历史流量数据
func (c *FrpsClient) GetProxyTraffic(proxyName string) (*TrafficData, error) {
	var data TrafficData
	path := fmt.Sprintf("/api/traffic/%s", proxyName)
	if err := c.doRequest("GET", path, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// GetClients 获取所有连接的客户端列表
func (c *FrpsClient) GetClients() (*ClientListResponse, error) {
	log.Printf("[DEBUG] GetClients - 开始查询客户端列表")
	log.Printf("[DEBUG] GetClients - Base URL: %s", c.baseURL)

	var response ClientListResponse
	if err := c.doRequest("GET", "/api/client", &response); err != nil {
		log.Printf("[DEBUG] GetClients - 查询失败: %v", err)
		return nil, err
	}

	log.Printf("[DEBUG] GetClients - 查询成功, 客户端数量: %d", len(response.Clients))
	for i, client := range response.Clients {
		log.Printf("[DEBUG] GetClients - 客户端[%d]: User=%s, Version=%s, RunID=%s",
			i, client.User, client.Version, client.RunID)
	}

	return &response, nil
}

// GetMetrics 获取Prometheus格式的指标数据
func (c *FrpsClient) GetMetrics() (*FrpsMetrics, error) {
	body, err := c.doRequestRaw("GET", "/metrics")
	if err != nil {
		return nil, err
	}

	return parsePrometheusMetrics(body)
}

// parsePrometheusMetrics 解析Prometheus格式的指标数据
func parsePrometheusMetrics(data string) (*FrpsMetrics, error) {
	metrics := &FrpsMetrics{
		ProxyCounts:   make(map[string]int64),
		ProxyTraffics: []ProxyTrafficData{},
	}

	parser := expfmt.TextParser{}
	families, err := parser.TextToMetricFamilies(strings.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("解析Prometheus指标失败: %w", err)
	}

	// 临时存储每个隧道的流量数据
	proxyTrafficMap := make(map[string]*ProxyTrafficData)

	for name, family := range families {
		switch name {
		case "frp_server_client_counts":
			metrics.ClientCounts = getGaugeValue(family)
		case "frp_server_proxy_counts":
			for _, m := range family.Metric {
				proxyType := getLabelValue(m.Label, "type")
				if proxyType != "" {
					metrics.ProxyCounts[proxyType] = int64(m.GetGauge().GetValue())
					metrics.TotalProxies += int64(m.GetGauge().GetValue())
				}
			}
		case "frp_server_traffic_in":
			for _, m := range family.Metric {
				trafficIn := int64(m.GetCounter().GetValue())
				metrics.TrafficIn += trafficIn
				// 提取单个隧道的流量
				proxyName := getLabelValue(m.Label, "name")
				proxyType := getLabelValue(m.Label, "type")
				if proxyName != "" {
					key := proxyName + ":" + proxyType
					if _, exists := proxyTrafficMap[key]; !exists {
						proxyTrafficMap[key] = &ProxyTrafficData{Name: proxyName, Type: proxyType}
					}
					proxyTrafficMap[key].TrafficIn = trafficIn
				}
			}
		case "frp_server_traffic_out":
			for _, m := range family.Metric {
				trafficOut := int64(m.GetCounter().GetValue())
				metrics.TrafficOut += trafficOut
				// 提取单个隧道的流量
				proxyName := getLabelValue(m.Label, "name")
				proxyType := getLabelValue(m.Label, "type")
				if proxyName != "" {
					key := proxyName + ":" + proxyType
					if _, exists := proxyTrafficMap[key]; !exists {
						proxyTrafficMap[key] = &ProxyTrafficData{Name: proxyName, Type: proxyType}
					}
					proxyTrafficMap[key].TrafficOut = trafficOut
				}
			}
		case "process_cpu_seconds_total":
			metrics.CpuSeconds = getCounterValue(family)
		case "process_resident_memory_bytes":
			metrics.MemoryBytes = int64(getGaugeValue(family))
		case "process_start_time_seconds":
			metrics.StartTime = float64(getGaugeValue(family))
			metrics.Uptime = time.Now().Unix() - int64(metrics.StartTime)
		case "go_goroutines":
			metrics.Goroutines = getGaugeValue(family)
		}
	}

	// 转换 map 为 slice
	for _, pt := range proxyTrafficMap {
		metrics.ProxyTraffics = append(metrics.ProxyTraffics, *pt)
	}

	return metrics, nil
}

func getGaugeValue(family *dto.MetricFamily) int64 {
	if len(family.Metric) > 0 && family.Metric[0].Gauge != nil {
		return int64(family.Metric[0].Gauge.GetValue())
	}
	return 0
}

func getCounterValue(family *dto.MetricFamily) float64 {
	if len(family.Metric) > 0 && family.Metric[0].Counter != nil {
		return family.Metric[0].Counter.GetValue()
	}
	return 0
}

func getLabelValue(labels []*dto.LabelPair, name string) string {
	for _, label := range labels {
		if label.GetName() == name {
			return label.GetValue()
		}
	}
	return ""
}
