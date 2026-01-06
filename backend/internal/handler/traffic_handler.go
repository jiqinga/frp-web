/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-14 16:03:26
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-30 16:30:42
 * @FilePath            : frp-web-test/backend/internal/handler/traffic_handler.go
 * @Description         : 流量统计处理器
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package handler

import (
	"frp-web-panel/internal/repository"
	"frp-web-panel/internal/service"
	"frp-web-panel/internal/util"
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type TrafficHandler struct {
	trafficService   *service.TrafficService
	proxyMetricsRepo *repository.ProxyMetricsRepository
	proxyRepo        *repository.ProxyRepository
	clientRepo       *repository.ClientRepository
}

func NewTrafficHandler() *TrafficHandler {
	return &TrafficHandler{
		trafficService:   service.NewTrafficService(),
		proxyMetricsRepo: repository.NewProxyMetricsRepository(),
		proxyRepo:        repository.NewProxyRepository(),
		clientRepo:       repository.NewClientRepository(),
	}
}

// GetTrafficSummary godoc
// @Summary 获取流量统计汇总
// @Description 获取系统整体流量统计汇总数据
// @Tags 流量统计
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} util.Response{data=map[string]interface{}} "流量统计汇总"
// @Failure 500 {object} util.Response "获取流量统计失败"
// @Router /api/traffic/summary [get]
func (h *TrafficHandler) GetTrafficSummary(c *gin.Context) {
	summary, err := h.trafficService.GetTrafficSummary()
	if err != nil {
		util.Error(c, 500, "获取流量统计失败")
		return
	}
	util.Success(c, summary)
}

// GetTrafficHistory godoc
// @Summary 获取代理流量历史
// @Description 获取指定代理的流量历史记录
// @Tags 流量统计
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "代理ID"
// @Param start query string false "开始时间(RFC3339格式)"
// @Param end query string false "结束时间(RFC3339格式)"
// @Success 200 {object} util.Response{data=[]map[string]interface{}} "流量历史记录"
// @Failure 404 {object} util.Response "代理或客户端不存在"
// @Failure 500 {object} util.Response "获取流量历史失败"
// @Router /api/traffic/proxy/{id} [get]
func (h *TrafficHandler) GetTrafficHistory(c *gin.Context) {
	proxyID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	startStr := c.DefaultQuery("start", time.Now().Add(-24*time.Hour).Format(time.RFC3339))
	endStr := c.DefaultQuery("end", time.Now().Format(time.RFC3339))

	start, _ := time.Parse(time.RFC3339, startStr)
	end, _ := time.Parse(time.RFC3339, endStr)

	log.Printf("[DEBUG GetTrafficHistory] proxyID=%d, start=%v, end=%v", proxyID, start, end)

	// 1. 根据 proxy_id 获取 proxy 信息
	proxy, err := h.proxyRepo.FindByID(uint(proxyID))
	if err != nil {
		log.Printf("[DEBUG GetTrafficHistory] 获取代理信息失败: %v", err)
		util.Error(c, 404, "代理不存在")
		return
	}

	// 2. 根据 client_id 获取 client 信息，获取 frp_server_id
	client, err := h.clientRepo.FindByID(proxy.ClientID)
	if err != nil {
		log.Printf("[DEBUG GetTrafficHistory] 获取客户端信息失败: %v", err)
		util.Error(c, 404, "客户端不存在")
		return
	}

	// 3. 检查 frp_server_id 是否存在
	if client.FrpServerID == nil {
		log.Printf("[DEBUG GetTrafficHistory] 客户端 %d 没有关联 FRP 服务器", client.ID)
		util.Success(c, []interface{}{})
		return
	}

	serverID := *client.FrpServerID

	// 4. 构建完整的代理名称（FRP 使用 clientName.proxyName 格式）
	fullProxyName := client.Name + "." + proxy.Name
	log.Printf("[DEBUG GetTrafficHistory] 查询 proxy_metrics_history: serverID=%d, fullProxyName=%s", serverID, fullProxyName)

	// 5. 从 proxy_metrics_history 表查询流量历史
	history, err := h.proxyMetricsRepo.GetHistory(serverID, fullProxyName, start, end)
	if err != nil {
		log.Printf("[DEBUG GetTrafficHistory] 查询失败: %v", err)
		util.Error(c, 500, "获取流量历史失败")
		return
	}

	log.Printf("[DEBUG GetTrafficHistory] 查询 proxy_metrics_history 表结果: %d 条记录", len(history))

	// 如果使用完整名称没有找到数据，尝试只用代理名称（兼容旧数据）
	if len(history) == 0 {
		log.Printf("[DEBUG GetTrafficHistory] 完整名称无结果，尝试使用短名称: %s", proxy.Name)
		history, err = h.proxyMetricsRepo.GetHistory(serverID, proxy.Name, start, end)
		if err == nil && len(history) > 0 {
			log.Printf("[DEBUG GetTrafficHistory] 使用短名称 %s 查询到 %d 条记录", proxy.Name, len(history))
		}
	}

	// 5. 转换为前端期望的格式
	var result []map[string]interface{}
	for _, h := range history {
		result = append(result, map[string]interface{}{
			"id":               h.ID,
			"proxy_id":         proxyID,
			"bytes_in":         h.TrafficIn,
			"bytes_out":        h.TrafficOut,
			"current_rate_in":  h.RateIn,
			"current_rate_out": h.RateOut,
			"record_time":      h.RecordTime,
			"created_at":       h.CreatedAt,
		})
	}

	util.Success(c, result)
}

