package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"frp-web-panel/internal/logger"
	"frp-web-panel/internal/model"
	"frp-web-panel/internal/repository"
	"net/http"
	"time"
)

// SystemEventNotifier 系统事件通知器
type SystemEventNotifier struct {
	alertRepo        *repository.AlertRepo
	emailService     *EmailService
	recipientService *AlertRecipientService
}

// NewSystemEventNotifier 创建系统事件通知器
func NewSystemEventNotifier(alertRepo *repository.AlertRepo) *SystemEventNotifier {
	return &SystemEventNotifier{
		alertRepo:        alertRepo,
		emailService:     NewEmailService(),
		recipientService: NewAlertRecipientService(),
	}
}

// CertEventData 证书事件数据
type CertEventData struct {
	Domain   string `json:"domain"`
	CertID   uint   `json:"cert_id,omitempty"`
	ProxyID  uint   `json:"proxy_id,omitempty"`
	Error    string `json:"error,omitempty"`
	ExpiryAt string `json:"expiry_at,omitempty"`
}

// DNSEventData DNS事件数据
type DNSEventData struct {
	Domain     string `json:"domain"`
	RecordType string `json:"record_type"`
	RecordID   string `json:"record_id,omitempty"`
	Error      string `json:"error,omitempty"`
}

// LoginEventData 登录事件数据
type LoginEventData struct {
	Username string `json:"username"`
	IP       string `json:"ip"`
	Error    string `json:"error,omitempty"`
}

// ConfigEventData 配置变更事件数据
type ConfigEventData struct {
	Key      string `json:"key"`
	OldValue string `json:"old_value,omitempty"`
	NewValue string `json:"new_value,omitempty"`
	Operator string `json:"operator,omitempty"`
}

// NotifyCertApply 证书申请事件通知
func (n *SystemEventNotifier) NotifyCertApply(domain string, certID uint, success bool, errMsg string) {
	ruleType := model.RuleTypeCertApplySuccess
	if !success {
		ruleType = model.RuleTypeCertApplyFailed
	}

	eventData := CertEventData{Domain: domain, CertID: certID, Error: errMsg}
	message := fmt.Sprintf("证书申请成功: %s", domain)
	if !success {
		message = fmt.Sprintf("证书申请失败: %s, 错误: %s", domain, errMsg)
	}

	n.notifySystemEvent(ruleType, message, eventData)
}

// NotifyCertRenew 证书续签事件通知
func (n *SystemEventNotifier) NotifyCertRenew(domain string, certID uint, success bool, errMsg string) {
	ruleType := model.RuleTypeCertRenewSuccess
	if !success {
		ruleType = model.RuleTypeCertRenewFailed
	}

	eventData := CertEventData{Domain: domain, CertID: certID, Error: errMsg}
	message := fmt.Sprintf("证书续签成功: %s", domain)
	if !success {
		message = fmt.Sprintf("证书续签失败: %s, 错误: %s", domain, errMsg)
	}

	n.notifySystemEvent(ruleType, message, eventData)
}

// NotifyCertExpiring 证书即将过期通知
func (n *SystemEventNotifier) NotifyCertExpiring(domain string, certID uint, expiryAt time.Time) {
	eventData := CertEventData{Domain: domain, CertID: certID, ExpiryAt: expiryAt.Format("2006-01-02")}
	message := fmt.Sprintf("证书即将过期: %s, 过期时间: %s", domain, expiryAt.Format("2006-01-02"))
	n.notifySystemEvent(model.RuleTypeCertExpiring, message, eventData)
}

// NotifyCertExpired 证书已过期通知
func (n *SystemEventNotifier) NotifyCertExpired(domain string, certID uint, expiryAt time.Time) {
	eventData := CertEventData{Domain: domain, CertID: certID, ExpiryAt: expiryAt.Format("2006-01-02")}
	message := fmt.Sprintf("证书已过期: %s, 过期时间: %s", domain, expiryAt.Format("2006-01-02"))
	n.notifySystemEvent(model.RuleTypeCertExpired, message, eventData)
}

// NotifyDNSSync DNS同步事件通知
func (n *SystemEventNotifier) NotifyDNSSync(domain string, recordType string, success bool, errMsg string) {
	ruleType := model.RuleTypeDNSSyncSuccess
	if !success {
		ruleType = model.RuleTypeDNSSyncFailed
	}

	eventData := DNSEventData{Domain: domain, RecordType: recordType, Error: errMsg}
	message := fmt.Sprintf("DNS记录同步成功: %s (%s)", domain, recordType)
	if !success {
		message = fmt.Sprintf("DNS记录同步失败: %s (%s), 错误: %s", domain, recordType, errMsg)
	}

	n.notifySystemEvent(ruleType, message, eventData)
}

