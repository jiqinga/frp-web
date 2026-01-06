/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-24 16:17:31
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-01 16:14:14
 * @FilePath            : frp-web-testbackendinternalutilfrpc_config_parser.go
 * @Description         : frpc配置文件解析器，支持TOML格式
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package util

import (
	"fmt"

	"github.com/pelletier/go-toml/v2"
)

// FrpcTomlConfig 表示frpc TOML配置文件结构
type FrpcTomlConfig struct {
	ServerAddr string         `toml:"serverAddr"`
	ServerPort int            `toml:"serverPort"`
	User       string         `toml:"user"`
	Auth       FrpcAuthConfig `toml:"auth"`
	WebServer  FrpcWebServer  `toml:"webServer"`
}

// FrpcAuthConfig 认证配置
type FrpcAuthConfig struct {
	Token string `toml:"token"`
}

// FrpcWebServer Web服务器配置（Admin API）
type FrpcWebServer struct {
	Addr     string `toml:"addr"`
	Port     int    `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
}

// ParsedFrpcConfig 解析后的配置结果
type ParsedFrpcConfig struct {
	ServerAddr    string `json:"server_addr"`
	ServerPort    int    `json:"server_port"`
	Token         string `json:"token"`
	FrpcAdminHost string `json:"frpc_admin_host"`
	FrpcAdminPort int    `json:"frpc_admin_port"`
	FrpcAdminUser string `json:"frpc_admin_user"`
	FrpcAdminPwd  string `json:"frpc_admin_pwd"`
}

// ParseFrpcConfig 解析frpc TOML配置
func ParseFrpcConfig(tomlContent string) (*ParsedFrpcConfig, error) {
	var cfg FrpcTomlConfig
	err := toml.Unmarshal([]byte(tomlContent), &cfg)
	if err != nil {
		return nil, fmt.Errorf("解析TOML失败: %w", err)
	}

	// 设置默认值
	serverPort := cfg.ServerPort
	if serverPort == 0 {
		serverPort = 7000
	}

	adminPort := cfg.WebServer.Port
	if adminPort == 0 {
		adminPort = 7400
	}

	result := &ParsedFrpcConfig{
		ServerAddr:    cfg.ServerAddr,
		ServerPort:    serverPort,
		Token:         cfg.Auth.Token,
		FrpcAdminHost: cfg.WebServer.Addr,
		FrpcAdminPort: adminPort,
		FrpcAdminUser: cfg.WebServer.User,
		FrpcAdminPwd:  cfg.WebServer.Password,
	}

	return result, nil
}
