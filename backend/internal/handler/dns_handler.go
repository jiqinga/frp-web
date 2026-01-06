package handler

import (
	"fmt"
	"frp-web-panel/internal/model"
	"frp-web-panel/internal/repository"
	"frp-web-panel/internal/service"
	"frp-web-panel/internal/util"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DNSHandler struct {
	providerRepo *repository.DNSProviderRepository
	recordRepo   *repository.DNSRecordRepository
	logService   *service.LogService
}

func NewDNSHandler() *DNSHandler {
	return &DNSHandler{
		providerRepo: repository.NewDNSProviderRepository(),
		recordRepo:   repository.NewDNSRecordRepository(),
		logService:   service.NewLogService(),
	}
}

// GetProviders godoc
// @Summary 获取 DNS 提供商列表
// @Description 获取所有配置的 DNS 提供商
// @Tags DNS管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} util.Response{data=[]object}
// @Failure 500 {object} util.Response
// @Router /api/dns/providers [get]
func (h *DNSHandler) GetProviders(c *gin.Context) {
	providers, err := h.providerRepo.FindAll()
	if err != nil {
		util.Error(c, 500, "获取DNS提供商列表失败")
		return
	}
	util.Success(c, providers)
}

// DNSProviderRequest 用于接收创建/更新 DNS 提供商的请求
type DNSProviderRequest struct {
	Name      string `json:"name" binding:"required"`
	Type      string `json:"type" binding:"required"`
	AccessKey string `json:"access_key" binding:"required"`
	SecretKey string `json:"secret_key"` // 创建时必填，更新时可选
	Enabled   bool   `json:"enabled"`
}

// CreateProvider godoc
// @Summary 创建 DNS 提供商
// @Description 创建新的 DNS 提供商配置
// @Tags DNS管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body DNSProviderRequest true "DNS 提供商信息"
// @Success 200 {object} util.Response{data=object}
// @Failure 400 {object} util.Response
// @Failure 500 {object} util.Response
// @Router /api/dns/providers [post]
func (h *DNSHandler) CreateProvider(c *gin.Context) {
	var req DNSProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	// 阿里云和腾讯云需要 SecretKey
	if req.Type != "cloudflare" && req.SecretKey == "" {
		util.Error(c, 400, "SecretKey 不能为空")
		return
	}

	provider := &model.DNSProvider{
		Name:      req.Name,
		Type:      model.DNSProviderType(req.Type),
		AccessKey: req.AccessKey,
		SecretKey: req.SecretKey,
		Enabled:   req.Enabled,
	}

	if err := h.providerRepo.Create(provider); err != nil {
		util.Error(c, 500, "创建DNS提供商失败")
		return
	}

	// 记录操作日志
	userID, _ := c.Get("user_id")
	h.logService.CreateLogAsync(userID.(uint), "create", "dns_provider", provider.ID,
		fmt.Sprintf("创建DNS提供商: %s", provider.Name), c.ClientIP())

	// 返回时清除 SecretKey
	provider.SecretKey = ""
	util.Success(c, provider)
}

// UpdateProvider godoc
// @Summary 更新 DNS 提供商
// @Description 更新指定的 DNS 提供商配置
// @Tags DNS管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "提供商 ID"
// @Param request body DNSProviderRequest true "DNS 提供商信息"
// @Success 200 {object} util.Response{data=object}
// @Failure 400 {object} util.Response
// @Failure 404 {object} util.Response
// @Failure 500 {object} util.Response
// @Router /api/dns/providers/{id} [put]
func (h *DNSHandler) UpdateProvider(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	// 先获取现有记录
	existing, err := h.providerRepo.FindByID(uint(id))
	if err != nil {
		util.Error(c, 404, "DNS提供商不存在")
		return
	}

	var req DNSProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	// 更新字段
	existing.Name = req.Name
	existing.Type = model.DNSProviderType(req.Type)
	existing.AccessKey = req.AccessKey
	existing.Enabled = req.Enabled

	// 只有当提供了新的 SecretKey 时才更新
	if req.SecretKey != "" {
		existing.SecretKey = req.SecretKey
	}

	if err := h.providerRepo.Update(existing); err != nil {
		util.Error(c, 500, "更新DNS提供商失败")
		return
	}

	// 记录操作日志
	userID, _ := c.Get("user_id")
	h.logService.CreateLogAsync(userID.(uint), "update", "dns_provider", existing.ID,
		fmt.Sprintf("更新DNS提供商: %s", existing.Name), c.ClientIP())

	// 返回时清除 SecretKey
	existing.SecretKey = ""
	util.Success(c, existing)
}

// DeleteProvider godoc
// @Summary 删除 DNS 提供商
// @Description 删除指定的 DNS 提供商
// @Tags DNS管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "提供商 ID"
// @Success 200 {object} util.Response
// @Failure 500 {object} util.Response
// @Router /api/dns/providers/{id} [delete]
func (h *DNSHandler) DeleteProvider(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	// 先获取提供商信息用于日志记录
	provider, _ := h.providerRepo.FindByID(uint(id))
	providerName := ""
	if provider != nil {
		providerName = provider.Name
	}

	if err := h.providerRepo.Delete(uint(id)); err != nil {
		util.Error(c, 500, "删除DNS提供商失败")
		return
	}

	// 记录操作日志
	userID, _ := c.Get("user_id")
	h.logService.CreateLogAsync(userID.(uint), "delete", "dns_provider", uint(id),
		fmt.Sprintf("删除DNS提供商: %s", providerName), c.ClientIP())

	util.Success(c, nil)
}

// TestProvider godoc
// @Summary 测试指定 DNS 提供商
// @Description 测试指定 DNS 提供商的连接是否正常
// @Tags DNS管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "提供商 ID"
// @Success 200 {object} util.Response{data=object{message=string}}
// @Failure 404 {object} util.Response
// @Failure 500 {object} util.Response
// @Router /api/dns/providers/{id}/test [post]
func (h *DNSHandler) TestProvider(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	provider, err := h.providerRepo.FindByID(uint(id))
	if err != nil {
		util.Error(c, 404, "DNS提供商不存在")
		return
	}
	dnsService := service.NewDNSService()
	if err := dnsService.TestProvider(provider); err != nil {
		util.Error(c, 500, err.Error())
		return
	}

	// 记录操作日志
	userID, _ := c.Get("user_id")
	h.logService.CreateLogAsync(userID.(uint), "test", "dns_provider", provider.ID,
		fmt.Sprintf("测试DNS提供商: %s", provider.Name), c.ClientIP())

	util.Success(c, gin.H{"message": "连接成功"})
}

// GetProviderSecret godoc
// @Summary 获取 DNS 提供商密钥
// @Description 获取指定 DNS 提供商的密钥（用于编辑时显示）
// @Tags DNS管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "提供商 ID"
// @Success 200 {object} util.Response{data=object{secret_key=string}}
// @Failure 404 {object} util.Response
// @Router /api/dns/providers/{id}/secret [get]
func (h *DNSHandler) GetProviderSecret(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	provider, err := h.providerRepo.FindByID(uint(id))
	if err != nil {
		util.Error(c, 404, "DNS提供商不存在")
		return
	}

	// 记录操作日志
	userID, _ := c.Get("user_id")
	h.logService.CreateLogAsync(userID.(uint), "view_secret", "dns_provider", provider.ID,
		fmt.Sprintf("查看DNS提供商密钥: %s", provider.Name), c.ClientIP())

	util.Success(c, gin.H{"secret_key": provider.SecretKey})
}

// TestProviderConfig godoc
// @Summary 测试 DNS 配置
// @Description 测试未保存的 DNS 提供商配置是否可用
// @Tags DNS管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{type=string,access_key=string,secret_key=string} true "DNS 配置信息"
// @Success 200 {object} util.Response{data=object{success=bool,message=string}}
// @Failure 400 {object} util.Response
// @Failure 500 {object} util.Response
// @Router /api/dns/providers/test [post]
func (h *DNSHandler) TestProviderConfig(c *gin.Context) {
	var req struct {
		Type      string `json:"type" binding:"required"`
		AccessKey string `json:"access_key" binding:"required"`
		SecretKey string `json:"secret_key"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	// 构造临时 provider 对象用于测试
	provider := &model.DNSProvider{
		Type:      model.DNSProviderType(req.Type),
		AccessKey: req.AccessKey,
		SecretKey: req.SecretKey,
	}

	dnsService := service.NewDNSService()
	if err := dnsService.TestProvider(provider); err != nil {
		util.Error(c, 500, err.Error())
		return
	}

	// 记录操作日志
	userID, _ := c.Get("user_id")
	h.logService.CreateLogAsync(userID.(uint), "test", "dns_provider", 0,
		"测试DNS配置", c.ClientIP())

	util.Success(c, gin.H{"success": true, "message": "连接成功"})
}

// GetRecords godoc
// @Summary 获取 DNS 记录列表
// @Description 获取所有 DNS 记录
// @Tags DNS管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} util.Response{data=[]object}
// @Failure 500 {object} util.Response
// @Router /api/dns/records [get]
func (h *DNSHandler) GetRecords(c *gin.Context) {
	records, err := h.recordRepo.FindAll()
	if err != nil {
		util.Error(c, 500, "获取DNS记录列表失败")
		return
	}
	util.Success(c, records)
}

// GetProviderDomains godoc
// @Summary 获取提供商域名列表
// @Description 获取指定 DNS 提供商下托管的所有域名
// @Tags DNS管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "提供商 ID"
// @Success 200 {object} util.Response{data=[]string}
// @Failure 404 {object} util.Response
// @Failure 500 {object} util.Response
// @Router /api/dns/providers/{id}/domains [get]
func (h *DNSHandler) GetProviderDomains(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	provider, err := h.providerRepo.FindByID(uint(id))
	if err != nil {
		util.Error(c, 404, "DNS提供商不存在")
		return
	}

	dnsService := service.NewDNSService()
	operator, err := dnsService.CreateDNSOperator(provider)
	if err != nil {
		util.Error(c, 500, "创建DNS操作实例失败: "+err.Error())
		return
	}

	domains, err := operator.ListDomains(c.Request.Context())
	if err != nil {
		util.Error(c, 500, "获取域名列表失败: "+err.Error())
		return
	}

	util.Success(c, domains)
}
