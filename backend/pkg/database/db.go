/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-14 15:22:11
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2026-01-07 14:09:35
 * @FilePath            : frp-web-testbackendpkgdatabasedb.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package database

import (
	"fmt"
	"frp-web-panel/internal/config"
	"frp-web-panel/internal/logger"
	"frp-web-panel/internal/model"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// zapWriter 适配 zap logger 到 GORM 的 Printf 接口
type zapWriter struct{}

func (w *zapWriter) Printf(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

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
		// SQLite 连接字符串添加 PRAGMA 优化参数
		dsn := fmt.Sprintf("%s?_journal_mode=WAL&_synchronous=NORMAL&_cache_size=-64000&_busy_timeout=5000&_foreign_keys=ON",
			cfg.Database.SQLite.Path)
		dialector = sqlite.Open(dsn)
	}

	// 根据日志级别配置 GORM Logger
	var gormLogLevel gormlogger.LogLevel
	if strings.ToLower(cfg.Log.Level) == "debug" {
		gormLogLevel = gormlogger.Info
	} else {
		gormLogLevel = gormlogger.Silent
	}

	gormLogger := gormlogger.New(
		&zapWriter{},
		gormlogger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  gormLogLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	var err error
	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return err
	}

	// 配置连接池
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

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
		logger.Warnf("创建默认管理员失败: %v", err)
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
			logger.Warnf("创建默认GitHub镜像失败: %v", err)
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
				logger.Warnf("创建默认设置 %s 失败: %v", setting.Key, err)
			}
		}
	}

	return nil
}
