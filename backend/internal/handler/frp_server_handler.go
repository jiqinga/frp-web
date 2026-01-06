package handler

import (
	"fmt"
	apperrors "frp-web-panel/internal/errors"
	"frp-web-panel/internal/middleware"
	"frp-web-panel/internal/model"
	"frp-web-panel/internal/repository"
	"frp-web-panel/internal/service"
	"frp-web-panel/internal/util"
	"strconv"

	"github.com/gin-gonic/gin"
)

type FrpServerHandler struct {
	service     *service.FrpServerService
	logService  *service.LogService
	metricsRepo *repository.ServerMetricsRepository
}

func NewFrpServerHandler(svc *service.FrpServerService, logSvc *service.LogService, metricsRepo *repository.ServerMetricsRepository) *FrpServerHandler {
	return &FrpServerHandler{
		service:     svc,
		logService:  logSvc,
		metricsRepo: metricsRepo,
	}
}

// GetAll godoc
// @Summary 获取 FRP 服务器列表
// @Description 获取所有 FRP 服务器的列表
// @Tags FRP服务器
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} util.Response{data=[]object}
// @Failure 500 {object} util.Response
// @Router /api/frp-servers [get]
func (h *FrpServerHandler) GetAll(c *gin.Context) {
	servers, err := h.service.GetAll()
	if err != nil {
		middleware.AbortWithAppError(c, apperrors.NewInternal("获取服务器列表失败", err))
		return
	}
	util.SuccessResponse(c, servers)
}

// GetByID godoc
// @Summary 获取单个 FRP 服务器
// @Description 根据 ID 获取 FRP 服务器详情
// @Tags FRP服务器
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "服务器ID"
// @Success 200 {object} util.Response{data=object}
// @Failure 400 {object} util.Response
// @Failure 404 {object} util.Response
// @Router /api/frp-servers/{id} [get]
func (h *FrpServerHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.AbortWithAppError(c, apperrors.NewBadRequest("无效的ID参数"))
		return
	}
	server, err := h.service.GetByID(uint(id))
	if err != nil {
		middleware.AbortWithAppError(c, apperrors.NewNotFound("服务器不存在"))
		return
	}
	util.SuccessResponse(c, server)
}

// Create godoc
// @Summary 创建 FRP 服务器
// @Description 创建新的 FRP 服务器配置
// @Tags FRP服务器
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param server body model.FrpServer true "服务器配置"
// @Success 200 {object} util.Response{data=object}
// @Failure 400 {object} util.Response
// @Failure 500 {object} util.Response
// @Router /api/frp-servers [post]
func (h *FrpServerHandler) Create(c *gin.Context) {
	var server model.FrpServer
	if err := c.ShouldBindJSON(&server); err != nil {
		middleware.AbortWithAppError(c, apperrors.NewBadRequest("参数错误"))
		return
	}

	if server.Token == "" {
		token, err := util.GenerateRandomToken(48)
		if err != nil {
			middleware.AbortWithAppError(c, apperrors.NewInternal("生成token失败", err))
			return
		}
		server.Token = token
	}

	if err := h.service.Create(&server); err != nil {
		middleware.AbortWithAppError(c, apperrors.NewInternal("创建服务器失败", err))
		return
	}

	userID, _ := c.Get("user_id")
	h.logService.CreateLogAsync(userID.(uint), "create", "frps", server.ID,
		fmt.Sprintf("创建FRP服务器: %s (%s:%d)", server.Name, server.Host, server.BindPort), c.ClientIP())

	util.SuccessResponse(c, server)
}

// Update godoc
// @Summary 更新 FRP 服务器
// @Description 更新指定 FRP 服务器的配置
// @Tags FRP服务器
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "服务器ID"
// @Param server body model.FrpServer true "服务器配置"
// @Success 200 {object} util.Response{data=object}
// @Failure 400 {object} util.Response
// @Failure 500 {object} util.Response
// @Router /api/frp-servers/{id} [put]
func (h *FrpServerHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.AbortWithAppError(c, apperrors.NewBadRequest("无效的ID参数"))
		return
	}
	var server model.FrpServer
	if err := c.ShouldBindJSON(&server); err != nil {
		middleware.AbortWithAppError(c, apperrors.NewBadRequest("参数错误"))
		return
	}

	server.ID = uint(id)
	if err := h.service.Update(&server); err != nil {
		middleware.AbortWithAppError(c, apperrors.NewInternal("更新服务器失败", err))
		return
	}

	userID, _ := c.Get("user_id")
	h.logService.CreateLogAsync(userID.(uint), "update", "frps", server.ID,
		fmt.Sprintf("更新FRP服务器: %s (%s:%d)", server.Name, server.Host, server.BindPort), c.ClientIP())

	util.SuccessResponse(c, server)
}

