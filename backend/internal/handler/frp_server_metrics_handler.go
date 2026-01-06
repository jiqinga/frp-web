/*
 * @Author              : 寂情�?
 * @Date                : 2025-12-29 16:41:20
 * @LastEditors         : 寂情�?
 * @LastEditTime        : 2025-12-30 16:19:26
 * @FilePath            : frp-web-testbackendinternalhandlerfrp_server_metrics_handler.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在�?
 */
package handler

import (
	"errors"
	apperrors "frp-web-panel/internal/errors"
	"frp-web-panel/internal/frp"
	"frp-web-panel/internal/middleware"
	"frp-web-panel/internal/util"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// GetMetrics godoc
// @Summary 获取服务器指�?
// @Description 获取指定 FRP 服务器的实时运行指标
// @Tags FRP服务�?
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "服务器ID"
// @Success 200 {object} util.Response{data=object}
// @Failure 400 {object} util.Response
// @Failure 500 {object} util.Response
// @Failure 501 {object} util.Response "服务器未开�?metrics 接口"
// @Router /api/frp-servers/{id}/metrics [get]
func (h *FrpServerHandler) GetMetrics(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.AbortWithAppError(c, apperrors.NewBadRequest("无效的ID参数"))
		return
	}
	metrics, err := h.service.GetMetrics(uint(id))
	if err != nil {
		if errors.Is(err, frp.ErrMetricsNotSupported) {
			middleware.AbortWithAppError(c, &apperrors.AppError{Code: http.StatusNotImplemented, Message: "该服务器未开�?metrics 接口"})
			return
		}
		middleware.AbortWithAppError(c, apperrors.NewInternal("获取指标失败: "+err.Error(), err))
		return
	}
	util.SuccessResponse(c, metrics)
}

// GetMetricsHistory godoc
// @Summary 获取服务器历史指�?
// @Description 获取指定 FRP 服务器的历史运行指标数据
// @Tags FRP服务�?
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "服务器ID"
// @Param days query int false "查询天数(1-7)" default(1)
// @Success 200 {object} util.Response{data=[]object}
// @Failure 400 {object} util.Response
// @Failure 500 {object} util.Response
// @Router /api/frp-servers/{id}/metrics-history [get]
func (h *FrpServerHandler) GetMetricsHistory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.AbortWithAppError(c, apperrors.NewBadRequest("无效的ID参数"))
		return
	}

	days := 1
	if d := c.Query("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed >= 1 && parsed <= 7 {
			days = parsed
		}
	}

	end := time.Now()
	start := end.AddDate(0, 0, -days)

	records, err := h.metricsRepo.GetHistory(uint(id), start, end)
	if err != nil {
		middleware.AbortWithAppError(c, apperrors.NewInternal("获取历史指标失败: "+err.Error(), err))
		return
	}
	util.SuccessResponse(c, records)
}
