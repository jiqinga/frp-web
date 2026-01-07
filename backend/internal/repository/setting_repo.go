/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-14 16:17:57
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2026-01-07 14:19:33
 * @FilePath            : frp-web-testbackendinternalrepositorysetting_repo.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package repository

import (
	"frp-web-panel/internal/model"
	"frp-web-panel/pkg/database"
	"sync"
	"time"
)

// settingCache 设置缓存
var settingCache sync.Map

// cacheEntry 缓存条目
type cacheEntry struct {
	value     string
	expiredAt time.Time
}

const settingCacheTTL = 5 * time.Minute

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
	// 更新缓存
	for _, s := range settings {
		settingCache.Store(s.Key, &cacheEntry{value: s.Value, expiredAt: time.Now().Add(settingCacheTTL)})
	}
	return settings, nil
}

func (r *SettingRepository) UpdateSetting(key, value string) error {
	db := database.DB
	var setting model.Setting
	err := db.Where("key = ?", key).First(&setting).Error
	if err != nil {
		setting = model.Setting{Key: key, Value: value}
		if err := db.Create(&setting).Error; err != nil {
			return err
		}
	} else {
		if err := db.Model(&setting).Update("value", value).Error; err != nil {
			return err
		}
	}
	// 更新缓存
	settingCache.Store(key, &cacheEntry{value: value, expiredAt: time.Now().Add(settingCacheTTL)})
	return nil
}

func (r *SettingRepository) GetSetting(key string) (string, error) {
	// 先查缓存
	if entry, ok := settingCache.Load(key); ok {
		e := entry.(*cacheEntry)
		if time.Now().Before(e.expiredAt) {
			return e.value, nil
		}
		settingCache.Delete(key)
	}
	// 缓存未命中，查数据库
	db := database.DB
	var setting model.Setting
	if err := db.Where("key = ?", key).First(&setting).Error; err != nil {
		return "", err
	}
	// 写入缓存
	settingCache.Store(key, &cacheEntry{value: setting.Value, expiredAt: time.Now().Add(settingCacheTTL)})
	return setting.Value, nil
}

func (r *SettingRepository) GetOrCreate(key, defaultValue, description string) (string, error) {
	// 先查缓存
	if entry, ok := settingCache.Load(key); ok {
		e := entry.(*cacheEntry)
		if time.Now().Before(e.expiredAt) {
			return e.value, nil
		}
		settingCache.Delete(key)
	}
	// 缓存未命中
	db := database.DB
	var setting model.Setting
	err := db.Where("key = ?", key).First(&setting).Error
	if err != nil {
		setting = model.Setting{Key: key, Value: defaultValue, Description: description}
		if err := db.Create(&setting).Error; err != nil {
			return "", err
		}
		settingCache.Store(key, &cacheEntry{value: defaultValue, expiredAt: time.Now().Add(settingCacheTTL)})
		return defaultValue, nil
	}
	settingCache.Store(key, &cacheEntry{value: setting.Value, expiredAt: time.Now().Add(settingCacheTTL)})
	return setting.Value, nil
}
