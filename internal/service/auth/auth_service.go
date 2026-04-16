// Package authservice 认证服务
package authservice

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	authpkg "yunyez/internal/pkg/auth"
	"yunyez/internal/model/auth"
)

// AuthService 认证服务接口
type AuthService interface {
	// Login 用户登录
	Login(ctx context.Context, req *authpkg.LoginRequest) (*authpkg.LoginResponse, error)
	
	// Logout 用户登出
	Logout(ctx context.Context, userID int64, token string) error
	
	// RefreshToken 刷新 Token
	RefreshToken(ctx context.Context, req *authpkg.RefreshTokenRequest) (*authpkg.TokenPair, error)
	
	// GetUserByID 根据 ID 获取用户
	GetUserByID(ctx context.Context, userID int64) (*auth.User, error)
	
	// GetUserRoles 获取用户角色
	GetUserRoles(ctx context.Context, userID int64) ([]string, error)
}

type authService struct {
	db            *gorm.DB
	redisClient   *redis.Client
	jwtManager    *authpkg.JWTManager
	blacklist     *authpkg.TokenBlacklist
	loginAttempts *authpkg.LoginAttemptManager
}

// NewAuthService 创建认证服务
func NewAuthService(
	db *gorm.DB,
	redisClient *redis.Client,
	jwtManager *authpkg.JWTManager,
	blacklist *authpkg.TokenBlacklist,
	loginAttempts *authpkg.LoginAttemptManager,
) AuthService {
	return &authService{
		db:            db,
		redisClient:   redisClient,
		jwtManager:    jwtManager,
		blacklist:     blacklist,
		loginAttempts: loginAttempts,
	}
}

// Login 用户登录
func (s *authService) Login(ctx context.Context, req *authpkg.LoginRequest) (*authpkg.LoginResponse, error) {
	// 1. 检查账户是否被锁定
	locked, err := s.loginAttempts.IsLocked(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("check lock status failed: %w", err)
	}
	if locked {
		s.recordLoginLog(ctx, req, 0, "account locked", nil)
		return nil, authpkg.NewAuthError(authpkg.CodeAccountLocked, "账户已被锁定，请稍后重试", authpkg.ErrAccountLocked)
	}
	
	// 2. 查询用户
	var user auth.User
	if err := s.db.WithContext(ctx).
		Where("username = ? AND deleted_at IS NULL", req.Username).
		First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			s.recordLoginLog(ctx, req, 0, "user not found", nil)
			return nil, authpkg.NewAuthError(authpkg.CodeInvalidCredentials, "用户名或密码错误", authpkg.ErrInvalidCredentials)
		}
		return nil, fmt.Errorf("query user failed: %w", err)
	}
	
	// 3. 检查账户状态
	if !user.IsActive() {
		s.recordLoginLog(ctx, req, 0, "account disabled", &user)
		return nil, authpkg.NewAuthError(authpkg.CodeAccountDisabled, "账户已被禁用", authpkg.ErrAccountDisabled)
	}
	
	// 4. 验证密码
	if err := authpkg.CheckPassword(req.Password, user.PasswordHash); err != nil {
		// 密码错误，记录失败尝试
		_ = s.loginAttempts.RecordFailedAttempt(ctx, req.Username)
		
		failedCount, _ := s.loginAttempts.GetFailedCount(ctx, req.Username)
		reason := fmt.Sprintf("invalid password (剩余 %d 次)", 5-failedCount)
		s.recordLoginLog(ctx, req, 0, reason, &user)
		
		return nil, authpkg.NewAuthError(authpkg.CodeInvalidCredentials, "用户名或密码错误", authpkg.ErrInvalidCredentials)
	}
	
	// 5. 清除失败尝试
	_ = s.loginAttempts.ClearFailedAttempts(ctx, req.Username)
	
	// 6. 获取用户角色
	roleCodes, err := s.GetUserRoles(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("get user roles failed: %w", err)
	}
	
	// 7. 生成 Token
	claims := authpkg.StandardClaims{
		UserID:       user.ID,
		Username:     user.Username,
		RoleCodes:    roleCodes,
		PlatformType: "admin",
	}
	
	tokenPair, err := s.jwtManager.GenerateTokenPair(claims, req.Remember)
	if err != nil {
		return nil, fmt.Errorf("generate token failed: %w", err)
	}
	
	// 8. 更新最后登录信息
	now := time.Now()
	s.db.WithContext(ctx).Model(&user).Updates(map[string]interface{}{
		"last_login_at": now,
		"last_login_ip": req.IP,
	})
	
	// 9. 记录登录成功日志
	s.recordLoginLog(ctx, req, 1, "", &user)
	
	// 10. 构造响应
	return &authpkg.LoginResponse{
		TokenPair: *tokenPair,
		User: authpkg.UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Nickname: user.Nickname,
			Email:    user.Email,
			Phone:    user.Phone,
			Avatar:   user.AvatarURL,
			Roles:    roleCodes,
		},
	}, nil
}

