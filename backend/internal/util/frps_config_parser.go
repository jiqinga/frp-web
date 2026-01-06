/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-24 16:05:16
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-11-24 16:05:28
 * @FilePath            : frp-web-testbackendinternalutilfrps_config_parser.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package util

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// FrpsConfig 表示frps配置文件结构
type FrpsConfig struct {
	BindPort  int             `yaml:"bindPort"`
	Auth      AuthConfig      `yaml:"auth"`
	WebServer WebServerConfig `yaml:"webServer"`
	Log       LogConfig       `yaml:"log"`
}

type AuthConfig struct {
	Method string `yaml:"method"`
	Token  string `yaml:"token"`
}

type WebServerConfig struct {
	Addr     string `yaml:"addr"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type LogConfig struct {
	To    string `yaml:"to"`
	Level string `yaml:"level"`
}

// ParsedFrpsConfig 解析后的配置结果
type ParsedFrpsConfig struct {
	BindPort      int    `json:"bind_port"`
	Token         string `json:"token"`
	Host          string `json:"host"`
	DashboardPort int    `json:"dashboard_port"`
	DashboardUser string `json:"dashboard_user"`
	DashboardPwd  string `json:"dashboard_pwd"`
}

// ParseFrpsConfig 解析frps YAML配置
func ParseFrpsConfig(yamlContent string) (*ParsedFrpsConfig, error) {
	var config FrpsConfig
	if err := yaml.Unmarshal([]byte(yamlContent), &config); err != nil {
		return nil, fmt.Errorf("解析YAML失败: %w", err)
	}

	result := &ParsedFrpsConfig{
		BindPort:      config.BindPort,
		Token:         config.Auth.Token,
		Host:          config.WebServer.Addr,
		DashboardPort: config.WebServer.Port,
		DashboardUser: config.WebServer.User,
		DashboardPwd:  config.WebServer.Password,
	}

	return result, nil
}
