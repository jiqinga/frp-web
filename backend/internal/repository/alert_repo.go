/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-17 16:38:45
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-26 14:27:41
 * @FilePath            : frp-web-testbackendinternalrepositoryalert_repo.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package repository

import (
	"frp-web-panel/internal/model"
	"time"

	"gorm.io/gorm"
)

type AlertRepo struct {
	db *gorm.DB
}

func NewAlertRepo(db *gorm.DB) *AlertRepo {
	return &AlertRepo{db: db}
}

func (r *AlertRepo) CreateRule(rule *model.AlertRule) error {
	return r.db.Create(rule).Error
}

func (r *AlertRepo) GetEnabledRules() ([]model.AlertRule, error) {
	var rules []model.AlertRule
	err := r.db.Where("enabled = ?", true).Find(&rules).Error
	return rules, err
}

func (r *AlertRepo) GetEnabledRulesByTargetType(targetType model.AlertTargetType) ([]model.AlertRule, error) {
	var rules []model.AlertRule
	err := r.db.Where("enabled = ? AND target_type = ?", true, targetType).Find(&rules).Error
	return rules, err
}

func (r *AlertRepo) GetAllRules() ([]model.AlertRule, error) {
	var rules []model.AlertRule
	err := r.db.Find(&rules).Error
	return rules, err
}

func (r *AlertRepo) GetRulesByProxyID(proxyID uint) ([]model.AlertRule, error) {
	var rules []model.AlertRule
	err := r.db.Where("proxy_id = ?", proxyID).Find(&rules).Error
	return rules, err
}

func (r *AlertRepo) GetRulesByTargetTypeAndID(targetType model.AlertTargetType, targetID uint) ([]model.AlertRule, error) {
	var rules []model.AlertRule
	err := r.db.Where("target_type = ? AND target_id = ?", targetType, targetID).Find(&rules).Error
	return rules, err
}

func (r *AlertRepo) UpdateRule(rule *model.AlertRule) error {
	return r.db.Save(rule).Error
}

func (r *AlertRepo) DeleteRule(id uint) error {
	return r.db.Delete(&model.AlertRule{}, id).Error
}

func (r *AlertRepo) CreateAlert(alert *model.AlertLog) error {
	return r.db.Create(alert).Error
}

func (r *AlertRepo) GetAlertLogs(limit int) ([]model.AlertLog, error) {
	var logs []model.AlertLog
	err := r.db.Order("created_at DESC").Limit(limit).Find(&logs).Error
	return logs, err
}

func (r *AlertRepo) GetRecentAlert(ruleID uint, duration time.Duration) (*model.AlertLog, error) {
	var log model.AlertLog
	cutoff := time.Now().Add(-duration)
	err := r.db.Where("rule_id = ? AND created_at > ?", ruleID, cutoff).
		Order("created_at DESC").First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *AlertRepo) GetRecentAlertByTarget(targetType model.AlertTargetType, targetID uint, duration time.Duration) (*model.AlertLog, error) {
	var log model.AlertLog
	cutoff := time.Now().Add(-duration)
	err := r.db.Where("target_type = ? AND target_id = ? AND created_at > ?", targetType, targetID, cutoff).
		Order("created_at DESC").First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *AlertRepo) MarkAsNotified(id uint) error {
	return r.db.Model(&model.AlertLog{}).Where("id = ?", id).Update("notified", true).Error
}

// GetSystemAlertRulesByRuleType 获取指定规则类型的系统告警规则
func (r *AlertRepo) GetSystemAlertRulesByRuleType(ruleType string) ([]model.AlertRule, error) {
	var rules []model.AlertRule
	err := r.db.Where("enabled = ? AND target_type = ? AND rule_type = ?", true, model.AlertTargetSystem, ruleType).Find(&rules).Error
	return rules, err
}

// GetRecentSystemAlert 获取最近的系统告警（用于冷却检查）
func (r *AlertRepo) GetRecentSystemAlert(ruleType string, duration time.Duration) (*model.AlertLog, error) {
	var log model.AlertLog
	cutoff := time.Now().Add(-duration)
	err := r.db.Where("target_type = ? AND alert_type = ? AND created_at > ?", model.AlertTargetSystem, ruleType, cutoff).
		Order("created_at DESC").First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}
