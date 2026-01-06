package handler

import (
	"fmt"
	"frp-web-panel/internal/service"
	"frp-web-panel/internal/util"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SettingHandler struct {
	settingService   *service.SettingService
	realtimeService  *service.RealtimeService
	metricsCollector *service.MetricsCollector
	logService       *service.LogService
	emailService     *service.EmailService
}

func NewSettingHandler() *SettingHandler {
	return &SettingHandler{
		settingService: service.NewSettingService(),
		logService:     service.NewLogService(),
		emailService:   service.NewEmailService(),
	}
}

func NewSettingHandlerWithService(realtimeService *service.RealtimeService, metricsCollector *service.MetricsCollector) *SettingHandler {
	return &SettingHandler{
		settingService:   service.NewSettingService(),
		realtimeService:  realtimeService,
		metricsCollector: metricsCollector,
		logService:       service.NewLogService(),
		emailService:     service.NewEmailService(),
	}
}

// GetSettings godoc
// @Summary 获取系统设置
// @Description 获取所有系统设置项
// @Tags 系统设置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} util.Response{data=map[string]interface{}} "设置列表"
// @Failure 500 {object} util.Response "获取设置失败"
// @Router /api/settings [get]
func (h *SettingHandler) GetSettings(c *gin.Context) {
	settings, err := h.settingService.GetAllSettings()
	if err != nil {
		util.ErrorResponse(c, http.StatusInternalServerError, "获取设置失败")
		return
	}
	util.SuccessResponse(c, settings)
}

// UpdateSetting godoc
// @Summary 更新系统设置
// @Description 更新指定的系统设置项
// @Tags 系统设置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object true "设置信息" SchemaExample({"key": "setting_key", "value": "setting_value"})
// @Success 200 {object} util.Response "更新成功"
// @Failure 400 {object} util.Response "参数错误"
// @Failure 500 {object} util.Response "更新设置失败"
// @Router /api/settings [put]
func (h *SettingHandler) UpdateSetting(c *gin.Context) {
	var req struct {
		Key   string `json:"key" binding:"required"`
		Value string `json:"value" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "参数错误")
		return
	}

	if err := h.settingService.UpdateSetting(req.Key, req.Value); err != nil {
		util.ErrorResponse(c, http.StatusInternalServerError, "更新设置失败")
		return
	}

	if seconds, err := strconv.Atoi(req.Value); err == nil {
		switch req.Key {
		case "server_status_check_interval":
			if h.realtimeService != nil {
				h.realtimeService.UpdateCheckInterval(seconds)
			}
		case "traffic_interval":
			if h.metricsCollector != nil {
				h.metricsCollector.UpdateInterval(seconds)
			}
		}
	}

	userID, _ := c.Get("user_id")
	h.logService.CreateLogAsync(userID.(uint), "update", "setting", 0,
		fmt.Sprintf("更新系统设置: %s = %s", req.Key, req.Value), c.ClientIP())

	util.SuccessResponse(c, nil)
}

// TestEmail godoc
// @Summary 测试邮件配置
// @Description 发送测试邮件以验证邮件配置是否正确
// @Tags 系统设置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object true "测试邮件请求" SchemaExample({"to": "test@example.com"})
// @Success 200 {object} util.Response{data=object} "测试邮件已发送"
// @Failure 400 {object} util.Response "请输入有效的邮箱地址"
// @Failure 500 {object} util.Response "发送测试邮件失败"
// @Router /api/settings/test-email [post]
func (h *SettingHandler) TestEmail(c *gin.Context) {
	var req struct {
		To string `json:"to" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "请输入有效的邮箱地址")
		return
	}

	if err := h.emailService.TestEmail(req.To); err != nil {
		util.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("发送测试邮件失败: %v", err))
		return
	}

	userID, _ := c.Get("user_id")
	h.logService.CreateLogAsync(userID.(uint), "test", "email", 0,
		fmt.Sprintf("发送测试邮件到: %s", req.To), c.ClientIP())

	util.SuccessResponse(c, gin.H{"message": "测试邮件已发送"})
}
