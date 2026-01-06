package handler

import (
	"fmt"
	"frp-web-panel/internal/errors"
	"frp-web-panel/internal/middleware"
	"frp-web-panel/internal/service"
	"frp-web-panel/internal/util"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Heartbeat godoc
// @Summary 客户端心跳
// @Description 客户端心跳上报接口（无需认证）
// @Tags 客户端管理
// @Accept json
// @Produce json
// @Param request body object{client_id=int,token=string,client_name=string,timestamp=string} true "心跳数据"
// @Success 200 {object} util.Response{data=object{message=string}}
// @Failure 400 {object} util.Response
// @Failure 401 {object} util.Response
// @Failure 404 {object} util.Response
// @Failure 500 {object} util.Response
// @Router /api/clients/heartbeat [post]
func (h *ClientHandler) Heartbeat(c *gin.Context) {
	var req struct {
		ClientID   uint   `json:"client_id" binding:"required"`
		Token      string `json:"token" binding:"required"`
		ClientName string `json:"client_name"`
		Timestamp  string `json:"timestamp"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithAppError(c, errors.NewBadRequest("参数错误"))
		return
	}

	client, err := h.clientService.GetClient(req.ClientID)
	if err != nil {
		middleware.AbortWithAppError(c, errors.NewNotFound("客户端不存在"))
		return
	}
	if client.Token != req.Token {
		middleware.AbortWithAppError(c, errors.NewUnauthorized("Token验证失败"))
		return
	}

	if err := h.clientService.UpdateHeartbeat(req.ClientID); err != nil {
		middleware.AbortWithAppError(c, errors.NewInternal("更新心跳失败", err))
		return
	}

	util.Success(c, gin.H{"message": "心跳上报成功"})
}

// UpdateClientSoftware godoc
// @Summary 更新客户端软件
// @Description 向指定客户端发送软件更新命令
// @Tags 客户端管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "客户端ID"
// @Param request body object{update_type=string,version=string,mirror_id=int} true "更新配置，update_type: frpc/daemon"
// @Success 200 {object} util.Response{data=object{message=string}}
// @Failure 400 {object} util.Response
// @Failure 401 {object} util.Response
// @Failure 500 {object} util.Response
// @Router /api/clients/{id}/update [post]
func (h *ClientHandler) UpdateClientSoftware(c *gin.Context) {
	if h.clientUpdateService == nil {
		middleware.AbortWithAppError(c, errors.NewInternal("更新服务未初始化", nil))
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.AbortWithAppError(c, errors.NewBadRequest("无效的客户端ID"))
		return
	}
	var req struct {
		UpdateType string `json:"update_type" binding:"required"`
		Version    string `json:"version"`
		MirrorID   *uint  `json:"mirror_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithAppError(c, errors.NewBadRequest("参数错误: "+err.Error()))
		return
	}

	if req.UpdateType != "frpc" && req.UpdateType != "daemon" {
		middleware.AbortWithAppError(c, errors.NewBadRequest("无效的更新类型，必须是 frpc 或 daemon"))
		return
	}

	updateReq := &service.UpdateRequest{
		ClientID:   uint(id),
		UpdateType: service.UpdateType(req.UpdateType),
		Version:    req.Version,
		MirrorID:   req.MirrorID,
	}

	if err := h.clientUpdateService.UpdateClient(updateReq); err != nil {
		middleware.AbortWithAppError(c, errors.NewInternal("更新失败: "+err.Error(), err))
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		middleware.AbortWithAppError(c, errors.NewUnauthorized("用户未认证"))
		return
	}
	uid, ok := userID.(uint)
	if !ok {
		middleware.AbortWithAppError(c, errors.NewInternal("用户ID类型错误", nil))
		return
	}
	client, _ := h.clientService.GetClient(uint(id))
	clientName := ""
	if client != nil {
		clientName = client.Name
	}
	h.logService.CreateLogAsync(uid, "update_software", "client", uint(id),
		fmt.Sprintf("更新客户端软件: %s (类型: %s, 版本: %s)", clientName, req.UpdateType, req.Version), c.ClientIP())

	util.Success(c, gin.H{"message": "更新命令已发送"})
}

// BatchUpdateClientsSoftware godoc
// @Summary 批量更新客户端软件
// @Description 向多个客户端发送软件更新命令
// @Tags 客户端管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{client_ids=[]int,update_type=string,version=string,mirror_id=int} true "批量更新配置"
// @Success 200 {object} util.Response{data=object{message=string,success_count=int,failed_clients=[]int,total=int}}
// @Failure 400 {object} util.Response
// @Failure 401 {object} util.Response
// @Failure 500 {object} util.Response
// @Router /api/clients/batch-update-software [post]
func (h *ClientHandler) BatchUpdateClientsSoftware(c *gin.Context) {
	if h.clientUpdateService == nil {
		middleware.AbortWithAppError(c, errors.NewInternal("更新服务未初始化", nil))
		return
	}

	var req struct {
		ClientIDs  []uint `json:"client_ids" binding:"required"`
		UpdateType string `json:"update_type" binding:"required"`
		Version    string `json:"version"`
		MirrorID   *uint  `json:"mirror_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithAppError(c, errors.NewBadRequest("参数错误: "+err.Error()))
		return
	}

	if req.UpdateType != "frpc" && req.UpdateType != "daemon" {
		middleware.AbortWithAppError(c, errors.NewBadRequest("无效的更新类型，必须是 frpc 或 daemon"))
		return
	}

	batchReq := &service.BatchUpdateRequest{
		ClientIDs:  req.ClientIDs,
		UpdateType: service.UpdateType(req.UpdateType),
		Version:    req.Version,
		MirrorID:   req.MirrorID,
	}

	successCount, failedClients, err := h.clientUpdateService.BatchUpdateClients(batchReq)
	if err != nil {
		middleware.AbortWithAppError(c, errors.NewInternal("批量更新失败: "+err.Error(), err))
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		middleware.AbortWithAppError(c, errors.NewUnauthorized("用户未认证"))
		return
	}
	uid, ok := userID.(uint)
	if !ok {
		middleware.AbortWithAppError(c, errors.NewInternal("用户ID类型错误", nil))
		return
	}
	h.logService.CreateLogAsync(uid, "batch_update_software", "client", 0,
		fmt.Sprintf("批量更新客户端软件 (类型: %s, 版本: %s, 总数: %d, 成功: %d)", req.UpdateType, req.Version, len(req.ClientIDs), successCount), c.ClientIP())

	util.Success(c, gin.H{
		"message":        "批量更新命令已发送",
		"success_count":  successCount,
		"failed_clients": failedClients,
		"total":          len(req.ClientIDs),
	})
}
