/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-24 15:27:02
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2026-01-07 11:09:38
 * @FilePath            : frp-web-testbackendinternalserviceclient_status_checker.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package service

import (
	"context"
	"frp-web-panel/internal/logger"
	"time"
)

type ClientStatusChecker struct {
	clientService *ClientService
	interval      time.Duration
	stopChan      chan struct{}
}

func NewClientStatusChecker(clientService *ClientService) *ClientStatusChecker {
	return &ClientStatusChecker{
		clientService: clientService,
		interval:      30 * time.Second,
		stopChan:      make(chan struct{}),
	}
}

func (c *ClientStatusChecker) Start() {
	c.StartWithContext(context.Background())
}

func (c *ClientStatusChecker) StartWithContext(ctx context.Context) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	logger.Info("客户端状态检测 服务已启动，检测间隔: 30秒（仅同步WS连接状态）")

	for {
		select {
		case <-ticker.C:
			if err := c.clientService.SyncAllClientsWSStatus(); err != nil {
				logger.Errorf("客户端状态检测 WS状态同步失败: %v", err)
			}
		case <-c.stopChan:
			logger.Info("客户端状态检测 服务已停止")
			return
		case <-ctx.Done():
			logger.Info("客户端状态检测 服务已停止 (context cancelled)")
			return
		}
	}
}

func (c *ClientStatusChecker) Stop() {
	close(c.stopChan)
}
