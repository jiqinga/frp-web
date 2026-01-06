/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-14 15:21:43
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-29 16:50:01
 * @FilePath            : frp-web-testbackendinternalconfigconfig.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Security SecurityConfig `mapstructure:"security"`
	Frps     FrpsConfig     `mapstructure:"frps"`
	FRP      FRPConfig      `mapstructure:"frp"`
	Log      LogConfig      `mapstructure:"log"`
}

type ServerConfig struct {
	Port      int    `mapstructure:"port"`
	Mode      string `mapstructure:"mode"`
	PublicURL string `mapstructure:"public_url"`
}

type DatabaseConfig struct {
	Type     string         `mapstructure:"type"`
	SQLite   SQLiteConfig   `mapstructure:"sqlite"`
	Postgres PostgresConfig `mapstructure:"postgres"`
}

type SQLiteConfig struct {
	Path string `mapstructure:"path"`
}

type PostgresConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
}

type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	ExpireHours int    `mapstructure:"expire_hours"`
}

type SecurityConfig struct {
	EncryptionKey string `mapstructure:"encryption_key"`
}

type FrpsConfig struct {
	ConfigPath     string `mapstructure:"config_path"`
	BinaryDir      string `mapstructure:"binary_dir"`
	ConfigDir      string `mapstructure:"config_dir"`
	LogDir         string `mapstructure:"log_dir"`
	DefaultVersion string `mapstructure:"default_version"`
	GithubAPI      string `mapstructure:"github_api"`
}

type FRPConfig struct {
	Client FRPClientConfig `mapstructure:"client"`
}

type FRPClientConfig struct {
	Timeout       string `mapstructure:"timeout"`
	RetryTimes    int    `mapstructure:"retry_times"`
	RetryInterval string `mapstructure:"retry_interval"`
}

type LogConfig struct {
	Level string `mapstructure:"level"`
	File  string `mapstructure:"file"`
}

var GlobalConfig *Config

func LoadConfig(path string) error {
	viper.SetConfigFile(path)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	GlobalConfig = &Config{}
	return viper.Unmarshal(GlobalConfig)
}

// 默认敏感配置值（用于检测是否使用默认值）
var defaultSensitiveValues = map[string]string{
	"jwt_secret":     "your-secret-key-change-in-production",
	"encryption_key": "12345678901234567890123456789012",
}

// Validate 验证配置的有效性
func (c *Config) Validate() error {
	var errs []string

	// 验证服务器配置
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		errs = append(errs, fmt.Sprintf("server.port must be between 1 and 65535, got %d", c.Server.Port))
	}

	// 验证数据库配置
	if c.Database.Type == "" {
		errs = append(errs, "database.type is required")
	} else if c.Database.Type == "sqlite" {
		if c.Database.SQLite.Path == "" {
			errs = append(errs, "database.sqlite.path is required when using sqlite")
		} else {
			// 检查目录是否存在或可创建
			dir := filepath.Dir(c.Database.SQLite.Path)
			if dir != "." && dir != "" {
				if _, err := os.Stat(dir); os.IsNotExist(err) {
					if err := os.MkdirAll(dir, 0755); err != nil {
						errs = append(errs, fmt.Sprintf("database.sqlite.path directory cannot be created: %s", dir))
					}
				}
			}
		}
	} else if c.Database.Type == "postgres" {
		if c.Database.Postgres.Host == "" {
			errs = append(errs, "database.postgres.host is required when using postgres")
		}
		if c.Database.Postgres.Port < 1 || c.Database.Postgres.Port > 65535 {
			errs = append(errs, fmt.Sprintf("database.postgres.port must be between 1 and 65535, got %d", c.Database.Postgres.Port))
		}
		if c.Database.Postgres.DBName == "" {
			errs = append(errs, "database.postgres.dbname is required when using postgres")
		}
	}

	// 验证 JWT 配置
	if c.JWT.Secret == "" {
		errs = append(errs, "jwt.secret is required")
	}
	if c.JWT.ExpireHours <= 0 {
		errs = append(errs, "jwt.expire_hours must be positive")
	}

	// 验证安全配置
	if c.Security.EncryptionKey == "" {
		errs = append(errs, "security.encryption_key is required")
	} else if len(c.Security.EncryptionKey) != 32 {
		errs = append(errs, "security.encryption_key must be exactly 32 characters for AES-256")
	}

	// 敏感配置默认值警告
	c.warnDefaultSensitiveValues()

	if len(errs) > 0 {
		return errors.New("配置验证失败:\n  - " + strings.Join(errs, "\n  - "))
	}
	return nil
}

// warnDefaultSensitiveValues 检查敏感配置是否使用默认值并输出警告
func (c *Config) warnDefaultSensitiveValues() {
	if c.JWT.Secret == defaultSensitiveValues["jwt_secret"] {
		log.Println("[WARNING] jwt.secret is using default value, please change it in production!")
	}
	if c.Security.EncryptionKey == defaultSensitiveValues["encryption_key"] {
		log.Println("[WARNING] security.encryption_key is using default value, please change it in production!")
	}
}
