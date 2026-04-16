package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// JWTManager JWT 管理器
type JWTManager struct {
	accessSecret  []byte
	refreshSecret []byte
	accessExpire  time.Duration
	refreshExpire time.Duration
	issuer        string
}

// NewJWTManager 创建 JWT 管理器
func NewJWTManager(config JWTConfig) *JWTManager {
	refreshSecret := config.AccessSecret
	if config.RefreshSecret != "" {
		refreshSecret = config.RefreshSecret
	}
	
	return &JWTManager{
		accessSecret:  []byte(config.AccessSecret),
		refreshSecret: []byte(refreshSecret),
		accessExpire:  time.Duration(config.AccessExpire) * time.Second,
		refreshExpire: time.Duration(config.RefreshExpire) * time.Second,
		issuer:        config.Issuer,
	}
}

// GenerateTokenPair 生成 Token 对 (Access + Refresh)
func (m *JWTManager) GenerateTokenPair(claims StandardClaims, remember bool) (*TokenPair, error) {
	accessToken, expiresIn, err := m.generateAccessToken(claims)
	if err != nil {
		return nil, err
	}
	
	refreshToken, _, err := m.generateRefreshToken(claims, remember)
	if err != nil {
		return nil, err
	}
	
	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
	}, nil
}

// generateAccessToken 生成 Access Token
func (m *JWTManager) generateAccessToken(claims StandardClaims) (string, int64, error) {
	now := time.Now()
	
	claims.Issuer = m.issuer
	claims.IssuedAt = jwt.NewNumericDate(now)
	claims.ExpiresAt = jwt.NewNumericDate(now.Add(m.accessExpire))
	claims.NotBefore = jwt.NewNumericDate(now)
	claims.ID = uuid.New().String()
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(m.accessSecret)
	
	var expiresIn int64
	if err == nil {
		expiresIn = int64(m.accessExpire.Seconds())
	}
	
	return tokenString, expiresIn, err
}

// generateRefreshToken 生成 Refresh Token
func (m *JWTManager) generateRefreshToken(claims StandardClaims, remember bool) (string, int64, error) {
	now := time.Now()
	expire := m.refreshExpire
	if remember {
		expire = time.Duration(claims.RegisteredClaims.ExpiresAt.Sub(now))
	}
	
	refreshClaims := StandardClaims{
		UserID:       claims.UserID,
		Username:     claims.Username,
		RoleCodes:    claims.RoleCodes,
		PlatformType: claims.PlatformType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expire)),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	tokenString, err := token.SignedString(m.refreshSecret)
	
	var expiresIn int64
	if err == nil {
		expiresIn = int64(expire.Seconds())
	}
	
	return tokenString, expiresIn, err
}

// ParseAccessToken 解析 Access Token
func (m *JWTManager) ParseAccessToken(tokenString string) (*StandardClaims, error) {
	return m.parseToken(tokenString, m.accessSecret)
}

// ParseRefreshToken 解析 Refresh Token
func (m *JWTManager) ParseRefreshToken(tokenString string) (*StandardClaims, error) {
	return m.parseToken(tokenString, m.refreshSecret)
}

// parseToken 解析 Token
func (m *JWTManager) parseToken(tokenString string, secret []byte) (*StandardClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, NewAuthError(CodeTokenMalformed, "unexpected signing method", nil)
		}
		return secret, nil
	})
	
	if err != nil {
		if err.Error() == "token is expired" || err.Error() == "token has invalid claims" {
			return nil, NewAuthError(CodeExpiredToken, "token has expired", err)
		}
		return nil, NewAuthError(CodeInvalidToken, "invalid token", err)
	}
	
	if claims, ok := token.Claims.(*StandardClaims); ok && token.Valid {
		return claims, nil
	}
	
	return nil, NewAuthError(CodeInvalidToken, "invalid token claims", nil)
}

// GetExpireTime 获取 Token 过期时间
func (m *JWTManager) GetExpireTime(tokenString string) (time.Time, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &StandardClaims{})
	if err != nil {
		return time.Time{}, err
	}
	
	if claims, ok := token.Claims.(*StandardClaims); ok {
		if claims.ExpiresAt != nil {
			return claims.ExpiresAt.Time, nil
		}
	}
	
	return time.Time{}, NewAuthError(CodeTokenMalformed, "invalid token claims", nil)
}
