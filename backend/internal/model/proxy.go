/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-14 15:24:35
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2026-01-07 14:30:24
 * @FilePath            : frp-web-testbackendinternalmodelproxy.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package model

import (
	"time"
)

type Proxy struct {
	ID                  uint       `json:"id" gorm:"primaryKey"`
	ClientID            uint       `json:"client_id" gorm:"not null;index"`
	Name                string     `json:"name" gorm:"size:100;not null;index"`
	Type                string     `json:"type" gorm:"size:20;not null"`
	Enabled             bool       `json:"enabled" gorm:"default:true;index"`
	LocalIP             string     `json:"local_ip" gorm:"size:50;default:127.0.0.1"`
	LocalPort           int        `json:"local_port" gorm:"not null"`
	RemotePort          int        `json:"remote_port"`
	CustomDomains       string     `json:"custom_domains" gorm:"type:text"`
	Subdomain           string     `json:"subdomain" gorm:"size:100"`
	Locations           string     `json:"locations" gorm:"type:text"`
	HostHeaderRewrite   string     `json:"host_header_rewrite" gorm:"size:200"`
	HttpUser            string     `json:"http_user" gorm:"size:100"`
	HttpPassword        string     `json:"http_password" gorm:"size:100"`
	SecretKey           string     `json:"secret_key" gorm:"size:100"`
	AllowUsers          string     `json:"allow_users" gorm:"type:text"`
	UseEncryption       bool       `json:"use_encryption" gorm:"default:false"`
	UseCompression      bool       `json:"use_compression" gorm:"default:false"`
	HealthCheckType     string     `json:"health_check_type" gorm:"size:20"`
	HealthCheckTimeout  int        `json:"health_check_timeout"`
	HealthCheckInterval int        `json:"health_check_interval"`
	BandwidthLimit      string     `json:"bandwidth_limit" gorm:"size:20"`
	BandwidthLimitMode  string     `json:"bandwidth_limit_mode" gorm:"size:10;default:client"`
	TotalBytesIn        int64      `json:"total_bytes_in" gorm:"default:0"`
	TotalBytesOut       int64      `json:"total_bytes_out" gorm:"default:0"`
	LastOnlineTime      *time.Time `json:"last_online_time"`
	CurrentBytesInRate  int64      `json:"current_bytes_in_rate" gorm:"default:0"`
	CurrentBytesOutRate int64      `json:"current_bytes_out_rate" gorm:"default:0"`
	LastTrafficUpdate   *time.Time `json:"last_traffic_update"`
	FrpStatus           string     `json:"frp_status" gorm:"size:20;default:unknown;index"`
	FrpCurConns         int        `json:"frp_cur_conns" gorm:"default:0"`
	FrpLastStartTime    *time.Time `json:"frp_last_start_time"`
	FrpLastCloseTime    *time.Time `json:"frp_last_close_time"`
	// 插件配置字段
	PluginType   string `json:"plugin_type" gorm:"size:50"`     // 插件类型: http_proxy, socks5, static_file, unix_domain_socket
	PluginConfig string `json:"plugin_config" gorm:"type:text"` // 插件配置 JSON
	// DNS同步字段
	EnableDNSSync bool   `json:"enable_dns_sync" gorm:"default:false"` // 是否启用DNS同步
	DNSProviderID *uint  `json:"dns_provider_id" gorm:"index"`         // DNS提供商ID
	DNSRootDomain string `json:"dns_root_domain" gorm:"size:100"`      // 根域名
	// 自动证书字段
	AutoCert  bool      `json:"auto_cert" gorm:"default:false"` // 是否自动申请证书
	CertID    *uint     `json:"cert_id" gorm:"index"`           // 关联的证书ID
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// 插件类型常量
const (
	PluginTypeHTTPProxy        = "http_proxy"
	PluginTypeSocks5           = "socks5"
	PluginTypeStaticFile       = "static_file"
	PluginTypeUnixDomainSocket = "unix_domain_socket"
	PluginTypeHTTPS2HTTP       = "https2http"
	PluginTypeHTTPS2HTTPS      = "https2https"
)

// HTTPProxyPluginConfig HTTP代理插件配置
type HTTPProxyPluginConfig struct {
	HttpUser     string `json:"httpUser"`
	HttpPassword string `json:"httpPassword"`
}

// Socks5PluginConfig SOCKS5代理插件配置
type Socks5PluginConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// StaticFilePluginConfig 静态文件服务插件配置
type StaticFilePluginConfig struct {
	LocalPath    string `json:"localPath"`
	StripPrefix  string `json:"stripPrefix"`
	HttpUser     string `json:"httpUser"`
	HttpPassword string `json:"httpPassword"`
}

// UnixDomainSocketPluginConfig Unix域套接字插件配置
type UnixDomainSocketPluginConfig struct {
	UnixPath string `json:"unixPath"`
}

// HTTPS2HTTPPluginConfig HTTPS转HTTP插件配置
type HTTPS2HTTPPluginConfig struct {
	LocalAddr         string `json:"localAddr"`         // 本地HTTP服务地址，如 127.0.0.1:8080
	CrtPath           string `json:"crtPath"`           // TLS证书文件路径
	KeyPath           string `json:"keyPath"`           // TLS私钥文件路径
	HostHeaderRewrite string `json:"hostHeaderRewrite"` // Host Header重写（可选）
}

// HTTPS2HTTPSPluginConfig HTTPS转HTTPS插件配置
type HTTPS2HTTPSPluginConfig struct {
	LocalAddr         string `json:"localAddr"`         // 本地HTTPS服务地址
	CrtPath           string `json:"crtPath"`           // TLS证书文件路径
	KeyPath           string `json:"keyPath"`           // TLS私钥文件路径
	HostHeaderRewrite string `json:"hostHeaderRewrite"` // Host Header重写（可选）
}
