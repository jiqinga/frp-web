/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-17 16:38:11
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-26 14:27:14
 * @FilePath            : frp-web-testbackendinternalmodelalert.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package model

import "time"

// AlertTargetType 告警目标类型
type AlertTargetType string

const (
	AlertTargetProxy  AlertTargetType = "proxy"  // 代理流量告警
	AlertTargetFrpc   AlertTargetType = "frpc"   // frpc 离线告警
	AlertTargetFrps   AlertTargetType = "frps"   // frps 离线告警
	AlertTargetSystem AlertTargetType = "system" // 系统级告警
)

// 系统级告警规则类型常量
const (
	// 证书相关
	RuleTypeCertApplySuccess = "cert_apply_success" // 证书申请成功
	RuleTypeCertApplyFailed  = "cert_apply_failed"  // 证书申请失败
	RuleTypeCertExpiring     = "cert_expiring"      // 证书即将过期
	RuleTypeCertExpired      = "cert_expired"       // 证书已过期
	RuleTypeCertRenewSuccess = "cert_renew_success" // 证书续签成功
	RuleTypeCertRenewFailed  = "cert_renew_failed"  // 证书续签失败
	// DNS相关
	RuleTypeDNSSyncSuccess = "dns_sync_success" // DNS同步成功
	RuleTypeDNSSyncFailed  = "dns_sync_failed"  // DNS同步失败
	// 安全相关
	RuleTypeLoginFailed = "login_failed" // 登录失败
	// 配置相关
	RuleTypeConfigChanged = "config_changed" // 配置变更
)

type AlertRule struct {
	ID                  uint            `json:"id" gorm:"primaryKey"`
	TargetType          AlertTargetType `json:"target_type" gorm:"type:varchar(20);not null;default:'proxy'"` // proxy, frpc, frps
	TargetID            uint            `json:"target_id" gorm:"not null"`                                    // 对应 proxy_id, client_id, frp_server_id
	ProxyID             uint            `json:"proxy_id" gorm:"not null"`                                     // 保留兼容旧数据
	RuleType            string          `json:"rule_type" gorm:"type:varchar(20);not null"`                   // daily, monthly, rate, offline
	ThresholdValue      int64           `json:"threshold_value" gorm:"default:0"`                             // 流量告警阈值，离线告警不需要
	ThresholdUnit       string          `json:"threshold_unit" gorm:"type:varchar(10);default:bytes"`         // bytes, MB, GB
	CooldownMinutes     int             `json:"cooldown_minutes" gorm:"default:60"`                           // 告警冷却时间（分钟）
	OfflineDelaySeconds int             `json:"offline_delay_seconds" gorm:"default:60"`                      // 离线延迟确认时间（秒），防止网络波动误报
	NotifyOnRecovery    bool            `json:"notify_on_recovery" gorm:"default:true"`                       // 恢复在线时是否发送通知
	Enabled             bool            `json:"enabled" gorm:"default:true"`
	NotifyRecipientIDs  string          `json:"notify_recipient_ids" gorm:"type:varchar(500)"` // 接收人ID列表，逗号分隔
	NotifyGroupIDs      string          `json:"notify_group_ids" gorm:"type:varchar(500)"`     // 分组ID列表，逗号分隔
	NotifyWebhook       string          `json:"notify_webhook" gorm:"type:varchar(500)"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
}

type AlertLog struct {
	ID             uint            `json:"id" gorm:"primaryKey"`
	RuleID         uint            `json:"rule_id" gorm:"not null"`
	TargetType     AlertTargetType `json:"target_type" gorm:"type:varchar(20);not null;default:'proxy'"` // proxy, frpc, frps, system
	TargetID       uint            `json:"target_id" gorm:"not null"`                                    // 对应 proxy_id, client_id, frp_server_id，系统告警为0
	ProxyID        uint            `json:"proxy_id" gorm:"not null"`                                     // 保留兼容旧数据
	AlertType      string          `json:"alert_type" gorm:"type:varchar(30);not null"`
	CurrentValue   int64           `json:"current_value" gorm:"default:0"` // 流量告警使用，其他告警为0
	ThresholdValue int64           `json:"threshold_value" gorm:"default:0"`
	Message        string          `json:"message" gorm:"type:text"`
	EventData      string          `json:"event_data" gorm:"type:text"` // 事件详情JSON，用于系统告警
	Notified       bool            `json:"notified" gorm:"default:false"`
	CreatedAt      time.Time       `json:"created_at"`
}
