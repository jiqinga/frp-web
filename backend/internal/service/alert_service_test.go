package service

import (
	"frp-web-panel/internal/model"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestConvertToBytes 测试阈值单位转换
func TestConvertToBytes(t *testing.T) {
	s := &AlertService{}

	tests := []struct {
		name     string
		value    int64
		unit     string
		expected int64
	}{
		{"bytes单位保持不变", 1024, "bytes", 1024},
		{"MB转换为bytes", 1, "MB", 1024 * 1024},
		{"GB转换为bytes", 1, "GB", 1024 * 1024 * 1024},
		{"TB转换为bytes", 1, "TB", 1024 * 1024 * 1024 * 1024},
		{"未知单位保持原值", 100, "unknown", 100},
		{"空单位保持原值", 500, "", 500},
		{"10GB转换", 10, "GB", 10 * 1024 * 1024 * 1024},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.convertToBytes(tt.value, tt.unit)
			assert.Equal(t, tt.expected, result, "转换结果应该正确")
		})
	}
}

// TestGetRuleTypeName 测试规则类型名称获取
func TestGetRuleTypeName(t *testing.T) {
	s := &AlertService{}

	tests := []struct {
		name     string
		ruleType string
		expected string
	}{
		{"每日流量类型", "daily", "每日"},
		{"每月流量类型", "monthly", "每月"},
		{"实时速率类型", "rate", "实时速率"},
		{"未知类型返回空字符串", "unknown", ""},
		{"空类型返回空字符串", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.getRuleTypeName(tt.ruleType)
			assert.Equal(t, tt.expected, result, "规则类型名称应该正确")
		})
	}
}

// TestParseIDList 测试ID列表解析
func TestParseIDList(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []uint
	}{
		{"空字符串返回nil", "", nil},
		{"单个ID", "1", []uint{1}},
		{"多个ID逗号分隔", "1,2,3", []uint{1, 2, 3}},
		{"带空格的ID列表", "1, 2, 3", []uint{1, 2, 3}},
		{"包含无效ID时跳过", "1,abc,3", []uint{1, 3}},
		{"大数字ID", "100,200,300", []uint{100, 200, 300}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseIDList(tt.input)
			assert.Equal(t, tt.expected, result, "ID列表解析应该正确")
		})
	}
}

// TestFormatBytes 测试字节格式化
func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{"小于1KB显示B", 500, "500 B"},
		{"1KB", 1024, "1.00 KB"},
		{"1MB", 1024 * 1024, "1.00 MB"},
		{"1GB", 1024 * 1024 * 1024, "1.00 GB"},
		{"1.5GB", int64(1.5 * 1024 * 1024 * 1024), "1.50 GB"},
		{"0字节", 0, "0 B"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatBytes(tt.bytes)
			assert.Equal(t, tt.expected, result, "字节格式化应该正确")
		})
	}
}

// TestHandleOfflineState 测试离线状态处理
func TestHandleOfflineState(t *testing.T) {
	t.Run("首次检测到离线_记录时间", func(t *testing.T) {
		s := &AlertService{
			pendingOffline: make(map[string]time.Time),
			alertingState:  make(map[string]bool),
		}

		rule := &model.AlertRule{ID: 1, OfflineDelaySeconds: 60}
		s.handleOfflineState("frpc:1", true, rule, "test-client", "frpc")

		_, exists := s.pendingOffline["frpc:1"]
		assert.True(t, exists, "应该记录首次离线时间")
		assert.False(t, s.alertingState["frpc:1"], "不应该立即发送告警")
	})

	t.Run("在线状态_清除离线记录", func(t *testing.T) {
		s := &AlertService{
			pendingOffline: make(map[string]time.Time),
			alertingState:  make(map[string]bool),
		}
		s.pendingOffline["frpc:1"] = time.Now().Add(-30 * time.Second)

		rule := &model.AlertRule{ID: 1, OfflineDelaySeconds: 60, NotifyOnRecovery: false}
		s.handleOfflineState("frpc:1", false, rule, "test-client", "frpc")

		_, exists := s.pendingOffline["frpc:1"]
		assert.False(t, exists, "应该清除离线记录")
	})

	t.Run("延迟确认期内_不发送告警", func(t *testing.T) {
		s := &AlertService{
			pendingOffline: make(map[string]time.Time),
			alertingState:  make(map[string]bool),
		}
		s.pendingOffline["frpc:1"] = time.Now().Add(-30 * time.Second)

		rule := &model.AlertRule{ID: 1, OfflineDelaySeconds: 60}
		s.handleOfflineState("frpc:1", true, rule, "test-client", "frpc")

		assert.False(t, s.alertingState["frpc:1"], "延迟期内不应该发送告警")
	})

	t.Run("默认延迟时间_60秒", func(t *testing.T) {
		s := &AlertService{
			pendingOffline: make(map[string]time.Time),
			alertingState:  make(map[string]bool),
		}

		rule := &model.AlertRule{ID: 1, OfflineDelaySeconds: 0}
		s.handleOfflineState("frpc:1", true, rule, "test-client", "frpc")

		_, exists := s.pendingOffline["frpc:1"]
		assert.True(t, exists, "应该使用默认60秒延迟")
	})
}

// TestAlertService_NilRepos 测试空仓库时的行为
func TestAlertService_NilRepos(t *testing.T) {
	t.Run("空clientRepo不panic", func(t *testing.T) {
		s := &AlertService{
			clientRepo:     nil,
			pendingOffline: make(map[string]time.Time),
			alertingState:  make(map[string]bool),
		}
		s.checkFrpcOfflineAlerts()
		assert.True(t, true)
	})

	t.Run("空frpServerRepo不panic", func(t *testing.T) {
		s := &AlertService{
			frpServerRepo:  nil,
			pendingOffline: make(map[string]time.Time),
			alertingState:  make(map[string]bool),
		}
		s.checkFrpsOfflineAlerts()
		assert.True(t, true)
	})
}
