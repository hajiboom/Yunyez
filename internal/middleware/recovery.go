// Package middleware 提供Gin框架的中间件函数
// RecoveryMiddleware 恢复中间件，用于捕获panic并记录错误日志
package middleware

import (
	"context"
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
	logger "yunyez/internal/pkg/logger"
)

// RecoveryMiddleware 恢复中间件，用于捕获panic并记录错误日志
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 获取堆栈信息
				stackBuf := make([]byte, 4096)
				stackSize := runtime.Stack(stackBuf, false)
				stackTrace := string(stackBuf[:stackSize])

				// 记录错误日志
				fields := map[string]interface{}{
					"error":       err,
					"stack_trace": stackTrace,
					"request_uri": c.Request.RequestURI,
					"method":      c.Request.Method,
					"client_ip":   c.ClientIP(),
				}

				ctx := context.Background()
				logger.Error(ctx, "Panic recovered", fields)

				// 返回500错误
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Internal Server Error",
				})

				// 终止请求处理
				c.Abort()
			}
		}()

		c.Next()
	}
}