/*
 * @Author              : 寂情啊
 * @Date                : 2025-12-22 15:44:35
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-22 15:44:46
 * @FilePath            : frp-web-testbackendinternalmodeldns_record.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package model

import "time"

// DNSRecordStatus DNS记录状态
type DNSRecordStatus string

const (
	DNSRecordStatusPending DNSRecordStatus = "pending"
	DNSRecordStatusSynced  DNSRecordStatus = "synced"
	DNSRecordStatusFailed  DNSRecordStatus = "failed"
)

// DNSRecord DNS记录
type DNSRecord struct {
	ID          uint            `json:"id" gorm:"primaryKey"`
	ProxyID     uint            `json:"proxy_id" gorm:"index"`
	ProviderID  uint            `json:"provider_id"`
	Domain      string          `json:"domain" gorm:"size:255"`
	RootDomain  string          `json:"root_domain" gorm:"size:100"`
	RecordType  string          `json:"record_type" gorm:"size:10;default:A"`
	RecordValue string          `json:"record_value" gorm:"size:50"`
	RecordID    string          `json:"record_id" gorm:"size:50"`
	Status      DNSRecordStatus `json:"status" gorm:"size:20;default:pending"`
	LastError   string          `json:"last_error" gorm:"type:text"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}
