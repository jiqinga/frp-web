/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-21 16:04:29
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-11-21 16:04:41
 * @FilePath            : frp-web-testbackendinternalrepositoryclient_register_token_repo.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package repository

import (
	"frp-web-panel/internal/model"
	"frp-web-panel/pkg/database"
	"time"
)

type ClientRegisterTokenRepository struct{}

func NewClientRegisterTokenRepository() *ClientRegisterTokenRepository {
	return &ClientRegisterTokenRepository{}
}

func (r *ClientRegisterTokenRepository) Create(token *model.ClientRegisterToken) error {
	return database.DB.Create(token).Error
}

func (r *ClientRegisterTokenRepository) FindByToken(token string) (*model.ClientRegisterToken, error) {
	var t model.ClientRegisterToken
	err := database.DB.Where("token = ?", token).First(&t).Error
	return &t, err
}

func (r *ClientRegisterTokenRepository) MarkAsUsed(id uint) error {
	now := time.Now()
	return database.DB.Model(&model.ClientRegisterToken{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"used":    true,
			"used_at": &now,
		}).Error
}

func (r *ClientRegisterTokenRepository) DeleteExpired() error {
	return database.DB.Where("expires_at < ? AND used = ?", time.Now(), false).
		Delete(&model.ClientRegisterToken{}).Error
}
