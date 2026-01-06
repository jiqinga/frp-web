/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-21 14:00:18
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-11-21 14:00:29
 * @FilePath            : frp-web-testbackendinternalmodelgithub_mirror.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package model

import (
	"time"
)

type GithubMirror struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"size:100;not null"`
	BaseURL     string    `json:"base_url" gorm:"size:255;not null"`
	IsDefault   bool      `json:"is_default" gorm:"default:false"`
	Enabled     bool      `json:"enabled" gorm:"default:true"`
	Description string    `json:"description" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
