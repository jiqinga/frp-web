/*
 * @Author              : 寂情啊
 * @Date                : 2025-12-23 13:55:06
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-24 17:19:30
 * @FilePath            : frp-web-testbackendinternalrepositorycertificate_repo.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package repository

import (
	"frp-web-panel/internal/model"
	"frp-web-panel/pkg/database"

	"gorm.io/gorm"
)

type CertificateRepository struct {
	db *gorm.DB
}

func NewCertificateRepository() *CertificateRepository {
	return &CertificateRepository{db: database.DB}
}

func (r *CertificateRepository) Create(cert *model.Certificate) error {
	return r.db.Create(cert).Error
}

func (r *CertificateRepository) Update(cert *model.Certificate) error {
	return r.db.Save(cert).Error
}

func (r *CertificateRepository) FindByID(id uint) (*model.Certificate, error) {
	var cert model.Certificate
	err := r.db.First(&cert, id).Error
	return &cert, err
}

func (r *CertificateRepository) FindByProxyID(proxyID uint) (*model.Certificate, error) {
	var cert model.Certificate
	err := r.db.Where("proxy_id = ?", proxyID).First(&cert).Error
	return &cert, err
}

func (r *CertificateRepository) FindByDomain(domain string) (*model.Certificate, error) {
	var cert model.Certificate
	err := r.db.Where("domain = ?", domain).First(&cert).Error
	return &cert, err
}

// FindActiveCertificates 获取所有有效证书
func (r *CertificateRepository) FindActiveCertificates() ([]model.Certificate, error) {
	var certs []model.Certificate
	err := r.db.Where("status = ?", model.CertStatusActive).Find(&certs).Error
	return certs, err
}

// FindByDomainPattern 根据域名模式匹配证书（支持通配符）
func (r *CertificateRepository) FindByDomainPattern(domain string) ([]model.Certificate, error) {
	var certs []model.Certificate
	// 精确匹配或通配符匹配
	// 例如：domain = "app.example.com" 可以匹配 "app.example.com" 或 "*.example.com"
	err := r.db.Where("status = ? AND (domain = ? OR domain = ?)",
		model.CertStatusActive, domain, "*."+getParentDomain(domain)).Find(&certs).Error
	return certs, err
}

// getParentDomain 获取父域名（去掉第一级子域名）
func getParentDomain(domain string) string {
	parts := splitDomain(domain)
	if len(parts) <= 2 {
		return domain
	}
	return joinDomain(parts[1:])
}

func splitDomain(domain string) []string {
	var parts []string
	current := ""
	for _, c := range domain {
		if c == '.' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

func joinDomain(parts []string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += "."
		}
		result += p
	}
	return result
}

func (r *CertificateRepository) FindExpiring() ([]model.Certificate, error) {
	var certs []model.Certificate
	err := r.db.Where("status = ? AND auto_renew = ?", model.CertStatusExpiring, true).Find(&certs).Error
	return certs, err
}

func (r *CertificateRepository) FindAll() ([]model.Certificate, error) {
	var certs []model.Certificate
	err := r.db.Find(&certs).Error
	return certs, err
}

func (r *CertificateRepository) Delete(id uint) error {
	return r.db.Delete(&model.Certificate{}, id).Error
}

func (r *CertificateRepository) DeleteByProxyID(proxyID uint) error {
	return r.db.Where("proxy_id = ?", proxyID).Delete(&model.Certificate{}).Error
}
