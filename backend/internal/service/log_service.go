/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-14 16:11:54
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-03 16:07:24
 * @FilePath            : frp-web-testbackendinternalservicelog_service.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */

package service

import (
	"frp-web-panel/internal/model"
	"frp-web-panel/internal/repository"
	"frp-web-panel/internal/util"
)

type LogService struct {
	logRepo *repository.LogRepository
}

func NewLogService() *LogService {
	return &LogService{
		logRepo: repository.NewLogRepository(),
	}
}

func (s *LogService) GetLogs(page, pageSize int, operationType, resourceType string) ([]model.OperationLog, int64, error) {
	return s.logRepo.GetLogs(page, pageSize, operationType, resourceType)
}

// CreateLog 创建操作日志，自动查询IP归属地
func (s *LogService) CreateLog(userID uint, operationType, resourceType string, resourceID uint, description, ipAddress string) error {
	// 异步查询IP归属地，避免阻塞主流程
	ipLocation := util.GetIPLocation(ipAddress)
	return s.logRepo.CreateLog(userID, operationType, resourceType, resourceID, description, ipAddress, ipLocation)
}

// CreateLogAsync 异步创建操作日志（不阻塞主流程）
func (s *LogService) CreateLogAsync(userID uint, operationType, resourceType string, resourceID uint, description, ipAddress string) {
	go func() {
		ipLocation := util.GetIPLocation(ipAddress)
		s.logRepo.CreateLog(userID, operationType, resourceType, resourceID, description, ipAddress, ipLocation)
	}()
}
