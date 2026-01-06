/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-14 16:00:42
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-09 16:05:02
 * @FilePath            : frp-web-testbackendinternalrepositorytraffic_repo.go
 * @Description         : 流量统计仓库
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package repository

import (
	"frp-web-panel/internal/model"
	"frp-web-panel/pkg/database"
	"time"
)

type TrafficRepository struct{}

func NewTrafficRepository() *TrafficRepository {
	return &TrafficRepository{}
}

// GetTrafficSummary 获取流量汇总统计
func (r *TrafficRepository) GetTrafficSummary() (*model.TrafficSummary, error) {
	var summary model.TrafficSummary

	err := database.DB.Model(&model.Proxy{}).
		Select("COALESCE(SUM(total_bytes_in), 0) as total_bytes_in, COALESCE(SUM(total_bytes_out), 0) as total_bytes_out, COALESCE(SUM(current_bytes_in_rate), 0) as current_rate_in, COALESCE(SUM(current_bytes_out_rate), 0) as current_rate_out, COUNT(*) as total_proxies, COUNT(CASE WHEN last_online_time IS NOT NULL AND last_online_time > ? THEN 1 END) as active_proxies", time.Now().Add(-5*time.Minute)).
		Scan(&summary).Error

	return &summary, err
}
