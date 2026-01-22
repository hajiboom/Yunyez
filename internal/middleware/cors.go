// Package middleware 提供了Gin框架的中间件函数，用于处理请求前的认证、日志记录、恢复等功能。
// CORSMiddleware CORS中间件，用于处理跨域请求
// 该中间件会设置响应头，允许所有域名的请求访问，同时支持携带凭证（如Cookie）。
// 它还会处理预检请求（OPTIONS方法），返回允许的方法和头信息。
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware CORS中间件，用于处理跨域请求
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*") // 生产环境中应指定具体的域名
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}