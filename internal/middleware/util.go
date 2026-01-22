package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"
)

// InitMiddlewares 根据配置初始化中间件
func InitMiddlewares() ([]gin.HandlerFunc, error) {
	// 从配置中读取JWT密钥
	jwtSecret := viper.GetString("jwt.secret")
	if jwtSecret == "" {
		// 如果配置中没有设置，则使用默认值（生产环境中不应这样做）
		jwtSecret = "default_secret_key_that_should_be_changed_in_production"
	}

	// 从配置中读取限流参数
	rateLimit := viper.GetFloat64("rate_limit.requests_per_second")
	if rateLimit == 0 {
		rateLimit = 10.0 // 默认每秒10个请求
	}
	burst := viper.GetInt("rate_limit.burst")
	if burst == 0 {
		burst = 20 // 默认突发容量为20
	}

	// 从配置中读取限流模式
	modeStr := viper.GetString("rate_limit.mode")
	var mode RateLimitMode
	if modeStr == "distributed" {
		mode = DistributedMode
	} else {
		mode = LocalMode // 默认使用本地模式
	}

	var middlewares []gin.HandlerFunc

	// 添加日志中间件
	middlewares = append(middlewares, LoggerToFile())

	// 添加恢复中间件
	middlewares = append(middlewares, RecoveryMiddleware())

	// 添加CORS中间件
	middlewares = append(middlewares, CORSMiddleware())

	// 添加安全头部中间件
	middlewares = append(middlewares, SecurityHeadersMiddleware())

	// 添加限流中间件
	middlewares = append(middlewares, RateLimitMiddleware(RateLimitConfig{
		Mode:  mode, // 从配置中读取模式
		Limit: rate.Limit(rateLimit), // 每秒请求数限制
		Burst: burst, // 突发容量
	}))

	// 注意：JWT认证中间件只应在需要保护的路由上使用，而不是全局使用
	// authMiddleware := AuthMiddleware(jwtSecret)

	return middlewares, nil
}

// SetupRouterWithMiddlewares 为路由器设置中间件的便捷函数
func SetupRouterWithMiddlewares(router *gin.Engine) error {
	middlewares, err := InitMiddlewares()
	if err != nil {
		return err
	}

	// 应用全局中间件
	for _, middleware := range middlewares {
		router.Use(middleware)
	}

	return nil
}