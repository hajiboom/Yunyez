package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LoggerToFile 自定义日志中间件
func LoggerToFile(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		c.Next()

		endTime := time.Now()
		latencyTime := endTime.Sub(startTime)
		reqMethod := c.Request.Method
		reqURI := c.Request.RequestURI
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()

		// 记录日志
		logger.Info("HTTP Request",
			zap.Int("status_code", statusCode),
			zap.String("client_ip", clientIP),
			zap.String("req_method", reqMethod),
			zap.String("req_uri", reqURI),
			zap.Duration("time", latencyTime),
			zap.String("user_agent", c.Request.UserAgent()),
		)

		// 如果有错误，也记录错误信息
		if len(c.Errors) > 0 {
			logger.Error("Request Error",
				zap.String("client_ip", clientIP),
				zap.String("req_method", reqMethod),
				zap.String("req_uri", reqURI),
				zap.String("error", c.Errors.Last().Error()),
			)
		}
	}
}