// NotifyLoginFailed 登录失败通知
func (n *SystemEventNotifier) NotifyLoginFailed(username string, ip string) {
	eventData := LoginEventData{Username: username, IP: ip}
	message := fmt.Sprintf("登录失败: 用户 %s, IP: %s", username, ip)
	n.notifySystemEvent(model.RuleTypeLoginFailed, message, eventData)
}

// NotifyConfigChanged 配置变更通知
func (n *SystemEventNotifier) NotifyConfigChanged(key string, oldValue string, newValue string, operator string) {
	// 敏感字段脱敏
	if isSensitiveKey(key) {
		oldValue = "******"
		newValue = "******"
	}
	eventData := ConfigEventData{Key: key, OldValue: oldValue, NewValue: newValue, Operator: operator}
	message := fmt.Sprintf("配置变更: %s", key)
	n.notifySystemEvent(model.RuleTypeConfigChanged, message, eventData)
}

func isSensitiveKey(key string) bool {
	sensitiveKeys := []string{"password", "secret", "token", "key"}
	for _, sk := range sensitiveKeys {
		if len(key) >= len(sk) && key[len(key)-len(sk):] == sk {
			return true
		}
	}
	return false
}

// notifySystemEvent 通用系统事件通知
func (n *SystemEventNotifier) notifySystemEvent(ruleType string, message string, eventData interface{}) {
	rules, err := n.alertRepo.GetSystemAlertRulesByRuleType(ruleType)
	if err != nil || len(rules) == 0 {
		return
	}

	eventDataJSON, _ := json.Marshal(eventData)

	for _, rule := range rules {
		// 检查冷却时间
		cooldown := time.Duration(rule.CooldownMinutes) * time.Minute
		if cooldown > 0 {
			recent, _ := n.alertRepo.GetRecentSystemAlert(ruleType, cooldown)
			if recent != nil {
				continue
			}
		}

		// 创建告警日志
		alert := &model.AlertLog{
			RuleID:     rule.ID,
			TargetType: model.AlertTargetSystem,
			TargetID:   0,
			AlertType:  ruleType,
			Message:    message,
			EventData:  string(eventDataJSON),
		}

		if err := n.alertRepo.CreateAlert(alert); err != nil {
			logger.Errorf("系统告警 创建告警日志失败: %v", err)
			continue
		}

		// 发送通知
		n.sendNotification(alert, &rule, ruleType)
	}
}

func (n *SystemEventNotifier) sendNotification(alert *model.AlertLog, rule *model.AlertRule, ruleType string) {
	emails := n.getNotifyEmails(rule)
	for _, email := range emails {
		go n.sendSystemAlertEmail(email, alert, ruleType)
	}
	if rule.NotifyWebhook != "" {
		go n.sendSystemAlertWebhook(rule.NotifyWebhook, alert, ruleType)
	}
	n.alertRepo.MarkAsNotified(alert.ID)
}

func (n *SystemEventNotifier) getNotifyEmails(rule *model.AlertRule) []string {
	recipientIDs := parseIDList(rule.NotifyRecipientIDs)
	groupIDs := parseIDList(rule.NotifyGroupIDs)
	return n.recipientService.GetEmailsByRecipientAndGroupIDs(recipientIDs, groupIDs)
}

func (n *SystemEventNotifier) sendSystemAlertEmail(to string, alert *model.AlertLog, ruleType string) error {
	subject := fmt.Sprintf("FRP系统告警 - %s", getRuleTypeName(ruleType))
	html, _, err := GenerateSystemAlertEmail(SystemAlertData{
		AlertType: getRuleTypeName(ruleType),
		Message:   alert.Message,
		EventData: alert.EventData,
		Time:      alert.CreatedAt,
	})
	if err != nil {
		return err
	}
	return n.emailService.SendHTMLEmail(to, subject, html, "text/html")
}

func (n *SystemEventNotifier) sendSystemAlertWebhook(url string, alert *model.AlertLog, ruleType string) error {
	payload := map[string]interface{}{
		"alert_type": ruleType,
		"message":    alert.Message,
		"event_data": alert.EventData,
		"timestamp":  alert.CreatedAt.Unix(),
	}
	jsonData, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func getRuleTypeName(ruleType string) string {
	names := map[string]string{
		model.RuleTypeCertApplySuccess: "证书申请成功",
		model.RuleTypeCertApplyFailed:  "证书申请失败",
		model.RuleTypeCertExpiring:     "证书即将过期",
		model.RuleTypeCertExpired:      "证书已过期",
		model.RuleTypeCertRenewSuccess: "证书续签成功",
		model.RuleTypeCertRenewFailed:  "证书续签失败",
		model.RuleTypeDNSSyncSuccess:   "DNS同步成功",
		model.RuleTypeDNSSyncFailed:    "DNS同步失败",
		model.RuleTypeLoginFailed:      "登录失败",
		model.RuleTypeConfigChanged:    "配置变更",
	}
	if name, ok := names[ruleType]; ok {
		return name
	}
	return ruleType
}
