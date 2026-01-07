/*
 * @Author              : 寂情啊
 * @Date                : 2025-12-29 15:03:57
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2026-01-07 11:21:53
 * @FilePath            : frp-web-testbackendcmdserverbootstrap.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package main

import (
	"fmt"
	"frp-web-panel/internal/config"
	"frp-web-panel/internal/container"
	"frp-web-panel/internal/logger"
	"frp-web-panel/internal/util"
	"frp-web-panel/pkg/database"
)

// BootstrapResult 包含启动初始化后的资源
type BootstrapResult struct {
	Container *container.Container
	Config    *config.Config
}

// Bootstrap 执行应用程序启动初始化
func Bootstrap(configPath string) (*BootstrapResult, error) {
	// 加载配置
	if err := config.LoadConfig(configPath); err != nil {
		return nil, fmt.Errorf("加载配置失败: %w", err)
	}

	// 初始化日志（必须在配置验证之前，因为验证可能输出警告日志）
	if err := logger.Init(config.GlobalConfig.Log.Level, config.GlobalConfig.Log.Format); err != nil {
		return nil, fmt.Errorf("初始化日志失败: %w", err)
	}

	// 验证配置
	if err := config.GlobalConfig.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	// 初始化数据库
	if err := database.InitDB(config.GlobalConfig); err != nil {
		return nil, fmt.Errorf("初始化数据库失败: %w", err)
	}

	// 初始化加密
	if err := util.InitEncryption(config.GlobalConfig.Security.EncryptionKey); err != nil {
		return nil, fmt.Errorf("初始化加密失败: %w", err)
	}

	// 初始化 IP 归属地查询服务
	if err := util.InitIPSearcher("./data"); err != nil {
		logger.Warnf("初始化IP查询服务失败: %v", err)
	}

	// 创建服务容器
	c := container.NewContainer(database.DB, config.GlobalConfig)

	return &BootstrapResult{
		Container: c,
		Config:    config.GlobalConfig,
	}, nil
}

// Cleanup 清理启动时初始化的资源
func Cleanup() {
	logger.Sync()
	util.CloseIPSearcher()
}
