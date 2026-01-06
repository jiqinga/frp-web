/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-14 15:22:11
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-29 16:49:23
 * @FilePath            : frp-web-testbackendpkgdatabasedb.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package database

import (
	"fmt"
	"frp-web-panel/internal/config"
	"frp-web-panel/internal/model"
	"log"

	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB(cfg *config.Config) error {
	var dialector gorm.Dialector

	if cfg.Database.Type == "postgres" {
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			cfg.Database.Postgres.Host,
			cfg.Database.Postgres.Port,
			cfg.Database.Postgres.User,
			cfg.Database.Postgres.Password,
			cfg.Database.Postgres.DBName,
		)
		dialector = postgres.Open(dsn)
	} else {
		dialector = sqlite.Open(cfg.Database.SQLite.Path)
	}

	var err error
	DB, err = gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return err
	}

	if err := autoMigrate(); err != nil {
		return err
	}

	if err := createDefaultAdmin(); err != nil {
		return err
	}

	if err := createDefaultGithubMirror(); err != nil {
		return err
	}

	return createDefaultSettings()
}

func autoMigrate() error {
	return DB.AutoMigrate(
		&model.User{},
		&model.Client{},
		&model.Proxy{},
		&model.OperationLog{},
		&model.Setting{},
		&model.AlertRule{},
		&model.AlertLog{},
		&model.FrpServer{},
		&model.GithubMirror{},
		&model.ClientRegisterToken{},
		&model.ServerMetricsHistory{},
		&model.ProxyMetricsHistory{},
		&model.AlertRecipient{},
		&model.AlertRecipientGroup{},
		&model.AlertGroupRecipient{},
		&model.DNSProvider{},
		&model.DNSRecord{},
		&model.Certificate{},
	)
}

func createDefaultAdmin() error {
	var count int64
	DB.Model(&model.User{}).Count(&count)
	if count > 0 {
		return nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := &model.User{
		Username: "admin",
		Password: string(hashedPassword),
		Nickname: "管理员",
		Role:     "admin",
	}

	if err := DB.Create(admin).Error; err != nil {
		log.Printf("警告: 创建默认管理员失败: %v", err)
	}

	return nil
}

func createDefaultGithubMirror() error {
	var count int64
	DB.Model(&model.GithubMirror{}).Count(&count)
	if count > 0 {
		return nil
	}

	mirrors := []model.GithubMirror{
		{
			Name:        "官方GitHub",
			BaseURL:     "https://github.com",
			IsDefault:   false,
			Enabled:     true,
			Description: "GitHub官方源",
		},
		{
			Name:        "加速源",
			BaseURL:     "https://xget.183321.xyz/gh",
			IsDefault:   true,
			Enabled:     true,
			Description: "GitHub加速源",
		},
	}

	for _, mirror := range mirrors {
		if err := DB.Create(&mirror).Error; err != nil {
			log.Printf("警告: 创建默认GitHub镜像失败: %v", err)
		}
	}

	return nil
}

func createDefaultSettings() error {
	// 使用 FirstOrCreate 确保每个设置项都存在
	settings := []model.Setting{
		{
			Key:         "server_status_check_interval",
			Value:       "10",
			Description: "服务器状态检查间隔(秒)",
		},
		{
			Key:         "public_url",
			Value:       "http://localhost:8080",
			Description: "公网访问地址(用于生成客户端注册脚本)",
		},
		{
			Key:         "traffic_interval",
			Value:       "30",
			Description: "流量采集间隔(秒)",
		},
		{
			Key:         "server_info_interval",
			Value:       "5",
			Description: "服务器信息采集间隔(秒)",
		},
		{
			Key:         "proxy_status_interval",
			Value:       "10",
			Description: "代理状态采集间隔(秒)",
		},
		{
			Key:         "client_check_interval",
			Value:       "15",
			Description: "客户端检查间隔(秒)",
		},
	}

	for _, setting := range settings {
		var existing model.Setting
		if err := DB.Where("key = ?", setting.Key).First(&existing).Error; err != nil {
			// 不存在则创建
			if err := DB.Create(&setting).Error; err != nil {
				log.Printf("警告: 创建默认设置 %s 失败: %v", setting.Key, err)
			}
		}
	}

	return nil
}
