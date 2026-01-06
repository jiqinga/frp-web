package handler

import (
	"fmt"
	"frp-web-panel/internal/errors"
	"frp-web-panel/internal/middleware"
	"frp-web-panel/internal/model"
	"frp-web-panel/internal/service"
	"frp-web-panel/internal/util"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ClientHandler struct {
	clientService         *service.ClientService
	clientRegisterService *service.ClientRegisterService
	clientUpdateService   *service.ClientUpdateService
	logService            *service.LogService
}

func NewClientHandler(clientSvc *service.ClientService, registerSvc *service.ClientRegisterService, updateSvc *service.ClientUpdateService, logSvc *service.LogService) *ClientHandler {
	return &ClientHandler{
		clientService:         clientSvc,
		clientRegisterService: registerSvc,
		clientUpdateService:   updateSvc,
		logService:            logSvc,
	}
}

// GetClients godoc
// @Summary 获取客户端列表
// @Description 分页获取客户端列表，支持关键词搜索
// @Tags 客户端管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param keyword query string false "搜索关键词"
// @Success 200 {object} util.Response{data=object{list=[]object,total=int64}}
// @Failure 500 {object} util.Response
// @Router /api/clients [get]
func (h *ClientHandler) GetClients(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	keyword := c.Query("keyword")

	clients, total, err := h.clientService.GetClients(page, pageSize, keyword)
	if err != nil {
		middleware.AbortWithAppError(c, errors.NewInternal("获取客户端列表失败", err))
		return
	}

	util.Success(c, gin.H{
		"list":  clients,
		"total": total,
	})
}

// GetClient godoc
// @Summary 获取客户端详情
// @Description 根据ID获取单个客户端的详细信息
// @Tags 客户端管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "客户端ID"
// @Success 200 {object} util.Response{data=object}
// @Failure 400 {object} util.Response
// @Failure 404 {object} util.Response
// @Router /api/clients/{id} [get]
func (h *ClientHandler) GetClient(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.AbortWithAppError(c, errors.NewBadRequest("无效的客户端ID"))
		return
	}
	client, err := h.clientService.GetClient(uint(id))
	if err != nil {
		middleware.AbortWithAppError(c, errors.NewNotFound("客户端不存在"))
		return
	}

	util.Success(c, client)
}

// CreateClient godoc
// @Summary 创建客户端
// @Description 创建新的客户端
// @Tags 客户端管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param client body model.Client true "客户端信息"
// @Success 200 {object} util.Response{data=object}
// @Failure 400 {object} util.Response
// @Failure 401 {object} util.Response
// @Failure 500 {object} util.Response
// @Router /api/clients [post]
func (h *ClientHandler) CreateClient(c *gin.Context) {
	var client model.Client
	if err := c.ShouldBindJSON(&client); err != nil {
		middleware.AbortWithAppError(c, errors.NewBadRequest("参数错误"))
		return
	}

	if err := h.clientService.CreateClient(&client); err != nil {
		middleware.AbortWithAppError(c, errors.NewInternal("创建客户端失败", err))
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
	h.logService.CreateLogAsync(uid, "create", "client", client.ID,
		fmt.Sprintf("创建客户端: %s", client.Name), c.ClientIP())

	util.Success(c, client)
}

// UpdateClient godoc
// @Summary 更新客户端
// @Description 更新指定客户端的信息
// @Tags 客户端管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "客户端ID"
// @Param client body model.Client true "客户端信息"
// @Success 200 {object} util.Response{data=object}
// @Failure 400 {object} util.Response
// @Failure 401 {object} util.Response
// @Failure 500 {object} util.Response
// @Router /api/clients/{id} [put]
func (h *ClientHandler) UpdateClient(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.AbortWithAppError(c, errors.NewBadRequest("无效的客户端ID"))
		return
	}
	var client model.Client
	if err := c.ShouldBindJSON(&client); err != nil {
		middleware.AbortWithAppError(c, errors.NewBadRequest("参数错误"))
		return
	}

	client.ID = uint(id)
	if err := h.clientService.UpdateClient(&client); err != nil {
		middleware.AbortWithAppError(c, errors.NewInternal("更新客户端失败", err))
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
	h.logService.CreateLogAsync(uid, "update", "client", client.ID,
		fmt.Sprintf("更新客户端: %s", client.Name), c.ClientIP())

	util.Success(c, client)
}

// DeleteClient godoc
// @Summary 删除客户端
// @Description 删除指定的客户端
// @Tags 客户端管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "客户端ID"
// @Success 200 {object} util.Response
// @Failure 400 {object} util.Response
// @Failure 401 {object} util.Response
// @Failure 500 {object} util.Response
// @Router /api/clients/{id} [delete]
func (h *ClientHandler) DeleteClient(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.AbortWithAppError(c, errors.NewBadRequest("无效的客户端ID"))
		return
	}

	client, _ := h.clientService.GetClient(uint(id))
	clientName := ""
	if client != nil {
		clientName = client.Name
	}

	if err := h.clientService.DeleteClient(uint(id)); err != nil {
		middleware.AbortWithAppError(c, errors.NewInternal("删除客户端失败", err))
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
	h.logService.CreateLogAsync(uid, "delete", "client", uint(id),
		fmt.Sprintf("删除客户端: %s (ID: %d)", clientName, id), c.ClientIP())

	util.Success(c, nil)
}

// GetProxiesByClient 获取客户端的代理列表
// 注意：此方法的 Swagger 文档在 ProxyHandler 中定义，避免重复声明
func (h *ClientHandler) GetProxiesByClient(c *gin.Context) {
	// This method is implemented in ProxyHandler
}

// GetClientVersions godoc
// @Summary 获取客户端版本列表
// @Description 获取指定客户端的历史版本信息
// @Tags 客户端管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "客户端ID"
// @Success 200 {object} util.Response{data=[]object}
// @Failure 400 {object} util.Response
// @Router /api/clients/{id}/versions [get]
func (h *ClientHandler) GetClientVersions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.AbortWithAppError(c, errors.NewBadRequest("无效的客户端ID"))
		return
	}
	versions, _, _, _, err := h.clientUpdateService.GetClientVersions(uint(id))
	if err != nil {
		middleware.AbortWithAppError(c, errors.NewInternal("获取版本列表失败", err))
		return
	}
	util.Success(c, versions)
}

// BatchUpdateClients godoc
// @Summary 批量更新客户端
// @Description 批量更新多个客户端的信息
// @Tags 客户端管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object true "批量更新请求"
// @Success 200 {object} util.Response
// @Failure 400 {object} util.Response
// @Router /api/clients/batch-update [post]
func (h *ClientHandler) BatchUpdateClients(c *gin.Context) {
	var req struct {
		IDs    []uint `json:"ids"`
		Action string `json:"action"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithAppError(c, errors.NewBadRequest("参数错误"))
		return
	}
	// TODO: Implement batch update logic
	util.Success(c, nil)
}

// GetOnlineClients godoc
// @Summary 获取在线客户端
// @Description 获取当前在线的客户端ID列表
// @Tags 客户端管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} util.Response{data=object{online_client_ids=[]int,count=int}}
// @Failure 500 {object} util.Response
// @Router /api/clients/online [get]
func (h *ClientHandler) GetOnlineClients(c *gin.Context) {
	onlineIDs, err := h.clientUpdateService.GetOnlineClients()
	if err != nil {
		middleware.AbortWithAppError(c, errors.NewInternal("获取在线客户端失败", err))
		return
	}
	util.Success(c, gin.H{
		"online_client_ids": onlineIDs,
		"count":             len(onlineIDs),
	})
}
