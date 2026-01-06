/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-14 16:01:12
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-09 16:05:13
 * @FilePath            : frp-web-testbackendinternalservicetraffic_service.go
 * @Description         : 流量统计服务
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package service

import (
	"frp-web-panel/internal/model"
	"frp-web-panel/internal/repository"
)

type TrafficService struct {
	trafficRepo *repository.TrafficRepository
}

func NewTrafficService() *TrafficService {
	return &TrafficService{
		trafficRepo: repository.NewTrafficRepository(),
	}
}

// GetTrafficSummary 获取流量汇总统计
func (s *TrafficService) GetTrafficSummary() (*model.TrafficSummary, error) {
	return s.trafficRepo.GetTrafficSummary()
}
