package handler

import (
	"fmt"
	"frp-web-panel/internal/repository"
	"frp-web-panel/internal/service"
	"frp-web-panel/internal/util"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CertificateHandler struct {
	certRepo    *repository.CertificateRepository
	acmeService *service.ACMEService
	logService  *service.LogService
}

func NewCertificateHandler(certRepo *repository.CertificateRepository, acmeService *service.ACMEService) *CertificateHandler {
	return &CertificateHandler{
		certRepo:    certRepo,
		acmeService: acmeService,
		logService:  service.NewLogService(),
	}
}

// ListCertificates godoc
// @Summary 获取证书列表
// @Description 获取所有 SSL 证书的列表，包含证书状态、域名、过期时间等信息
// @Tags 证书管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} util.Response{data=[]object} "成功返回证书列表"
// @Failure 500 {object} util.Response "服务器内部错误"
// @Router /api/certificates [get]
func (h *CertificateHandler) ListCertificates(c *gin.Context) {
	certs, err := h.certRepo.FindAll()
	if err != nil {
		util.Error(c, http.StatusInternalServerError, "获取证书列表失败: "+err.Error())
		return
	}
	util.Success(c, certs)
}

// GetCertificate godoc
// @Summary 获取证书详情
// @Description 根据证书 ID 获取单个证书的详细信息
// @Tags 证书管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "证书ID"
// @Success 200 {object} util.Response{data=object} "成功返回证书详情"
// @Failure 400 {object} util.Response "无效的证书ID"
// @Failure 404 {object} util.Response "证书不存在"
// @Router /api/certificates/{id} [get]
func (h *CertificateHandler) GetCertificate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		util.Error(c, http.StatusBadRequest, "无效的证书ID")
		return
	}

	cert, err := h.certRepo.FindByID(uint(id))
	if err != nil {
		util.Error(c, http.StatusNotFound, "证书不存在")
		return
	}
	util.Success(c, cert)
}

// RequestCertificateInput 申请证书请求
type RequestCertificateInput struct {
	ProxyID       uint   `json:"proxy_id"`
	Domain        string `json:"domain" binding:"required"`
	DNSProviderID uint   `json:"dns_provider_id" binding:"required"`
	AutoRenew     *bool  `json:"auto_renew"`
}

// RequestCertificateResponse 申请证书响应
type RequestCertificateResponse struct {
	TaskID string `json:"task_id"`
	Domain string `json:"domain"`
}

// RequestCertificate godoc
// @Summary 申请证书
// @Description 异步申请新的 SSL 证书，通过 DNS 验证方式完成域名所有权验证
// @Tags 证书管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body RequestCertificateInput true "证书申请信息"
// @Success 200 {object} util.Response{data=RequestCertificateResponse} "成功提交申请，返回任务ID"
// @Failure 400 {object} util.Response "参数错误"
// @Router /api/certificates [post]
func (h *CertificateHandler) RequestCertificate(c *gin.Context) {
	var input RequestCertificateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		util.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 生成任务ID
	taskID := uuid.New().String()

	// 处理自动续期参数，默认为true
	autoRenew := true
	if input.AutoRenew != nil {
		autoRenew = *input.AutoRenew
	}

	// 异步执行证书申请
	go func() {
		h.acmeService.RequestCertificateWithTaskAndAutoRenew(taskID, input.ProxyID, input.Domain, input.DNSProviderID, autoRenew)
	}()

	// 记录操作日志
	userID, _ := c.Get("user_id")
	h.logService.CreateLogAsync(userID.(uint), "request", "certificate", 0,
		fmt.Sprintf("申请证书: %s", input.Domain), c.ClientIP())

	util.Success(c, RequestCertificateResponse{
		TaskID: taskID,
		Domain: input.Domain,
	})
}

// RenewCertificate godoc
// @Summary 续期证书
// @Description 手动续期指定的 SSL 证书，重新向 ACME 服务器申请新证书
// @Tags 证书管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "证书ID"
// @Success 200 {object} util.Response "续期成功"
// @Failure 400 {object} util.Response "无效的证书ID"
// @Failure 500 {object} util.Response "续期证书失败"
// @Router /api/certificates/{id}/renew [post]
func (h *CertificateHandler) RenewCertificate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		util.Error(c, http.StatusBadRequest, "无效的证书ID")
		return
	}

	// 先获取证书信息用于日志记录
	cert, _ := h.certRepo.FindByID(uint(id))
	domain := ""
	if cert != nil {
		domain = cert.Domain
	}

	if err := h.acmeService.RenewCertificate(uint(id)); err != nil {
		util.Error(c, http.StatusInternalServerError, "续期证书失败: "+err.Error())
		return
	}

	// 记录操作日志
	userID, _ := c.Get("user_id")
	h.logService.CreateLogAsync(userID.(uint), "renew", "certificate", uint(id),
		fmt.Sprintf("续期证书: %s", domain), c.ClientIP())

	util.Success(c, nil)
}

