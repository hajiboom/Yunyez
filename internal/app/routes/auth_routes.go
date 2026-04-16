// Package routes 路由注册
package routes

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	authcontroller "yunyez/internal/controller/auth"
	"yunyez/internal/middleware"
	authpkg "yunyez/internal/pkg/auth"
	authservice "yunyez/internal/service/auth"
)

// AuthDependencies 认证服务依赖
type AuthDependencies struct {
	DB          *gorm.DB
	RedisClient *redis.Client
	AuthConfig  authpkg.AuthConfig
}

// SetupAuthRoutes 注册认证相关路由
func SetupAuthRoutes(r *gin.Engine, deps AuthDependencies) {
	// 1. 初始化认证组件
	jwtManager := authpkg.NewJWTManager(deps.AuthConfig.JWT)
	
	blacklist := authpkg.NewTokenBlacklist(deps.RedisClient, deps.AuthConfig.Redis)
	loginAttempts := authpkg.NewLoginAttemptManager(deps.RedisClient, deps.AuthConfig.LoginSafety)
	
	authSvc := authservice.NewAuthService(
		deps.DB,
		deps.RedisClient,
		jwtManager,
		blacklist,
		loginAttempts,
	)
	
	loginCtrl := authcontroller.NewLoginController(authSvc)
	
	// 2. 创建认证中间件配置
	authCfg := middleware.AuthMiddlewareConfig{
		JWTManager:  jwtManager,
		Blacklist:   blacklist,
		RedisClient: deps.RedisClient,
	}
	
	// 3. 注册路由
	api := r.Group("/api")
	{
		// 公开路由 (无需认证)
		publicAuth := api.Group("/auth")
		{
			publicAuth.POST("/login", loginCtrl.Login)           // 登录
			publicAuth.POST("/refresh", loginCtrl.RefreshToken)  // 刷新 Token
		}
		
		// 需要认证的路由
		protectedAuth := api.Group("/auth")
		protectedAuth.Use(middleware.AuthMiddleware(authCfg))
		{
			protectedAuth.POST("/logout", loginCtrl.Logout)          // 登出
			protectedAuth.GET("/userinfo", loginCtrl.GetUserInfo)    // 获取用户信息
		}
		
		// 示例: 需要特定角色的路由
		adminRoutes := api.Group("/admin")
		adminRoutes.Use(middleware.AuthMiddleware(middleware.AuthMiddlewareConfig{
			JWTManager:    jwtManager,
			Blacklist:     blacklist,
			RedisClient:   deps.RedisClient,
			RequiredRoles: []string{"super_admin", "admin"},
		}))
		{
			// adminRoutes.GET("/users", ...)  // 仅管理员可访问
		}
	}
}

// DefaultAuthConfig 返回默认认证配置 (开发环境)
// 生产环境应从配置文件读取
func DefaultAuthConfig() authpkg.AuthConfig {
	return authpkg.AuthConfig{
		Mode: "local",
		JWT: authpkg.JWTConfig{
			AccessSecret:   "yunyez-dev-secret-2026-change-in-production",
			RefreshSecret:  "yunyez-dev-refresh-secret-2026",
			AccessExpire:   7200,    // 2 小时
			RefreshExpire:  604800,  // 7 天
			RememberExpire: 2592000, // 30 天
			Issuer:         "yunyez-auth",
		},
		Redis: authpkg.RedisConfig{
			Enabled:   true,
			KeyPrefix: "auth:token:blacklist:",
		},
		LoginSafety: authpkg.LoginSafetyConfig{
			MaxAttempts:  5,
			LockDuration: 900, // 15 分钟
			RateLimit:    30,
		},
	}
}

// TokenCleanupTask Token 黑名单清理任务
// 定期清理过期的 Token (虽然 Redis 会自动过期，但可以做一些额外的清理工作)
func StartTokenCleanupTask(redisClient *redis.Client) {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			// 这里可以添加额外的清理逻辑
			// 例如: 统计黑名单大小、监控异常等
		}
	}()
}
