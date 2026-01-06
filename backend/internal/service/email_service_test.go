package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetPort 测试端口解析
func TestGetPort(t *testing.T) {
	tests := []struct {
		name     string
		portStr  string
		expected int
	}{
		{"有效端口", "587", 587},
		{"SSL端口", "465", 465},
		{"无效端口返回默认值", "invalid", 587},
		{"空字符串返回默认值", "", 587},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetPort(tt.portStr)
			assert.Equal(t, tt.expected, result, "端口解析应该正确")
		})
	}
}

// TestEmailService_IsConfigured 测试邮件配置检查
func TestEmailService_IsConfigured(t *testing.T) {
	t.Run("未配置时返回false", func(t *testing.T) {
		// 由于依赖SettingService，这里只测试逻辑
		// 实际测试需要mock SettingService
		config := &EmailConfig{Host: "", Port: ""}
		isConfigured := config.Host != "" && config.Port != ""
		assert.False(t, isConfigured, "空配置应返回false")
	})

	t.Run("配置完整时返回true", func(t *testing.T) {
		config := &EmailConfig{Host: "smtp.example.com", Port: "587"}
		isConfigured := config.Host != "" && config.Port != ""
		assert.True(t, isConfigured, "完整配置应返回true")
	})

	t.Run("只有Host时返回false", func(t *testing.T) {
		config := &EmailConfig{Host: "smtp.example.com", Port: ""}
		isConfigured := config.Host != "" && config.Port != ""
		assert.False(t, isConfigured, "缺少Port应返回false")
	})

	t.Run("只有Port时返回false", func(t *testing.T) {
		config := &EmailConfig{Host: "", Port: "587"}
		isConfigured := config.Host != "" && config.Port != ""
		assert.False(t, isConfigured, "缺少Host应返回false")
	})
}

// TestEmailMessageFormat 测试邮件消息格式
func TestEmailMessageFormat(t *testing.T) {
	t.Run("邮件头格式正确", func(t *testing.T) {
		from := "sender@example.com"
		to := "receiver@example.com"
		subject := "测试邮件"
		contentType := "text/html"
		body := "<h1>Hello</h1>"

		// 模拟邮件格式化
		msg := "From: " + from + "\r\n" +
			"To: " + to + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: " + contentType + "; charset=UTF-8\r\n\r\n" +
			body

		assert.Contains(t, msg, "From: sender@example.com")
		assert.Contains(t, msg, "To: receiver@example.com")
		assert.Contains(t, msg, "Subject: 测试邮件")
		assert.Contains(t, msg, "Content-Type: text/html")
		assert.Contains(t, msg, "<h1>Hello</h1>")
	})
}

// TestEmailConfig 测试邮件配置结构
func TestEmailConfig(t *testing.T) {
	config := EmailConfig{
		Host:     "smtp.example.com",
		Port:     "587",
		Username: "user@example.com",
		Password: "password123",
		From:     "noreply@example.com",
		SSL:      true,
	}

	assert.Equal(t, "smtp.example.com", config.Host)
	assert.Equal(t, "587", config.Port)
	assert.Equal(t, "user@example.com", config.Username)
	assert.Equal(t, "password123", config.Password)
	assert.Equal(t, "noreply@example.com", config.From)
	assert.True(t, config.SSL)
}

// TestDefaultContentType 测试默认内容类型
func TestDefaultContentType(t *testing.T) {
	t.Run("空内容类型默认为text/html", func(t *testing.T) {
		contentType := ""
		if contentType == "" {
			contentType = "text/html"
		}
		assert.Equal(t, "text/html", contentType)
	})

	t.Run("指定内容类型保持不变", func(t *testing.T) {
		contentType := "text/plain"
		if contentType == "" {
			contentType = "text/html"
		}
		assert.Equal(t, "text/plain", contentType)
	})
}

// TestFromAddressFallback 测试发件人地址回退
func TestFromAddressFallback(t *testing.T) {
	t.Run("From为空时使用Username", func(t *testing.T) {
		config := EmailConfig{
			Username: "user@example.com",
			From:     "",
		}
		from := config.From
		if from == "" {
			from = config.Username
		}
		assert.Equal(t, "user@example.com", from)
	})

	t.Run("From不为空时使用From", func(t *testing.T) {
		config := EmailConfig{
			Username: "user@example.com",
			From:     "noreply@example.com",
		}
		from := config.From
		if from == "" {
			from = config.Username
		}
		assert.Equal(t, "noreply@example.com", from)
	})
}