// Delete godoc
// @Summary 删除 FRP 服务器
// @Description 删除指定的 FRP 服务器
// @Tags FRP服务器
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "服务器ID"
// @Param remove_installation query bool false "是否同时删除远程安装"
// @Success 200 {object} util.Response
// @Failure 400 {object} util.Response
// @Failure 500 {object} util.Response
// @Router /api/frp-servers/{id} [delete]
func (h *FrpServerHandler) Delete(c *gin.Context) {
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

	removeInstallation := c.Query("remove_installation") == "true"
	if err := h.service.Delete(uint(id), removeInstallation); err != nil {
		middleware.AbortWithAppError(c, apperrors.NewInternal("删除服务器失败", err))
		return
	}

	userID, _ := c.Get("user_id")
	h.logService.CreateLogAsync(userID.(uint), "delete", "frps", uint(id),
		fmt.Sprintf("删除FRP服务器: %s (ID: %d)", serverName, id), c.ClientIP())

	util.SuccessResponse(c, nil)
}

// TestConnection godoc
// @Summary 测试服务器连接
// @Description 测试 FRP 服务器的 Dashboard 连接是否正常
// @Tags FRP服务器
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param server body model.FrpServer true "服务器配置"
// @Success 200 {object} util.Response{data=object{message=string}}
// @Failure 400 {object} util.Response
// @Router /api/frp-servers/test [post]
func (h *FrpServerHandler) TestConnection(c *gin.Context) {
	var server model.FrpServer
	if err := c.ShouldBindJSON(&server); err != nil {
		middleware.AbortWithAppError(c, apperrors.NewBadRequest("参数错误"))
		return
	}

	if err := h.service.TestConnection(&server); err != nil {
		middleware.AbortWithAppError(c, apperrors.NewBadRequest("连接失败: "+err.Error()))
		return
	}
	util.SuccessResponse(c, gin.H{"message": "连接成功"})
}

// ParseConfig godoc
// @Summary 解析服务器配置
// @Description 解析 FRP 服务器配置文件内容
// @Tags FRP服务器
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object true "配置内容"
// @Success 200 {object} util.Response{data=object}
// @Failure 400 {object} util.Response
// @Router /api/frp-servers/parse-config [post]
func (h *FrpServerHandler) ParseConfig(c *gin.Context) {
	var req struct {
		Config string `json:"config"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithAppError(c, apperrors.NewBadRequest("参数错误"))
		return
	}
	result, err := util.ParseFrpsConfig(req.Config)
	if err != nil {
		middleware.AbortWithAppError(c, apperrors.NewBadRequest("解析配置失败: "+err.Error()))
		return
	}
	util.SuccessResponse(c, result)
}

// Download godoc
// @Summary 下载 FRP 服务端
// @Description 下载指定版本的 FRP 服务端程序
// @Tags FRP服务器
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "服务器ID"
// @Param request body object true "下载参数"
// @Success 200 {object} util.Response
// @Failure 400 {object} util.Response
// @Failure 500 {object} util.Response
// @Router /api/frp-servers/{id}/download [post]
func (h *FrpServerHandler) Download(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.AbortWithAppError(c, apperrors.NewBadRequest("无效的ID参数"))
		return
	}
	var req struct {
		Version string `json:"version"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithAppError(c, apperrors.NewBadRequest("参数错误"))
		return
	}
	if err := h.service.Download(uint(id), req.Version); err != nil {
		middleware.AbortWithAppError(c, apperrors.NewInternal("下载失败", err))
		return
	}
	util.SuccessResponse(c, gin.H{"message": "下载任务已启动"})
}

// GetLocalVersion godoc
// @Summary 获取本地 FRP 版本
// @Description 获取本地已下载的 FRP 服务端版本
// @Tags FRP服务器
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "服务器ID"
// @Success 200 {object} util.Response{data=object}
// @Failure 400 {object} util.Response
// @Router /api/frp-servers/{id}/local-version [get]
func (h *FrpServerHandler) GetLocalVersion(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.AbortWithAppError(c, apperrors.NewBadRequest("无效的ID参数"))
		return
	}
	version, err := h.service.GetLocalVersion(uint(id))
	if err != nil {
		middleware.AbortWithAppError(c, apperrors.NewInternal("获取版本失败", err))
		return
	}
	util.SuccessResponse(c, gin.H{"version": version})
}

// GetRunningTask godoc
// @Summary 获取运行中的任务
// @Description 获取指定服务器当前运行中的任务状态
// @Tags FRP服务器
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "服务器ID"
// @Success 200 {object} util.Response{data=object}
// @Failure 400 {object} util.Response
// @Router /api/frp-servers/{id}/running-task [get]
func (h *FrpServerHandler) GetRunningTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.AbortWithAppError(c, apperrors.NewBadRequest("无效的ID参数"))
		return
	}
	operation, exists := h.service.GetRunningTask(uint(id))
	util.SuccessResponse(c, gin.H{"operation": operation, "exists": exists})
}
