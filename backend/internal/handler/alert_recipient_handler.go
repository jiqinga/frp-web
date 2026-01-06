package handler

import (
	"fmt"
	"frp-web-panel/internal/model"
	"frp-web-panel/internal/service"
	"frp-web-panel/internal/util"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AlertRecipientHandler struct {
	svc        *service.AlertRecipientService
	logService *service.LogService
}

func NewAlertRecipientHandler() *AlertRecipientHandler {
	return &AlertRecipientHandler{
		svc:        service.NewAlertRecipientService(),
		logService: service.NewLogService(),
	}
}

// CreateRecipient godoc
// @Summary 创建告警接收人
// @Description 创建新的告警接收人，用于接收告警通知
// @Tags 告警管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param recipient body model.AlertRecipient true "告警接收人信息"
// @Success 200 {object} util.Response{data=object} "创建成功"
// @Failure 400 {object} util.Response "参数错误"
// @Failure 500 {object} util.Response "创建失败"
// @Router /api/alerts/recipients [post]
func (h *AlertRecipientHandler) CreateRecipient(c *gin.Context) {
	var r model.AlertRecipient
	if err := c.ShouldBindJSON(&r); err != nil {
		util.Error(c, 4001, "参数错误")
		return
	}
	if err := h.svc.CreateRecipient(&r); err != nil {
		util.Error(c, 5001, "创建失败")
		return
	}
	userID, _ := c.Get("user_id")
	h.logService.CreateLogAsync(userID.(uint), "create", "alert_recipient", r.ID,
		fmt.Sprintf("创建告警接收人: %s (%s)", r.Name, r.Email), c.ClientIP())
	util.Success(c, r)
}

// GetAllRecipients godoc
// @Summary 获取所有告警接收人
// @Description 获取系统中所有的告警接收人列表
// @Tags 告警管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} util.Response{data=[]object} "获取成功"
// @Failure 500 {object} util.Response "获取失败"
// @Router /api/alerts/recipients [get]
func (h *AlertRecipientHandler) GetAllRecipients(c *gin.Context) {
	list, err := h.svc.GetAllRecipients()
	if err != nil {
		util.Error(c, 5002, "获取失败")
		return
	}
	util.Success(c, list)
}

// UpdateRecipient godoc
// @Summary 更新告警接收人
// @Description 更新指定告警接收人的信息
// @Tags 告警管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "接收人ID"
// @Param recipient body model.AlertRecipient true "告警接收人信息"
// @Success 200 {object} util.Response{data=object} "更新成功"
// @Failure 400 {object} util.Response "参数错误"
// @Failure 500 {object} util.Response "更新失败"
// @Router /api/alerts/recipients/{id} [put]
func (h *AlertRecipientHandler) UpdateRecipient(c *gin.Context) {
	var r model.AlertRecipient
	if err := c.ShouldBindJSON(&r); err != nil {
		util.Error(c, 4001, "参数错误")
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	r.ID = uint(id)
	if err := h.svc.UpdateRecipient(&r); err != nil {
		util.Error(c, 5003, "更新失败")
		return
	}
	userID, _ := c.Get("user_id")
	h.logService.CreateLogAsync(userID.(uint), "update", "alert_recipient", r.ID,
		fmt.Sprintf("更新告警接收人: %s (%s)", r.Name, r.Email), c.ClientIP())
	util.Success(c, r)
}

// DeleteRecipient godoc
// @Summary 删除告警接收人
// @Description 根据ID删除指定的告警接收人
// @Tags 告警管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "接收人ID"
// @Success 200 {object} util.Response "删除成功"
// @Failure 500 {object} util.Response "删除失败"
// @Router /api/alerts/recipients/{id} [delete]
func (h *AlertRecipientHandler) DeleteRecipient(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if err := h.svc.DeleteRecipient(uint(id)); err != nil {
		util.Error(c, 5004, "删除失败")
		return
	}
	userID, _ := c.Get("user_id")
	h.logService.CreateLogAsync(userID.(uint), "delete", "alert_recipient", uint(id),
		fmt.Sprintf("删除告警接收人 (ID: %d)", id), c.ClientIP())
	util.Success(c, nil)
}

// CreateGroup godoc
// @Summary 创建告警接收人分组
// @Description 创建新的告警接收人分组，用于批量管理接收人
// @Tags 告警管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param group body model.AlertRecipientGroup true "分组信息"
// @Success 200 {object} util.Response{data=object} "创建成功"
// @Failure 400 {object} util.Response "参数错误"
// @Failure 500 {object} util.Response "创建失败"
// @Router /api/alerts/groups [post]
func (h *AlertRecipientHandler) CreateGroup(c *gin.Context) {
	var g model.AlertRecipientGroup
	if err := c.ShouldBindJSON(&g); err != nil {
		util.Error(c, 4001, "参数错误")
		return
	}
	if err := h.svc.CreateGroup(&g); err != nil {
		util.Error(c, 5001, "创建失败")
		return
	}
	userID, _ := c.Get("user_id")
	h.logService.CreateLogAsync(userID.(uint), "create", "alert_recipient_group", g.ID,
		fmt.Sprintf("创建告警接收人分组: %s", g.Name), c.ClientIP())
	util.Success(c, g)
}

// GetAllGroups godoc
// @Summary 获取所有告警接收人分组
// @Description 获取系统中所有的告警接收人分组列表
// @Tags 告警管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} util.Response{data=[]object} "获取成功"
// @Failure 500 {object} util.Response "获取失败"
// @Router /api/alerts/groups [get]
func (h *AlertRecipientHandler) GetAllGroups(c *gin.Context) {
	list, err := h.svc.GetAllGroups()
	if err != nil {
		util.Error(c, 5002, "获取失败")
		return
	}
	util.Success(c, list)
}

// UpdateGroup godoc
// @Summary 更新告警接收人分组
// @Description 更新指定告警接收人分组的信息
// @Tags 告警管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "分组ID"
// @Param group body model.AlertRecipientGroup true "分组信息"
// @Success 200 {object} util.Response{data=object} "更新成功"
// @Failure 400 {object} util.Response "参数错误"
// @Failure 500 {object} util.Response "更新失败"
// @Router /api/alerts/groups/{id} [put]
func (h *AlertRecipientHandler) UpdateGroup(c *gin.Context) {
	var g model.AlertRecipientGroup
	if err := c.ShouldBindJSON(&g); err != nil {
		util.Error(c, 4001, "参数错误")
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	g.ID = uint(id)
	if err := h.svc.UpdateGroup(&g); err != nil {
		util.Error(c, 5003, "更新失败")
		return
	}
	userID, _ := c.Get("user_id")
	h.logService.CreateLogAsync(userID.(uint), "update", "alert_recipient_group", g.ID,
		fmt.Sprintf("更新告警接收人分组: %s", g.Name), c.ClientIP())
	util.Success(c, g)
}

// DeleteGroup godoc
// @Summary 删除告警接收人分组
// @Description 根据ID删除指定的告警接收人分组
// @Tags 告警管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "分组ID"
// @Success 200 {object} util.Response "删除成功"
// @Failure 500 {object} util.Response "删除失败"
// @Router /api/alerts/groups/{id} [delete]
func (h *AlertRecipientHandler) DeleteGroup(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if err := h.svc.DeleteGroup(uint(id)); err != nil {
		util.Error(c, 5004, "删除失败")
		return
	}
	userID, _ := c.Get("user_id")
	h.logService.CreateLogAsync(userID.(uint), "delete", "alert_recipient_group", uint(id),
		fmt.Sprintf("删除告警接收人分组 (ID: %d)", id), c.ClientIP())
	util.Success(c, nil)
}

// SetGroupRecipients godoc
// @Summary 设置分组成员
// @Description 设置指定分组包含的告警接收人列表
// @Tags 告警管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "分组ID"
// @Param request body object{recipient_ids=[]uint} true "接收人ID列表"
// @Success 200 {object} util.Response "设置成功"
// @Failure 400 {object} util.Response "参数错误"
// @Failure 500 {object} util.Response "设置失败"
// @Router /api/alerts/groups/{id}/recipients [put]
func (h *AlertRecipientHandler) SetGroupRecipients(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var req struct {
		RecipientIDs []uint `json:"recipient_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Error(c, 4001, "参数错误")
		return
	}
	if err := h.svc.SetGroupRecipients(uint(id), req.RecipientIDs); err != nil {
		util.Error(c, 5005, "设置失败")
		return
	}
	userID, _ := c.Get("user_id")
	h.logService.CreateLogAsync(userID.(uint), "update", "alert_recipient_group", uint(id),
		fmt.Sprintf("设置分组成员 (分组ID: %d, 成员数: %d)", id, len(req.RecipientIDs)), c.ClientIP())
	util.Success(c, nil)
}
