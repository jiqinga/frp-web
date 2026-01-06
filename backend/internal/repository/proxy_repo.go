/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-14 15:30:41
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-25 16:55:17
 * @FilePath            : frp-web-testbackendinternalrepositoryproxy_repo.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package repository

import (
	"frp-web-panel/internal/model"
	"frp-web-panel/pkg/database"
)

type ProxyRepository struct{}

func NewProxyRepository() *ProxyRepository {
	return &ProxyRepository{}
}

func (r *ProxyRepository) FindByClientID(clientID uint) ([]model.Proxy, error) {
	var proxies []model.Proxy
	err := database.DB.Where("client_id = ?", clientID).Find(&proxies).Error
	return proxies, err
}

func (r *ProxyRepository) FindByID(id uint) (*model.Proxy, error) {
	var proxy model.Proxy
	err := database.DB.First(&proxy, id).Error
	return &proxy, err
}

func (r *ProxyRepository) Create(proxy *model.Proxy) error {
	return database.DB.Create(proxy).Error
}

func (r *ProxyRepository) Update(proxy *model.Proxy) error {
	return database.DB.Save(proxy).Error
}

func (r *ProxyRepository) Delete(id uint) error {
	return database.DB.Delete(&model.Proxy{}, id).Error
}

func (r *ProxyRepository) Count() (int64, error) {
	var count int64
	err := database.DB.Model(&model.Proxy{}).Count(&count).Error
	return count, err
}

func (r *ProxyRepository) GetProxyTypeStats() (map[string]int64, error) {
	var results []struct {
		Type  string
		Count int64
	}
	err := database.DB.Model(&model.Proxy{}).
		Select("type, COUNT(*) as count").
		Group("type").
		Scan(&results).Error

	stats := make(map[string]int64)
	for _, r := range results {
		stats[r.Type] = r.Count
	}
	return stats, err
}

func (r *ProxyRepository) FindAll() ([]model.Proxy, error) {
	var proxies []model.Proxy
	err := database.DB.Find(&proxies).Error
	return proxies, err
}

func (r *ProxyRepository) GetByName(name string) (*model.Proxy, error) {
	var proxy model.Proxy
	err := database.DB.Where("name = ?", name).First(&proxy).Error
	return &proxy, err
}

// GetUsedRemotePorts 获取所有已使用的远程端口列表
func (r *ProxyRepository) GetUsedRemotePorts() ([]int, error) {
	var ports []int
	err := database.DB.Model(&model.Proxy{}).
		Where("remote_port > 0").
		Pluck("remote_port", &ports).Error
	return ports, err
}

// ToggleEnabled 切换代理的启用/禁用状态
func (r *ProxyRepository) ToggleEnabled(id uint) (*model.Proxy, error) {
	var proxy model.Proxy
	if err := database.DB.First(&proxy, id).Error; err != nil {
		return nil, err
	}
	proxy.Enabled = !proxy.Enabled
	if err := database.DB.Save(&proxy).Error; err != nil {
		return nil, err
	}
	return &proxy, nil
}

// FindEnabledByClientID 获取指定客户端的所有启用的代理
func (r *ProxyRepository) FindEnabledByClientID(clientID uint) ([]model.Proxy, error) {
	var proxies []model.Proxy
	err := database.DB.Where("client_id = ? AND enabled = ?", clientID, true).Find(&proxies).Error
	return proxies, err
}

// CountByCertIDAndClientID 统计同一客户端中使用指定证书的代理数量（排除指定代理）
func (r *ProxyRepository) CountByCertIDAndClientID(certID uint, clientID uint, excludeProxyID uint) (int64, error) {
	var count int64
	err := database.DB.Model(&model.Proxy{}).
		Where("cert_id = ? AND client_id = ? AND id != ?", certID, clientID, excludeProxyID).
		Count(&count).Error
	return count, err
}
