/*
 * @Author              : 寂情啊
 * @Date                : 2025-12-29 15:04:29
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2026-01-07 10:54:49
 * @FilePath            : frp-web-testbackendcmdserverscheduler.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package main

import (
	"context"
	"frp-web-panel/internal/container"
	"frp-web-panel/internal/logger"
	"time"
)

// RegisterScheduledTasks 注册所有定时任务和回调
func RegisterScheduledTasks(c *container.Container) {
	// 注册 WebSocket Hub 为一次性任务
	c.Services.TaskManager.RegisterOneShotTask("websocket-hub", func(ctx context.Context) {
		c.Hub.RunWithContext(ctx)
	})

	// 注册告警检测定时任务
	c.Services.TaskManager.RegisterPeriodicTask("alert-check", 5*time.Minute, c.Services.Alert.CheckAlerts)
	c.Services.TaskManager.RegisterPeriodicTask("offline-alert-check", 1*time.Minute, c.Services.Alert.CheckOfflineAlerts)
}

// RegisterCallbacks 注册所有回调函数
func RegisterCallbacks(c *container.Container) {
	// 设置 WebSocket 连接状态回调
	c.ClientDaemonHub.SetStatusCallback(func(clientID uint, online bool) {
		logger.Debugf("[WS状态回调] 客户端 %d WS连接状态变更: %v", clientID, online)
		if err := c.Services.Client.UpdateWSStatus(clientID, online); err != nil {
			logger.Errorf("[WS状态回调] 更新客户端 %d WS状态失败: %v", clientID, err)
		}
		if !online {
			if err := c.Services.Client.UpdateOnlineStatusDirectly(clientID, "offline"); err != nil {
				logger.Errorf("[WS状态回调] 更新客户端 %d 在线状态失败: %v", clientID, err)
			}
		}
	})

	// 设置配置同步结果回调
	c.ClientDaemonHub.SetConfigSyncResultCallback(func(clientID uint, success bool, errorMsg string, rolledBack bool) {
		logger.Debugf("[配置同步回调] 客户端 %d 同步结果: success=%v, error=%s, rolledBack=%v", clientID, success, errorMsg, rolledBack)
		if err := c.Services.Client.UpdateConfigSyncStatus(clientID, success, errorMsg, rolledBack); err != nil {
			logger.Errorf("[配置同步回调] 更新客户端 %d 配置同步状态失败: %v", clientID, err)
		}
	})

	// 设置frpc控制结果回调，广播给前端
	c.ClientDaemonHub.SetFrpcControlResultCallback(func(clientID uint, action string, success bool, message string) {
		logger.Debugf("[frpc控制回调] 客户端 %d 控制结果: action=%s, success=%v, message=%s", clientID, action, success, message)
		c.Hub.BroadcastFrpcControlResult(clientID, action, success, message)
	})
}

// StartServices 启动所有后台服务
func StartServices(c *container.Container) {
	// 服务启动时重置所有客户端状态为离线
	if err := c.Services.Client.ResetAllClientStatus(); err != nil {
		logger.Warnf("重置客户端状态失败: %v", err)
	}

	// 启动任务管理器管理的任务
	c.Services.TaskManager.Start()

	// 启动有自己 Stop 方法的服务
	c.Services.Realtime.Start()
	c.Services.FrpSync.Start()
	c.Services.MetricsCollector.Start()
	go c.Services.ClientStatusChecker.StartWithContext(c.Services.TaskManager.Context())
	c.Services.CertRenewal.Start()
}

// StopServices 停止所有后台服务
func StopServices(c *container.Container, shutdownTimeout time.Duration) {
	c.Services.CertRenewal.Stop()
	c.Services.MetricsCollector.Stop()
	c.Services.Realtime.Stop()
	c.Services.FrpSync.Stop()
	c.Services.ClientStatusChecker.Stop()

	if err := c.Services.TaskManager.Shutdown(shutdownTimeout); err != nil {
		logger.Errorf("任务管理器关闭错误: %v", err)
	}
}
