package service

import (
	"crypto/tls"
	"fmt"
	"frp-web-panel/internal/logger"
	"net/smtp"
	"strconv"
)

type EmailService struct {
	settingService *SettingService
}

func NewEmailService() *EmailService {
	return &EmailService{
		settingService: NewSettingService(),
	}
}

// SendEmail 发送纯文本邮件
func (s *EmailService) SendEmail(to, subject, body string) error {
	return s.SendHTMLEmail(to, subject, body, "text/plain")
}

// SendHTMLEmail 发送HTML邮件
func (s *EmailService) SendHTMLEmail(to, subject, body, contentType string) error {
	config, err := s.settingService.GetEmailConfig()
	if err != nil {
		return fmt.Errorf("获取邮件配置失败: %w", err)
	}

	if config.Host == "" {
		return fmt.Errorf("SMTP服务器未配置")
	}

	from := config.From
	if from == "" {
		from = config.Username
	}

	if contentType == "" {
		contentType = "text/html"
	}

	msg := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: %s; charset=UTF-8\r\n\r\n%s", from, to, subject, contentType, body))

	addr := fmt.Sprintf("%s:%s", config.Host, config.Port)

	if config.SSL {
		return s.sendWithTLS(addr, config, from, to, msg)
	}
	return s.sendWithSTARTTLS(addr, config, from, to, msg)
}

func (s *EmailService) sendWithSTARTTLS(addr string, config *EmailConfig, from, to string, msg []byte) error {
	var auth smtp.Auth
	if config.Username != "" {
		auth = smtp.PlainAuth("", config.Username, config.Password, config.Host)
	}
	return smtp.SendMail(addr, auth, from, []string{to}, msg)
}

func (s *EmailService) sendWithTLS(addr string, config *EmailConfig, from, to string, msg []byte) error {
	tlsConfig := &tls.Config{ServerName: config.Host}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, config.Host)
	if err != nil {
		return err
	}
	defer client.Close()

	if config.Username != "" {
		auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)
		if err := client.Auth(auth); err != nil {
			return err
		}
	}

	if err := client.Mail(from); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}

	w, err := client.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}

	return client.Quit()
}

// TestEmail 测试邮件配置
func (s *EmailService) TestEmail(to string) error {
	config, err := s.settingService.GetEmailConfig()
	if err != nil {
		return fmt.Errorf("获取邮件配置失败: %w", err)
	}

	logger.Infof("邮件测试 配置: Host=%s, Port=%s, SSL=%v, From=%s",
		config.Host, config.Port, config.SSL, config.From)

	subject := "FRP Panel 邮件测试"
	html, _, err := GenerateTestEmail(TestEmailData{
		Host: config.Host,
		Port: config.Port,
		SSL:  config.SSL,
	})
	if err != nil {
		return fmt.Errorf("生成邮件模板失败: %w", err)
	}

	return s.SendHTMLEmail(to, subject, html, "text/html")
}

// IsConfigured 检查邮件是否已配置
func (s *EmailService) IsConfigured() bool {
	config, err := s.settingService.GetEmailConfig()
	if err != nil {
		return false
	}
	return config.Host != "" && config.Port != ""
}

// GetPort 获取端口号
func GetPort(portStr string) int {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 587
	}
	return port
}
