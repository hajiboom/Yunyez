// Package auth_test 认证系统集成测试
package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	authpkg "yunyez/internal/pkg/auth"
	authcontroller "yunyez/internal/controller/auth"
	modelAuth "yunyez/internal/model/auth"
	authservice "yunyez/internal/service/auth"
)

var (
	testDB      *gorm.DB
	testRedis   *redis.Client
	testHandler *LoginTestHandler
)

// LoginTestHandler 测试用的处理器集合
type LoginTestHandler struct {
	AuthService authservice.AuthService
	AuthCtrl    *authcontroller.LoginController
	JWTManager  *authpkg.JWTManager
	Blacklist   *authpkg.TokenBlacklist
}

// TestMain 测试入口
func TestMain(m *testing.M) {
	var err error
	
	// 1. 初始化测试数据库连接
	testDB, err = initTestDB()
	if err != nil {
		fmt.Printf("Failed to init test DB: %v\n", err)
		os.Exit(1)
	}
	
	// 2. 初始化 Redis
	testRedis, err = initTestRedis()
	if err != nil {
		fmt.Printf("Failed to init test Redis: %v\n", err)
		os.Exit(1)
	}
	
	// 3. 执行数据库迁移
	if err := runMigration(testDB); err != nil {
		fmt.Printf("Failed to run migration: %v\n", err)
		os.Exit(1)
	}
	
	// 4. 创建测试用户
	if err := createTestUser(testDB); err != nil {
		// 如果失败，先清理再重试一次
		cleanupTestData(testDB)
		if err := createTestUser(testDB); err != nil {
			fmt.Printf("Failed to create test user: %v\n", err)
			os.Exit(1)
		}
	}
	
	// 5. 初始化认证组件
	jwtManager := authpkg.NewJWTManager(authpkg.JWTConfig{
		AccessSecret:   "test-secret-2026",
		RefreshSecret:  "test-refresh-secret-2026",
		AccessExpire:   3600,    // 1 小时
		RefreshExpire:  86400,   // 1 天
		Issuer:         "yunyez-test",
	})
	
	blacklist := authpkg.NewTokenBlacklist(testRedis, authpkg.RedisConfig{
		Enabled:   true,
		KeyPrefix: "auth:test:blacklist:",
	})
	
	loginAttempts := authpkg.NewLoginAttemptManager(testRedis, authpkg.LoginSafetyConfig{
		MaxAttempts:  5,
		LockDuration: 900,
	})
	
	authSvc := authservice.NewAuthService(testDB, testRedis, jwtManager, blacklist, loginAttempts)
	
	testHandler = &LoginTestHandler{
		AuthService: authSvc,
		AuthCtrl:    authcontroller.NewLoginController(authSvc),
		JWTManager:  jwtManager,
		Blacklist:   blacklist,
	}
	
	fmt.Println("Test setup completed successfully")
	
	// 运行测试
	code := m.Run()
	
	// 清理测试数据（总是执行）
	cleanupTestData(testDB)
	
	os.Exit(code)
}

// TestLogin_Success 测试登录成功
func TestLogin_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// 准备请求
	loginReq := authpkg.LoginRequest{
		Username: "testuser",
		Password: "Test123456!",
		Remember: false,
	}
	
	body, _ := json.Marshal(loginReq)
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "TestClient/1.0")
	
	w := httptest.NewRecorder()
	
	// 创建 Gin 上下文
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	
	// 调用登录接口
	testHandler.AuthCtrl.Login(c)
	
	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response authcontroller.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, 0, response.Code)
	assert.Equal(t, "success", response.Message)
	
	// 验证返回的数据
	data, ok := response.Data.(map[string]interface{})
	require.True(t, ok)
	
	accessToken, ok := data["access_token"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, accessToken)
	
	refreshToken, ok := data["refresh_token"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, refreshToken)
	
	tokenType, ok := data["token_type"].(string)
	require.True(t, ok)
	assert.Equal(t, "Bearer", tokenType)
	
	expiresIn, ok := data["expires_in"].(float64)
	require.True(t, ok)
	assert.Equal(t, float64(3600), expiresIn)
	
	// 验证用户信息
	user, ok := data["user"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "testuser", user["username"])
	assert.Equal(t, "测试用户", user["nickname"])
	
	// 验证 Token 可解析
	claims, err := testHandler.JWTManager.ParseAccessToken(accessToken)
	require.NoError(t, err)
	assert.Equal(t, "testuser", claims.Username)
	assert.Greater(t, claims.UserID, int64(0)) // 只验证 ID 大于 0
	
	fmt.Printf("✅ Login success - Access Token: %s...\n", accessToken[:20])
	fmt.Printf("✅ User: %+v\n", user)
}

// TestLogin_InvalidPassword 测试密码错误
func TestLogin_InvalidPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	loginReq := authpkg.LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}
	
	body, _ := json.Marshal(loginReq)
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	
	testHandler.AuthCtrl.Login(c)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	var response authcontroller.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, 40004, response.Code)
	assert.Contains(t, response.Message, "密码错误")
	
	fmt.Println("✅ Invalid password test passed")
}

