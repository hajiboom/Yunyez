package middleware

import (
	"time"
	"yunyez/internal/common/config"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// SetupMiddlewares 初始化并返回常用的中间件集合
func SetupMiddlewares() []gin.HandlerFunc {
	// 
	jwtSecret := config.GetStringWithDefault("jwt.secret", "test-secret")
	return []gin.HandlerFunc{
		LoggerToFile(),              // 日志记录
		RecoveryMiddleware(),        // Panic恢复
		CORSMiddleware(),            // CORS跨域
		SecurityHeadersMiddleware(), // 安全头部
		RateLimitMiddleware(RateLimitConfig{
			Mode:  LocalMode, // 使用本地模式，如果需要分布式则改为 DistributedMode
			Limit: rate.Every(1 * time.Second), // 每秒最多1个请求
			Burst: 10, // 突发容量为10
		}), // 限流: 每秒最多10个请求
		AuthMiddleware(jwtSecret),   // JWT认证 (可根据需要选择性使用)
	}
}

// ApplyMiddlewares 应用多个中间件到路由组
func ApplyMiddlewares(group *gin.RouterGroup, middlewares ...gin.HandlerFunc) {
	for _, middleware := range middlewares {
		group.Use(middleware)
	}
}