/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-14 16:17:28
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-26 14:31:44
 * @FilePath            : frp-web-testbackendinternalservicesetting_service.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package service

import (
	"frp-web-panel/internal/model"
	"frp-web-panel/internal/repository"
)

// 邮件配置相关的设置项 key
const (
	SettingSmtpHost     = "smtp_host"
	SettingSmtpPort     = "smtp_port"
	SettingSmtpUsername = "smtp_username"
	SettingSmtpPassword = "smtp_password"
	SettingSmtpFrom     = "smtp_from"
	SettingSmtpSSL      = "smtp_ssl"
	SettingPanelURL     = "panel_url"
	SettingAcmeEmail    = "acme_email"
)

// EmailConfig 邮件配置结构
type EmailConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
	SSL      bool
}

type SettingService struct {
	settingRepo   *repository.SettingRepository
	eventNotifier *SystemEventNotifier
}

func NewSettingService() *SettingService {
	return &SettingService{
		settingRepo: repository.NewSettingRepository(),
	}
}

// SetEventNotifier 设置系统事件通知器
func (s *SettingService) SetEventNotifier(notifier *SystemEventNotifier) {
	s.eventNotifier = notifier
}

func (s *SettingService) GetAllSettings() ([]model.Setting, error) {
	// 确保 panel_url 设置项存在
	s.settingRepo.GetOrCreate(SettingPanelURL, "http://localhost:3000", "FRP Panel 访问地址")
	return s.settingRepo.GetAllSettings()
}

func (s *SettingService) UpdateSetting(key, value string) error {
	// 获取旧值用于通知
	oldValue, _ := s.settingRepo.GetSetting(key)
	if err := s.settingRepo.UpdateSetting(key, value); err != nil {
		return err
	}
	// 发送配置变更通知
	if s.eventNotifier != nil && oldValue != value {
		go s.eventNotifier.NotifyConfigChanged(key, oldValue, value, "")
	}
	return nil
}

func (s *SettingService) GetSetting(key string) (string, error) {
	return s.settingRepo.GetSetting(key)
}

func (s *SettingService) GetOrCreate(key, defaultValue, description string) (string, error) {
	return s.settingRepo.GetOrCreate(key, defaultValue, description)
}

// GetPanelURL 获取面板访问地址
func (s *SettingService) GetPanelURL() string {
	url, _ := s.settingRepo.GetOrCreate(SettingPanelURL, "http://localhost:3000", "FRP Panel 访问地址")
	return url
}

// GetAcmeEmail 获取 ACME 证书申请邮箱
func (s *SettingService) GetAcmeEmail() string {
	email, _ := s.settingRepo.GetOrCreate(SettingAcmeEmail, "", "ACME证书申请邮箱(Let's Encrypt)")
	return email
}

// GetEmailConfig 获取邮件配置
func (s *SettingService) GetEmailConfig() (*EmailConfig, error) {
	host, _ := s.settingRepo.GetOrCreate(SettingSmtpHost, "", "SMTP服务器地址")
	port, _ := s.settingRepo.GetOrCreate(SettingSmtpPort, "587", "SMTP端口")
	username, _ := s.settingRepo.GetOrCreate(SettingSmtpUsername, "", "SMTP用户名")
	password, _ := s.settingRepo.GetOrCreate(SettingSmtpPassword, "", "SMTP密码")
	from, _ := s.settingRepo.GetOrCreate(SettingSmtpFrom, "", "发件人地址")
	ssl, _ := s.settingRepo.GetOrCreate(SettingSmtpSSL, "false", "启用SSL/TLS")

	return &EmailConfig{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		From:     from,
		SSL:      ssl == "true",
	}, nil
}
