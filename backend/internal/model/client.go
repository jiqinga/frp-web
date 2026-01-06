/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-14 15:24:09
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-26 15:18:15
 * @FilePath            : frp-web-testbackendinternalmodelclient.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package model

import (
	"time"
)

type Client struct {
	ID             uint       `json:"id" gorm:"primaryKey"`
	Name           string     `json:"name" gorm:"uniqueIndex;size:100;not null"`
	Remark         string     `json:"remark" gorm:"type:text"`
	ServerAddr     string     `json:"server_addr" gorm:"size:255;not null"`
	ServerPort     int        `json:"server_port" gorm:"not null"`
	Token          string     `json:"token" gorm:"size:255"`
	Protocol       string     `json:"protocol" gorm:"size:10;default:tcp"`
	FrpcAdminHost  string     `json:"frpc_admin_host" gorm:"size:255"`
	FrpcAdminPort  int        `json:"frpc_admin_port"`
	FrpcAdminUser  string     `json:"frpc_admin_user" gorm:"size:100"`
	FrpcAdminPwd   string     `json:"frpc_admin_pwd" gorm:"size:255"`
	FrpServerID    *uint      `json:"frp_server_id"`
	OnlineStatus   string     `json:"online_status" gorm:"size:20;default:unknown"`
	LastHeartbeat  *time.Time `json:"last_heartbeat"`
	ConfigVersion  int        `json:"config_version" gorm:"default:1"`
	WsConnected    bool       `json:"ws_connected" gorm:"default:false"`
	LastConfigSync *time.Time `json:"last_config_sync"`
	// 配置同步状态字段
	ConfigSyncStatus string     `json:"config_sync_status" gorm:"size:20;default:pending"` // synced/failed/pending/rolled_back
	ConfigSyncError  string     `json:"config_sync_error" gorm:"type:text"`
	ConfigSyncTime   *time.Time `json:"config_sync_time"`
	// 版本信息字段
	FrpcVersion   string    `json:"frpc_version" gorm:"size:50"`
	DaemonVersion string    `json:"daemon_version" gorm:"size:50"`
	OS            string    `json:"os" gorm:"size:20"`
	Arch          string    `json:"arch" gorm:"size:20"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Proxies       []Proxy   `json:"proxies,omitempty" gorm:"foreignKey:ClientID"`
}
