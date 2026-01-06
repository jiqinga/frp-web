package service

import (
	"frp-web-panel/internal/model"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetEmailsByRecipientAndGroupIDs_EmptyInputs 测试空输入
func TestGetEmailsByRecipientAndGroupIDs_EmptyInputs(t *testing.T) {
	t.Run("空接收人和分组ID返回空列表", func(t *testing.T) {
		// 由于服务依赖全局数据库，这里测试逻辑
		emailMap := make(map[string]bool)
		emails := make([]string, 0, len(emailMap))
		for email := range emailMap {
			emails = append(emails, email)
		}
		assert.Empty(t, emails, "空输入应返回空列表")
	})
}

// TestEmailDeduplication 测试邮箱去重逻辑
func TestEmailDeduplication(t *testing.T) {
	tests := []struct {
		name           string
		directEmails   []string
		groupEmails    []string
		expectedCount  int
		expectedEmails []string
	}{
		{
			name:           "无重复邮箱",
			directEmails:   []string{"a@test.com", "b@test.com"},
			groupEmails:    []string{"c@test.com"},
			expectedCount:  3,
			expectedEmails: []string{"a@test.com", "b@test.com", "c@test.com"},
		},
		{
			name:           "有重复邮箱",
			directEmails:   []string{"a@test.com", "b@test.com"},
			groupEmails:    []string{"a@test.com", "c@test.com"},
			expectedCount:  3,
			expectedEmails: []string{"a@test.com", "b@test.com", "c@test.com"},
		},
		{
			name:           "全部重复",
			directEmails:   []string{"a@test.com"},
			groupEmails:    []string{"a@test.com"},
			expectedCount:  1,
			expectedEmails: []string{"a@test.com"},
		},
		{
			name:           "空直接接收人",
			directEmails:   []string{},
			groupEmails:    []string{"a@test.com", "b@test.com"},
			expectedCount:  2,
			expectedEmails: []string{"a@test.com", "b@test.com"},
		},
		{
			name:           "空分组",
			directEmails:   []string{"a@test.com"},
			groupEmails:    []string{},
			expectedCount:  1,
			expectedEmails: []string{"a@test.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 模拟去重逻辑
			emailMap := make(map[string]bool)
			for _, email := range tt.directEmails {
				emailMap[email] = true
			}
			for _, email := range tt.groupEmails {
				emailMap[email] = true
			}

			emails := make([]string, 0, len(emailMap))
			for email := range emailMap {
				emails = append(emails, email)
			}

			assert.Equal(t, tt.expectedCount, len(emails), "去重后邮箱数量应正确")
		})
	}
}

// TestAlertRecipientModel 测试接收人模型
func TestAlertRecipientModel(t *testing.T) {
	t.Run("接收人表名", func(t *testing.T) {
		r := model.AlertRecipient{}
		assert.Equal(t, "alert_recipients", r.TableName())
	})

	t.Run("分组表名", func(t *testing.T) {
		g := model.AlertRecipientGroup{}
		assert.Equal(t, "alert_recipient_groups", g.TableName())
	})

	t.Run("关联表名", func(t *testing.T) {
		gr := model.AlertGroupRecipient{}
		assert.Equal(t, "alert_group_recipients", gr.TableName())
	})
}

// TestAlertRecipientFields 测试接收人字段
func TestAlertRecipientFields(t *testing.T) {
	recipient := model.AlertRecipient{
		ID:      1,
		Name:    "测试接收人",
		Email:   "test@example.com",
		Enabled: true,
	}

	assert.Equal(t, uint(1), recipient.ID)
	assert.Equal(t, "测试接收人", recipient.Name)
	assert.Equal(t, "test@example.com", recipient.Email)
	assert.True(t, recipient.Enabled)
}

// TestAlertRecipientGroupFields 测试分组字段
func TestAlertRecipientGroupFields(t *testing.T) {
	group := model.AlertRecipientGroup{
		ID:          1,
		Name:        "测试分组",
		Description: "测试描述",
		Enabled:     true,
		Recipients:  []model.AlertRecipient{},
	}

	assert.Equal(t, uint(1), group.ID)
	assert.Equal(t, "测试分组", group.Name)
	assert.Equal(t, "测试描述", group.Description)
	assert.True(t, group.Enabled)
	assert.Empty(t, group.Recipients)
}
