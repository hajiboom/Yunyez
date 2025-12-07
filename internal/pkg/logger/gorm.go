// 数据库日志记录器
// 该日志记录器默认输出到标准错误流（stderr）
// 记录数据库操作日志，sql执行及时间
// 实现 gorm.Interface 接口的 LogMode 方法
// 日志级别：Debug < Info < Warn < Error
// @date: 2025-11-17
// @version: 1.0.0

package logger

import (
	"context"
	"errors"
	"time"

	gorm_logger "gorm.io/gorm/logger"
)

// sql日志记录器
type SQLLogger struct {
	logger *Logger
}

// LogMode implements gorm_logger.Interface.
// 返回日志级别（由gorm控制）
func (sl *SQLLogger) LogMode(gorm_logger.LogLevel) gorm_logger.Interface {
	return sl
}


func unmarshal(data ...interface{}) map[string]interface{} {
	fields := map[string]interface{}{}
	if len(data) > 0 {
		if len(data)%2 == 0 { // 如果是偶数个，假设是kv结构
			for i := 0; i < len(data); i += 2 {
				if key, ok := data[i].(string); ok {
					fields[key] = data[i+1]
				}
			}
		} else { // 如果是奇数个，假设是args
			fields["args"] = data
		}
	}
	return fields
} 

// Info implements gorm_logger.Interface.
// Info 记录Info级别日志信息
func (sl *SQLLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	fields := unmarshal(data...)
	sl.logger.Info(ctx, msg, fields)
}

// Warn implements gorm_logger.Interface.
// Warn 记录Warn级别日志信息
func (sl *SQLLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	fields := unmarshal(data...)
	sl.logger.Warn(ctx, msg, fields)
}

// Error implements gorm_logger.Interface.
// Error 记录Error级别日志信息
func (sl *SQLLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	fields := unmarshal(data...)
	sl.logger.Error(ctx, msg, fields)
}

// Trace implements gorm_logger.Interface.
// Trace 实现sql跟踪
func (sl *SQLLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	// 计算执行时间
	elapsed := time.Since(begin)

	// 获取SQL和影响行数
	sql, rows := fc()

	// 构建日志字段
	fields := map[string]interface{}{
		"elapsed": elapsed.String(),
		"rows":    rows,
		"sql":     sql,
	}

	// 根据执行时间和错误情况记录不同级别的日志
	switch {
	case err != nil && !errors.Is(err, gorm_logger.ErrRecordNotFound):
		fields["error"] = err.Error()
		sl.logger.Error(ctx, "SQL execution failed", fields)
	case elapsed > time.Second:
		sl.logger.Warn(ctx, "SQL execution is slow", fields)
	default:
		sl.logger.Debug(ctx, "SQL execution completed", fields)
	}
}

func NewSQLLogger(logger *Logger) gorm_logger.Interface {
	return &SQLLogger{logger: logger}
}