// GetProxyRates godoc
// @Summary 获取服务器下所有隧道的实时速率
// @Description 获取指定FRP服务器下所有代理隧道的最新速率数据
// @Tags 流量统计
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param server_id path int true "FRP服务器ID"
// @Success 200 {object} util.Response{data=[]object} "隧道速率列表"
// @Failure 400 {object} util.Response "无效的服务器ID"
// @Failure 500 {object} util.Response "获取隧道速率失败"
// @Router /api/traffic/rates/{server_id} [get]
func (h *TrafficHandler) GetProxyRates(c *gin.Context) {
	serverID, err := strconv.ParseUint(c.Param("server_id"), 10, 32)
	if err != nil {
		util.Error(c, 400, "无效的服务器ID")
		return
	}

	rates, err := h.proxyMetricsRepo.GetLatestByServer(uint(serverID))
	if err != nil {
		util.Error(c, 500, "获取隧道速率失败")
		return
	}
	util.Success(c, rates)
}

// GetProxyRateHistory godoc
// @Summary 获取单个隧道的速率历史
// @Description 获取指定隧道在时间范围内的速率历史记录
// @Tags 流量统计
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param server_id path int true "FRP服务器ID"
// @Param proxy_name path string true "隧道名称"
// @Param start query string false "开始时间(RFC3339格式)"
// @Param end query string false "结束时间(RFC3339格式)"
// @Success 200 {object} util.Response{data=[]object} "速率历史记录"
// @Failure 400 {object} util.Response "无效的服务器ID或隧道名称"
// @Failure 500 {object} util.Response "获取隧道速率历史失败"
// @Router /api/traffic/rates/{server_id}/{proxy_name} [get]
func (h *TrafficHandler) GetProxyRateHistory(c *gin.Context) {
	serverID, err := strconv.ParseUint(c.Param("server_id"), 10, 32)
	if err != nil {
		util.Error(c, 400, "无效的服务器ID")
		return
	}

	proxyName := c.Param("proxy_name")
	if proxyName == "" {
		util.Error(c, 400, "隧道名称不能为空")
		return
	}

	startStr := c.DefaultQuery("start", time.Now().Add(-24*time.Hour).Format(time.RFC3339))
	endStr := c.DefaultQuery("end", time.Now().Format(time.RFC3339))

	start, _ := time.Parse(time.RFC3339, startStr)
	end, _ := time.Parse(time.RFC3339, endStr)

	history, err := h.proxyMetricsRepo.GetHistory(uint(serverID), proxyName, start, end)
	if err != nil {
		util.Error(c, 500, "获取隧道速率历史失败")
		return
	}
	util.Success(c, history)
}

// GetTrafficTrend godoc
// @Summary 获取流量趋势
// @Description 获取指定时间范围内的流量趋势数据，默认24小时
// @Tags 流量统计
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param hours query int false "时间范围(小时)" default(24)
// @Success 200 {object} util.Response{data=[]repository.TrafficTrendPoint} "流量趋势数据"
// @Failure 500 {object} util.Response "获取流量趋势失败"
// @Router /api/traffic/trend [get]
func (h *TrafficHandler) GetTrafficTrend(c *gin.Context) {
	hours, _ := strconv.Atoi(c.DefaultQuery("hours", "24"))
	if hours <= 0 {
		hours = 24
	}

	trend, err := h.proxyMetricsRepo.GetHourlyTrafficTrend(hours)
	if err != nil {
		util.Error(c, 500, "获取流量趋势失败")
		return
	}
	util.Success(c, trend)
}

// GetProxiesTrafficSummary godoc
// @Summary 批量获取代理流量汇总
// @Description 获取所有代理在指定时间范围内的流量汇总数据
// @Tags 流量统计
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param hours query int false "时间范围(小时)" default(24)
// @Success 200 {object} util.Response{data=map[string]interface{}} "代理流量汇总，键为代理ID"
// @Failure 500 {object} util.Response "获取流量汇总失败"
// @Router /api/traffic/proxies/summary [get]
func (h *TrafficHandler) GetProxiesTrafficSummary(c *gin.Context) {
	hours, _ := strconv.Atoi(c.DefaultQuery("hours", "24"))
	if hours <= 0 {
		hours = 24
	}

	// 获取所有代理
	proxies, err := h.proxyRepo.FindAll()
	if err != nil {
		util.Error(c, 500, "获取代理列表失败")
		return
	}

	// 构建代理名称列表（格式：clientName.proxyName）
	proxyNames := make([]string, 0, len(proxies))
	proxyNameMap := make(map[string]uint) // 映射 fullName -> proxyID

	for _, proxy := range proxies {
		client, err := h.clientRepo.FindByID(proxy.ClientID)
		if err != nil || client == nil {
			continue
		}
		fullName := client.Name + "." + proxy.Name
		proxyNames = append(proxyNames, fullName)
		proxyNameMap[fullName] = proxy.ID
	}

	// 获取流量汇总（将小时转换为天数，向上取整）
	days := (hours + 23) / 24
	if days < 1 {
		days = 1
	}
	summary, err := h.proxyMetricsRepo.GetTrafficSummaryByProxyNames(proxyNames, days)
	if err != nil {
		util.Error(c, 500, "获取流量汇总失败")
		return
	}

	// 转换为以 proxyID 为键的结果
	result := make(map[string]interface{})
	for fullName, traffic := range summary {
		if proxyID, ok := proxyNameMap[fullName]; ok {
			result[strconv.FormatUint(uint64(proxyID), 10)] = traffic
		}
	}

	util.Success(c, result)
}
