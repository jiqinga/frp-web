package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"frp-web-panel/internal/logger"
	"frp-web-panel/internal/model"
	"frp-web-panel/internal/repository"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// httpClient 带超时的 HTTP 客户端，防止 goroutine 泄漏
var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

type AlertService struct {
	alertRepo        *repository.AlertRepo
	trafficRepo      *repository.TrafficRepository
	proxyRepo        *repository.ProxyRepository
	clientRepo       *repository.ClientRepository
	frpServerRepo    *repository.FrpServerRepository
	emailService     *EmailService
	recipientService *AlertRecipientService

	// 离线状态追踪（内存中，不持久化）
	pendingOffline map[string]time.Time // key: "frpc:id" 或 "frps:id", value: 首次检测到离线的时间
	alertingState  map[string]bool      // key: "frpc:id" 或 "frps:id", value: 是否已发送告警
	stateMutex     sync.RWMutex
}

func NewAlertService(alertRepo *repository.AlertRepo, trafficRepo *repository.TrafficRepository, proxyRepo *repository.ProxyRepository) *AlertService {
	return &AlertService{
		alertRepo:        alertRepo,
		trafficRepo:      trafficRepo,
		proxyRepo:        proxyRepo,
		emailService:     NewEmailService(),
		recipientService: NewAlertRecipientService(),
		pendingOffline:   make(map[string]time.Time),
		alertingState:    make(map[string]bool),
	}
}

// SetClientRepo 设置客户端仓库（用于离线告警）
func (s *AlertService) SetClientRepo(repo *repository.ClientRepository) {
	s.clientRepo = repo
}

// SetFrpServerRepo 设置FRP服务器仓库（用于离线告警）
func (s *AlertService) SetFrpServerRepo(repo *repository.FrpServerRepository) {
	s.frpServerRepo = repo
}

func (s *AlertService) CheckAlerts() {
	// 只获取 proxy 类型的流量告警规则，避免查询 ProxyID=0 的无效记录
	rules, err := s.alertRepo.GetEnabledRulesByTargetType(model.AlertTargetProxy)
	if err != nil {
		return
	}

	for _, rule := range rules {
		if s.shouldSkipAlert(rule.ID) {
			continue
		}

		proxy, err := s.proxyRepo.FindByID(rule.ProxyID)
		if err != nil {
			continue
		}

		var currentValue int64
		switch rule.RuleType {
		case "rate":
			currentValue = proxy.CurrentBytesInRate + proxy.CurrentBytesOutRate
		case "daily", "monthly":
			currentValue = proxy.TotalBytesIn + proxy.TotalBytesOut
		}

		thresholdBytes := s.convertToBytes(rule.ThresholdValue, rule.ThresholdUnit)
		if currentValue > thresholdBytes {
			alert := &model.AlertLog{
				RuleID:         rule.ID,
				ProxyID:        rule.ProxyID,
				AlertType:      rule.RuleType,
				CurrentValue:   currentValue,
				ThresholdValue: thresholdBytes,
				Message:        fmt.Sprintf("代理 %s 的%s流量已超过阈值", proxy.Name, s.getRuleTypeName(rule.RuleType)),
			}
			if err := s.alertRepo.CreateAlert(alert); err == nil {
				s.sendNotification(alert, &rule, proxy.Name)
			}
		}
	}
}

func (s *AlertService) shouldSkipAlert(ruleID uint) bool {
	recentAlert, err := s.alertRepo.GetRecentAlert(ruleID, 1*time.Hour)
	return err == nil && recentAlert != nil
}

func (s *AlertService) sendNotification(alert *model.AlertLog, rule *model.AlertRule, proxyName string) {
	emails := s.getNotifyEmails(rule)
	for _, email := range emails {
		go s.sendEmail(email, alert, proxyName)
	}
	if rule.NotifyWebhook != "" {
		go s.sendWebhook(rule.NotifyWebhook, alert, proxyName)
	}
	s.alertRepo.MarkAsNotified(alert.ID)
}

