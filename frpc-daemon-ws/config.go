/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-25 16:58:39
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-29 15:58:44
 * @FilePath            : frp-web-testfrpc-daemon-wsconfig.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ClientID     int    `yaml:"client_id"`
	Token        string `yaml:"token"`
	ServerURL    string `yaml:"server_url"`
	FrpcPath     string `yaml:"frpc_path"`
	FrpcConfig   string `yaml:"frpc_config"`
	LogFile      string `yaml:"log_file"`
	HeartbeatSec int    `yaml:"heartbeat_sec"`
	// 安装目录，用于存放日志等文件
	InstallDir string `yaml:"install_dir"`

	// frpc Admin API 配置
	FrpcAdminAddr     string `yaml:"frpc_admin_addr"`
	FrpcAdminPort     int    `yaml:"frpc_admin_port"`
	FrpcAdminUser     string `yaml:"frpc_admin_user"`
	FrpcAdminPassword string `yaml:"frpc_admin_password"`

	// systemctl 服务名称（Linux 系统使用 systemctl 管理 frpc 服务）
	// 如果配置了此项，降级方案将使用 systemctl restart 而不是直接操作进程
	FrpcServiceName string `yaml:"frpc_service_name"`

	// daemon 自身的 systemctl 服务名称（Linux 系统使用 systemctl 管理 daemon 服务）
	// 如果配置了此项，daemon 自更新时将使用 systemctl restart 而不是直接启动进程
	DaemonServiceName string `yaml:"daemon_service_name"`
}

// ValidateConfig 验证配置必填字段
func (c *Config) Validate() error {
	if c.ClientID == 0 {
		return fmt.Errorf("client_id 是必填字段")
	}
	if c.Token == "" {
		return fmt.Errorf("token 是必填字段")
	}
	if c.ServerURL == "" {
		return fmt.Errorf("server_url 是必填字段")
	}
	if c.FrpcPath == "" {
		return fmt.Errorf("frpc_path 是必填字段")
	}
	if c.FrpcConfig == "" {
		return fmt.Errorf("frpc_config 是必填字段")
	}
	return nil
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	// 验证必填字段
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}
	if cfg.HeartbeatSec == 0 {
		cfg.HeartbeatSec = 30
	}
	// 如果没有指定日志文件，且指定了安装目录，则使用安装目录下的日志文件
	if cfg.LogFile == "" && cfg.InstallDir != "" {
		cfg.LogFile = filepath.Join(cfg.InstallDir, "frpc-daemon.log")
	}
	// Admin API 默认配置
	if cfg.FrpcAdminAddr == "" {
		cfg.FrpcAdminAddr = "127.0.0.1"
	}
	if cfg.FrpcAdminPort == 0 {
		cfg.FrpcAdminPort = 7400
	}
	if cfg.FrpcAdminUser == "" {
		cfg.FrpcAdminUser = "admin"
	}
	if cfg.FrpcAdminPassword == "" {
		cfg.FrpcAdminPassword = "admin"
	}
	return &cfg, nil
}
