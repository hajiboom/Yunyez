// Package middleware 提供了Gin框架的中间件函数，用于处理请求前的认证、日志记录、恢复等功能。
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	authpkg "yunyez/internal/pkg/auth"
)

// AuthMiddlewareConfig 认证中间件配置
type AuthMiddlewareConfig struct {
	JWTManager    *authpkg.JWTManager
	Blacklist     *authpkg.TokenBlacklist
	RedisClient   *redis.Client
	RequiredRoles []string // 要求的角色 (空则不检查)
}

// AuthMiddleware 认证中间件 (重构版)
// 支持:
// - 标准化 Claims 提取
// - Redis Token 黑名单检查
// - 角色权限验证
// - 统一错误响应
func AuthMiddleware(cfg AuthMiddlewareConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 提取 Token
		tokenString, err := extractToken(c)
		if err != nil {
			abortWithError(c, http.StatusUnauthorized, authpkg.CodeTokenMissing, err.Error())
			return
		}
		
		// 2. 解析 Token
		claims, err := cfg.JWTManager.ParseAccessToken(tokenString)
		if err != nil {
			if authErr, ok := err.(*authpkg.AuthError); ok {
				abortWithError(c, http.StatusUnauthorized, authErr.Code, authErr.Error())
				return
			}
			abortWithError(c, http.StatusUnauthorized, authpkg.CodeInvalidToken, "Invalid token")
			return
		}
		
		// 3. 检查 Token 黑名单
		if cfg.Blacklist != nil {
			isBlacklisted, checkErr := cfg.Blacklist.IsBlacklisted(c.Request.Context(), claims.ID)
			if checkErr != nil {
				abortWithError(c, http.StatusInternalServerError, authpkg.CodeInternalError, "Token verification failed")
				return
			}
			if isBlacklisted {
				abortWithError(c, http.StatusUnauthorized, authpkg.CodeBlacklistedToken, "Token has been revoked")
				return
			}
		}
		
		// 4. 角色权限检查
		if len(cfg.RequiredRoles) > 0 {
			hasRole := false
			for _, requiredRole := range cfg.RequiredRoles {
				for _, userRole := range claims.RoleCodes {
					if userRole == requiredRole {
						hasRole = true
						break
					}
				}
				if hasRole {
					break
				}
			}
			
			if !hasRole {
				abortWithError(c, http.StatusForbidden, 40301, "Insufficient permissions")
				return
			}
		}
		
		// 5. 将用户信息注入 Context (标准化)
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role_codes", claims.RoleCodes)
		c.Set("platform_type", claims.PlatformType)
		c.Set("claims", claims)
		
		c.Next()
	}
}

// extractToken 从请求头中提取 Token
func extractToken(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", authpkg.ErrTokenMissing
	}
	
	// 支持 Bearer 和 Token 前缀
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer "), nil
	}
	if strings.HasPrefix(authHeader, "Token ") {
		return strings.TrimPrefix(authHeader, "Token "), nil
	}
	
	// 无前缀，直接使用
	return authHeader, nil
}

// abortWithError 中止请求并返回错误响应
func abortWithError(c *gin.Context, httpStatus int, code int, message string) {
	c.JSON(httpStatus, gin.H{
		"code":    code,
		"message": message,
	})
	c.Abort()
}

// OptionalAuthMiddleware 可选认证中间件 (不强制要求 Token)
// 如果有 Token 则解析并注入用户信息，没有则跳过
func OptionalAuthMiddleware(cfg AuthMiddlewareConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := extractToken(c)
		if err != nil {
			// 没有 Token，继续
			c.Next()
			return
		}
		
		claims, err := cfg.JWTManager.ParseAccessToken(tokenString)
		if err != nil {
			// Token 无效，继续 (可选认证不阻断)
			c.Next()
			return
		}
		
		// 检查黑名单
		if cfg.Blacklist != nil {
			isBlacklisted, _ := cfg.Blacklist.IsBlacklisted(c.Request.Context(), claims.ID)
			if isBlacklisted {
				c.Next()
				return
			}
		}
		
		// 注入用户信息
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role_codes", claims.RoleCodes)
		c.Set("platform_type", claims.PlatformType)
		c.Set("claims", claims)
		
		c.Next()
	}
}
