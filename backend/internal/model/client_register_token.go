/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-21 16:04:00
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-11-24 16:39:56
 * @FilePath            : frp-web-testbackendinternalmodelclient_register_token.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package model

import (
	"time"
)

// ClientRegisterToken 客户端注册Token
type ClientRegisterToken struct {
	ID            uint       `json:"id" gorm:"primaryKey"`
	Token         string     `json:"token" gorm:"uniqueIndex;size:64;not null"`
	ClientName    string     `json:"client_name" gorm:"size:100;not null"`
	FrpServerID   uint       `json:"frp_server_id" gorm:"default:1"`
	ServerAddr    string     `json:"server_addr" gorm:"size:255;not null"`
	ServerPort    int        `json:"server_port" gorm:"not null"`
	TokenStr      string     `json:"token_str" gorm:"size:255"`
	AdminPassword string     `json:"admin_password" gorm:"size:64"`
	Protocol      string     `json:"protocol" gorm:"size:10;default:tcp"`
	Remark        string     `json:"remark" gorm:"type:text"`
	ExpiresAt     time.Time  `json:"expires_at" gorm:"not null"`
	Used          bool       `json:"used" gorm:"default:false"`
	UsedAt        *time.Time `json:"used_at"`
	CreatedBy     uint       `json:"created_by"`
	CreatedAt     time.Time  `json:"created_at"`
}