// DeleteCertificate godoc
// @Summary 删除证书
// @Description 删除指定的 SSL 证书，删除后无法恢复
// @Tags 证书管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "证书ID"
// @Success 200 {object} util.Response "删除成功"
// @Failure 400 {object} util.Response "无效的证书ID"
// @Failure 404 {object} util.Response "证书不存在"
// @Failure 500 {object} util.Response "删除证书失败"
// @Router /api/certificates/{id} [delete]
func (h *CertificateHandler) DeleteCertificate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		util.Error(c, http.StatusBadRequest, "无效的证书ID")
		return
	}

	// 检查证书是否存在并获取信息用于日志记录
	cert, err := h.certRepo.FindByID(uint(id))
	if err != nil {
		util.Error(c, http.StatusNotFound, "证书不存在")
		return
	}
	domain := cert.Domain

	if err := h.certRepo.Delete(uint(id)); err != nil {
		util.Error(c, http.StatusInternalServerError, "删除证书失败: "+err.Error())
		return
	}

	// 记录操作日志
	userID, _ := c.Get("user_id")
	h.logService.CreateLogAsync(userID.(uint), "delete", "certificate", uint(id),
		fmt.Sprintf("删除证书: %s", domain), c.ClientIP())

	util.Success(c, nil)
}

// GetCertificatesByDomain godoc
// @Summary 按域名查询证书
// @Description 根据精确域名查询对应的证书列表
// @Tags 证书管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param domain query string true "域名"
// @Success 200 {object} util.Response{data=[]object} "成功返回证书列表"
// @Failure 400 {object} util.Response "域名不能为空"
// @Failure 500 {object} util.Response "查询证书失败"
// @Router /api/certificates/by-domain [get]
func (h *CertificateHandler) GetCertificatesByDomain(c *gin.Context) {
	domain := c.Query("domain")
	if domain == "" {
		util.Error(c, http.StatusBadRequest, "域名不能为空")
		return
	}

	certs, err := h.certRepo.FindByDomain(domain)
	if err != nil {
		util.Error(c, http.StatusInternalServerError, "查询证书失败: "+err.Error())
		return
	}
	util.Success(c, certs)
}

// GetExpiringCertificates godoc
// @Summary 获取即将过期的证书
// @Description 获取所有即将在 30 天内过期的证书列表，用于提醒续期
// @Tags 证书管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} util.Response{data=[]object} "成功返回即将过期的证书列表"
// @Failure 500 {object} util.Response "查询证书失败"
// @Router /api/certificates/expiring [get]
func (h *CertificateHandler) GetExpiringCertificates(c *gin.Context) {
	certs, err := h.certRepo.FindExpiring()
	if err != nil {
		util.Error(c, http.StatusInternalServerError, "查询证书失败: "+err.Error())
		return
	}
	util.Success(c, certs)
}

// GetActiveCertificates godoc
// @Summary 获取活跃证书
// @Description 获取所有状态为有效的证书列表，不包含已过期或申请失败的证书
// @Tags 证书管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} util.Response{data=[]object} "成功返回活跃证书列表"
// @Failure 500 {object} util.Response "查询证书失败"
// @Router /api/certificates/active [get]
func (h *CertificateHandler) GetActiveCertificates(c *gin.Context) {
	certs, err := h.certRepo.FindActiveCertificates()
	if err != nil {
		util.Error(c, http.StatusInternalServerError, "查询证书失败: "+err.Error())
		return
	}
	util.Success(c, certs)
}

// GetMatchingCertificates godoc
// @Summary 匹配证书
// @Description 根据域名模式匹配证书，支持通配符证书匹配（如 *.example.com 可匹配 sub.example.com）
// @Tags 证书管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param domain query string true "域名"
// @Success 200 {object} util.Response{data=[]object} "成功返回匹配的证书列表"
// @Failure 400 {object} util.Response "域名不能为空"
// @Failure 500 {object} util.Response "查询证书失败"
// @Router /api/certificates/match [get]
func (h *CertificateHandler) GetMatchingCertificates(c *gin.Context) {
	domain := c.Query("domain")
	if domain == "" {
		util.Error(c, http.StatusBadRequest, "域名不能为空")
		return
	}

	certs, err := h.certRepo.FindByDomainPattern(domain)
	if err != nil {
		util.Error(c, http.StatusInternalServerError, "查询证书失败: "+err.Error())
		return
	}
	util.Success(c, certs)
}

// UpdateAutoRenewInput 更新自动续期请求
type UpdateAutoRenewInput struct {
	AutoRenew bool `json:"auto_renew"`
}

