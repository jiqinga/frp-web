/*
 * @Author              : 寂情啊
 * @Date                : 2025-12-22 15:44:11
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-22 16:24:54
 * @FilePath            : frp-web-testbackendinternalmodeldns_provider.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package model

import "time"

// DNSProviderType DNS提供商类型
type DNSProviderType string

const (
	DNSProviderTypeAliyun     DNSProviderType = "aliyun"
	DNSProviderTypeCloudflare DNSProviderType = "cloudflare"
	DNSProviderTypeTencent    DNSProviderType = "tencent"
)

// 保留旧常量以兼容
const DNSProviderAliyun = DNSProviderTypeAliyun

// DNSProvider DNS提供商配置
type DNSProvider struct {
	ID        uint            `json:"id" gorm:"primaryKey"`
	Name      string          `json:"name" gorm:"size:100;not null"`
	Type      DNSProviderType `json:"type" gorm:"size:20;not null"`
	AccessKey string          `json:"access_key" gorm:"size:100"`
	SecretKey string          `json:"-" gorm:"size:200"` // 不返回给前端
	Enabled   bool            `json:"enabled" gorm:"default:true"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}
