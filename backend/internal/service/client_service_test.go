/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-18 12:15:00
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-11-18 13:34:40
 * @FilePath            : frp-web-testbackendinternalserviceclient_service_test.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package service

import (
	"frp-web-panel/internal/model"
	"frp-web-panel/internal/repository"
	"frp-web-panel/pkg/database"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupServiceTestDB(t *testing.T) {
	var err error
	database.DB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = database.DB.AutoMigrate(&model.Client{}, &model.Proxy{})
	assert.NoError(t, err)
}

func TestClientService_GetClients(t *testing.T) {
	setupServiceTestDB(t)
	service := NewClientService()

	repo := repository.NewClientRepository()
	repo.Create(&model.Client{Name: "client1", ServerAddr: "127.0.0.1", ServerPort: 7000})
	repo.Create(&model.Client{Name: "client2", ServerAddr: "127.0.0.1", ServerPort: 7000})

	clients, total, err := service.GetClients(1, 10, "")
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, clients, 2)
}

func TestClientService_GetClientByID(t *testing.T) {
	setupServiceTestDB(t)
	service := NewClientService()

	repo := repository.NewClientRepository()
	client := &model.Client{Name: "test-client", ServerAddr: "127.0.0.1", ServerPort: 7000}
	repo.Create(client)

	found, err := service.GetClient(client.ID)
	assert.NoError(t, err)
	assert.Equal(t, "test-client", found.Name)
}

func TestClientService_CreateClient(t *testing.T) {
	setupServiceTestDB(t)
	service := NewClientService()

	client := &model.Client{
		Name:       "new-client",
		ServerAddr: "127.0.0.1",
		ServerPort: 7000,
	}

	err := service.CreateClient(client)
	assert.NoError(t, err)
	assert.NotZero(t, client.ID)
}

func TestClientService_UpdateClient(t *testing.T) {
	setupServiceTestDB(t)
	service := NewClientService()

	repo := repository.NewClientRepository()
	client := &model.Client{Name: "test-client", ServerAddr: "127.0.0.1", ServerPort: 7000}
	repo.Create(client)

	client.Name = "updated-client"
	err := service.UpdateClient(client)
	assert.NoError(t, err)

	found, _ := service.GetClient(client.ID)
	assert.Equal(t, "updated-client", found.Name)
}

func TestClientService_DeleteClient(t *testing.T) {
	setupServiceTestDB(t)
	service := NewClientService()

	repo := repository.NewClientRepository()
	client := &model.Client{Name: "test-client", ServerAddr: "127.0.0.1", ServerPort: 7000}
	repo.Create(client)

	err := service.DeleteClient(client.ID)
	assert.NoError(t, err)

	_, err = service.GetClient(client.ID)
	assert.Error(t, err)
}
