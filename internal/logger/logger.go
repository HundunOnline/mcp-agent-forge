package logger

import (
	"os"
	"path/filepath"

	"agent-forge/internal/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var log *zap.Logger

// InitLogger 初始化日志系统
func InitLogger(cfg *config.Config) error {
	// 设置日志级别
	level := zap.InfoLevel
	if err := level.UnmarshalText([]byte(cfg.Log.Level)); err != nil {
		return err
	}

	// 创建编码器配置
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	var core zapcore.Core
	if cfg.Log.Enabled {
		// 确保日志目录存在
		logDir := filepath.Dir(cfg.Log.File)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return err
		}

		// 配置文件日志输出
		writer := zapcore.AddSync(&lumberjack.Logger{
			Filename:   cfg.Log.File,
			MaxSize:    cfg.Log.MaxSize,
			MaxBackups: cfg.Log.MaxBackups,
			MaxAge:     cfg.Log.MaxAge,
			Compress:   cfg.Log.Compress,
		})

		core = zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			writer,
			level,
		)
	} else {
		// 使用标准错误输出
		core = zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(os.Stderr),
			level,
		)
	}

	// 创建日志记录器
	log = zap.New(core)
	return nil
}

// GetLogger 获取日志记录器实例
func GetLogger() *zap.Logger {
	if log == nil {
		// 如果日志记录器未初始化，使用标准错误的默认配置
		config := zap.NewProductionConfig()
		config.OutputPaths = []string{"stderr"}
		log, _ = config.Build()
	}
	return log
}

// Info 记录信息级别的日志
func Info(msg string, fields ...zap.Field) {
	GetLogger().Info(msg, fields...)
}

// Error 记录错误级别的日志
func Error(msg string, fields ...zap.Field) {
	GetLogger().Error(msg, fields...)
}

// Fatal 记录致命错误级别的日志并退出程序
func Fatal(msg string, fields ...zap.Field) {
	GetLogger().Fatal(msg, fields...)
}

// Debug 记录调试级别的日志
func Debug(msg string, fields ...zap.Field) {
	GetLogger().Debug(msg, fields...)
}

// Warn 记录警告级别的日志
func Warn(msg string, fields ...zap.Field) {
	GetLogger().Warn(msg, fields...)
}
