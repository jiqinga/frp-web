/*
 * @Author              : 寂情�?
 * @Date                : 2025-11-14 16:11:12
 * @LastEditors         : 寂情�?
 * @LastEditTime        : 2025-12-30 16:30:28
 * @FilePath            : frp-web-testbackendinternalhandlerlog_handler.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在�?
 */
package handler

import (
	"frp-web-panel/internal/service"
	"frp-web-panel/internal/util"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type LogHandler struct {
	logService *service.LogService
}

func NewLogHandler() *LogHandler {
	return &LogHandler{
		logService: service.NewLogService(),
	}
}

// GetLogs godoc
// @Summary 获取操作日志列表
// @Description 分页获取系统操作日志，支持按操作类型和资源类型筛�?
// @Tags 日志管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param operation_type query string false "操作类型"
// @Param resource_type query string false "资源类型"
// @Success 200 {object} util.Response{data=object} "日志列表和总数"
// @Failure 500 {object} util.Response "获取日志失败"
// @Router /api/logs [get]
func (h *LogHandler) GetLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	operationType := c.Query("operation_type")
	resourceType := c.Query("resource_type")

	logs, total, err := h.logService.GetLogs(page, pageSize, operationType, resourceType)
	if err != nil {
		util.ErrorResponse(c, http.StatusInternalServerError, "获取日志失败")
		return
	}

	util.SuccessResponse(c, gin.H{
		"list":  logs,
		"total": total,
	})
}

// CreateLog godoc
// @Summary 创建操作日志
// @Description 手动创建一条操作日志记�?
// @Tags 日志管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object true "日志信息" SchemaExample({"operation_type": "create", "resource_type": "proxy", "resource_id": 1, "description": "创建代理"})
// @Success 200 {object} util.Response "创建成功"
// @Failure 400 {object} util.Response "参数错误"
// @Failure 500 {object} util.Response "创建日志失败"
// @Router /api/logs [post]
func (h *LogHandler) CreateLog(c *gin.Context) {
	var req struct {
		OperationType string `json:"operation_type" binding:"required"`
		ResourceType  string `json:"resource_type" binding:"required"`
		ResourceID    uint   `json:"resource_id"`
		Description   string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "参数错误")
		return
	}

	userID, _ := c.Get("user_id")
	if err := h.logService.CreateLog(userID.(uint), req.OperationType, req.ResourceType, req.ResourceID, req.Description, c.ClientIP()); err != nil {
		util.ErrorResponse(c, http.StatusInternalServerError, "创建日志失败")
		return
	}

	util.SuccessResponse(c, nil)
}
