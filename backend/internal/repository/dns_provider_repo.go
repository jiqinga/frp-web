/*
 * @Author              : 寂情啊
 * @Date                : 2025-12-22 15:45:32
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-22 15:45:44
 * @FilePath            : frp-web-testbackendinternalrepositorydns_provider_repo.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package repository

import (
	"frp-web-panel/internal/model"
	"frp-web-panel/pkg/database"
)

type DNSProviderRepository struct{}

func NewDNSProviderRepository() *DNSProviderRepository {
	return &DNSProviderRepository{}
}

func (r *DNSProviderRepository) Create(provider *model.DNSProvider) error {
	return database.DB.Create(provider).Error
}

func (r *DNSProviderRepository) FindByID(id uint) (*model.DNSProvider, error) {
	var provider model.DNSProvider
	err := database.DB.First(&provider, id).Error
	return &provider, err
}

func (r *DNSProviderRepository) FindAll() ([]model.DNSProvider, error) {
	var providers []model.DNSProvider
	err := database.DB.Find(&providers).Error
	return providers, err
}

func (r *DNSProviderRepository) FindEnabled() ([]model.DNSProvider, error) {
	var providers []model.DNSProvider
	err := database.DB.Where("enabled = ?", true).Find(&providers).Error
	return providers, err
}

func (r *DNSProviderRepository) Update(provider *model.DNSProvider) error {
	return database.DB.Save(provider).Error
}

func (r *DNSProviderRepository) Delete(id uint) error {
	return database.DB.Delete(&model.DNSProvider{}, id).Error
}
