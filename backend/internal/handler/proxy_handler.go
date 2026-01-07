/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-14 15:33:43
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2026-01-07 11:02:25
 * @FilePath            : frp-web-testbackendinternalhandlerproxy_handler.go
 * @Description         : 代理处理器
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package handler

import (
	"fmt"
	"frp-web-panel/internal/logger"
	"frp-web-panel/internal/model"
	"frp-web-panel/internal/repository"
	"frp-web-panel/internal/service"
	"frp-web-panel/internal/util"
	"frp-web-panel/internal/websocket"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProxyHandler struct {
	proxyService  *service.ProxyService
	clientService *service.ClientService
	logService    *service.LogService
	certRepo      *repository.CertificateRepository
	proxyRepo     *repository.ProxyRepository
}

func NewProxyHandler() *ProxyHandler {
	return &ProxyHandler{
		proxyService:  service.NewProxyService(),
		clientService: service.NewClientService(),
		logService:    service.NewLogService(),
		certRepo:      repository.NewCertificateRepository(),
		proxyRepo:     repository.NewProxyRepository(),
	}
}

// pushConfigUpdate 推送配置更新到客户端
func (h *ProxyHandler) pushConfigUpdate(clientID uint) {
	logger.Debugf("[配置推送] 开始推送配置到客户端 ID=%d", clientID)

	// 检查客户端是否在线
	isOnline := websocket.ClientDaemonHubInstance.IsClientOnline(clientID)
	logger.Debugf("[配置推送] 客户端 ID=%d 在线状态: %v", clientID, isOnline)

	if !isOnline {
		logger.Warnf("[配置推送] 客户端 ID=%d 不在线，跳过配置推送", clientID)
		return
	}

	// 获取客户端信息
	client, err := h.clientService.GetClient(clientID)
	if err != nil {
		logger.Errorf("[配置推送] 获取客户端信息失败: %v", err)
		return
	}
	logger.Debugf("[配置推送] 客户端 %s 当前配置版本: %d", client.Name, client.ConfigVersion)

	// 🔧 修复：在推送配置前，先同步所需的证书
	h.syncCertificatesForClient(clientID)

	// 生成配置
	config, err := h.proxyService.ExportClientConfig(clientID)
	if err != nil {
		logger.Errorf("[配置推送] 生成配置失败: %v", err)
		// 生成配置失败，更新状态为 failed
		h.clientService.UpdateConfigSyncStatus(clientID, false, fmt.Sprintf("生成配置失败: %v", err), false)
		return
	}
	logger.Debugf("[配置推送] 生成的配置内容:\n%s", config)

	// 递增版本号
	newVersion := client.ConfigVersion + 1
	logger.Debugf("[配置推送] 新版本号: %d", newVersion)

	// 推送前设置状态为 pending
	h.clientService.SetConfigSyncPending(clientID)

	// 推送配置
	if err := websocket.ClientDaemonHubInstance.PushConfigUpdate(clientID, config, newVersion); err != nil {
		logger.Errorf("[配置推送] 推送配置失败: %v", err)
		// 推送失败，更新状态为 failed
		h.clientService.UpdateConfigSyncStatus(clientID, false, fmt.Sprintf("推送配置失败: %v", err), false)
		return
	}
	logger.Infof("[配置推送] 配置已推送到客户端 ID=%d，等待 daemon 返回同步结果", clientID)

	// 更新配置版本号到数据库
	h.clientService.UpdateConfigSync(clientID, newVersion, nil)
}

// syncCertificatesForClient 同步客户端所需的所有证书
func (h *ProxyHandler) syncCertificatesForClient(clientID uint) {
	logger.Debugf("[证书同步] 开始同步客户端 ID=%d 所需的证书", clientID)

	// 获取该客户端所有启用的代理
	proxies, err := h.proxyService.GetProxiesByClient(clientID)
	if err != nil {
		logger.Errorf("[证书同步] 获取代理列表失败: %v", err)
		return
	}

	// 收集所有需要同步的证书ID（去重）
	certIDs := make(map[uint]bool)
	for _, proxy := range proxies {
		if proxy.Enabled && proxy.CertID != nil {
			certIDs[*proxy.CertID] = true
		}
	}

	if len(certIDs) == 0 {
		logger.Debugf("[证书同步] 客户端 ID=%d 没有需要同步的证书", clientID)
		return
	}

	logger.Debugf("[证书同步] 客户端 ID=%d 需要同步 %d 个证书", clientID, len(certIDs))

	// 同步每个证书
	for certID := range certIDs {
		cert, err := h.certRepo.FindByID(certID)
		if err != nil {
			logger.Errorf("[证书同步] 获取证书 ID=%d 失败: %v", certID, err)
			continue
		}
		if cert == nil {
			logger.Warnf("[证书同步] 证书 ID=%d 不存在", certID)
			continue
		}
		if cert.Status != model.CertStatusActive {
			logger.Warnf("[证书同步] 证书 ID=%d 状态不是 active (当前=%s)，跳过", certID, cert.Status)
			continue
		}
		if cert.CertPEM == "" || cert.KeyPEM == "" {
			logger.Warnf("[证书同步] 证书 ID=%d 内容为空，跳过", certID)
			continue
		}

		// 推送证书到客户端
		if err := websocket.ClientDaemonHubInstance.PushCertSync(clientID, cert.Domain, cert.CertPEM, cert.KeyPEM); err != nil {
			logger.Errorf("[证书同步] 推送证书 %s 失败: %v", cert.Domain, err)
		} else {
			logger.Infof("[证书同步] 证书 %s 已推送到客户端 ID=%d", cert.Domain, clientID)
		}
	}
}

// GetAllProxies godoc
// @Summary 获取所有代理列表
// @Description 获取系统中所有代理配置的列表
// @Tags 代理管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} util.Response{data=[]object} "代理列表"
// @Failure 500 {object} util.Response "获取代理列表失败"
// @Router /api/proxies [get]
func (h *ProxyHandler) GetAllProxies(c *gin.Context) {
	proxies, err := h.proxyService.GetAllProxies()
	if err != nil {
		util.Error(c, 500, "获取代理列表失败")
		return
	}

	util.Success(c, proxies)
}

// GetProxiesByClient godoc
// @Summary 获取客户端代理列表
// @Description 获取指定客户端下的所有代理配置
// @Tags 代理管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "客户端ID"
// @Success 200 {object} util.Response{data=[]object} "代理列表"
// @Failure 500 {object} util.Response "获取代理列表失败"
// @Router /api/clients/{id}/proxies [get]
func (h *ProxyHandler) GetProxiesByClient(c *gin.Context) {
	clientID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	proxies, err := h.proxyService.GetProxiesByClient(uint(clientID))
	if err != nil {
		util.Error(c, 500, "获取代理列表失败")
		return
	}

	util.Success(c, proxies)
}

// checkClientOnline 检查客户端是否在线
func (h *ProxyHandler) checkClientOnline(clientID uint) bool {
	return websocket.ClientDaemonHubInstance.IsClientOnline(clientID)
}

// CreateProxy godoc
// @Summary 创建代理
// @Description 创建新的代理配置，客户端必须在线才能创建
// @Tags 代理管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.Proxy true "代理配置信息"
// @Success 200 {object} util.Response{data=object} "创建成功"
// @Failure 400 {object} util.Response "参数错误或客户端离线"
// @Router /api/proxies [post]
func (h *ProxyHandler) CreateProxy(c *gin.Context) {
	var proxy model.Proxy
	if err := c.ShouldBindJSON(&proxy); err != nil {
		util.Error(c, 400, "参数错误")
		return
	}

	// 校验客户端是否在线
	if !h.checkClientOnline(proxy.ClientID) {
		logger.Warnf("[代理创建] 客户端 ID=%d 离线，拒绝创建代理", proxy.ClientID)
		util.Error(c, 400, "客户端离线，无法创建代理")
		return
	}

	if err := h.proxyService.CreateProxy(&proxy); err != nil {
		logger.Errorf("[代理创建] 创建失败: %v", err)
		util.Error(c, 400, err.Error())
		return
	}

	// 推送配置更新
	h.pushConfigUpdate(proxy.ClientID)

	// 记录操作日志
	userID, _ := c.Get("user_id")
	h.logService.CreateLogAsync(userID.(uint), "create", "proxy", proxy.ID,
		fmt.Sprintf("创建代理: %s (类型: %s, 端口: %d)", proxy.Name, proxy.Type, proxy.RemotePort), c.ClientIP())

	util.Success(c, proxy)
}

// UpdateProxy godoc
// @Summary 更新代理
// @Description 更新指定代理的配置，客户端必须在线才能更新
// @Tags 代理管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "代理ID"
// @Param request body model.Proxy true "代理配置信息"
// @Success 200 {object} util.Response{data=object} "更新成功"
// @Failure 400 {object} util.Response "参数错误或客户端离线"
// @Failure 500 {object} util.Response "更新代理失败"
// @Router /api/proxies/{id} [put]
func (h *ProxyHandler) UpdateProxy(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	// 获取更新前的代理信息
	oldProxy, err := h.proxyService.GetProxy(uint(id))
	if err != nil {
		logger.Errorf("[代理更新] 获取代理信息失败: %v", err)
		util.Error(c, 500, "获取代理信息失败")
		return
	}

	// 校验客户端是否在线
	if !h.checkClientOnline(oldProxy.ClientID) {
		logger.Warnf("[代理更新] 客户端 ID=%d 离线，拒绝更新代理", oldProxy.ClientID)
		util.Error(c, 400, "客户端离线，无法更新代理")
		return
	}

	var proxy model.Proxy
	if err := c.ShouldBindJSON(&proxy); err != nil {
		util.Error(c, 400, "参数错误")
		return
	}

	logger.Debugf("[代理更新] ID=%d, Name=%s, RemotePort=%d -> %d",
		id, proxy.Name, oldProxy.RemotePort, proxy.RemotePort)

	// 🔧 修复：更新代理时保留原有的 enabled 状态和运行时统计数据
	// 问题原因：前端编辑代理时没有传递 enabled 字段，Go 的 bool 零值是 false
	// 导致 GORM Save 时将 enabled 设置为 false，然后 ExportClientConfig 只获取 enabled=true 的代理
	// 解决方案：更新时保留原有的 enabled 状态，除非通过专门的 toggle 接口来切换
	proxy.Enabled = oldProxy.Enabled

	// 如果前端没有传递插件配置，保留原有的插件配置
	// 注意：如果前端明确传递了空字符串，表示要清除插件配置
	if proxy.PluginType == "" && oldProxy.PluginType != "" {
		// 检查是否是前端故意清除插件配置（通过检查请求体）
		// 如果前端没有传递 plugin_type 字段，则保留原有配置
		// 这里简化处理：如果新值为空且旧值不为空，保留旧值
		// 前端如果要清除插件，需要明确传递 plugin_type: ""
	}

	// 🔧 修复：如果前端没有传递 DNS 相关字段，保留原有值
	// 问题原因：前端编辑代理时可能没有传递这些字段，导致被覆盖为零值
	if proxy.DNSProviderID == nil && oldProxy.DNSProviderID != nil {
		proxy.DNSProviderID = oldProxy.DNSProviderID
	}
	if proxy.DNSRootDomain == "" && oldProxy.DNSRootDomain != "" {
		proxy.DNSRootDomain = oldProxy.DNSRootDomain
	}

	// 保留运行时统计数据，这些数据不应该被前端更新覆盖
	proxy.TotalBytesIn = oldProxy.TotalBytesIn
	proxy.TotalBytesOut = oldProxy.TotalBytesOut
	proxy.CurrentBytesInRate = oldProxy.CurrentBytesInRate
	proxy.CurrentBytesOutRate = oldProxy.CurrentBytesOutRate
	proxy.LastOnlineTime = oldProxy.LastOnlineTime
	proxy.LastTrafficUpdate = oldProxy.LastTrafficUpdate
	proxy.FrpStatus = oldProxy.FrpStatus
	proxy.FrpCurConns = oldProxy.FrpCurConns
	proxy.FrpLastStartTime = oldProxy.FrpLastStartTime
	proxy.FrpLastCloseTime = oldProxy.FrpLastCloseTime
	proxy.CreatedAt = oldProxy.CreatedAt

	proxy.ID = uint(id)
	if err := h.proxyService.UpdateProxy(&proxy); err != nil {
		logger.Errorf("[代理更新] 更新失败: %v", err)
		util.Error(c, 500, "更新代理失败")
		return
	}

	logger.Infof("[代理更新] 更新成功, 推送配置到客户端 ClientID=%d", proxy.ClientID)

	// 推送配置更新
	h.pushConfigUpdate(proxy.ClientID)

	// 记录操作日志
	userID, _ := c.Get("user_id")
	h.logService.CreateLogAsync(userID.(uint), "update", "proxy", proxy.ID,
		fmt.Sprintf("更新代理: %s (类型: %s, 端口: %d)", proxy.Name, proxy.Type, proxy.RemotePort), c.ClientIP())

	util.Success(c, proxy)
}

// DeleteProxy godoc
// @Summary 删除代理
// @Description 删除指定的代理配置，可选择是否同时删除关联的DNS记录，客户端必须在线才能删除
// @Tags 代理管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "代理ID"
// @Param deleteDNS query bool false "是否同时删除DNS记录" default(true)
// @Success 200 {object} util.Response "删除成功"
// @Failure 400 {object} util.Response "客户端离线"
// @Failure 500 {object} util.Response "删除代理失败"
// @Router /api/proxies/{id} [delete]
func (h *ProxyHandler) DeleteProxy(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	// 读取 deleteDNS 查询参数，默认为 true
	deleteDNSStr := c.DefaultQuery("deleteDNS", "true")
	deleteDNS := deleteDNSStr == "true" || deleteDNSStr == "1"

	// 先获取代理信息以获取ClientID
	proxy, err := h.proxyService.GetProxy(uint(id))
	if err != nil {
		util.Error(c, 500, "获取代理信息失败")
		return
	}

	// 校验客户端是否在线
	if !h.checkClientOnline(proxy.ClientID) {
		logger.Warnf("[代理删除] 客户端 ID=%d 离线，拒绝删除代理", proxy.ClientID)
		util.Error(c, 400, "客户端离线，无法删除代理")
		return
	}

	// 保存代理信息用于日志记录和证书删除
	proxyName := proxy.Name
	proxyType := proxy.Type
	proxyPort := proxy.RemotePort
	clientID := proxy.ClientID
	certID := proxy.CertID

	logger.Debugf("[代理删除] 删除代理 ID=%d, deleteDNS=%v", id, deleteDNS)

	if err := h.proxyService.DeleteProxy(uint(id), deleteDNS); err != nil {
		util.Error(c, 500, "删除代理失败")
		return
	}

	// 检查是否需要删除客户端的证书文件
	h.cleanupCertificateIfNeeded(clientID, certID, uint(id))

	// 推送配置更新
	h.pushConfigUpdate(clientID)

	// 记录操作日志
	userID, _ := c.Get("user_id")
	dnsInfo := ""
	if deleteDNS {
		dnsInfo = ", 同时删除DNS记录"
	}
	h.logService.CreateLogAsync(userID.(uint), "delete", "proxy", uint(id),
		fmt.Sprintf("删除代理: %s (类型: %s, 端口: %d%s)", proxyName, proxyType, proxyPort, dnsInfo), c.ClientIP())

	util.Success(c, nil)
}

// cleanupCertificateIfNeeded 检查并清理不再使用的证书
func (h *ProxyHandler) cleanupCertificateIfNeeded(clientID uint, certID *uint, deletedProxyID uint) {
	if certID == nil {
		return
	}

	// 检查同客户端的其他代理是否还在使用该证书
	count, err := h.proxyRepo.CountByCertIDAndClientID(*certID, clientID, deletedProxyID)
	if err != nil {
		logger.Errorf("[证书清理] 检查证书使用情况失败: %v", err)
		return
	}

	if count > 0 {
		logger.Debugf("[证书清理] 证书 ID=%d 仍被 %d 个代理使用，跳过删除", *certID, count)
		return
	}

	// 获取证书信息以获取域名
	cert, err := h.certRepo.FindByID(*certID)
	if err != nil || cert == nil {
		logger.Warnf("[证书清理] 获取证书信息失败: %v", err)
		return
	}

	// 推送证书删除命令到客户端
	if err := websocket.ClientDaemonHubInstance.PushCertDelete(clientID, cert.Domain); err != nil {
		logger.Errorf("[证书清理] 推送证书删除失败: %v", err)
	} else {
		logger.Infof("[证书清理] 已推送证书删除命令: domain=%s, clientID=%d", cert.Domain, clientID)
	}
}

// ToggleProxy godoc
// @Summary 切换代理启用/禁用状态
// @Description 切换指定代理的启用/禁用状态，客户端必须在线才能切换
// @Tags 代理管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "代理ID"
// @Success 200 {object} util.Response{data=object} "切换成功"
// @Failure 400 {object} util.Response "客户端离线"
// @Failure 500 {object} util.Response "切换代理状态失败"
// @Router /api/proxies/{id}/toggle [put]
func (h *ProxyHandler) ToggleProxy(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	// 先获取代理信息以校验客户端在线状态
	existingProxy, err := h.proxyService.GetProxy(uint(id))
	if err != nil {
		util.Error(c, 500, "获取代理信息失败")
		return
	}

	// 校验客户端是否在线
	if !h.checkClientOnline(existingProxy.ClientID) {
		logger.Warnf("[代理状态切换] 客户端 ID=%d 离线，拒绝切换状态", existingProxy.ClientID)
		util.Error(c, 400, "客户端离线，无法切换代理状态")
		return
	}

	proxy, err := h.proxyService.ToggleProxy(uint(id))
	if err != nil {
		util.Error(c, 500, "切换代理状态失败")
		return
	}

	logger.Debugf("[代理状态切换] 代理 ID=%d, Name=%s, Enabled=%v", proxy.ID, proxy.Name, proxy.Enabled)

	// 推送配置更新
	h.pushConfigUpdate(proxy.ClientID)

	// 记录操作日志
	userID, _ := c.Get("user_id")
	status := "禁用"
	if proxy.Enabled {
		status = "启用"
	}
	h.logService.CreateLogAsync(userID.(uint), "update", "proxy", proxy.ID,
		fmt.Sprintf("%s代理: %s", status, proxy.Name), c.ClientIP())

	util.Success(c, proxy)
}

// ExportConfig godoc
// @Summary 导出客户端配置
// @Description 导出指定客户端的FRP配置文件(frpc.toml格式)
// @Tags 代理管理
// @Security BearerAuth
// @Param id path int true "客户端ID"
// @Produce text/plain
// @Success 200 {string} string "FRP配置文件内容"
// @Failure 500 {object} util.Response "导出配置失败"
// @Router /api/clients/{id}/export [get]
func (h *ProxyHandler) ExportConfig(c *gin.Context) {
	clientID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	config, err := h.proxyService.ExportClientConfig(uint(clientID))
	if err != nil {
		util.Error(c, 500, "导出配置失败")
		return
	}

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=frpc.toml")
	c.String(200, config)
}