// Logout 用户登出
func (s *authService) Logout(ctx context.Context, userID int64, token string) error {
	// 解析 Token 获取 JTI
	claims, err := s.jwtManager.ParseAccessToken(token)
	if err != nil {
		return fmt.Errorf("parse token failed: %w", err)
	}
	
	// 将 Token 加入黑名单
	expireTime, _ := s.jwtManager.GetExpireTime(token)
	ttl := time.Until(expireTime)
	if ttl > 0 {
		if err := s.blacklist.AddToBlacklist(ctx, claims.ID, ttl); err != nil {
			return fmt.Errorf("add to blacklist failed: %w", err)
		}
	}
	
	return nil
}

// RefreshToken 刷新 Token
func (s *authService) RefreshToken(ctx context.Context, req *authpkg.RefreshTokenRequest) (*authpkg.TokenPair, error) {
	// 1. 解析 Refresh Token
	claims, err := s.jwtManager.ParseRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, authpkg.NewAuthError(authpkg.CodeInvalidToken, "无效的 Refresh Token", authpkg.ErrInvalidToken)
	}
	
	// 2. 检查 Token 是否在黑名单中
	isBlacklisted, err := s.blacklist.IsBlacklisted(ctx, claims.ID)
	if err != nil {
		return nil, fmt.Errorf("check blacklist failed: %w", err)
	}
	if isBlacklisted {
		return nil, authpkg.NewAuthError(authpkg.CodeBlacklistedToken, "Token 已被吊销", authpkg.ErrBlacklistedToken)
	}
	
	// 3. 查询用户
	user, err := s.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, authpkg.NewAuthError(authpkg.CodeUserNotFound, "用户不存在", authpkg.ErrUserNotFound)
	}
	
	// 4. 检查账户状态
	if !user.IsActive() {
		return nil, authpkg.NewAuthError(authpkg.CodeAccountDisabled, "账户已被禁用", authpkg.ErrAccountDisabled)
	}
	
	// 5. 获取最新角色
	roleCodes, err := s.GetUserRoles(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("get user roles failed: %w", err)
	}
	
	// 6. 生成新 Token 对
	newClaims := authpkg.StandardClaims{
		UserID:       user.ID,
		Username:     user.Username,
		RoleCodes:    roleCodes,
		PlatformType: claims.PlatformType,
	}
	
	tokenPair, err := s.jwtManager.GenerateTokenPair(newClaims, false)
	if err != nil {
		return nil, fmt.Errorf("generate token failed: %w", err)
	}
	
	// 7. 将旧的 Refresh Token 加入黑名单
	expireTime, _ := s.jwtManager.GetExpireTime(req.RefreshToken)
	ttl := time.Until(expireTime)
	if ttl > 0 {
		_ = s.blacklist.AddToBlacklist(ctx, claims.ID, ttl)
	}
	
	return tokenPair, nil
}

// GetUserByID 根据 ID 获取用户
func (s *authService) GetUserByID(ctx context.Context, userID int64) (*auth.User, error) {
	var user auth.User
	if err := s.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", userID).
		First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, authpkg.NewAuthError(authpkg.CodeUserNotFound, "用户不存在", authpkg.ErrUserNotFound)
		}
		return nil, fmt.Errorf("query user failed: %w", err)
	}
	return &user, nil
}

// GetUserRoles 获取用户角色
func (s *authService) GetUserRoles(ctx context.Context, userID int64) ([]string, error) {
	var roleCodes []string
	
	err := s.db.WithContext(ctx).
		Table("auth.roles r").
		Select("r.role_code").
		Joins("INNER JOIN auth.user_roles ur ON ur.role_id = r.id").
		Where("ur.user_id = ? AND r.status = 1 AND r.deleted_at IS NULL", userID).
		Pluck("r.role_code", &roleCodes).Error
	
	if err != nil {
		return nil, fmt.Errorf("query user roles failed: %w", err)
	}
	
	if len(roleCodes) == 0 {
		return []string{}, nil
	}
	
	return roleCodes, nil
}

// recordLoginLog 记录登录日志
func (s *authService) recordLoginLog(ctx context.Context, req *authpkg.LoginRequest, status int8, failureReason string, user *auth.User) {
	log := auth.LoginLog{
		Username:      req.Username,
		LoginType:     "password",
		Status:        status,
		FailureReason: failureReason,
		IPAddress:     req.IP,
		UserAgent:     req.UserAgent,
	}
	
	if user != nil {
		log.UserID = &user.ID
	}
	
	// 异步写入，不阻塞主流程
	go func() {
		if err := s.db.Create(&log).Error; err != nil {
			// 日志写入失败仅记录，不影响主流程
			fmt.Printf("failed to record login log: %v\n", err)
		}
	}()
}
