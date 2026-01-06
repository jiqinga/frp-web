/*
 * @Author              : 寂情啊
 * @Date                : 2025-12-12 14:19:35
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-12 14:19:47
 * @FilePath            : frp-web-testbackendinternalmodelalert_recipient.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package model

import "time"

// AlertRecipient 告警接收人
type AlertRecipient struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"type:varchar(100);not null"`
	Email     string    `json:"email" gorm:"type:varchar(255);not null"`
	Enabled   bool      `json:"enabled" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AlertRecipientGroup 告警接收人分组
type AlertRecipientGroup struct {
	ID          uint             `json:"id" gorm:"primaryKey"`
	Name        string           `json:"name" gorm:"type:varchar(100);not null;uniqueIndex"`
	Description string           `json:"description" gorm:"type:varchar(500)"`
	Enabled     bool             `json:"enabled" gorm:"default:true"`
	Recipients  []AlertRecipient `json:"recipients" gorm:"-"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// AlertGroupRecipient 分组-接收人关联表
type AlertGroupRecipient struct {
	ID          uint `json:"id" gorm:"primaryKey"`
	GroupID     uint `json:"group_id" gorm:"not null;index:idx_group_recipient,unique"`
	RecipientID uint `json:"recipient_id" gorm:"not null;index:idx_group_recipient,unique"`
}

func (AlertRecipient) TableName() string {
	return "alert_recipients"
}

func (AlertRecipientGroup) TableName() string {
	return "alert_recipient_groups"
}

func (AlertGroupRecipient) TableName() string {
	return "alert_group_recipients"
}
