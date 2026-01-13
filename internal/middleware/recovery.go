// Package middleware 提供Gin框架的中间件函数
// RecoveryMiddleware 恢复中间件，用于捕获panic并记录错误日志
package middleware

import (
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RecoveryMiddleware 恢复中间件，用于捕获panic并记录错误日志
func RecoveryMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 获取堆栈信息
				stackBuf := make([]byte, 4096)
				stackSize := runtime.Stack(stackBuf, false)
				stackTrace := string(stackBuf[:stackSize])

				// 记录错误日志
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("stack_trace", stackTrace),
					zap.String("request_uri", c.Request.RequestURI),
					zap.String("method", c.Request.Method),
					zap.String("client_ip", c.ClientIP()),
				)

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