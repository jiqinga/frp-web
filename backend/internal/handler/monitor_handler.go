/*
 * @Author              : 寂情�?
 * @Date                : 2025-11-14 16:22:31
 * @LastEditors         : 寂情�?
 * @LastEditTime        : 2025-12-30 16:31:46
 * @FilePath            : frp-web-testbackendinternalhandlermonitor_handler.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在�?
 */
package handler

import (
	"frp-web-panel/internal/service"
	"frp-web-panel/internal/util"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MonitorHandler struct {
	monitorService *service.MonitorService
}

func NewMonitorHandler() *MonitorHandler {
	return &MonitorHandler{
		monitorService: service.NewMonitorService(),
	}
}

// GetOverview godoc
// @Summary 获取监控概览
// @Description 获取系统监控概览数据，包括服务器数量、客户端数量、代理数量等统计信息
// @Tags 监控
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} util.Response{data=map[string]interface{}} "监控概览数据"
// @Failure 500 {object} util.Response "获取监控概览失败"
// @Router /api/monitor/overview [get]
func (h *MonitorHandler) GetOverview(c *gin.Context) {
	overview, err := h.monitorService.GetOverview()
	if err != nil {
		util.ErrorResponse(c, http.StatusInternalServerError, "获取监控概览失败")
		return
	}
	util.SuccessResponse(c, overview)
}

// GetStats godoc
// @Summary 获取统计数据
// @Description 获取系统详细统计数据，包括流量统计、连接数�?
// @Tags 监控
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} util.Response{data=map[string]interface{}} "统计数据"
// @Failure 500 {object} util.Response "获取统计数据失败"
// @Router /api/monitor/stats [get]
func (h *MonitorHandler) GetStats(c *gin.Context) {
	stats, err := h.monitorService.GetStats()
	if err != nil {
		util.ErrorResponse(c, http.StatusInternalServerError, "获取统计数据失败")
		return
	}
	util.SuccessResponse(c, stats)
}
