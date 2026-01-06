/*
 * @Author              : 寂情�?
 * @Date                : 2025-11-17 16:41:08
 * @LastEditors         : 寂情�?
 * @LastEditTime        : 2025-12-30 16:27:19
 * @FilePath            : frp-web-testbackendinternalhandleralert_handler.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在�?
 */
package handler

import (
	"fmt"
	"frp-web-panel/internal/model"
	"frp-web-panel/internal/service"
	"frp-web-panel/internal/util"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AlertHandler struct {
	alertService *service.AlertService
	logService   *service.LogService
}

func NewAlertHandler(alertService *service.AlertService) *AlertHandler {
	return &AlertHandler{
		alertService: alertService,
		logService:   service.NewLogService(),
	}
}

// CreateRule godoc
// @Summary 创建告警规则
// @Description 创建新的告警规则，支持多种规则类型和目标类型
// @Tags 告警管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param rule body model.AlertRule true "告警规则信息"
// @Success 200 {object} util.Response{data=object} "创建成功"
// @Failure 400 {object} util.Response "参数错误"
// @Failure 500 {object} util.Response "创建失败"
// @Router /api/alerts/rules [post]
func (h *AlertHandler) CreateRule(c *gin.Context) {
	var rule model.AlertRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		util.Error(c, 4001, "参数错误")
		return
	}

	if err := h.alertService.CreateRule(&rule); err != nil {
		util.Error(c, 4002, "创建告警规则失败")
		return
	}

	// 记录操作日志
	userID, _ := c.Get("user_id")
	logMsg := fmt.Sprintf("创建告警规则 (目标类型: %s, 目标ID: %d, 规则类型: %s)", rule.TargetType, rule.TargetID, rule.RuleType)
	h.logService.CreateLogAsync(userID.(uint), "create", "alert_rule", rule.ID, logMsg, c.ClientIP())

	util.Success(c, rule)
}

// GetRulesByProxyID godoc
// @Summary 按代理ID获取告警规则
// @Description 获取指定代理关联的所有告警规�?
// @Tags 告警管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "代理ID"
// @Success 200 {object} util.Response{data=[]object} "获取成功"
// @Failure 500 {object} util.Response "获取失败"
// @Router /api/alerts/rules/proxy/{id} [get]
func (h *AlertHandler) GetRulesByProxyID(c *gin.Context) {
	proxyID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	rules, err := h.alertService.GetRulesByProxyID(uint(proxyID))
	if err != nil {
		util.Error(c, 4003, "获取告警规则失败")
		return
	}

	util.Success(c, rules)
}

// UpdateRule godoc
// @Summary 更新告警规则
// @Description 更新现有的告警规则配�?
// @Tags 告警管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param rule body model.AlertRule true "告警规则信息"
// @Success 200 {object} util.Response{data=object} "更新成功"
// @Failure 400 {object} util.Response "参数错误"
// @Failure 500 {object} util.Response "更新失败"
// @Router /api/alerts/rules [put]
func (h *AlertHandler) UpdateRule(c *gin.Context) {
	var rule model.AlertRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		util.Error(c, 4001, "参数错误")
		return
	}

	if err := h.alertService.UpdateRule(&rule); err != nil {
		util.Error(c, 4004, "更新告警规则失败")
		return
	}

	// 记录操作日志
	userID, _ := c.Get("user_id")
	logMsg := fmt.Sprintf("更新告警规则 (ID: %d, 目标类型: %s, 规则类型: %s)", rule.ID, rule.TargetType, rule.RuleType)
	h.logService.CreateLogAsync(userID.(uint), "update", "alert_rule", rule.ID, logMsg, c.ClientIP())

	util.Success(c, rule)
}

// DeleteRule godoc
// @Summary 删除告警规则
// @Description 根据ID删除指定的告警规�?
// @Tags 告警管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "规则ID"
// @Success 200 {object} util.Response "删除成功"
// @Failure 500 {object} util.Response "删除失败"
// @Router /api/alerts/rules/{id} [delete]
func (h *AlertHandler) DeleteRule(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if err := h.alertService.DeleteRule(uint(id)); err != nil {
		util.Error(c, 4005, "删除告警规则失败")
		return
	}

	// 记录操作日志
	userID, _ := c.Get("user_id")
	h.logService.CreateLogAsync(userID.(uint), "delete", "alert_rule", uint(id),
		fmt.Sprintf("删除告警规则 (ID: %d)", id), c.ClientIP())

	util.Success(c, nil)
}

// GetAllRules godoc
// @Summary 获取所有告警规�?
// @Description 获取系统中所有的告警规则列表
// @Tags 告警管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} util.Response{data=[]object} "获取成功"
// @Failure 500 {object} util.Response "获取失败"
// @Router /api/alerts/rules [get]
func (h *AlertHandler) GetAllRules(c *gin.Context) {
	rules, err := h.alertService.GetAllRules()
	if err != nil {
		util.Error(c, 4007, "获取告警规则失败")
		return
	}

	util.Success(c, rules)
}

// GetAlertLogs godoc
// @Summary 获取告警日志
// @Description 获取告警触发的历史日志记�?
// @Tags 告警管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "返回记录数量限制" default(100)
// @Success 200 {object} util.Response{data=[]object} "获取成功"
// @Failure 500 {object} util.Response "获取失败"
// @Router /api/alerts/logs [get]
func (h *AlertHandler) GetAlertLogs(c *gin.Context) {
	limit := 100
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	logs, err := h.alertService.GetAlertLogs(limit)
	if err != nil {
		util.Error(c, 4006, "获取告警日志失败")
		return
	}

	util.Success(c, logs)
}
