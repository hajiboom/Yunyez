// Package auth 提供认证相关的公共功能
package auth

import "errors"

// 认证相关错误定义
var (
	ErrInvalidToken      = errors.New("invalid token")
	ErrExpiredToken      = errors.New("token has expired")
	ErrBlacklistedToken  = errors.New("token has been blacklisted")
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrAccountDisabled   = errors.New("account has been disabled")
	ErrAccountLocked     = errors.New("account has been locked")
	ErrUserNotFound      = errors.New("user not found")
	ErrRoleNotFound      = errors.New("role not found")
	ErrTokenMissing      = errors.New("authorization header is required")
	ErrTokenMalformed    = errors.New("malformed token")
	ErrTokenInvalid      = errors.New("token is invalid")
)

// 认证相关错误码
const (
	CodeSuccess             = 0
	CodeInvalidToken        = 40001
	CodeExpiredToken        = 40002
	CodeBlacklistedToken    = 40003
	CodeInvalidCredentials  = 40004
	CodeAccountDisabled     = 40005
	CodeAccountLocked       = 40006
	CodeUserNotFound        = 40007
	CodeTokenMissing        = 40008
	CodeTokenMalformed      = 40009
	CodeInternalError       = 50001
)

// AuthError 认证错误类型
type AuthError struct {
	Code    int
	Message string
	Err     error
}

func (e *AuthError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return "unknown auth error"
}

func (e *AuthError) Unwrap() error {
	return e.Err
}

// NewAuthError 创建认证错误
func NewAuthError(code int, message string, err error) *AuthError {
	return &AuthError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}
