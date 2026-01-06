package handler

import (
	"fmt"
	apperrors "frp-web-panel/internal/errors"
	"frp-web-panel/internal/middleware"
	"frp-web-panel/internal/util"
	"strconv"

	"github.com/gin-gonic/gin"
)

// TestSSH godoc
// @Summary 测试 SSH 连接
// @Description 测试与远程服务器的 SSH 连接是否正常
// @Tags FRP服务器
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "服务器ID"
// @Success 200 {object} util.Response{data=object{message=string}}
// @Failure 400 {object} util.Response
// @Router /api/frp-servers/{id}/test-ssh [post]
func (h *FrpServerHandler) TestSSH(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.AbortWithAppError(c, apperrors.NewBadRequest("无效的ID参数"))
		return
	}
	if err := h.service.TestSSH(uint(id)); err != nil {
		middleware.AbortWithAppError(c, apperrors.NewBadRequest("SSH连接失败: "+err.Error()))
		return
	}
	util.SuccessResponse(c, gin.H{"message": "SSH连接成功"})
}

// RemoteInstall godoc
// @Summary 远程安装 frps
// @Description 通过 SSH 在远程服务器上安装 frps
// @Tags FRP服务器
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "服务器ID"
// @Param request body object{mirror_id=int} false "安装参数"
// @Success 200 {object} util.Response{data=object{message=string}}
// @Failure 400 {object} util.Response
// @Router /api/frp-servers/{id}/remote-install [post]
func (h *FrpServerHandler) RemoteInstall(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.AbortWithAppError(c, apperrors.NewBadRequest("无效的ID参数"))
		return
	}
	var req struct {
		MirrorID *uint `json:"mirror_id"`
	}
	c.ShouldBindJSON(&req)

	userID, _ := c.Get("user_id")
	server, _ := h.service.GetByID(uint(id))
	serverName := ""
	if server != nil {
		serverName = server.Name
	}
	h.logService.CreateLogAsync(userID.(uint), "remote_install", "frps", uint(id),
		fmt.Sprintf("远程安装FRP服务器: %s", serverName), c.ClientIP())

	go func() {
		h.service.RemoteInstall(uint(id), req.MirrorID)
	}()
	util.SuccessResponse(c, gin.H{"message": "安装任务已启动"})
}

// RemoteStart godoc
// @Summary 远程启动 frps
// @Description 通过 SSH 在远程服务器上启动 frps 服务
// @Tags FRP服务器
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "服务器ID"
// @Success 200 {object} util.Response{data=object{message=string}}
// @Failure 400 {object} util.Response
// @Failure 500 {object} util.Response
// @Router /api/frp-servers/{id}/remote-start [post]
func (h *FrpServerHandler) RemoteStart(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.AbortWithAppError(c, apperrors.NewBadRequest("无效的ID参数"))
		return
	}
	if err := h.service.RemoteStart(uint(id)); err != nil {
		middleware.AbortWithAppError(c, apperrors.NewInternal("远程启动失败: "+err.Error(), err))
		return
	}

	userID, _ := c.Get("user_id")
	server, _ := h.service.GetByID(uint(id))
	serverName := ""
	if server != nil {
		serverName = server.Name
	}
	h.logService.CreateLogAsync(userID.(uint), "remote_start", "frps", uint(id),
		fmt.Sprintf("远程启动FRP服务器: %s", serverName), c.ClientIP())

	util.SuccessResponse(c, gin.H{"message": "远程启动成功"})
}

