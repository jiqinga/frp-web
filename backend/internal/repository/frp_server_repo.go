/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-19 17:09:56
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-11-19 17:24:41
 * @FilePath            : frp-web-testbackendinternalrepositoryfrp_server_repo.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package repository

import (
	"frp-web-panel/internal/model"

	"gorm.io/gorm"
)

type FrpServerRepository struct {
	db *gorm.DB
}

func NewFrpServerRepository(db *gorm.DB) *FrpServerRepository {
	return &FrpServerRepository{db: db}
}

func (r *FrpServerRepository) Create(server *model.FrpServer) error {
	return r.db.Create(server).Error
}

func (r *FrpServerRepository) Update(server *model.FrpServer) error {
	return r.db.Save(server).Error
}

func (r *FrpServerRepository) Delete(id uint) error {
	return r.db.Delete(&model.FrpServer{}, id).Error
}

func (r *FrpServerRepository) GetByID(id uint) (*model.FrpServer, error) {
	var server model.FrpServer
	if err := r.db.First(&server, id).Error; err != nil {
		return nil, err
	}
	return &server, nil
}

func (r *FrpServerRepository) GetAll() ([]model.FrpServer, error) {
	var servers []model.FrpServer
	if err := r.db.Find(&servers).Error; err != nil {
		return nil, err
	}
	return servers, nil
}

func (r *FrpServerRepository) GetEnabled() ([]model.FrpServer, error) {
	var servers []model.FrpServer
	if err := r.db.Where("enabled = ?", true).Find(&servers).Error; err != nil {
		return nil, err
	}
	return servers, nil
}