// getNotifyEmails 获取规则的所有通知邮箱
func (s *AlertService) getNotifyEmails(rule *model.AlertRule) []string {
	recipientIDs := parseIDList(rule.NotifyRecipientIDs)
	groupIDs := parseIDList(rule.NotifyGroupIDs)
	return s.recipientService.GetEmailsByRecipientAndGroupIDs(recipientIDs, groupIDs)
}

func parseIDList(str string) []uint {
	if str == "" {
		return nil
	}
	parts := strings.Split(str, ",")
	ids := make([]uint, 0, len(parts))
	for _, p := range parts {
		if id, err := strconv.ParseUint(strings.TrimSpace(p), 10, 32); err == nil {
			ids = append(ids, uint(id))
		}
	}
	return ids
}

func (s *AlertService) sendEmail(to string, alert *model.AlertLog, proxyName string) error {
	subject := fmt.Sprintf("FRP流量告警 - %s", proxyName)
	html, _, err := GenerateTrafficAlertEmail(TrafficAlertData{
		ProxyName:    proxyName,
		AlertType:    s.getRuleTypeName(alert.AlertType),
		CurrentValue: formatBytes(alert.CurrentValue),
		Threshold:    formatBytes(alert.ThresholdValue),
		Time:         alert.CreatedAt,
	})
	if err != nil {
		return err
	}
	return s.emailService.SendHTMLEmail(to, subject, html, "text/html")
}