// RemoteStop godoc
// @Summary 远程停止 frps
// @Description 通过 SSH 在远程服务器上停止 frps 服务
// @Tags FRP服务器
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "服务器ID"
// @Success 200 {object} util.Response{data=object{message=string}}
// @Failure 400 {object} util.Response
// @Failure 500 {object} util.Response
// @Router /api/frp-servers/{id}/remote-stop [post]
func (h *FrpServerHandler) RemoteStop(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.AbortWithAppError(c, apperrors.NewBadRequest("无效的ID参数"))
		return
	}
	if err := h.service.RemoteStop(uint(id)); err != nil {
		middleware.AbortWithAppError(c, apperrors.NewInternal("远程停止失败: "+err.Error(), err))
		return
	}

	userID, _ := c.Get("user_id")
	server, _ := h.service.GetByID(uint(id))
	serverName := ""
	if server != nil {
		serverName = server.Name
	}
	h.logService.CreateLogAsync(userID.(uint), "remote_stop", "frps", uint(id),
		fmt.Sprintf("远程停止FRP服务器: %s", serverName), c.ClientIP())

	util.SuccessResponse(c, gin.H{"message": "远程停止成功"})
}

// RemoteRestart godoc
// @Summary 远程重启 frps
// @Description 通过 SSH 在远程服务器上重启 frps 服务
// @Tags FRP服务器
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "服务器ID"
// @Success 200 {object} util.Response{data=object{message=string}}
// @Failure 400 {object} util.Response
// @Failure 500 {object} util.Response
// @Router /api/frp-servers/{id}/remote-restart [post]
func (h *FrpServerHandler) RemoteRestart(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.AbortWithAppError(c, apperrors.NewBadRequest("无效的ID参数"))
		return
	}
	if err := h.service.RemoteRestart(uint(id)); err != nil {
		middleware.AbortWithAppError(c, apperrors.NewInternal("远程重启失败: "+err.Error(), err))
		return
	}

	userID, _ := c.Get("user_id")
	server, _ := h.service.GetByID(uint(id))
	serverName := ""
	if server != nil {
		serverName = server.Name
	}
	h.logService.CreateLogAsync(userID.(uint), "remote_restart", "frps", uint(id),
		fmt.Sprintf("远程重启FRP服务器: %s", serverName), c.ClientIP())

	util.SuccessResponse(c, gin.H{"message": "远程重启成功"})
}

// RemoteUninstall godoc
// @Summary 远程卸载 frps
// @Description 通过 SSH 在远程服务器上卸载 frps
// @Tags FRP服务器
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "服务器ID"
// @Success 200 {object} util.Response{data=object{message=string}}
// @Failure 400 {object} util.Response
// @Failure 500 {object} util.Response
// @Router /api/frp-servers/{id}/remote-uninstall [post]
func (h *FrpServerHandler) RemoteUninstall(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.AbortWithAppError(c, apperrors.NewBadRequest("无效的ID参数"))
		return
	}

	server, _ := h.service.GetByID(uint(id))
	serverName := ""
	if server != nil {
		serverName = server.Name
	}

	if err := h.service.RemoteUninstall(uint(id)); err != nil {
		middleware.AbortWithAppError(c, apperrors.NewInternal("远程卸载失败: "+err.Error(), err))
		return
	}

	userID, _ := c.Get("user_id")
	h.logService.CreateLogAsync(userID.(uint), "remote_uninstall", "frps", uint(id),
		fmt.Sprintf("远程卸载FRP服务器: %s", serverName), c.ClientIP())

	util.SuccessResponse(c, gin.H{"message": "远程卸载成功"})
}

// RemoteGetLogs godoc
// @Summary 获取远程 frps 日志
// @Description 通过 SSH 获取远程服务器上的 frps 运行日志
// @Tags FRP服务器
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "服务器ID"
// @Param lines query int false "日志行数" default(100)
// @Success 200 {object} util.Response{data=object{logs=string}}
// @Failure 400 {object} util.Response
// @Failure 500 {object} util.Response
// @Router /api/frp-servers/{id}/remote-logs [get]
func (h *FrpServerHandler) RemoteGetLogs(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.AbortWithAppError(c, apperrors.NewBadRequest("无效的ID参数"))
		return
	}
	lines := 100
	if l := c.Query("lines"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			lines = parsed
		}
	}

	logs, err := h.service.RemoteGetLogs(uint(id), lines)
	if err != nil {
		middleware.AbortWithAppError(c, apperrors.NewInternal("获取日志失败: "+err.Error(), err))
		return
	}
	util.SuccessResponse(c, gin.H{"logs": logs})
}

