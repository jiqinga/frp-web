/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-18 11:09:52
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-11-18 15:16:58
 * @FilePath            : frp-web-testbackendinternalserviceauth_service_test.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package service

import (
	"testing"

	"frp-web-panel/internal/config"
	"frp-web-panel/internal/model"
	"frp-web-panel/pkg/database"

	"golang.org/x/crypto/bcrypt"
)

func setupTestDB(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type:   "sqlite",
			SQLite: config.SQLiteConfig{Path: ":memory:"},
		},
		JWT: config.JWTConfig{
			Secret:      "test-secret-key",
			ExpireHours: 24,
		},
	}
	database.InitDB(cfg)
	config.GlobalConfig = cfg
}

func TestLogin(t *testing.T) {
	setupTestDB(t)
	database.DB.AutoMigrate(&model.User{})

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("test123"), bcrypt.DefaultCost)
	user := &model.User{
		Username: "testuser",
		Password: string(hashedPassword),
		Role:     "admin",
	}
	database.DB.Create(user)

	authService := NewAuthService()

	tests := []struct {
		name     string
		username string
		password string
		wantErr  bool
	}{
		{"valid credentials", "testuser", "test123", false},
		{"invalid password", "testuser", "wrong", true},
		{"invalid username", "nouser", "test123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := authService.Login(tt.username, tt.password, "127.0.0.1")
			if (err != nil) != tt.wantErr {
				t.Errorf("Login() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