// TestLogin_UserNotFound 测试用户不存在
func TestLogin_UserNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	loginReq := authpkg.LoginRequest{
		Username: "nonexistent",
		Password: "password",
	}
	
	body, _ := json.Marshal(loginReq)
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	
	testHandler.AuthCtrl.Login(c)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	var response authcontroller.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, 40004, response.Code)
	
	fmt.Println("✅ User not found test passed")
}

// TestRefreshToken 测试刷新 Token
func TestRefreshToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// 1. 先登录获取 Token
	loginResp := doLogin(t, "testuser", "Test123456!")
	refreshToken := loginResp.RefreshToken
	
	// 2. 使用 Refresh Token 获取新 Token
	refreshReq := authpkg.RefreshTokenRequest{
		RefreshToken: refreshToken,
	}
	
	body, _ := json.Marshal(refreshReq)
	req, _ := http.NewRequest("POST", "/api/auth/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	
	testHandler.AuthCtrl.RefreshToken(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response authcontroller.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, 0, response.Code)
	
	data, ok := response.Data.(map[string]interface{})
	require.True(t, ok)
	
	newAccessToken, ok := data["access_token"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, newAccessToken)
	
	// 验证新 Token 可解析
	claims, err := testHandler.JWTManager.ParseAccessToken(newAccessToken)
	require.NoError(t, err)
	assert.Equal(t, "testuser", claims.Username)
	
	fmt.Println("✅ Refresh token test passed")
}

// TestLogout 测试登出
func TestLogout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// 1. 先登录
	loginResp := doLogin(t, "testuser", "Test123456!")
	accessToken := loginResp.AccessToken
	userID := loginResp.User.ID
	
	// 2. 登出
	logoutReq := authpkg.LogoutRequest{
		Token: accessToken,
	}
	
	body, _ := json.Marshal(logoutReq)
	req, _ := http.NewRequest("POST", "/api/auth/logout", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	
	// 注入用户信息到 Context
	c.Set("user_id", userID)
	
	testHandler.AuthCtrl.Logout(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	// 3. 验证 Token 已被加入黑名单
	isBlacklisted, err := testHandler.Blacklist.IsBlacklisted(context.Background(), loginResp.JTI)
	require.NoError(t, err)
	assert.True(t, isBlacklisted, "Token should be blacklisted after logout")
	
	fmt.Println("✅ Logout test passed")
}

// TestLogout_InvalidToken 测试使用已登出的 Token 访问
func TestLogout_TokenBlacklisted(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// 1. 登录
	loginResp := doLogin(t, "testuser", "Test123456!")
	accessToken := loginResp.AccessToken
	userID := loginResp.User.ID
	
	// 2. 登出
	logoutReq := authpkg.LogoutRequest{
		Token: accessToken,
	}
	body, _ := json.Marshal(logoutReq)
	req, _ := http.NewRequest("POST", "/api/auth/logout", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", userID)
	testHandler.AuthCtrl.Logout(c)
	
	// 3. 尝试使用已登出的 Token 访问受保护资源
	req2, _ := http.NewRequest("GET", "/api/auth/userinfo", nil)
	req2.Header.Set("Authorization", "Bearer "+accessToken)
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = req2
	
	testHandler.AuthCtrl.GetUserInfo(c2)
	
	// 应该返回 401 或从 Context 中找不到 user_id
	// 注意：这里直接调用控制器，中间件未执行，所以需要手动测试中间件
	// 这里我们验证 Token 确实在黑名单中
	isBlacklisted, err := testHandler.Blacklist.IsBlacklisted(context.Background(), loginResp.JTI)
	require.NoError(t, err)
	assert.True(t, isBlacklisted)
	
	fmt.Println("✅ Token blacklist test passed")
}

// TestGetUserInfo 测试获取用户信息
func TestGetUserInfo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// 1. 登录
	loginResp := doLogin(t, "testuser", "Test123456!")
	userID := loginResp.User.ID
	
	// 2. 获取用户信息
	req, _ := http.NewRequest("GET", "/api/auth/userinfo", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", userID)
	
	testHandler.AuthCtrl.GetUserInfo(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response authcontroller.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, 0, response.Code)
	
	user, ok := response.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "testuser", user["username"])
	assert.Equal(t, "测试用户", user["nickname"])
	
	fmt.Printf("✅ Get user info: %+v\n", user)
}

// ========== 辅助函数 ==========

// LoginResponse 登录响应结构
type LoginResponse struct {
	AccessToken  string
	RefreshToken string
	User         authpkg.UserInfo
	JTI          string
}

// doLogin 执行登录并返回 Token
func doLogin(t *testing.T, username, password string) *LoginResponse {
	gin.SetMode(gin.TestMode)
	
	loginReq := authpkg.LoginRequest{
		Username: username,
		Password: password,
	}
	
	body, _ := json.Marshal(loginReq)
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	
	testHandler.AuthCtrl.Login(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response authcontroller.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	data := response.Data.(map[string]interface{})
	userData := data["user"].(map[string]interface{})
	
	return &LoginResponse{
		AccessToken:  data["access_token"].(string),
		RefreshToken: data["refresh_token"].(string),
		User: authpkg.UserInfo{
			ID:       int64(userData["id"].(float64)),
			Username: userData["username"].(string),
			Nickname: userData["nickname"].(string),
		},
		JTI: extractJTI(t, data["access_token"].(string)),
	}
}

// extractJTI 从 Token 中提取 JTI
func extractJTI(t *testing.T, tokenString string) string {
	claims, err := testHandler.JWTManager.ParseAccessToken(tokenString)
	require.NoError(t, err)
	return claims.ID
}

// initTestDB 初始化测试数据库连接
func initTestDB() (*gorm.DB, error) {
	dsn := "host=localhost port=5432 user=postgres password=root dbname=yunyez sslmode=disable TimeZone=Asia/Shanghai"
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}
	
	// 设置 search_path 包含 auth schema
	if err := db.Exec("SET search_path TO auth, public").Error; err != nil {
		return nil, fmt.Errorf("failed to set search_path: %w", err)
	}
	
	return db, nil
}

// initTestRedis 初始化测试 Redis 连接
func initTestRedis() (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}
	
	return rdb, nil
}

