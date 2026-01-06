package repository

import (
	"frp-web-panel/internal/model"
	"frp-web-panel/pkg/database"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupClientTestDB(t *testing.T) {
	var err error
	database.DB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = database.DB.AutoMigrate(&model.Client{}, &model.Proxy{})
	assert.NoError(t, err)
}

func TestClientRepository_Create(t *testing.T) {
	setupClientTestDB(t)
	repo := NewClientRepository()

	client := &model.Client{
		Name:       "test-client",
		ServerAddr: "127.0.0.1",
		ServerPort: 7000,
		Token:      "test-token",
	}

	err := repo.Create(client)
	assert.NoError(t, err)
	assert.NotZero(t, client.ID)
}

func TestClientRepository_FindByID(t *testing.T) {
	setupClientTestDB(t)
	repo := NewClientRepository()

	client := &model.Client{
		Name:       "test-client",
		ServerAddr: "127.0.0.1",
		ServerPort: 7000,
	}
	repo.Create(client)

	found, err := repo.FindByID(client.ID)
	assert.NoError(t, err)
	assert.Equal(t, "test-client", found.Name)
}

func TestClientRepository_FindAll(t *testing.T) {
	setupClientTestDB(t)
	repo := NewClientRepository()

	repo.Create(&model.Client{Name: "client1", ServerAddr: "127.0.0.1", ServerPort: 7000})
	repo.Create(&model.Client{Name: "client2", ServerAddr: "127.0.0.1", ServerPort: 7000})

	clients, total, err := repo.FindAll(1, 10, "")
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, clients, 2)
}

func TestClientRepository_Update(t *testing.T) {
	setupClientTestDB(t)
	repo := NewClientRepository()

	client := &model.Client{
		Name:       "test-client",
		ServerAddr: "127.0.0.1",
		ServerPort: 7000,
	}
	repo.Create(client)

	client.Name = "updated-client"
	err := repo.Update(client)
	assert.NoError(t, err)

	found, _ := repo.FindByID(client.ID)
	assert.Equal(t, "updated-client", found.Name)
}

func TestClientRepository_Delete(t *testing.T) {
	setupClientTestDB(t)
	repo := NewClientRepository()

	client := &model.Client{
		Name:       "test-client",
		ServerAddr: "127.0.0.1",
		ServerPort: 7000,
	}
	repo.Create(client)

	err := repo.Delete(client.ID)
	assert.NoError(t, err)

	_, err = repo.FindByID(client.ID)
	assert.Error(t, err)
}
