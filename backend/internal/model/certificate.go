/*
 * @Author              : 寂情啊
 * @Date                : 2025-12-23 13:54:36
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-23 13:54:50
 * @FilePath            : frp-web-testbackendinternalmodelcertificate.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package model

import (
	"time"
)

// 证书状态常量
const (
	CertStatusPending  = "pending"  // 申请中
	CertStatusActive   = "active"   // 有效
	CertStatusExpiring = "expiring" // 即将过期
	CertStatusExpired  = "expired"  // 已过期
	CertStatusFailed   = "failed"   // 申请失败
)

// Certificate SSL证书模型
type Certificate struct {
	ID            uint       `json:"id" gorm:"primaryKey"`
	ProxyID       uint       `json:"proxy_id" gorm:"index"`                 // 关联的代理ID
	Domain        string     `json:"domain" gorm:"size:255;not null"`       // 域名
	ProviderID    uint       `json:"provider_id" gorm:"index"`              // DNS提供商ID
	Status        string     `json:"status" gorm:"size:20;default:pending"` // 状态
	CertPEM       string     `json:"cert_pem" gorm:"type:text"`             // 证书内容(PEM格式)
	KeyPEM        string     `json:"-" gorm:"type:text"`                    // 私钥内容(PEM格式，不返回给前端)
	IssuerCertPEM string     `json:"issuer_cert_pem" gorm:"type:text"`      // 颁发者证书
	NotBefore     *time.Time `json:"not_before"`                            // 生效时间
	NotAfter      *time.Time `json:"not_after"`                             // 过期时间
	LastError     string     `json:"last_error" gorm:"type:text"`           // 最后错误信息
	AutoRenew     bool       `json:"auto_renew" gorm:"default:true"`        // 是否自动续期
	AcmeAccountID string     `json:"acme_account_id" gorm:"size:255"`       // ACME账户ID
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// IsExpiringSoon 检查证书是否即将过期（30天内）
func (c *Certificate) IsExpiringSoon() bool {
	if c.NotAfter == nil {
		return false
	}
	return time.Until(*c.NotAfter) < 30*24*time.Hour
}

// IsExpired 检查证书是否已过期
func (c *Certificate) IsExpired() bool {
	if c.NotAfter == nil {
		return false
	}
	return time.Now().After(*c.NotAfter)
}
