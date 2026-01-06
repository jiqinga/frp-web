package service

import (
	"frp-web-panel/internal/model"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestIsSensitiveKey 测试敏感字段检测
func TestIsSensitiveKey(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{"password结尾", "smtp_password", true},
		{"secret结尾", "api_secret", true},
		{"token结尾", "access_token", true},
		{"key结尾", "api_key", true},
		{"普通字段", "smtp_host", false},
		{"普通字段2", "email_from", false},
		{"空字符串", "", false},
		{"PASSWORD大写不匹配", "PASSWORD", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSensitiveKey(tt.key)
			assert.Equal(t, tt.expected, result, "敏感字段检测应该正确")
		})
	}
}

// TestGetRuleTypeName 测试规则类型名称映射
func TestGetRuleTypeName_SystemEvents(t *testing.T) {
	tests := []struct {
		name     string
		ruleType string
		expected string
	}{
		{"证书申请成功", model.RuleTypeCertApplySuccess, "证书申请成功"},
		{"证书申请失败", model.RuleTypeCertApplyFailed, "证书申请失败"},
		{"证书即将过期", model.RuleTypeCertExpiring, "证书即将过期"},
		{"证书已过期", model.RuleTypeCertExpired, "证书已过期"},
		{"证书续签成功", model.RuleTypeCertRenewSuccess, "证书续签成功"},
		{"证书续签失败", model.RuleTypeCertRenewFailed, "证书续签失败"},
		{"DNS同步成功", model.RuleTypeDNSSyncSuccess, "DNS同步成功"},
		{"DNS同步失败", model.RuleTypeDNSSyncFailed, "DNS同步失败"},
		{"登录失败", model.RuleTypeLoginFailed, "登录失败"},
		{"配置变更", model.RuleTypeConfigChanged, "配置变更"},
		{"未知类型", "unknown_type", "unknown_type"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getRuleTypeName(tt.ruleType)
			assert.Equal(t, tt.expected, result, "规则类型名称应该正确")
		})
	}
}

// TestCertEventData 测试证书事件数据结构
func TestCertEventData(t *testing.T) {
	data := CertEventData{
		Domain:   "example.com",
		CertID:   1,
		ProxyID:  2,
		Error:    "test error",
		ExpiryAt: "2024-12-31",
	}

	assert.Equal(t, "example.com", data.Domain)
	assert.Equal(t, uint(1), data.CertID)
	assert.Equal(t, uint(2), data.ProxyID)
	assert.Equal(t, "test error", data.Error)
	assert.Equal(t, "2024-12-31", data.ExpiryAt)
}

// TestDNSEventData 测试DNS事件数据结构
func TestDNSEventData(t *testing.T) {
	data := DNSEventData{
		Domain:     "example.com",
		RecordType: "A",
		RecordID:   "123",
		Error:      "dns error",
	}

	assert.Equal(t, "example.com", data.Domain)
	assert.Equal(t, "A", data.RecordType)
	assert.Equal(t, "123", data.RecordID)
	assert.Equal(t, "dns error", data.Error)
}

// TestLoginEventData 测试登录事件数据结构
func TestLoginEventData(t *testing.T) {
	data := LoginEventData{
		Username: "admin",
		IP:       "192.168.1.1",
		Error:    "invalid password",
	}

	assert.Equal(t, "admin", data.Username)
	assert.Equal(t, "192.168.1.1", data.IP)
	assert.Equal(t, "invalid password", data.Error)
}

// TestConfigEventData 测试配置变更事件数据结构
func TestConfigEventData(t *testing.T) {
	data := ConfigEventData{
		Key:      "smtp_host",
		OldValue: "old.smtp.com",
		NewValue: "new.smtp.com",
		Operator: "admin",
	}

	assert.Equal(t, "smtp_host", data.Key)
	assert.Equal(t, "old.smtp.com", data.OldValue)
	assert.Equal(t, "new.smtp.com", data.NewValue)
	assert.Equal(t, "admin", data.Operator)
}

// TestAlertTargetTypes 测试告警目标类型常量
func TestAlertTargetTypes(t *testing.T) {
	assert.Equal(t, model.AlertTargetType("proxy"), model.AlertTargetProxy)
	assert.Equal(t, model.AlertTargetType("frpc"), model.AlertTargetFrpc)
	assert.Equal(t, model.AlertTargetType("frps"), model.AlertTargetFrps)
	assert.Equal(t, model.AlertTargetType("system"), model.AlertTargetSystem)
}

// TestRuleTypeConstants 测试规则类型常量
func TestRuleTypeConstants(t *testing.T) {
	assert.Equal(t, "cert_apply_success", model.RuleTypeCertApplySuccess)
	assert.Equal(t, "cert_apply_failed", model.RuleTypeCertApplyFailed)
	assert.Equal(t, "cert_expiring", model.RuleTypeCertExpiring)
	assert.Equal(t, "cert_expired", model.RuleTypeCertExpired)
	assert.Equal(t, "cert_renew_success", model.RuleTypeCertRenewSuccess)
	assert.Equal(t, "cert_renew_failed", model.RuleTypeCertRenewFailed)
	assert.Equal(t, "dns_sync_success", model.RuleTypeDNSSyncSuccess)
	assert.Equal(t, "dns_sync_failed", model.RuleTypeDNSSyncFailed)
	assert.Equal(t, "login_failed", model.RuleTypeLoginFailed)
	assert.Equal(t, "config_changed", model.RuleTypeConfigChanged)
}
