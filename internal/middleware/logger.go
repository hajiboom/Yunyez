package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	logger "yunyez/internal/pkg/logger"
)

// LoggerToFile 自定义日志中间件
func LoggerToFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		c.Next()

		endTime := time.Now()
		latencyTime := endTime.Sub(startTime)
		reqMethod := c.Request.Method
		reqURI := c.Request.RequestURI
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// 记录日志
		fields := map[string]interface{}{
			"status_code": statusCode,
			"client_ip":   clientIP,
			"req_method":  reqMethod,
			"req_uri":     reqURI,
			"time":        latencyTime.String(),
			"user_agent":  userAgent,
		}

		ctx := context.Background()
		logger.Info(ctx, "HTTP Request", fields)

		// 如果有错误，也记录错误信息
		if len(c.Errors) > 0 {
			errorFields := map[string]interface{}{
				"client_ip":  clientIP,
				"req_method": reqMethod,
				"req_uri":    reqURI,
				"error":      c.Errors.Last().Error(),
			}
			logger.Error(ctx, "Request Error", errorFields)
		}
	}
}