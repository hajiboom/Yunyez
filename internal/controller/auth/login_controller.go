// Package authcontroller 认证控制器
package authcontroller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	authpkg "yunyez/internal/pkg/auth"
	"yunyez/internal/service/auth"
)

// LoginController 登录控制器
type LoginController struct {
	authService authservice.AuthService
}

// NewLoginController 创建登录控制器
func NewLoginController(authService authservice.AuthService) *LoginController {
	return &LoginController{
		authService: authService,
	}
}

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    authpkg.CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

// Error 错误响应
func Error(c *gin.Context, httpStatus int, code int, message string) {
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: message,
	})
}

// Login 用户登录
// POST /api/auth/login
func (ctrl *LoginController) Login(c *gin.Context) {
	var req authpkg.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, authpkg.CodeTokenMalformed, "请求参数错误: "+err.Error())
		return
	}
	
	// 提取 IP 和 User-Agent
	req.IP = c.ClientIP()
	req.UserAgent = c.GetHeader("User-Agent")
	
	// 调用登录服务
	resp, err := ctrl.authService.Login(c.Request.Context(), &req)
	if err != nil {
		// 判断错误类型
		if authErr, ok := err.(*authpkg.AuthError); ok {
			Error(c, http.StatusUnauthorized, authErr.Code, authErr.Error())
			return
		}
		Error(c, http.StatusInternalServerError, authpkg.CodeInternalError, "登录失败: "+err.Error())
		return
	}
	
	Success(c, resp)
}

// Logout 用户登出
// POST /api/auth/logout
func (ctrl *LoginController) Logout(c *gin.Context) {
	// 从 Context 中获取用户信息 (由 AuthMiddleware 注入)
	userID, exists := c.Get("user_id")
	if !exists {
		Error(c, http.StatusUnauthorized, authpkg.CodeTokenMissing, "未登录")
		return
	}
	
	// 获取 Token
	token := c.GetHeader("Authorization")
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}
	
	// 调用登出服务
	if err := ctrl.authService.Logout(c.Request.Context(), userID.(int64), token); err != nil {
		Error(c, http.StatusInternalServerError, authpkg.CodeInternalError, "登出失败: "+err.Error())
		return
	}
	
	Success(c, nil)
}

// RefreshToken 刷新 Token
// POST /api/auth/refresh
func (ctrl *LoginController) RefreshToken(c *gin.Context) {
	var req authpkg.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, authpkg.CodeTokenMalformed, "请求参数错误: "+err.Error())
		return
	}
	
	// 调用刷新服务
	tokenPair, err := ctrl.authService.RefreshToken(c.Request.Context(), &req)
	if err != nil {
		if authErr, ok := err.(*authpkg.AuthError); ok {
			Error(c, http.StatusUnauthorized, authErr.Code, authErr.Error())
			return
		}
		Error(c, http.StatusInternalServerError, authpkg.CodeInternalError, "刷新 Token 失败: "+err.Error())
		return
	}
	
	Success(c, tokenPair)
}

// GetUserInfo 获取当前用户信息
// GET /api/auth/userinfo
func (ctrl *LoginController) GetUserInfo(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		Error(c, http.StatusUnauthorized, authpkg.CodeTokenMissing, "未登录")
		return
	}
	
	user, err := ctrl.authService.GetUserByID(c.Request.Context(), userID.(int64))
	if err != nil {
		if authErr, ok := err.(*authpkg.AuthError); ok {
			Error(c, http.StatusUnauthorized, authErr.Code, authErr.Error())
			return
		}
		Error(c, http.StatusInternalServerError, authpkg.CodeInternalError, "获取用户信息失败: "+err.Error())
		return
	}
	
	// 获取角色
	roles, err := ctrl.authService.GetUserRoles(c.Request.Context(), user.ID)
	if err != nil {
		Error(c, http.StatusInternalServerError, authpkg.CodeInternalError, "获取角色信息失败: "+err.Error())
		return
	}
	
	Success(c, authpkg.UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Email:    user.Email,
		Phone:    user.Phone,
		Avatar:   user.AvatarURL,
		Roles:    roles,
	})
}
