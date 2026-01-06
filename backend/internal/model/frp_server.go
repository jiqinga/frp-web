/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-19 17:06:28
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-11-21 14:04:04
 * @FilePath            : frp-web-testbackendinternalmodelfrp_server.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package model

import "time"

type FrpServerStatus string

const (
	StatusStopped  FrpServerStatus = "stopped"
	StatusStarting FrpServerStatus = "starting"
	StatusRunning  FrpServerStatus = "running"
	StatusStopping FrpServerStatus = "stopping"
	StatusError    FrpServerStatus = "error"
)

type ServerType string

const (
	ServerTypeLocal  ServerType = "local"
	ServerTypeRemote ServerType = "remote"
)

type FrpServer struct {
	ID            uint            `json:"id" gorm:"primaryKey"`
	Name          string          `json:"name" gorm:"uniqueIndex;size:100;not null"`
	ServerType    ServerType      `json:"server_type" gorm:"size:20;default:'local'"`
	Host          string          `json:"host" gorm:"size:255;not null"`
	DashboardPort int             `json:"dashboard_port" gorm:"not null;default:7500"`
	DashboardUser string          `json:"dashboard_user" gorm:"size:100"`
	DashboardPwd  string          `json:"dashboard_pwd" gorm:"size:255"`
	BindPort      int             `json:"bind_port" gorm:"default:7000"`
	Token         string          `json:"token" gorm:"size:64"`
	SSHHost       string          `json:"ssh_host" gorm:"size:255"`
	SSHPort       int             `json:"ssh_port" gorm:"default:22"`
	SSHUser       string          `json:"ssh_user" gorm:"size:100"`
	SSHPassword   string          `json:"ssh_password" gorm:"size:500"`
	InstallPath   string          `json:"install_path" gorm:"size:500;default:'/opt/frps'"`
	MirrorID      *uint           `json:"mirror_id"`
	Enabled       bool            `json:"enabled" gorm:"default:true"`
	Status        FrpServerStatus `json:"status" gorm:"size:20;default:'stopped'"`
	PID           int             `json:"pid" gorm:"default:0"`
	Version       string          `json:"version" gorm:"size:50"`
	BinaryPath    string          `json:"binary_path" gorm:"size:500"`
	ConfigPath    string          `json:"config_path" gorm:"size:500"`
	LastSyncTime  *time.Time      `json:"last_sync_time"`
	LastError     string          `json:"last_error" gorm:"type:text"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}
