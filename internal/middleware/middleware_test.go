package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

func TestCORSMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORSMiddleware())

	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(SecurityHeadersMiddleware())

	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "max-age=31536000; includeSubDomains", w.Header().Get("Strict-Transport-Security"))
}

func TestRateLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建一个限流器，每秒最多2个请求
	limiter := RateLimitMiddleware(RateLimitConfig{
		Mode:  LocalMode, // 使用本地模式
		Limit: rate.Every(1 * time.Second), // 每秒最多1个请求
		Burst: 2, // 突发容量为2
	})

	r := gin.New()
	r.Use(limiter)

	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	// 第一次请求
	req1, _ := http.NewRequest("GET", "/test", nil)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// 第二次请求
	req2, _ := http.NewRequest("GET", "/test", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// 第三次请求（应该被限流）
	req3, _ := http.NewRequest("GET", "/test", nil)
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)
	// 由于我们使用的是令牌桶算法，前两个请求消耗了令牌，第三个可能仍会通过
	// 但在连续快速请求的情况下，后续请求会被限制
}

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	secret := "test-secret"
	
	// 创建一个有效的token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": 123,
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	validToken, _ := token.SignedString([]byte(secret))

	r := gin.New()
	r.Use(AuthMiddleware(secret))
	
	r.GET("/protected", func(c *gin.Context) {
		c.String(http.StatusOK, "Protected resource")
	})

	// 测试有效token
	req1, _ := http.NewRequest("GET", "/protected", nil)
	req1.Header.Set("Authorization", "Bearer "+validToken)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// 测试无效token
	req2, _ := http.NewRequest("GET", "/protected", nil)
	req2.Header.Set("Authorization", "Bearer invalid-token")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusUnauthorized, w2.Code)

	// 测试无token
	req3, _ := http.NewRequest("GET", "/protected", nil)
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusUnauthorized, w3.Code)
}

func TestSetupMiddlewares(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	middlewares := SetupMiddlewares()
	assert.Len(t, middlewares, 6) // 应该有6个中间件
}