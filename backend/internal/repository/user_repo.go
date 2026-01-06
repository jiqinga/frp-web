/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-14 15:27:48
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-03 15:56:36
 * @FilePath            : frp-web-testbackendinternalrepositoryuser_repo.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package repository

import (
	"frp-web-panel/internal/model"
	"frp-web-panel/pkg/database"
)

type UserRepository struct{}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

func (r *UserRepository) FindByUsername(username string) (*model.User, error) {
	var user model.User
	err := database.DB.Where("username = ?", username).First(&user).Error
	return &user, err
}

func (r *UserRepository) FindByID(id uint) (*model.User, error) {
	var user model.User
	err := database.DB.First(&user, id).Error
	return &user, err
}

func (r *UserRepository) Create(user *model.User) error {
	return database.DB.Create(user).Error
}

// UpdatePassword 更新用户密码
func (r *UserRepository) UpdatePassword(id uint, hashedPassword string) error {
	return database.DB.Model(&model.User{}).Where("id = ?", id).Update("password", hashedPassword).Error
}
