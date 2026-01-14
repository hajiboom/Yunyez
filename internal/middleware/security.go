// Package middleware 提供Gin框架的中间件函数
// SecurityHeadersMiddleware 安全头部中间件，用于设置常见的安全头部
package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeadersMiddleware 安全头部中间件，用于设置常见的安全头部
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置X-Frame-Options头部，防止点击劫持
		c.Header("X-Frame-Options", "DENY")
		
		// 设置X-Content-Type-Options头部，防止MIME类型嗅探
		c.Header("X-Content-Type-Options", "nosniff")
		
		// 设置X-XSS-Protection头部，启用浏览器的XSS过滤器
		c.Header("X-XSS-Protection", "1; mode=block")
		
		// 设置Strict-Transport-Security头部，强制使用HTTPS
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		
		// 设置Content-Security-Policy头部，防止各种注入攻击
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self'; frame-src 'self'; object-src 'none'")
		
		// 设置Referrer-Policy头部，控制Referer头部的发送
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// 设置Permissions-Policy头部，控制浏览器功能的使用
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		
		c.Next()
	}
}