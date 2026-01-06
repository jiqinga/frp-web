package handler

import (
	"fmt"
	apperrors "frp-web-panel/internal/errors"
	"frp-web-panel/internal/middleware"
	"frp-web-panel/internal/util"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Start godoc
// @Summary 启动本地 FRP 服务
// @Description 启动指定的本地 FRP 服务器进程
// @Tags FRP服务器
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "服务器ID"
// @Success 200 {object} util.Response{data=object{message=string}}
// @Failure 400 {object} util.Response
// @Failure 500 {object} util.Response
// @Router /api/frp-servers/{id}/start [post]
func (h *FrpServerHandler) Start(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.AbortWithAppError(c, apperrors.NewBadRequest("无效的ID参数"))
		return
	}
	if err := h.service.Start(uint(id)); err != nil {
		middleware.AbortWithAppError(c, apperrors.NewInternal("启动失败: "+err.Error(), err))
		return
	}

	userID, _ := c.Get("user_id")
	server, _ := h.service.GetByID(uint(id))
	serverName := ""
	if server != nil {
		serverName = server.Name
	}
	h.logService.CreateLogAsync(userID.(uint), "start", "frps", uint(id),
		fmt.Sprintf("启动FRP服务器: %s", serverName), c.ClientIP())

	util.SuccessResponse(c, gin.H{"message": "启动成功"})
}

// Stop godoc
// @Summary 停止本地 FRP 服务
// @Description 停止指定的本地 FRP 服务器进程
// @Tags FRP服务器
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "服务器ID"
// @Success 200 {object} util.Response{data=object{message=string}}
// @Failure 400 {object} util.Response
// @Failure 500 {object} util.Response
// @Router /api/frp-servers/{id}/stop [post]
func (h *FrpServerHandler) Stop(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.AbortWithAppError(c, apperrors.NewBadRequest("无效的ID参数"))
		return
	}
	if err := h.service.Stop(uint(id)); err != nil {
		middleware.AbortWithAppError(c, apperrors.NewInternal("停止失败: "+err.Error(), err))
		return
	}

	userID, _ := c.Get("user_id")
	server, _ := h.service.GetByID(uint(id))
	serverName := ""
	if server != nil {
		serverName = server.Name
	}
	h.logService.CreateLogAsync(userID.(uint), "stop", "frps", uint(id),
		fmt.Sprintf("停止FRP服务器: %s", serverName), c.ClientIP())

	util.SuccessResponse(c, gin.H{"message": "停止成功"})
}

// Restart godoc
// @Summary 重启本地 FRP 服务
// @Description 重启指定的本地 FRP 服务器进程
// @Tags FRP服务器
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "服务器ID"
// @Success 200 {object} util.Response{data=object{message=string}}
// @Failure 400 {object} util.Response
// @Failure 500 {object} util.Response
// @Router /api/frp-servers/{id}/restart [post]
func (h *FrpServerHandler) Restart(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.AbortWithAppError(c, apperrors.NewBadRequest("无效的ID参数"))
		return
	}
	if err := h.service.Restart(uint(id)); err != nil {
		middleware.AbortWithAppError(c, apperrors.NewInternal("重启失败: "+err.Error(), err))
		return
	}

	userID, _ := c.Get("user_id")
	server, _ := h.service.GetByID(uint(id))
	serverName := ""
	if server != nil {
		serverName = server.Name
	}
	h.logService.CreateLogAsync(userID.(uint), "restart", "frps", uint(id),
		fmt.Sprintf("重启FRP服务器: %s", serverName), c.ClientIP())

	util.SuccessResponse(c, gin.H{"message": "重启成功"})
}

// GetStatus godoc
// @Summary 获取服务运行状态
// @Description 获取指定 FRP 服务器的运行状态
// @Tags FRP服务器
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "服务器ID"
// @Success 200 {object} util.Response{data=object{status=string}}
// @Failure 400 {object} util.Response
// @Failure 500 {object} util.Response
// @Router /api/frp-servers/{id}/status [get]
func (h *FrpServerHandler) GetStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.AbortWithAppError(c, apperrors.NewBadRequest("无效的ID参数"))
		return
	}
	status, err := h.service.GetStatus(uint(id))
	if err != nil {
		middleware.AbortWithAppError(c, apperrors.NewInternal("获取状态失败", err))
		return
	}
	util.SuccessResponse(c, gin.H{"status": status})
}
