// Package middleware 提供了Gin框架的中间件函数，用于处理请求前的认证、日志记录、恢复等功能。
// AuthMiddleware 认证中间件，用于验证JWT令牌
// 该中间件会检查请求头中的Authorization字段，提取JWT令牌并验证其有效性。
// 如果令牌有效，会将解析后的声明存储在上下文（context）中，供后续处理器使用。
// 如果令牌无效或缺失，会返回401 Unauthorized错误。
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// AuthMiddleware 认证中间件，用于验证JWT令牌
func AuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// 检查Authorization头部格式 (Bearer <token>)
		tokenString := ""
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		} else if strings.HasPrefix(authHeader, "Token ") {
			tokenString = strings.TrimPrefix(authHeader, "Token ")
		} else {
			// 如果没有Bearer前缀，则整个字符串就是token
			tokenString = authHeader
		}

		// 解析JWT令牌
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// 验证签名方法
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrInvalidKeyType
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			c.Abort()
			return
		}

		// 将解析后的声明存储在上下文中，供后续处理器使用
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			c.Set("claims", claims)
		}

		c.Next()
	}
}