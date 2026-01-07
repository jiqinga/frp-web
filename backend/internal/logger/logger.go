/*
 * @Author              : 寂情啊
 * @Date                : 2026-01-07 10:43:31
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2026-01-07 14:50:57
 * @FilePath            : frp-web-testbackendinternalloggerlogger.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package logger

import (
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ANSI 颜色码
const (
	colorReset   = "\033[0m"
	colorCyan    = "\033[36m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorMagenta = "\033[35m"
)

var (
	L *zap.Logger
	S *zap.SugaredLogger
)

// cyanTimeEncoder 青色时间编码器
func cyanTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(colorCyan + t.Format("2006-01-02T15:04:05.000Z0700") + colorReset)
}

// Init 初始化日志模块
// level: debug, info, warn, error
// format: console, json
func Init(level, format string) error {
	// 环境变量优先
	if envLevel := os.Getenv("LOG_LEVEL"); envLevel != "" {
		level = envLevel
	}

	// 解析日志等级
	var zapLevel zapcore.Level
	switch strings.ToLower(level) {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	// 构建配置
	if strings.ToLower(format) == "json" {
		cfg := zap.NewProductionConfig()
		cfg.Level = zap.NewAtomicLevelAt(zapLevel)
		var err error
		L, err = cfg.Build(zap.AddCallerSkip(1))
		if err != nil {
			return err
		}
	} else {
		// 自定义彩色 console encoder
		encoderConfig := zapcore.EncoderConfig{
			TimeKey:          "T",
			LevelKey:         "L",
			MessageKey:       "M",
			EncodeTime:       cyanTimeEncoder,
			EncodeLevel:      zapcore.CapitalColorLevelEncoder,
			ConsoleSeparator: "\t",
		}
		core := zapcore.NewCore(
			newColoredMessageEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			zapLevel,
		)
		L = zap.New(core, zap.AddCallerSkip(1))
	}
	S = L.Sugar()
	return nil
}

// Sync 刷新日志缓冲
func Sync() {
	if L != nil {
		_ = L.Sync()
	}
}

// 便捷函数
func Debug(msg string, fields ...zap.Field) { L.Debug(msg, fields...) }
func Info(msg string, fields ...zap.Field)  { L.Info(msg, fields...) }
func Warn(msg string, fields ...zap.Field)  { L.Warn(msg, fields...) }
func Error(msg string, fields ...zap.Field) { L.Error(msg, fields...) }

func Debugf(template string, args ...interface{}) { S.Debugf(template, args...) }
func Infof(template string, args ...interface{})  { S.Infof(template, args...) }
func Warnf(template string, args ...interface{})  { S.Warnf(template, args...) }
func Errorf(template string, args ...interface{}) { S.Errorf(template, args...) }
func Fatal(msg string, fields ...zap.Field)       { L.Fatal(msg, fields...) }
func Fatalf(template string, args ...interface{}) { S.Fatalf(template, args...) }
