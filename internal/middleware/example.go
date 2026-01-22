package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/time/rate"
)

// ExampleUsage 示例：如何在应用中使用中间件
func ExampleUsage() {
	// 初始化Gin路由器
	r := gin.Default()

	// 在这里应该初始化zap logger
	// logger, _ := zap.NewProduction()
	// defer logger.Sync()

	// 使用CORS中间件
	r.Use(CORSMiddleware())

	// 使用日志中间件 (需要logger实例)
	// r.Use(LoggerToFile(logger))

	// 使用恢复中间件 (需要logger实例)
	// r.Use(RecoveryMiddleware(logger))

	// 使用安全头部中间件
	r.Use(SecurityHeadersMiddleware())

	// 使用限流中间件 (每秒最多5个请求，突发容量为10)
	rateLimiter := RateLimitMiddleware(RateLimitConfig{
		Mode:  LocalMode, // 使用本地模式
		Limit: rate.Every(1 * time.Second), // 每秒最多1个请求
		Burst: 5, // 突发容量为5
	})
	r.Use(rateLimiter)

	// 创建需要认证的路由组
	protected := r.Group("/api/protected")
	// 使用JWT认证中间件 (需要提供密钥)
	// protected.Use(AuthMiddleware("your-jwt-secret-key"))
	protected.GET("/data", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "This is protected data",
			"user":    c.MustGet("claims"), // 从JWT token获取的用户信息
		})
	})

	// 公共路由
	public := r.Group("/api/public")
	public.GET("/info", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "This is public data",
		})
	})

	// 演示如何生成JWT token
	GenerateTokenExample()
}

// GenerateTokenExample 示例：如何生成JWT token
func GenerateTokenExample() {
	// 创建一个示例token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": 123,
		"email":   "user@example.com",
		"exp":     time.Now().Add(time.Hour * 72).Unix(), // 72小时后过期
	})

	// 使用密钥签名token
	tokenString, err := token.SignedString([]byte("your-jwt-secret-key"))
	if err != nil {
		// 处理错误
		return
	}

	// tokenString即为生成的JWT token
	_ = tokenString
}