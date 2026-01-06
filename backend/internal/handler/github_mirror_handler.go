/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-21 14:02:27
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-30 16:51:55
 * @FilePath            : frp-web-testbackendinternalhandlergithub_mirror_handler.go
 * @Description         : GitHub加速源处理器
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package handler

import (
	"fmt"
	"frp-web-panel/internal/model"
	"frp-web-panel/internal/service"
	"frp-web-panel/internal/util"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type GithubMirrorHandler struct {
	service    *service.GithubMirrorService
	logService *service.LogService
}

func NewGithubMirrorHandler() *GithubMirrorHandler {
	return &GithubMirrorHandler{
		service:    service.NewGithubMirrorService(),
		logService: service.NewLogService(),
	}
}

// GetAll godoc
// @Summary 获取所有GitHub加速源
// @Description 获取所有已配置的GitHub加速源列表
// @Tags GitHub加速源
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} util.Response{data=[]object} "加速源列表"
// @Failure 500 {object} util.Response "获取加速源列表失败"
// @Router /api/github-mirrors [get]
func (h *GithubMirrorHandler) GetAll(c *gin.Context) {
	mirrors, err := h.service.GetAll()
	if err != nil {
		util.ErrorResponse(c, http.StatusInternalServerError, "获取加速源列表失败")
		return
	}
	util.SuccessResponse(c, mirrors)
}

// GetByID godoc
// @Summary 获取GitHub加速源详情
// @Description 根据ID获取指定的GitHub加速源详情
// @Tags GitHub加速源
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "加速源ID"
// @Success 200 {object} util.Response{data=object} "加速源详情"
// @Failure 404 {object} util.Response "加速源不存在"
// @Router /api/github-mirrors/{id} [get]
func (h *GithubMirrorHandler) GetByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	mirror, err := h.service.GetByID(uint(id))
	if err != nil {
		util.ErrorResponse(c, http.StatusNotFound, "加速源不存在")
		return
	}
	util.SuccessResponse(c, mirror)
}

// Create godoc
// @Summary 创建GitHub加速源
// @Description 创建新的GitHub加速源配置
// @Tags GitHub加速源
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.GithubMirror true "加速源信息"
// @Success 200 {object} util.Response{data=object} "创建成功"
// @Failure 400 {object} util.Response "参数错误"
// @Failure 500 {object} util.Response "创建加速源失败"
// @Router /api/github-mirrors [post]
func (h *GithubMirrorHandler) Create(c *gin.Context) {
	var mirror model.GithubMirror
	if err := c.ShouldBindJSON(&mirror); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "参数错误")
		return
	}

	if err := h.service.Create(&mirror); err != nil {
		util.ErrorResponse(c, http.StatusInternalServerError, "创建加速源失败")
		return
	}

	// 记录操作日志
	userID, _ := c.Get("user_id")
	h.logService.CreateLogAsync(userID.(uint), "create", "github_mirror", mirror.ID,
		fmt.Sprintf("创建GitHub加速源: %s", mirror.Name), c.ClientIP())

	util.SuccessResponse(c, mirror)
}

// Update godoc
// @Summary 更新GitHub加速源
// @Description 更新指定的GitHub加速源配置
// @Tags GitHub加速源
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "加速源ID"
// @Param request body model.GithubMirror true "加速源信息"
// @Success 200 {object} util.Response{data=object} "更新成功"
// @Failure 400 {object} util.Response "参数错误"
// @Failure 500 {object} util.Response "更新加速源失败"
// @Router /api/github-mirrors/{id} [put]
func (h *GithubMirrorHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var mirror model.GithubMirror
	if err := c.ShouldBindJSON(&mirror); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "参数错误")
		return
	}

	mirror.ID = uint(id)
	if err := h.service.Update(&mirror); err != nil {
		util.ErrorResponse(c, http.StatusInternalServerError, "更新加速源失败")
		return
	}

	// 记录操作日志
	userID, _ := c.Get("user_id")
	h.logService.CreateLogAsync(userID.(uint), "update", "github_mirror", mirror.ID,
		fmt.Sprintf("更新GitHub加速源: %s", mirror.Name), c.ClientIP())

	util.SuccessResponse(c, mirror)
}

// Delete godoc
// @Summary 删除GitHub加速源
// @Description 删除指定的GitHub加速源配置
// @Tags GitHub加速源
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "加速源ID"
// @Success 200 {object} util.Response "删除成功"
// @Failure 500 {object} util.Response "删除加速源失败"
// @Router /api/github-mirrors/{id} [delete]
func (h *GithubMirrorHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	// 先获取加速源信息用于日志记录
	mirror, _ := h.service.GetByID(uint(id))
	mirrorName := ""
	if mirror != nil {
		mirrorName = mirror.Name
	}

	if err := h.service.Delete(uint(id)); err != nil {
		util.ErrorResponse(c, http.StatusInternalServerError, "删除加速源失败")
		return
	}

	// 记录操作日志
	userID, _ := c.Get("user_id")
	h.logService.CreateLogAsync(userID.(uint), "delete", "github_mirror", uint(id),
		fmt.Sprintf("删除GitHub加速源: %s (ID: %d)", mirrorName, id), c.ClientIP())

	util.SuccessResponse(c, nil)
}

// SetDefault godoc
// @Summary 设置默认GitHub加速源
// @Description 将指定的GitHub加速源设置为默认使用的加速源
// @Tags GitHub加速源
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "加速源ID"
// @Success 200 {object} util.Response{data=object} "设置成功"
// @Failure 500 {object} util.Response "设置默认加速源失败"
// @Router /api/github-mirrors/{id}/set-default [post]
func (h *GithubMirrorHandler) SetDefault(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if err := h.service.SetDefault(uint(id)); err != nil {
		util.ErrorResponse(c, http.StatusInternalServerError, "设置默认加速源失败")
		return
	}

	// 记录操作日志
	userID, _ := c.Get("user_id")
	mirror, _ := h.service.GetByID(uint(id))
	mirrorName := ""
	if mirror != nil {
		mirrorName = mirror.Name
	}
	h.logService.CreateLogAsync(userID.(uint), "set_default", "github_mirror", uint(id),
		fmt.Sprintf("设置默认GitHub加速源: %s", mirrorName), c.ClientIP())

	util.SuccessResponse(c, gin.H{"message": "设置成功"})
}