// runMigration 执行数据库迁移
func runMigration(db *gorm.DB) error {
	// 创建 auth schema
	if err := db.Exec("CREATE SCHEMA IF NOT EXISTS auth").Error; err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}
	
	// 自动迁移表
	if err := db.AutoMigrate(
		&modelAuth.User{},
		&modelAuth.Role{},
		&modelAuth.UserRole{},
		&modelAuth.LoginLog{},
	); err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}
	
	fmt.Println("Database migration completed")
	return nil
}

// createTestUser 创建或获取测试用户
func createTestUser(db *gorm.DB) error {
	// 创建角色
	roles := []modelAuth.Role{
		{RoleCode: "super_admin", RoleName: "超级管理员", Status: 1},
		{RoleCode: "admin", RoleName: "管理员", Status: 1},
		{RoleCode: "test_user", RoleName: "测试用户", Status: 1},
	}
	
	for _, role := range roles {
		db.Where("role_code = ?", role.RoleCode).FirstOrCreate(&role, role)
	}
	
	// 加密密码
	passwordHash, err := authpkg.HashPassword("Test123456!")
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	
	// 先查找已存在的用户
	var user modelAuth.User
	err = db.Where("username = ?", "testuser").First(&user).Error
	if err == nil {
		// 用户存在，检查是否有角色关联
		var testRole modelAuth.Role
		if err := db.Where("role_code = ?", "test_user").First(&testRole).Error; err != nil {
			return fmt.Errorf("failed to find test role: %w", err)
		}
		
		// 检查角色关联
		var count int64
		db.Model(&modelAuth.UserRole{}).Where("user_id = ? AND role_id = ?", user.ID, testRole.ID).Count(&count)
		if count == 0 {
			userRole := modelAuth.UserRole{
				UserID: user.ID,
				RoleID: testRole.ID,
			}
			_ = db.Create(&userRole)
		}
		
		fmt.Printf("Test user already exists (ID: %d)\n", user.ID)
		return nil
	}
	
	// 用户不存在，创建新用户（使用唯一值避免约束冲突）
	user = modelAuth.User{
		Username:     "testuser",
		PasswordHash: passwordHash,
		Nickname:     "测试用户",
		Email:        fmt.Sprintf("testuser_%d@example.com", time.Now().Unix()),
		Phone:        fmt.Sprintf("138%08d", time.Now().Unix()%100000000),
		Status:       1,
	}
	
	if err := db.Create(&user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	
	// 查找角色
	var testRole modelAuth.Role
	if err := db.Where("role_code = ?", "test_user").First(&testRole).Error; err != nil {
		return fmt.Errorf("failed to find test role: %w", err)
	}
	
	// 关联角色
	userRole := modelAuth.UserRole{
		UserID: user.ID,
		RoleID: testRole.ID,
	}
	
	if err := db.Create(&userRole).Error; err != nil {
		fmt.Printf("User role relation exists: %v\n", err)
	}
	
	fmt.Printf("Test user created (ID: %d)\n", user.ID)
	return nil
}

// cleanupTestData 清理测试数据
func cleanupTestData(db *gorm.DB) {
	// 清理登录日志
	db.Where("username = ?", "testuser").Delete(&modelAuth.LoginLog{})
	
	// 清理用户角色
	db.Where("user_id IN (SELECT id FROM auth.users WHERE username = ?)", "testuser").
		Delete(&modelAuth.UserRole{})
	
	// 清理用户
	db.Where("username = ?", "testuser").Delete(&modelAuth.User{})
	
	fmt.Println("Test data cleaned")
}
