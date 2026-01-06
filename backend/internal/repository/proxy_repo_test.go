package repository

import (
	"frp-web-panel/internal/model"
	"frp-web-panel/pkg/database"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupProxyTestDB(t *testing.T) {
	var err error
	database.DB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = database.DB.AutoMigrate(&model.Proxy{}, &model.Client{})
	assert.NoError(t, err)
}

func TestProxyRepository_Create(t *testing.T) {
	setupProxyTestDB(t)
	repo := NewProxyRepository()

	proxy := &model.Proxy{
		Name:      "test-proxy",
		Type:      "tcp",
		LocalIP:   "127.0.0.1",
		LocalPort: 22,
		ClientID:  1,
	}

	err := repo.Create(proxy)
	assert.NoError(t, err)
	assert.NotZero(t, proxy.ID)
}

func TestProxyRepository_FindByClientID(t *testing.T) {
	setupProxyTestDB(t)
	repo := NewProxyRepository()

	repo.Create(&model.Proxy{Name: "proxy1", Type: "tcp", LocalIP: "127.0.0.1", LocalPort: 22, ClientID: 1})
	repo.Create(&model.Proxy{Name: "proxy2", Type: "http", LocalIP: "127.0.0.1", LocalPort: 80, ClientID: 1})

	proxies, err := repo.FindByClientID(1)
	assert.NoError(t, err)
	assert.Len(t, proxies, 2)
}

func TestProxyRepository_FindByID(t *testing.T) {
	setupProxyTestDB(t)
	repo := NewProxyRepository()

	proxy := &model.Proxy{
		Name:      "test-proxy",
		Type:      "tcp",
		LocalIP:   "127.0.0.1",
		LocalPort: 22,
		ClientID:  1,
	}
	repo.Create(proxy)

	found, err := repo.FindByID(proxy.ID)
	assert.NoError(t, err)
	assert.Equal(t, "test-proxy", found.Name)
}

func TestProxyRepository_Update(t *testing.T) {
	setupProxyTestDB(t)
	repo := NewProxyRepository()

	proxy := &model.Proxy{
		Name:      "test-proxy",
		Type:      "tcp",
		LocalIP:   "127.0.0.1",
		LocalPort: 22,
		ClientID:  1,
	}
	repo.Create(proxy)

	proxy.Name = "updated-proxy"
	err := repo.Update(proxy)
	assert.NoError(t, err)

	found, _ := repo.FindByID(proxy.ID)
	assert.Equal(t, "updated-proxy", found.Name)
}

func TestProxyRepository_Delete(t *testing.T) {
	setupProxyTestDB(t)
	repo := NewProxyRepository()

	proxy := &model.Proxy{
		Name:      "test-proxy",
		Type:      "tcp",
		LocalIP:   "127.0.0.1",
		LocalPort: 22,
		ClientID:  1,
	}
	repo.Create(proxy)

	err := repo.Delete(proxy.ID)
	assert.NoError(t, err)

	_, err = repo.FindByID(proxy.ID)
	assert.Error(t, err)
}

func TestProxyRepository_GetProxyTypeStats(t *testing.T) {
	setupProxyTestDB(t)
	repo := NewProxyRepository()

	repo.Create(&model.Proxy{Name: "proxy1", Type: "tcp", LocalIP: "127.0.0.1", LocalPort: 22, ClientID: 1})
	repo.Create(&model.Proxy{Name: "proxy2", Type: "tcp", LocalIP: "127.0.0.1", LocalPort: 23, ClientID: 1})
	repo.Create(&model.Proxy{Name: "proxy3", Type: "http", LocalIP: "127.0.0.1", LocalPort: 80, ClientID: 1})

	stats, err := repo.GetProxyTypeStats()
	assert.NoError(t, err)
	assert.Equal(t, int64(2), stats["tcp"])
	assert.Equal(t, int64(1), stats["http"])
}