// RemoteGetVersion godoc
// @Summary 获取远程 frps 版本
// @Description 通过 SSH 获取远程服务器上安装的 frps 版本
// @Tags FRP服务器
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "服务器ID"
// @Success 200 {object} util.Response{data=object{version=string}}
// @Failure 400 {object} util.Response
// @Failure 500 {object} util.Response
// @Router /api/frp-servers/{id}/remote-version [get]
func (h *FrpServerHandler) RemoteGetVersion(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.AbortWithAppError(c, apperrors.NewBadRequest("无效的ID参数"))
		return
	}
	version, err := h.service.RemoteGetVersion(uint(id))
	if err != nil {
		middleware.AbortWithAppError(c, apperrors.NewInternal("获取版本失败: "+err.Error(), err))
		return
	}
	util.SuccessResponse(c, gin.H{"version": version})
}

// RemoteReinstall godoc
// @Summary 远程重装 frps
// @Description 通过 SSH 在远程服务器上重新安装 frps
// @Tags FRP服务器
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "服务器ID"
// @Param request body object{regenerate_auth=bool,mirror_id=int} false "重装参数"
// @Success 200 {object} util.Response{data=object{message=string}}
// @Failure 400 {object} util.Response
// @Router /api/frp-servers/{id}/remote-reinstall [post]
func (h *FrpServerHandler) RemoteReinstall(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.AbortWithAppError(c, apperrors.NewBadRequest("无效的ID参数"))
		return
	}
	var req struct {
		RegenerateAuth bool  `json:"regenerate_auth"`
		MirrorID       *uint `json:"mirror_id"`
	}
	c.ShouldBindJSON(&req)

	userID, _ := c.Get("user_id")
	server, _ := h.service.GetByID(uint(id))
	serverName := ""
	if server != nil {
		serverName = server.Name
	}
	h.logService.CreateLogAsync(userID.(uint), "remote_reinstall", "frps", uint(id),
		fmt.Sprintf("远程重装FRP服务器: %s (重新生成认证: %v)", serverName, req.RegenerateAuth), c.ClientIP())

	go func() {
		h.service.RemoteReinstall(uint(id), req.RegenerateAuth, req.MirrorID)
	}()
	util.SuccessResponse(c, gin.H{"message": "重装任务已启动"})
}

// RemoteUpgrade godoc
// @Summary 远程升级 frps
// @Description 通过 SSH 在远程服务器上升级 frps 到指定版本
// @Tags FRP服务器
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "服务器ID"
// @Param request body object{version=string,mirror_id=int} false "升级参数"
// @Success 200 {object} util.Response{data=object{message=string}}
// @Failure 400 {object} util.Response
// @Router /api/frp-servers/{id}/remote-upgrade [post]
func (h *FrpServerHandler) RemoteUpgrade(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.AbortWithAppError(c, apperrors.NewBadRequest("无效的ID参数"))
		return
	}
	var req struct {
		Version  string `json:"version"`
		MirrorID *uint  `json:"mirror_id"`
	}
	c.ShouldBindJSON(&req)

	userID, _ := c.Get("user_id")
	server, _ := h.service.GetByID(uint(id))
	serverName := ""
	if server != nil {
		serverName = server.Name
	}
	h.logService.CreateLogAsync(userID.(uint), "remote_upgrade", "frps", uint(id),
		fmt.Sprintf("远程升级FRP服务器: %s (目标版本: %s)", serverName, req.Version), c.ClientIP())

	go func() {
		h.service.RemoteUpgrade(uint(id), req.Version, req.MirrorID)
	}()
	util.SuccessResponse(c, gin.H{"message": "升级任务已启动"})
}
