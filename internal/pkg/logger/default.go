// Package logger 默认日志记录器
// 该日志记录器默认输出到标准错误流（stderr）
// 记录基本的业务日志
// 日志级别：Debug < Info < Warn < Error
// @date: 2025-11-17
// @version: 1.0.0
package logger

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	config "yunyez/internal/common/config"
	tools "yunyez/internal/common/tools"
)

// Logger 是日志记录器结构体
type Logger struct {
	zapLogger *zap.Logger
	sugar     *zap.SugaredLogger // 添加 SugaredLogger 以支持 ...interface{} 变参
	expire    int                // 日志过期时间，单位：天
}

// DefaultLogger 是默认的日志记录器实例
var DefaultLogger *Logger

// getLogFilePath 获取日志文件路径
func getLogFilePath() string {
	wd := tools.GetRootDir()
	logFilePath := config.GetString("logger.storage")
	date := time.Now().Format("2006-01-02")
	logFile := fmt.Sprintf("app-%s.log", date)

	completedLogFile := filepath.Join(wd, logFilePath, logFile)

	return completedLogFile
}

// encoderConfig 是日志编码器配置
var encoderConfig = zapcore.EncoderConfig{
	TimeKey:        "time",
	LevelKey:       "level",
	NameKey:        "logger",
	CallerKey:      "caller",
	MessageKey:     "msg",
	StacktraceKey:  "stacktrace",
	LineEnding:     zapcore.DefaultLineEnding,
	EncodeLevel:    zapcore.CapitalLevelEncoder,
	EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
	EncodeDuration: zapcore.SecondsDurationEncoder,
	EncodeCaller:   zapcore.ShortCallerEncoder,
}

// 初始化默认日志记录器
func init() {
	DefaultLogger = New()
}

// New 创建一个新的Logger实例
func New() *Logger {
	// 获取日志文件路径
	logFilePath := getLogFilePath()
	log.Printf("log file path: %s\n", logFilePath)

	// 创建文件输出
	var fileSync zapcore.WriteSyncer = zapcore.AddSync(os.Stderr)
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		fileSync = zapcore.AddSync(file)
	} else {
		fmt.Printf("failed to open log file: %v\n", err)
	}

	// 创建编码器
	fileEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	// 创建核心 - 只写入文件
	fileCore := zapcore.NewCore(fileEncoder, fileSync, zapcore.DebugLevel)
	var cores []zapcore.Core
	cores = append(cores, fileCore)
	if config.GetBool("app.debug") { // 开发模式追加日志打印到控制台
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // 开发模式下使用彩色日志
		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
		consoleCore := zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), zapcore.DebugLevel)
		cores = append(cores, consoleCore)
	}
	// 创建日志记录器
	logger := zap.New(zapcore.NewTee(cores...),
					zap.AddCaller(), 
					zap.AddCallerSkip(2), // 跳过调用者栈帧，避免记录 logger 包的封装
					zap.Development())

	return &Logger{
		zapLogger: logger,
		sugar:     logger.Sugar(),
	}
}

func toZapField(ctx context.Context, fields map[string]interface{}) []zap.Field {
	if fields == nil {
		fields = make(map[string]interface{})
	}
	// 添加 trace_id 到日志字段
	traceID := tools.GetTraceID(ctx)
	fields["trace_id"] = traceID

	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}

	return zapFields
}

// --- 日志记录方法 ---

// Debug 记录调试级别日志，支持键值对参数 (key1, val1, key2, val2, ...)
func (l *Logger) Debug(ctx context.Context, msg string, fields map[string]interface{}) {
	l.zapLogger.Debug(msg, toZapField(ctx, fields)...)
}

// Info 记录信息级别日志，支持键值对参数
func (l *Logger) Info(ctx context.Context, msg string, fields map[string]interface{}) {
	l.zapLogger.Info(msg, toZapField(ctx, fields)...)
}

// Warn 记录警告级别日志，支持键值对参数
func (l *Logger) Warn(ctx context.Context, msg string, fields map[string]interface{}) {
	l.zapLogger.Warn(msg, toZapField(ctx, fields)...)
}

// Error 记录错误级别日志，支持键值对参数
func (l *Logger) Error(ctx context.Context, msg string, fields map[string]interface{}) {
	l.zapLogger.Error(msg, toZapField(ctx, fields)...)
}

// DPanic 记录灾难性错误级别日志，在开发模式下会panic，支持键值对参数
func (l *Logger) DPanic(ctx context.Context, msg string, fields map[string]interface{}) {
	l.zapLogger.DPanic(msg, toZapField(ctx, fields)...)
}

// Panic 记录恐慌级别日志并panic，支持键值对参数
func (l *Logger) Panic(ctx context.Context, msg string, fields map[string]interface{}) {
	l.zapLogger.Panic(msg, toZapField(ctx, fields)...)
}

// Fatal 记录致命错误级别日志并退出程序，支持键值对参数
func (l *Logger) Fatal(ctx context.Context, msg string, fields map[string]interface{}) {
	l.zapLogger.Fatal(msg, toZapField(ctx, fields)...)
}

// With 返回一个新的Logger，包含指定的键值对上下文
func (l *Logger) With(ctx context.Context, fields map[string]interface{}) *Logger {
	newSugar := l.sugar.With(toZapField(ctx, fields))
	return &Logger{
		zapLogger: newSugar.Desugar(), // Desugar if you need the underlying *zap.Logger later
		sugar:     newSugar,
	}
}

// Sync 刷新缓冲区中的日志
func (l *Logger) Sync() error {
	return l.zapLogger.Sync()
}

// --- 全局便捷函数 ---

// Debug 记录调试级别日志
func Debug(ctx context.Context, msg string, fields map[string]interface{}) {
	DefaultLogger.Debug(ctx, msg, fields)
}

// Info 记录信息级别日志
func Info(ctx context.Context, msg string, fields map[string]interface{}) {
	DefaultLogger.Info(ctx, msg, fields)
}

// Warn 记录警告级别日志
func Warn(ctx context.Context, msg string, fields map[string]interface{}) {
	DefaultLogger.Warn(ctx, msg, fields)
}

// Error 记录错误级别日志
func Error(ctx context.Context, msg string, fields map[string]interface{}) {
	DefaultLogger.Error(ctx, msg, fields)
}

// DPanic 记录灾难性错误级别日志，在开发模式下会panic
func DPanic(ctx context.Context, msg string, fields map[string]interface{}) {
	DefaultLogger.DPanic(ctx, msg, fields)
}

// Panic 记录恐慌级别日志并panic
func Panic(ctx context.Context, msg string, fields map[string]interface{}) {
	DefaultLogger.Panic(ctx, msg, fields)
}

// Fatal 记录致命错误级别日志并退出程序
func Fatal(ctx context.Context, msg string, fields map[string]interface{}) {
	DefaultLogger.Fatal(ctx, msg, fields)
}
