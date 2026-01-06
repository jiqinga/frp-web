/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-14 15:25:08
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-03 16:36:04
 * @FilePath            : frp-web-testbackendinternalmodellog.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package model

import (
	"time"
)

type OperationLog struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	UserID        uint      `json:"user_id" gorm:"index"`
	Username      string    `json:"username" gorm:"-"` // 非数据库字段，用于返回用户名
	OperationType string    `json:"operation_type" gorm:"size:50;not null"`
	ResourceType  string    `json:"resource_type" gorm:"size:50;not null"`
	ResourceID    uint      `json:"resource_id"`
	Description   string    `json:"description" gorm:"type:text"`
	IPAddress     string    `json:"ip_address" gorm:"size:50"`
	IPLocation    string    `json:"ip_location" gorm:"size:100"` // IP归属地
	CreatedAt     time.Time `json:"created_at"`
}
