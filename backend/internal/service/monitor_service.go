/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-14 16:23:04
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-11-14 16:24:37
 * @FilePath            : frp-web-testbackendinternalservicemonitor_service.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package service

import (
	"frp-web-panel/internal/repository"
)

type MonitorService struct {
	clientRepo  *repository.ClientRepository
	proxyRepo   *repository.ProxyRepository
	trafficRepo *repository.TrafficRepository
}

func NewMonitorService() *MonitorService {
	return &MonitorService{
		clientRepo:  repository.NewClientRepository(),
		proxyRepo:   repository.NewProxyRepository(),
		trafficRepo: repository.NewTrafficRepository(),
	}
}

type OverviewData struct {
	TotalClients   int64 `json:"total_clients"`
	TotalProxies   int64 `json:"total_proxies"`
	ActiveProxies  int64 `json:"active_proxies"`
	TotalBytesIn   int64 `json:"total_bytes_in"`
	TotalBytesOut  int64 `json:"total_bytes_out"`
	CurrentRateIn  int64 `json:"current_rate_in"`
	CurrentRateOut int64 `json:"current_rate_out"`
}

type StatsData struct {
	ProxyTypeStats map[string]int64 `json:"proxy_type_stats"`
	RecentLogs     []interface{}    `json:"recent_logs"`
}

func (s *MonitorService) GetOverview() (*OverviewData, error) {
	totalClients, err := s.clientRepo.Count()
	if err != nil {
		return nil, err
	}

	totalProxies, err := s.proxyRepo.Count()
	if err != nil {
		return nil, err
	}

	summary, err := s.trafficRepo.GetTrafficSummary()
	if err != nil {
		return nil, err
	}

	return &OverviewData{
		TotalClients:   totalClients,
		TotalProxies:   totalProxies,
		ActiveProxies:  int64(summary.ActiveProxies),
		TotalBytesIn:   summary.TotalBytesIn,
		TotalBytesOut:  summary.TotalBytesOut,
		CurrentRateIn:  summary.CurrentRateIn,
		CurrentRateOut: summary.CurrentRateOut,
	}, nil
}

func (s *MonitorService) GetStats() (*StatsData, error) {
	proxyTypeStats, err := s.proxyRepo.GetProxyTypeStats()
	if err != nil {
		return nil, err
	}

	return &StatsData{
		ProxyTypeStats: proxyTypeStats,
		RecentLogs:     []interface{}{},
	}, nil
}
