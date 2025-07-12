package logger

import (
	"context"
	"github.com/kokp520/banking-system/server/pkg/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"path/filepath"
)

var Logger *zap.Logger

// Init Logger
// level: debug, info, warn, error
// logDir: log輸出位置
func Init(level, format, logDir string) error {

	var logLevel zapcore.Level
	switch level {
	case "debug":
		logLevel = zapcore.DebugLevel
	case "info":
		logLevel = zapcore.InfoLevel
	case "warn":
		logLevel = zapcore.WarnLevel
	case "error":
		logLevel = zapcore.ErrorLevel
	default:
		logLevel = zapcore.InfoLevel
	}

	var encoder zapcore.Encoder
	if format == "json" {
		encoder = zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	} else {
		encoder = zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	}

	var cores []zapcore.Core

	consoleCore := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		logLevel,
	)
	cores = append(cores, consoleCore)

	// 文件輸出（如果指定了目錄）
	if logDir != "" {
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return err
		}

		appLogFile := &lumberjack.Logger{
			Filename:   filepath.Join(logDir, "app.log"),
			MaxSize:    100, // MB
			MaxBackups: 3,
			MaxAge:     30, // days
			Compress:   true,
		}

		errorLogFile := &lumberjack.Logger{
			Filename:   filepath.Join(logDir, "error.log"),
			MaxSize:    100, // MB
			MaxBackups: 3,
			MaxAge:     30, // days
			Compress:   true,
		}

		appCore := zapcore.NewCore(
			encoder,
			zapcore.AddSync(appLogFile),
			logLevel,
		)
		cores = append(cores, appCore)

		errorCore := zapcore.NewCore(
			encoder,
			zapcore.AddSync(errorLogFile),
			zapcore.ErrorLevel,
		)
		cores = append(cores, errorCore)
	}

	core := zapcore.NewTee(cores...)
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	Logger = logger
	return nil
}

func WithTraceID(ctx context.Context) *zap.Logger {
	traceID := trace.GetTraceID(ctx)
	if traceID == "" {
		return Logger
	}
	return Logger.With(zap.String(trace.Key, traceID))
}

func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	Logger.Debug(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Logger.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Logger.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	Logger.Fatal(msg, fields...)
}

//func Sync() {
//	Logger.Sync()
//}
