/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-14 16:17:57
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-11 17:19:59
 * @FilePath            : frp-web-testbackendinternalrepositorysetting_repo.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package repository

import (
	"frp-web-panel/internal/model"
	"frp-web-panel/pkg/database"
)

type SettingRepository struct{}

func NewSettingRepository() *SettingRepository {
	return &SettingRepository{}
}

func (r *SettingRepository) GetAllSettings() ([]model.Setting, error) {
	db := database.DB
	var settings []model.Setting
	if err := db.Find(&settings).Error; err != nil {
		return nil, err
	}
	return settings, nil
}

func (r *SettingRepository) UpdateSetting(key, value string) error {
	db := database.DB
	var setting model.Setting
	err := db.Where("key = ?", key).First(&setting).Error
	if err != nil {
		// 不存在则创建
		setting = model.Setting{
			Key:   key,
			Value: value,
		}
		return db.Create(&setting).Error
	}
	// 存在则更新
	return db.Model(&setting).Update("value", value).Error
}

func (r *SettingRepository) GetSetting(key string) (string, error) {
	db := database.DB
	var setting model.Setting
	if err := db.Where("key = ?", key).First(&setting).Error; err != nil {
		return "", err
	}
	return setting.Value, nil
}

func (r *SettingRepository) GetOrCreate(key, defaultValue, description string) (string, error) {
	db := database.DB
	var setting model.Setting
	err := db.Where("key = ?", key).First(&setting).Error
	if err != nil {
		// 不存在则创建
		setting = model.Setting{
			Key:         key,
			Value:       defaultValue,
			Description: description,
		}
		if err := db.Create(&setting).Error; err != nil {
			return "", err
		}
		return defaultValue, nil
	}
	return setting.Value, nil
}
