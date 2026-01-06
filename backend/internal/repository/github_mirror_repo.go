/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-21 14:00:48
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-11-21 14:01:00
 * @FilePath            : frp-web-testbackendinternalrepositorygithub_mirror_repo.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package repository

import (
	"frp-web-panel/internal/model"
	"frp-web-panel/pkg/database"

	"gorm.io/gorm"
)

type GithubMirrorRepository struct{}

func NewGithubMirrorRepository() *GithubMirrorRepository {
	return &GithubMirrorRepository{}
}

func (r *GithubMirrorRepository) GetAll() ([]model.GithubMirror, error) {
	var mirrors []model.GithubMirror
	err := database.DB.Order("is_default DESC, id ASC").Find(&mirrors).Error
	return mirrors, err
}

func (r *GithubMirrorRepository) GetByID(id uint) (*model.GithubMirror, error) {
	var mirror model.GithubMirror
	err := database.DB.First(&mirror, id).Error
	return &mirror, err
}

func (r *GithubMirrorRepository) GetDefault() (*model.GithubMirror, error) {
	var mirror model.GithubMirror
	err := database.DB.Where("is_default = ? AND enabled = ?", true, true).First(&mirror).Error
	return &mirror, err
}

func (r *GithubMirrorRepository) Create(mirror *model.GithubMirror) error {
	return database.DB.Create(mirror).Error
}

func (r *GithubMirrorRepository) Update(mirror *model.GithubMirror) error {
	return database.DB.Save(mirror).Error
}

func (r *GithubMirrorRepository) Delete(id uint) error {
	return database.DB.Delete(&model.GithubMirror{}, id).Error
}

func (r *GithubMirrorRepository) SetDefault(id uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.GithubMirror{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
			return err
		}
		return tx.Model(&model.GithubMirror{}).Where("id = ?", id).Update("is_default", true).Error
	})
}
