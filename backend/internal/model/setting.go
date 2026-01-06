/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-14 15:25:33
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-11-14 15:25:43
 * @FilePath            : frp-web-testbackendinternalmodelsetting.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package model

import (
	"time"
)

type Setting struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Key         string    `json:"key" gorm:"uniqueIndex;size:100;not null"`
	Value       string    `json:"value" gorm:"type:text"`
	Description string    `json:"description" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
