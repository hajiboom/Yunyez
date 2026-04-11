package auth

import "github.com/golang-jwt/jwt/v4"

// StandardClaims 标准化 JWT Claims
type StandardClaims struct {
	jwt.RegisteredClaims
	
	// UserID 用户 ID
	UserID int64 `json:"user_id"`
	
	// Username 用户名
	Username string `json:"username"`
	
	// RoleCodes 角色代码列表
	RoleCodes []string `json:"role_codes"`
	
	// PlatformType 平台类型 (admin, open, device)
	PlatformType string `json:"platform_type"`
}

// TokenPair Token 对 (Access + Refresh)
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	
	// TokenType Token 类型 (Bearer)
	TokenType string `json:"token_type"`
	
	// ExpiresIn Access Token 过期时间 (秒)
	ExpiresIn int64 `json:"expires_in"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	
	// Remember 是否记住登录
	Remember bool `json:"remember"`
	
	// IP 登录 IP (由中间件提取)
	IP string `json:"-"`
	
	// UserAgent User-Agent (由中间件提取)
	UserAgent string `json:"-"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	TokenPair
	
	// User 用户信息
	User UserInfo `json:"user"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID       int64    `json:"id"`
	Username string   `json:"username"`
	Nickname string   `json:"nickname"`
	Email    string   `json:"email"`
	Phone    string   `json:"phone"`
	Avatar   string   `json:"avatar"`
	Roles    []string `json:"roles"`
}

// RefreshTokenRequest 刷新 Token 请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// LogoutRequest 登出请求
type LogoutRequest struct {
	// Token 要吊销的 Token (可选，不填则使用请求头中的)
	Token string `json:"token"`
}
