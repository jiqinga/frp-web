package handler

import (
	"fmt"
	"frp-web-panel/internal/errors"
	"frp-web-panel/internal/middleware"
	"frp-web-panel/internal/model"
	"frp-web-panel/internal/util"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GenerateRegisterToken godoc
// @Summary 生成注册令牌
// @Description 生成客户端注册令牌，用于自动注册客户端
// @Tags 客户端管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.ClientRegisterToken true "注册令牌配置"
// @Success 200 {object} util.Response{data=object}
// @Failure 400 {object} util.Response
// @Failure 401 {object} util.Response
// @Failure 500 {object} util.Response
// @Router /api/clients/register/token [post]
func (h *ClientHandler) GenerateRegisterToken(c *gin.Context) {
	var req model.ClientRegisterToken
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithAppError(c, errors.NewBadRequest("参数错误"))
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
	token, err := h.clientRegisterService.GenerateToken(&req, uid)
	if err != nil {
		middleware.AbortWithAppError(c, errors.NewInternal("生成Token失败", err))
		return
	}

	h.logService.CreateLogAsync(uid, "generate_token", "client_register_token", token.ID,
		fmt.Sprintf("生成客户端注册Token: %s (客户端名称: %s)", token.Token[:8]+"...", req.ClientName), c.ClientIP())

	util.Success(c, token)
}

// GenerateRegisterScript godoc
// @Summary 生成注册脚本
// @Description 根据令牌生成客户端安装注册脚本
// @Tags 客户端管理
// @Accept json
// @Produce plain
// @Security BearerAuth
// @Param token query string true "注册令牌"
// @Param type query string false "脚本类型" default(bash) Enums(bash, powershell)
// @Param mirror query int true "镜像ID"
// @Success 200 {string} string "安装脚本内容"
// @Failure 400 {object} util.Response
// @Router /api/clients/register/script [get]
func (h *ClientHandler) GenerateRegisterScript(c *gin.Context) {
	token := c.Query("token")
	scriptType := c.DefaultQuery("type", "bash")
	mirrorIDStr := c.Query("mirror")

	mirrorID, err := strconv.ParseUint(mirrorIDStr, 10, 32)
	if err != nil {
		middleware.AbortWithAppError(c, errors.NewBadRequest("无效的镜像ID"))
		return
	}

	script, err := h.clientRegisterService.GenerateScript(token, scriptType, uint(mirrorID))
	if err != nil {
		middleware.AbortWithAppError(c, errors.NewBadRequest(err.Error()))
		return
	}

	c.Header("Content-Type", "text/plain")
	c.String(200, script)
}

// RegisterClient godoc
// @Summary 客户端注册
// @Description 客户端使用令牌进行自动注册（无需认证）
// @Tags 客户端管理
// @Accept json
// @Produce json
// @Param request body object{token=string} true "注册令牌"
// @Success 200 {object} util.Response{data=object}
// @Failure 400 {object} util.Response
// @Router /api/clients/register [post]
func (h *ClientHandler) RegisterClient(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithAppError(c, errors.NewBadRequest("参数错误"))
		return
	}

	client, err := h.clientRegisterService.RegisterClient(req.Token)
	if err != nil {
		middleware.AbortWithAppError(c, errors.NewBadRequest(err.Error()))
		return
	}

	util.Success(c, client)
}

// GetInstallScript godoc
// @Summary 获取安装脚本
// @Description 根据令牌获取客户端安装脚本（公开接口）
// @Tags 客户端管理
// @Accept json
// @Produce plain
// @Param token path string true "注册令牌"
// @Param type query string false "脚本类型" default(bash) Enums(bash, powershell)
// @Param mirror query int false "镜像ID" default(1)
// @Success 200 {string} string "安装脚本内容"
// @Failure 400 {string} string "错误信息"
// @Router /install/{token} [get]
func (h *ClientHandler) GetInstallScript(c *gin.Context) {
	token := c.Param("token")
	scriptType := c.DefaultQuery("type", "bash")
	mirrorIDStr := c.DefaultQuery("mirror", "1")

	mirrorID, err := strconv.ParseUint(mirrorIDStr, 10, 32)
	if err != nil {
		c.String(400, "无效的镜像ID")
		return
	}

	script, err := h.clientRegisterService.GetInstallScript(token, scriptType, uint(mirrorID))
	if err != nil {
		c.String(400, err.Error())
		return
	}

	c.Header("Content-Type", "text/plain")
	c.String(200, script)
}

// ParseConfig godoc
// @Summary 解析客户端配置
// @Description 解析 frpc 配置文件内容
// @Tags 客户端管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{config=string} true "配置内容"
// @Success 200 {object} util.Response{data=object}
// @Failure 400 {object} util.Response
// @Router /api/clients/parse-config [post]
func (h *ClientHandler) ParseConfig(c *gin.Context) {
	var req struct {
		Config string `json:"config" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithAppError(c, errors.NewBadRequest("参数错误"))
		return
	}

	result, err := util.ParseFrpcConfig(req.Config)
	if err != nil {
		middleware.AbortWithAppError(c, errors.NewBadRequest(err.Error()))
		return
	}

	util.Success(c, result)
}
