package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HTTP请求认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查请求头是否包含认证信息
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "auth header is empty"})
			c.Abort()
			return
		}

		// TODO: 解析认证信息，验证合法性
		

		// 认证通过，继续处理请求
		c.Next()
	}
}