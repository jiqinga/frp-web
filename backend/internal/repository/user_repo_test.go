/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-18 12:13:36
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-11-18 12:13:46
 * @FilePath            : frp-web-testbackendinternalrepositoryuser_repo_test.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package repository

import (
	"frp-web-panel/internal/model"
	"frp-web-panel/pkg/database"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) {
	var err error
	database.DB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = database.DB.AutoMigrate(&model.User{})
	assert.NoError(t, err)
}

func TestUserRepository_Create(t *testing.T) {
	setupTestDB(t)
	repo := NewUserRepository()

	user := &model.User{
		Username: "testuser",
		Password: "hashedpassword",
	}

	err := repo.Create(user)
	assert.NoError(t, err)
	assert.NotZero(t, user.ID)
}

func TestUserRepository_FindByUsername(t *testing.T) {
	setupTestDB(t)
	repo := NewUserRepository()

	user := &model.User{
		Username: "testuser",
		Password: "hashedpassword",
	}
	repo.Create(user)

	found, err := repo.FindByUsername("testuser")
	assert.NoError(t, err)
	assert.Equal(t, "testuser", found.Username)
}

func TestUserRepository_FindByID(t *testing.T) {
	setupTestDB(t)
	repo := NewUserRepository()

	user := &model.User{
		Username: "testuser",
		Password: "hashedpassword",
	}
	repo.Create(user)

	found, err := repo.FindByID(user.ID)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
}
