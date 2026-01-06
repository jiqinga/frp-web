/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-14 15:30:09
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-26 15:18:40
 * @FilePath            : frp-web-testbackendinternalrepositoryclient_repo.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package repository

import (
	"frp-web-panel/internal/model"
	"frp-web-panel/pkg/database"
	"log"
	"time"
)

type ClientRepository struct{}

func NewClientRepository() *ClientRepository {
	return &ClientRepository{}
}

func (r *ClientRepository) FindAll(page, pageSize int, keyword string) ([]model.Client, int64, error) {
	var clients []model.Client
	var total int64

	query := database.DB.Model(&model.Client{})
	if keyword != "" {
		query = query.Where("name LIKE ? OR remark LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Preload("Proxies").Find(&clients).Error

	return clients, total, err
}

func (r *ClientRepository) FindByID(id uint) (*model.Client, error) {
	var client model.Client
	err := database.DB.Preload("Proxies").First(&client, id).Error
	return &client, err
}

func (r *ClientRepository) Create(client *model.Client) error {
	return database.DB.Create(client).Error
}

func (r *ClientRepository) Update(client *model.Client) error {
	return database.DB.Save(client).Error
}

func (r *ClientRepository) Delete(id uint) error {
	return database.DB.Delete(&model.Client{}, id).Error
}

func (r *ClientRepository) Count() (int64, error) {
	var count int64
	err := database.DB.Model(&model.Client{}).Count(&count).Error
	return count, err
}

func (r *ClientRepository) UpdateOnlineStatus(id uint, status string, heartbeat time.Time) error {
	return database.DB.Model(&model.Client{}).Where("id = ?", id).Updates(map[string]interface{}{
		"online_status":  status,
		"last_heartbeat": heartbeat,
	}).Error
}

func (r *ClientRepository) GetAllForStatusCheck() ([]model.Client, error) {
	var clients []model.Client
	err := database.DB.Find(&clients).Error
	return clients, err
}

// UpdateWSStatus 更新客户端WebSocket连接状态
func (r *ClientRepository) UpdateWSStatus(id uint, connected bool) error {
	log.Printf("[ClientRepo] 更新客户端 %d 的WS状态: %v", id, connected)
	err := database.DB.Model(&model.Client{}).Where("id = ?", id).Update("ws_connected", connected).Error
	if err != nil {
		log.Printf("[ClientRepo] ❌ 更新失败: %v", err)
	} else {
		log.Printf("[ClientRepo] ✅ 更新成功")
	}
	return err
}

// UpdateConfigSync 更新客户端配置同步信息
func (r *ClientRepository) UpdateConfigSync(id uint, version int, syncTime *time.Time) error {
	return database.DB.Model(&model.Client{}).Where("id = ?", id).Updates(map[string]interface{}{
		"config_version":   version,
		"last_config_sync": syncTime,
	}).Error
}

// UpdateVersionInfo 更新客户端版本信息
func (r *ClientRepository) UpdateVersionInfo(id uint, frpcVersion, daemonVersion, os, arch string) error {
	updates := map[string]interface{}{}
	if frpcVersion != "" {
		updates["frpc_version"] = frpcVersion
	}
	if daemonVersion != "" {
		updates["daemon_version"] = daemonVersion
	}
	if os != "" {
		updates["os"] = os
	}
	if arch != "" {
		updates["arch"] = arch
	}
	if len(updates) == 0 {
		return nil
	}
	log.Printf("[ClientRepo] 更新客户端 %d 版本信息: %+v", id, updates)
	return database.DB.Model(&model.Client{}).Where("id = ?", id).Updates(updates).Error
}

// FindByIDs 根据ID列表查询客户端
func (r *ClientRepository) FindByIDs(ids []uint) ([]model.Client, error) {
	var clients []model.Client
	err := database.DB.Where("id IN ?", ids).Find(&clients).Error
	return clients, err
}

// FindByFrpServerID 根据FRP服务器ID查询关联的客户端
func (r *ClientRepository) FindByFrpServerID(frpServerID uint) ([]model.Client, error) {
	var clients []model.Client
	err := database.DB.Where("frp_server_id = ?", frpServerID).Find(&clients).Error
	return clients, err
}

// ResetAllClientStatus 重置所有客户端状态为离线
func (r *ClientRepository) ResetAllClientStatus() error {
	return database.DB.Model(&model.Client{}).Where("1 = 1").Updates(map[string]interface{}{
		"online_status": "offline",
		"ws_connected":  false,
	}).Error
}

// UpdateConfigSyncStatus 更新客户端配置同步状态
func (r *ClientRepository) UpdateConfigSyncStatus(id uint, status string, errorMsg string, syncTime time.Time) error {
	log.Printf("[ClientRepo] 更新客户端 %d 配置同步状态: status=%s, error=%s", id, status, errorMsg)
	return database.DB.Model(&model.Client{}).Where("id = ?", id).Updates(map[string]interface{}{
		"config_sync_status": status,
		"config_sync_error":  errorMsg,
		"config_sync_time":   syncTime,
	}).Error
}
