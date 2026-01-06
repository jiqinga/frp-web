/*
 * @Author              : 寂情啊
 * @Date                : 2025-12-05 11:05:44
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-25 17:40:52
 * @FilePath            : frp-web-testbackendinternalrepositoryproxy_metrics_repo.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package repository

import (
	"frp-web-panel/internal/model"
	"frp-web-panel/pkg/database"
	"time"
)

type ProxyMetricsRepository struct{}

func NewProxyMetricsRepository() *ProxyMetricsRepository {
	return &ProxyMetricsRepository{}
}

// Create 创建隧道指标记录
func (r *ProxyMetricsRepository) Create(metrics *model.ProxyMetricsHistory) error {
	return database.DB.Create(metrics).Error
}

// BatchCreate 批量创建隧道指标记录
func (r *ProxyMetricsRepository) BatchCreate(metrics []model.ProxyMetricsHistory) error {
	if len(metrics) == 0 {
		return nil
	}
	return database.DB.Create(&metrics).Error
}

// GetHistory 获取指定隧道的历史指标
func (r *ProxyMetricsRepository) GetHistory(serverID uint, proxyName string, start, end time.Time) ([]model.ProxyMetricsHistory, error) {
	var records []model.ProxyMetricsHistory
	err := database.DB.Where("server_id = ? AND proxy_name = ? AND record_time BETWEEN ? AND ?", serverID, proxyName, start, end).
		Order("record_time ASC").Find(&records).Error
	return records, err
}

// GetLatestByServer 获取服务器下所有隧道的最新指标
func (r *ProxyMetricsRepository) GetLatestByServer(serverID uint) ([]model.ProxyMetricsHistory, error) {
	var records []model.ProxyMetricsHistory
	subQuery := database.DB.Model(&model.ProxyMetricsHistory{}).
		Select("MAX(id)").
		Where("server_id = ?", serverID).
		Group("proxy_name")
	err := database.DB.Where("id IN (?)", subQuery).Find(&records).Error
	return records, err
}

// DeleteOlderThan 删除指定时间之前的记录
func (r *ProxyMetricsRepository) DeleteOlderThan(before time.Time) (int64, error) {
	result := database.DB.Where("record_time < ?", before).Delete(&model.ProxyMetricsHistory{})
	return result.RowsAffected, result.Error
}

// GetDistinctProxyNames 获取指定服务器下所有不同的 proxy_name
func (r *ProxyMetricsRepository) GetDistinctProxyNames(serverID uint) ([]string, error) {
	var names []string
	err := database.DB.Model(&model.ProxyMetricsHistory{}).
		Where("server_id = ?", serverID).
		Distinct("proxy_name").
		Pluck("proxy_name", &names).Error
	return names, err
}

// TrafficTrendPoint 流量趋势数据点
type TrafficTrendPoint struct {
	Time     string `json:"time"`
	Inbound  int64  `json:"inbound"`
	Outbound int64  `json:"outbound"`
}

// GetHourlyTrafficTrend 获取按小时聚合的流量趋势
func (r *ProxyMetricsRepository) GetHourlyTrafficTrend(hours int) ([]TrafficTrendPoint, error) {
	startTime := time.Now().Add(-time.Duration(hours) * time.Hour)

	type Result struct {
		HourKey  string
		Hour     string
		TotalIn  int64
		TotalOut int64
	}

	var results []Result
	err := database.DB.Model(&model.ProxyMetricsHistory{}).
		Select("strftime('%Y-%m-%d %H', record_time) as hour_key, strftime('%H:00', record_time) as hour, SUM(traffic_in) as total_in, SUM(traffic_out) as total_out").
		Where("record_time >= ?", startTime).
		Group("hour_key").
		Order("hour_key ASC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// 创建数据映射
	dataMap := make(map[string]TrafficTrendPoint)
	for _, r := range results {
		dataMap[r.HourKey] = TrafficTrendPoint{
			Time:     r.Hour,
			Inbound:  r.TotalIn,
			Outbound: r.TotalOut,
		}
	}

	// 生成完整的24小时时间轴
	trend := make([]TrafficTrendPoint, 0, hours)
	now := time.Now()
	for i := hours - 1; i >= 0; i-- {
		t := now.Add(-time.Duration(i) * time.Hour)
		hourKey := t.Format("2006-01-02 15")
		hourLabel := t.Format("15:00")
		if data, ok := dataMap[hourKey]; ok {
			trend = append(trend, data)
		} else {
			trend = append(trend, TrafficTrendPoint{
				Time:     hourLabel,
				Inbound:  0,
				Outbound: 0,
			})
		}
	}

	return trend, nil
}

// GetTrafficSummaryByProxyNames 获取指定代理名称列表在时间范围内的流量汇总
func (r *ProxyMetricsRepository) GetTrafficSummaryByProxyNames(proxyNames []string, days int) (map[string]map[string]int64, error) {
	if len(proxyNames) == 0 {
		return make(map[string]map[string]int64), nil
	}

	startTime := time.Now().Add(-time.Duration(days) * 24 * time.Hour)

	type Result struct {
		ProxyName string
		TotalIn   int64
		TotalOut  int64
	}

	var results []Result
	err := database.DB.Model(&model.ProxyMetricsHistory{}).
		Select("proxy_name, COALESCE(SUM(traffic_in), 0) as total_in, COALESCE(SUM(traffic_out), 0) as total_out").
		Where("proxy_name IN ? AND record_time >= ?", proxyNames, startTime).
		Group("proxy_name").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	summary := make(map[string]map[string]int64)
	for _, r := range results {
		summary[r.ProxyName] = map[string]int64{
			"total_in":  r.TotalIn,
			"total_out": r.TotalOut,
		}
	}

	return summary, nil
}