func (s *AlertService) sendWebhook(url string, alert *model.AlertLog, proxyName string) error {
	payload := map[string]interface{}{
		"proxy_name":      proxyName,
		"alert_type":      alert.AlertType,
		"current_value":   alert.CurrentValue,
		"threshold_value": alert.ThresholdValue,
		"message":         alert.Message,
		"timestamp":       alert.CreatedAt.Unix(),
	}

	jsonData, _ := json.Marshal(payload)
	resp, err := httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func (s *AlertService) convertToBytes(value int64, unit string) int64 {
	switch unit {
	case "MB":
		return value * 1024 * 1024
	case "GB":
		return value * 1024 * 1024 * 1024
	case "TB":
		return value * 1024 * 1024 * 1024 * 1024
	default:
		return value
	}
}

func (s *AlertService) getRuleTypeName(ruleType string) string {
	switch ruleType {
	case "daily":
		return "每日"
	case "monthly":
		return "每月"
	case "rate":
		return "实时速率"
	default:
		return ""
	}
}

func (s *AlertService) CreateRule(rule *model.AlertRule) error {
	return s.alertRepo.CreateRule(rule)
}

func (s *AlertService) GetRulesByProxyID(proxyID uint) ([]model.AlertRule, error) {
	return s.alertRepo.GetRulesByProxyID(proxyID)
}

func (s *AlertService) UpdateRule(rule *model.AlertRule) error {
	return s.alertRepo.UpdateRule(rule)
}

func (s *AlertService) DeleteRule(id uint) error {
	return s.alertRepo.DeleteRule(id)
}

func (s *AlertService) GetAllRules() ([]model.AlertRule, error) {
	return s.alertRepo.GetAllRules()
}

func (s *AlertService) GetAlertLogs(limit int) ([]model.AlertLog, error) {
	return s.alertRepo.GetAlertLogs(limit)
}

// CheckOfflineAlerts 检查 frpc 和 frps 离线告警
func (s *AlertService) CheckOfflineAlerts() {
	s.checkFrpcOfflineAlerts()
	s.checkFrpsOfflineAlerts()
}

// checkFrpcOfflineAlerts 检查 frpc 离线告警（带延迟确认）
func (s *AlertService) checkFrpcOfflineAlerts() {
	if s.clientRepo == nil {
		return
	}

	rules, err := s.alertRepo.GetEnabledRulesByTargetType(model.AlertTargetFrpc)
	if err != nil {
		logger.Errorf("离线告警 获取 frpc 告警规则失败: %v", err)
		return
	}

	for _, rule := range rules {
		client, err := s.clientRepo.FindByID(rule.TargetID)
		if err != nil {
			continue
		}

		targetKey := fmt.Sprintf("frpc:%d", rule.TargetID)
		isOffline := client.OnlineStatus == "offline"

		s.handleOfflineState(targetKey, isOffline, &rule, client.Name, "frpc")
	}
}

// checkFrpsOfflineAlerts 检查 frps 离线告警（带延迟确认）
func (s *AlertService) checkFrpsOfflineAlerts() {
	if s.frpServerRepo == nil {
		return
	}

	rules, err := s.alertRepo.GetEnabledRulesByTargetType(model.AlertTargetFrps)
	if err != nil {
		logger.Errorf("离线告警 获取 frps 告警规则失败: %v", err)
		return
	}

	for _, rule := range rules {
		server, err := s.frpServerRepo.GetByID(rule.TargetID)
		if err != nil {
			continue
		}

		targetKey := fmt.Sprintf("frps:%d", rule.TargetID)
		isOffline := server.Status != model.StatusRunning

		s.handleOfflineState(targetKey, isOffline, &rule, server.Name, "frps")
	}
}

// handleOfflineState 处理离线状态（延迟确认 + 恢复通知）
func (s *AlertService) handleOfflineState(targetKey string, isOffline bool, rule *model.AlertRule, targetName, targetType string) {
	s.stateMutex.Lock()
	defer s.stateMutex.Unlock()

	delaySeconds := rule.OfflineDelaySeconds
	if delaySeconds <= 0 {
		delaySeconds = 60 // 默认60秒
	}

	if isOffline {
		// 目标离线
		firstOfflineTime, pending := s.pendingOffline[targetKey]
		if !pending {
			// 首次检测到离线，记录时间
			s.pendingOffline[targetKey] = time.Now()
			logger.Infof("离线告警 %s %s 检测到离线，等待 %d 秒确认", targetType, targetName, delaySeconds)
			return
		}

		// 检查是否超过延迟确认时间
		if time.Since(firstOfflineTime) < time.Duration(delaySeconds)*time.Second {
			return // 还在延迟确认期内
		}

		// 检查是否已发送告警
		if s.alertingState[targetKey] {
			return // 已发送告警，等待恢复
		}

		// 检查冷却时间
		cooldown := time.Duration(rule.CooldownMinutes) * time.Minute
		var alertTargetType model.AlertTargetType
		if targetType == "frpc" {
			alertTargetType = model.AlertTargetFrpc
		} else {
			alertTargetType = model.AlertTargetFrps
		}
		if s.shouldSkipAlertByTargetUnlocked(alertTargetType, rule.TargetID, cooldown) {
			return
		}

		// 发送离线告警
		alert := &model.AlertLog{
			RuleID:     rule.ID,
			TargetType: alertTargetType,
			TargetID:   rule.TargetID,
			ProxyID:    0,
			AlertType:  "offline",
			Message:    fmt.Sprintf("%s %s 已离线（持续 %d 秒）", targetType, targetName, int(time.Since(firstOfflineTime).Seconds())),
		}
		if err := s.alertRepo.CreateAlert(alert); err == nil {
			s.sendOfflineNotification(alert, rule, targetName, targetType)
			s.alertingState[targetKey] = true
			logger.Infof("离线告警 %s %s 离线告警已发送", targetType, targetName)
		}
	} else {
		// 目标在线
		wasAlerting := s.alertingState[targetKey]
		_, wasPending := s.pendingOffline[targetKey]

		// 清除状态
		delete(s.pendingOffline, targetKey)
		delete(s.alertingState, targetKey)

		// 如果之前已发送告警且启用了恢复通知，发送恢复通知
		if wasAlerting && rule.NotifyOnRecovery {
			s.sendRecoveryNotification(rule, targetName, targetType)
			logger.Infof("离线告警 %s %s 已恢复在线，恢复通知已发送", targetType, targetName)
		} else if wasPending {
			logger.Infof("离线告警 %s %s 在延迟确认期内恢复，取消告警", targetType, targetName)
		}
	}
}

// shouldSkipAlertByTargetUnlocked 检查是否应跳过告警（不加锁版本）
func (s *AlertService) shouldSkipAlertByTargetUnlocked(targetType model.AlertTargetType, targetID uint, cooldown time.Duration) bool {
	recentAlert, err := s.alertRepo.GetRecentAlertByTarget(targetType, targetID, cooldown)
	return err == nil && recentAlert != nil
}

// sendRecoveryNotification 发送恢复通知
func (s *AlertService) sendRecoveryNotification(rule *model.AlertRule, targetName, targetType string) {
	emails := s.getNotifyEmails(rule)
	for _, email := range emails {
		go s.sendRecoveryEmail(email, targetName, targetType)
	}
	if rule.NotifyWebhook != "" {
		go s.sendRecoveryWebhook(rule.NotifyWebhook, targetName, targetType)
	}
}

func (s *AlertService) sendRecoveryEmail(to, targetName, targetType string) error {
	subject := fmt.Sprintf("FRP恢复通知 - %s %s 已恢复在线", targetType, targetName)
	html, _, err := GenerateRecoveryAlertEmail(RecoveryAlertData{
		TargetType: targetType,
		TargetName: targetName,
		Time:       time.Now(),
	})
	if err != nil {
		return err
	}
	return s.emailService.SendHTMLEmail(to, subject, html, "text/html")
}

func (s *AlertService) sendRecoveryWebhook(url, targetName, targetType string) error {
	payload := map[string]interface{}{
		"target_type": targetType,
		"target_name": targetName,
		"alert_type":  "recovery",
		"message":     fmt.Sprintf("%s %s 已恢复在线", targetType, targetName),
		"timestamp":   time.Now().Unix(),
	}

	jsonData, _ := json.Marshal(payload)
	resp, err := httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (s *AlertService) sendOfflineNotification(alert *model.AlertLog, rule *model.AlertRule, targetName, targetType string) {
	emails := s.getNotifyEmails(rule)
	for _, email := range emails {
		go s.sendOfflineEmail(email, alert, targetName, targetType)
	}
	if rule.NotifyWebhook != "" {
		go s.sendOfflineWebhook(rule.NotifyWebhook, alert, targetName, targetType)
	}
	s.alertRepo.MarkAsNotified(alert.ID)
}

func (s *AlertService) sendOfflineEmail(to string, alert *model.AlertLog, targetName, targetType string) error {
	subject := fmt.Sprintf("FRP离线告警 - %s %s", targetType, targetName)
	html, _, err := GenerateOfflineAlertEmail(OfflineAlertData{
		TargetType: targetType,
		TargetName: targetName,
		Message:    alert.Message,
		Time:       alert.CreatedAt,
	})
	if err != nil {
		return err
	}
	return s.emailService.SendHTMLEmail(to, subject, html, "text/html")
}

func (s *AlertService) sendOfflineWebhook(url string, alert *model.AlertLog, targetName, targetType string) error {
	payload := map[string]interface{}{
		"target_type": targetType,
		"target_name": targetName,
		"alert_type":  "offline",
		"message":     alert.Message,
		"timestamp":   alert.CreatedAt.Unix(),
	}

	jsonData, _ := json.Marshal(payload)
	resp, err := httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// GetRulesByTargetType 根据目标类型获取规则
func (s *AlertService) GetRulesByTargetType(targetType model.AlertTargetType) ([]model.AlertRule, error) {
	return s.alertRepo.GetEnabledRulesByTargetType(targetType)
}
