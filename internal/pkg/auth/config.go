package auth

// AuthConfig 认证配置
type AuthConfig struct {
	// Mode 认证模式: local, multi_secret, oidc, introspection
	Mode string `mapstructure:"mode"`
	
	// JWT 配置
	JWT JWTConfig `mapstructure:"jwt"`
	
	// Redis 配置 (用于 Token 黑名单)
	Redis RedisConfig `mapstructure:"redis"`
	
	// 登录安全配置
	LoginSafety LoginSafetyConfig `mapstructure:"login_safety"`
}

// JWTConfig JWT 配置
type JWTConfig struct {
	// AccessSecret Access Token 签名密钥
	AccessSecret string `mapstructure:"access_secret"`
	
	// RefreshSecret Refresh Token 签名密钥 (可选，不填则使用 AccessSecret)
	RefreshSecret string `mapstructure:"refresh_secret"`
	
	// AccessExpire Access Token 过期时间 (秒)
	AccessExpire int64 `mapstructure:"access_expire"`
	
	// RefreshExpire Refresh Token 过期时间 (秒)
	RefreshExpire int64 `mapstructure:"refresh_expire"`
	
	// RememberExpire "记住登录" 时的过期时间 (秒)
	RememberExpire int64 `mapstructure:"remember_expire"`
	
	// Issuer Token 签发者
	Issuer string `mapstructure:"issuer"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	// Enabled 是否启用 Redis
	Enabled bool `mapstructure:"enabled"`
	
	// KeyPrefix Token 黑名单的 Key 前缀
	KeyPrefix string `mapstructure:"key_prefix"`
}

// LoginSafetyConfig 登录安全配置
type LoginSafetyConfig struct {
	// MaxAttempts 最大失败次数
	MaxAttempts int `mapstructure:"max_attempts"`
	
	// LockDuration 锁定时长 (秒)
	LockDuration int64 `mapstructure:"lock_duration"`
	
	// RateLimit 单 IP 登录限流 (每分钟请求数)
	RateLimit int `mapstructure:"rate_limit"`
}

// DefaultAuthConfig 返回默认配置
func DefaultAuthConfig() AuthConfig {
	return AuthConfig{
		Mode: "local",
		JWT: JWTConfig{
			AccessSecret:   "default-secret-change-in-production",
			AccessExpire:   7200,     // 2 小时
			RefreshExpire:  604800,   // 7 天
			RememberExpire: 2592000,  // 30 天
			Issuer:         "yunyez-auth",
		},
		Redis: RedisConfig{
			Enabled:   false,
			KeyPrefix: "auth:token:blacklist:",
		},
		LoginSafety: LoginSafetyConfig{
			MaxAttempts:  5,
			LockDuration: 900, // 15 分钟
			RateLimit:    30,
		},
	}
}
