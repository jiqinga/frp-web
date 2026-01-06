/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-14 16:12:36
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-03 16:36:32
 * @FilePath            : frp-web-testbackendinternalrepositorylog_repo.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package repository

import (
	"frp-web-panel/internal/model"
	"frp-web-panel/pkg/database"
)

type LogRepository struct{}

func NewLogRepository() *LogRepository {
	return &LogRepository{}
}

func (r *LogRepository) GetLogs(page, pageSize int, operationType, resourceType string) ([]model.OperationLog, int64, error) {
	db := database.DB
	var logs []model.OperationLog
	var total int64

	query := db.Model(&model.OperationLog{})

	if operationType != "" {
		query = query.Where("operation_type = ?", operationType)
	}
	if resourceType != "" {
		query = query.Where("resource_type = ?", resourceType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	// 收集所有非零的 user_id
	userIDs := make([]uint, 0)
	for _, log := range logs {
		if log.UserID > 0 {
			userIDs = append(userIDs, log.UserID)
		}
	}

	// 批量查询用户名
	userMap := make(map[uint]string)
	if len(userIDs) > 0 {
		var users []model.User
		if err := db.Where("id IN ?", userIDs).Find(&users).Error; err == nil {
			for _, user := range users {
				userMap[user.ID] = user.Username
			}
		}
	}

	// 填充 Username 字段
	for i := range logs {
		if logs[i].UserID == 0 {
			logs[i].Username = "系统"
		} else if username, ok := userMap[logs[i].UserID]; ok {
			logs[i].Username = username
		} else {
			logs[i].Username = "已删除用户"
		}
	}

	return logs, total, nil
}

func (r *LogRepository) CreateLog(userID uint, operationType, resourceType string, resourceID uint, description, ipAddress, ipLocation string) error {
	db := database.DB
	log := &model.OperationLog{
		UserID:        userID,
		OperationType: operationType,
		ResourceType:  resourceType,
		ResourceID:    resourceID,
		Description:   description,
		IPAddress:     ipAddress,
		IPLocation:    ipLocation,
	}
	return db.Create(log).Error
}
