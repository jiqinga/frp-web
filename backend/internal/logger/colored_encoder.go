package logger

import (
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

// coloredMessageEncoder 包装 console encoder，为消息添加颜色
type coloredMessageEncoder struct {
	zapcore.Encoder
	cfg zapcore.EncoderConfig
}

// newColoredMessageEncoder 创建彩色消息编码器
func newColoredMessageEncoder(cfg zapcore.EncoderConfig) zapcore.Encoder {
	return &coloredMessageEncoder{
		Encoder: zapcore.NewConsoleEncoder(cfg),
		cfg:     cfg,
	}
}

// Clone 实现 Encoder 接口
func (e *coloredMessageEncoder) Clone() zapcore.Encoder {
	return &coloredMessageEncoder{
		Encoder: e.Encoder.Clone(),
		cfg:     e.cfg,
	}
}

// EncodeEntry 编码日志条目，为消息添加颜色
func (e *coloredMessageEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	// 根据日志等级选择颜色
	var color string
	switch entry.Level {
	case zapcore.DebugLevel:
		color = colorMagenta
	case zapcore.InfoLevel:
		color = colorGreen
	case zapcore.WarnLevel:
		color = colorYellow
	case zapcore.ErrorLevel, zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		color = colorRed
	default:
		color = ""
	}

	// 为消息添加颜色
	if color != "" {
		entry.Message = color + entry.Message + colorReset
	}

	return e.Encoder.EncodeEntry(entry, fields)
}
