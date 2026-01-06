/*
 * @Author              : 寂情啊
 * @Date                : 2025-12-04 17:27:28
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-04 17:27:41
 * @FilePath            : frp-web-testbackendinternalrepositoryserver_metrics_repo.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package repository

import (
	"frp-web-panel/internal/model"
	"frp-web-panel/pkg/database"
	"time"
)

type ServerMetricsRepository struct{}

func NewServerMetricsRepository() *ServerMetricsRepository {
	return &ServerMetricsRepository{}
}

// Create 创建指标记录
func (r *ServerMetricsRepository) Create(metrics *model.ServerMetricsHistory) error {
	return database.DB.Create(metrics).Error
}

// GetHistory 获取指定服务器的历史指标
func (r *ServerMetricsRepository) GetHistory(serverID uint, start, end time.Time) ([]model.ServerMetricsHistory, error) {
	var records []model.ServerMetricsHistory
	err := database.DB.Where("server_id = ? AND record_time BETWEEN ? AND ?", serverID, start, end).
		Order("record_time ASC").Find(&records).Error
	return records, err
}

// DeleteOlderThan 删除指定时间之前的记录
func (r *ServerMetricsRepository) DeleteOlderThan(before time.Time) (int64, error) {
	result := database.DB.Where("record_time < ?", before).Delete(&model.ServerMetricsHistory{})
	return result.RowsAffected, result.Error
}
