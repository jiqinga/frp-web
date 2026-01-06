/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-14 15:32:45
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-30 16:44:44
 * @FilePath            : frp-web-testbackendinternalhandlerauth_handler.go
 * @Description         : 认证处理器
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package handler

import (
	"fmt"
	"frp-web-panel/internal/errors"
	"frp-web-panel/internal/middleware"
	"frp-web-panel/internal/service"
	"frp-web-panel/internal/util"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
	logService  *service.LogService
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		authService: service.NewAuthService(),
		logService:  service.NewLogService(),
	}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"admin"`
	Password string `json:"password" binding:"required" example:"admin123"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required" example:"oldpass123"`
	NewPassword string `json:"new_password" binding:"required,min=6" example:"newpass123"`
}

// Login godoc
// @Summary 用户登录
// @Description 用户登录接口，验证用户名和密码后返回JWT Token
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录信息"
// @Success 200 {object} util.Response{data=object} "token和用户信息"
// @Failure 400 {object} util.Response "参数错误"
// @Failure 401 {object} util.Response "用户名或密码错误"
// @Router /api/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithAppError(c, errors.NewBadRequest("参数错误"))
		return
	}

	token, user, err := h.authService.Login(req.Username, req.Password, c.ClientIP())
	if err != nil {
		h.logService.CreateLogAsync(0, "login_failed", "user", 0,
			fmt.Sprintf("用户 %s 登录失败: %s", req.Username, err.Error()), c.ClientIP())
		middleware.AbortWithAppError(c, errors.NewUnauthorized(err.Error()))
		return
	}

	h.logService.CreateLogAsync(user.ID, "login", "user", user.ID,
		fmt.Sprintf("用户 %s 登录成功", user.Username), c.ClientIP())

	util.Success(c, gin.H{
		"token": token,
		"user":  user,
	})
}

// GetProfile godoc
// @Summary 获取用户信息
// @Description 获取当前登录用户的详细信息
// @Tags 认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} util.Response{data=object} "用户信息"
// @Failure 404 {object} util.Response "用户不存在"
// @Router /api/auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := c.GetUint("user_id")
	user, err := h.authService.GetProfile(userID)
	if err != nil {
		middleware.AbortWithAppError(c, errors.NewNotFound("用户不存在"))
		return
	}

	util.Success(c, user)
}

// ChangePassword godoc
// @Summary 修改密码
// @Description 修改当前登录用户的密码，需要提供旧密码验证
// @Tags 认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ChangePasswordRequest true "密码修改信息"
// @Success 200 {object} util.Response{data=object} "密码修改成功"
// @Failure 400 {object} util.Response "参数错误或旧密码错误"
// @Router /api/auth/password [put]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithAppError(c, errors.NewValidation("参数错误：新密码至少需要6个字符"))
		return
	}

	userID := c.GetUint("user_id")
	if err := h.authService.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		middleware.AbortWithAppError(c, errors.NewBadRequest(err.Error()))
		return
	}

	h.logService.CreateLogAsync(userID, "change_password", "user", userID,
		"用户修改密码", c.ClientIP())

	util.Success(c, gin.H{"message": "密码修改成功"})
}