// UpdateAutoRenew godoc
// @Summary 设置自动续期
// @Description 更新证书的自动续期状态，开启后系统将在证书即将过期时自动续期
// @Tags 证书管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "证书ID"
// @Param input body UpdateAutoRenewInput true "自动续期状态"
// @Success 200 {object} util.Response "更新成功"
// @Failure 400 {object} util.Response "无效的证书ID或参数错误"
// @Failure 404 {object} util.Response "证书不存在"
// @Failure 500 {object} util.Response "更新失败"
// @Router /api/certificates/{id}/auto-renew [put]
func (h *CertificateHandler) UpdateAutoRenew(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		util.Error(c, http.StatusBadRequest, "无效的证书ID")
		return
	}

	var input UpdateAutoRenewInput
	if err := c.ShouldBindJSON(&input); err != nil {
		util.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	cert, err := h.certRepo.FindByID(uint(id))
	if err != nil {
		util.Error(c, http.StatusNotFound, "证书不存在")
		return
	}

	cert.AutoRenew = input.AutoRenew
	if err := h.certRepo.Update(cert); err != nil {
		util.Error(c, http.StatusInternalServerError, "更新失败: "+err.Error())
		return
	}

	// 记录操作日志
	userID, _ := c.Get("user_id")
	h.logService.CreateLogAsync(userID.(uint), "update", "certificate", uint(id),
		fmt.Sprintf("更新证书自动续期: %s", cert.Domain), c.ClientIP())

	util.Success(c, nil)
}

// ReapplyCertificate godoc
// @Summary 重新申请证书
// @Description 重新申请之前申请失败的证书，使用原有的域名和 DNS 提供商配置
// @Tags 证书管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "证书ID"
// @Success 200 {object} util.Response "重新申请成功"
// @Failure 400 {object} util.Response "无效的证书ID"
// @Failure 500 {object} util.Response "重新申请证书失败"
// @Router /api/certificates/{id}/reapply [post]
func (h *CertificateHandler) ReapplyCertificate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		util.Error(c, http.StatusBadRequest, "无效的证书ID")
		return
	}

	// 先获取证书信息用于日志记录
	cert, _ := h.certRepo.FindByID(uint(id))
	domain := ""
	if cert != nil {
		domain = cert.Domain
	}

	if err := h.acmeService.ReapplyCertificate(uint(id)); err != nil {
		util.Error(c, http.StatusInternalServerError, "重新申请证书失败: "+err.Error())
		return
	}

	// 记录操作日志
	userID, _ := c.Get("user_id")
	h.logService.CreateLogAsync(userID.(uint), "reapply", "certificate", uint(id),
		fmt.Sprintf("重新申请证书: %s", domain), c.ClientIP())

	util.Success(c, nil)
}

// DownloadCertificateResponse 下载证书响应
type DownloadCertificateResponse struct {
	Domain        string `json:"domain"`
	CertPem       string `json:"cert_pem"`
	IssuerCertPem string `json:"issuer_cert_pem"`
	FullChainPem  string `json:"full_chain_pem"`
}

// DownloadCertificate godoc
// @Summary 下载证书
// @Description 下载证书内容，包含证书 PEM、颁发者证书 PEM 和完整证书链
// @Tags 证书管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "证书ID"
// @Success 200 {object} util.Response{data=DownloadCertificateResponse} "成功返回证书内容"
// @Failure 400 {object} util.Response "无效的证书ID或证书状态无效"
// @Failure 404 {object} util.Response "证书不存在"
// @Router /api/certificates/{id}/download [get]
func (h *CertificateHandler) DownloadCertificate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		util.Error(c, http.StatusBadRequest, "无效的证书ID")
		return
	}

	cert, err := h.certRepo.FindByID(uint(id))
	if err != nil {
		util.Error(c, http.StatusNotFound, "证书不存在")
		return
	}

	if cert.Status != "active" && cert.Status != "expiring" {
		util.Error(c, http.StatusBadRequest, "证书状态无效，无法下载")
		return
	}

	if cert.CertPEM == "" {
		util.Error(c, http.StatusBadRequest, "证书内容为空")
		return
	}

	// 构建完整证书链
	fullChain := cert.CertPEM
	if cert.IssuerCertPEM != "" {
		fullChain = cert.CertPEM + "\n" + cert.IssuerCertPEM
	}

	// 记录操作日志
	userID, _ := c.Get("user_id")
	h.logService.CreateLogAsync(userID.(uint), "download", "certificate", uint(id),
		fmt.Sprintf("下载证书: %s", cert.Domain), c.ClientIP())

	util.Success(c, DownloadCertificateResponse{
		Domain:        cert.Domain,
		CertPem:       cert.CertPEM,
		IssuerCertPem: cert.IssuerCertPEM,
		FullChainPem:  fullChain,
	})
}
