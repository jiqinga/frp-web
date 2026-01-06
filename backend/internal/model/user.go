/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-14 15:23:20
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-11-14 15:23:30
 * @FilePath            : frp-web-testbackendinternalmodeluser.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package model

import (
	"time"
)

type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"uniqueIndex;size:50;not null"`
	Password  string    `json:"-" gorm:"size:255;not null"`
	Nickname  string    `json:"nickname" gorm:"size:100"`
	Role      string    `json:"role" gorm:"size:20;default:admin"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
