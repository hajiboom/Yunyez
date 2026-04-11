# Middleware SSO 适配指南

| 版本 | 日期 | 作者 | 说明 |
|------|------|------|------|
| v1.0 | 2026-04-07 | Qwen | 定义中间件SSO适配的详细实施方案 |

---

## 1. 当前 Middleware 架构分析

### 1.1 文件结构

```
internal/middleware/
├── auth.go           # JWT认证中间件（需重点改造）
├── cors.go           # CORS跨域中间件
├── example.go        # 示例中间件
├── init.go           # 中间件初始化逻辑
├── logger.go         # 日志中间件
├── middleware_test.go # 测试文件
├── rate_limit.go     # 限流中间件（支持本地/分布式）
├── recovery.go       # Panic恢复中间件
├── security.go       # 安全头部中间件
└── util.go           # 工具函数
```

### 1.2 当前问题清单

| 优先级 | 文件 | 问题 | 影响 |
|--------|------|------|------|
| **P0** | `auth.go` | 硬编码单Secret，不支持多服务 | 无法SSO集成 |
| **P0** | `auth.go` | 仅支持HS256对称加密 | 无法验证OIDC公钥 |
| **P0** | `auth.go` | 无Redis黑名单检查 | Token无法吊销 |
| **P1** | `auth.go` | Claims结构未标准化 | 微服务无法统一使用 |
| **P1** | `auth.go` | 错误码不规范 | 与业务错误码不统一 |
| **P1** | `init.go` | JWT Secret从配置直接读取 | 不支持动态密钥轮换 |
| **P2** | `rate_limit.go` | 限流逻辑与认证分离 | 无法针对认证用户限流 |
| **P2** | `security.go` | CSP策略过于严格 | 可能阻止SSO登录页面 |

---

## 2. Auth Middleware 改造方案

### 2.1 设计目标

1. **向后兼容**：现有代码无需修改即可使用
2. **多模式支持**：Local → Multi-Secret → OIDC → Introspection
3. **可扩展**：支持自定义验证逻辑
4. **高性能**：支持本地缓存减少RPC调用

### 2.2 架构设计

```
┌─────────────────────────────────────────────┐
│          AuthMiddleware (统一入口)           │
└──────────────────┬──────────────────────────┘
                   │
        ┌──────────┼──────────┐
        │          │          │
        ▼          ▼          ▼
   ┌────────┐ ┌────────┐ ┌──────────┐
   │Local   │ │ OIDC   │ │Introspect│
   │Validator│ │Validator│ │Validator │
   └────────┘ └────────┘ └──────────┘
        │          │          │
        └──────────┼──────────┘
                   ▼
          ┌────────────────┐
          │  Claims标准化   │
          │  错误码映射     │
          └────────────────┘
```

### 2.3 代码实现

#### 步骤1: 定义配置结构

**文件**: `internal/service/auth/config.go` (新建)

```go
package auth

import (
	"crypto/rsa"
	"time"
	
	"github.com/redis/go-redis/v9"
)

// AuthMode 定义认证模式
type AuthMode string

const (
	// ModeLocal 本地模式，使用单一Secret验证
	ModeLocal AuthMode = "local"
	
	// ModeMultiSecret 多Secret模式，支持多端不同密钥
	ModeMultiSecret AuthMode = "multi_secret"
	
	// ModeOIDC OIDC模式，使用公钥验证JWT
	ModeOIDC AuthMode = "oidc"
	
	// ModeIntrospection 自检模式，调用用户中心API验证
	ModeIntrospection AuthMode = "introspection"
	
	// ModeTrustGateway 信任Gateway模式，从Header提取用户信息
	ModeTrustGateway AuthMode = "trust_gateway"
)

// AuthConfig 认证配置
type AuthConfig struct {
	// 认证模式
	Mode AuthMode
	
	// ===== 本地模式配置 =====
	// Secret用于HS256验证（向后兼容）
	Secret string
	
	// ===== 多Secret模式配置 =====
	// Secrets映射表，支持不同端使用不同密钥
	// 例如: map[string]string{"admin": "xxx", "open": "yyy"}
	Secrets map[string]string
	
	// ===== OIDC模式配置 =====
	// OIDC公钥，用于RS256验证
	PublicKey *rsa.PublicKey
	
	// OIDC发行者，用于验证iss字段
	OIDCIssuer string
	
	// ===== Introspection模式配置 =====
	// 用户中心自检端点URL
	IntrospectionURL string
	
	// OAuth2客户端ID
	ClientID string
	
	// OAuth2客户端密钥
	ClientSecret string
	
	// ===== 通用配置 =====
	// Redis客户端，用于黑名单检查
	RedisClient *redis.Client
	
	// 黑名单Key前缀
	BlacklistPrefix string
	
	// Token信息缓存时间（减少重复验证）
	CacheTTL time.Duration
	
	// 是否允许缺少Token（用于公开端点）
	Optional bool
}

// DefaultAuthConfig 返回默认配置（向后兼容）
func DefaultAuthConfig(secret string) AuthConfig {
	return AuthConfig{
		Mode:            ModeLocal,
		Secret:          secret,
		BlacklistPrefix: "auth:token:blacklist:",
		CacheTTL:        5 * time.Minute,
		Optional:        false,
	}
}
```

---

#### 步骤2: 定义标准化Claims

**文件**: `internal/service/auth/claims.go` (新建)

```go
package auth

import (
	"github.com/golang-jwt/jwt/v4"
)

// StandardClaims 标准化的用户声明
type StandardClaims struct {
	// JWT标准字段
	TokenID    string   `json:"jti"`    // Token唯一ID
	Issuer     string   `json:"iss"`    // 发行者
	Subject    string   `json:"sub"`    // 主题（用户ID）
	Audience   []string `json:"aud"`    // 受众
	ExpiresAt  int64    `json:"exp"`    // 过期时间
	IssuedAt   int64    `json:"iat"`    // 签发时间
	
	// 业务字段
	UserID     int64    `json:"user_id"`
	Username   string   `json:"username"`
	Email      string   `json:"email"`
	Roles      []string `json:"roles"`
	PlatformType string `json:"platform_type"`
	
	// 原始Claims（保留所有自定义字段）
	Raw jwt.MapClaims
}

// ParseStandardClaims 从JWT解析标准化Claims
func ParseStandardClaims(token *jwt.Token) (*StandardClaims, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims type")
	}
	
	stdClaims := &StandardClaims{
		Raw: claims,
	}
	
	// 解析标准字段
	if jti, ok := claims["jti"].(string); ok {
		stdClaims.TokenID = jti
	}
	if iss, ok := claims["iss"].(string); ok {
		stdClaims.Issuer = iss
	}
	if sub, ok := claims["sub"].(string); ok {
		stdClaims.Subject = sub
	}
	
	// 解析业务字段
	if userID, ok := claims["user_id"].(float64); ok {
		stdClaims.UserID = int64(userID)
	}
	if username, ok := claims["username"].(string); ok {
		stdClaims.Username = username
	}
	if roles, ok := claims["roles"].([]interface{}); ok {
		for _, role := range roles {
			if r, ok := role.(string); ok {
				stdClaims.Roles = append(stdClaims.Roles, r)
			}
		}
	}
	if platform, ok := claims["platform_type"].(string); ok {
		stdClaims.PlatformType = platform
	}
	
	return stdClaims, nil
}
```

---

#### 步骤3: 定义验证器接口

**文件**: `internal/service/auth/validator.go` (新建)

```go
package auth

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net/http"
	"strings"
	"time"
	
	"github.com/golang-jwt/jwt/v4"
)

// Validator 验证器接口
type Validator interface {
	// Validate 验证Token并返回标准化Claims
	Validate(ctx context.Context, tokenString string) (*StandardClaims, error)
}

// LocalValidator 本地验证器（向后兼容）
type LocalValidator struct {
	secret        string
	redisClient   interface{} // 可选，用于黑名单检查
	blacklistPrefix string
}

// NewLocalValidator 创建本地验证器
func NewLocalValidator(secret string, redisClient interface{}, blacklistPrefix string) *LocalValidator {
	return &LocalValidator{
		secret:        secret,
		redisClient:   redisClient,
		blacklistPrefix: blacklistPrefix,
	}
}

// Validate 实现Validator接口
func (v *LocalValidator) Validate(ctx context.Context, tokenString string) (*StandardClaims, error) {
	// 1. 解析JWT
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(v.secret), nil
	})
	
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	
	// 2. 检查黑名单（如果Redis客户端可用）
	claims, _ := ParseStandardClaims(token)
	if claims != nil && claims.TokenID != "" && v.redisClient != nil {
		// TODO: 实现黑名单检查逻辑
		// blacklistKey := v.blacklistPrefix + claims.TokenID
		// exists, err := v.redisClient.Exists(ctx, blacklistKey).Result()
		// if err == nil && exists > 0 {
		//     return nil, fmt.Errorf("token has been revoked")
		// }
	}
	
	// 3. 返回标准化Claims
	return claims, nil
}

// OIDCValidator OIDC验证器（使用公钥验证）
type OIDCValidator struct {
	publicKey  *rsa.PublicKey
	issuer     string
}

// NewOIDCValidator 创建OIDC验证器
func NewOIDCValidator(publicKey *rsa.PublicKey, issuer string) *OIDCValidator {
	return &OIDCValidator{
		publicKey: publicKey,
		issuer:    issuer,
	}
}

// Validate 实现Validator接口
func (v *OIDCValidator) Validate(ctx context.Context, tokenString string) (*StandardClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return v.publicKey, nil
	})
	
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid OIDC token: %w", err)
	}
	
	claims, err := ParseStandardClaims(token)
	if err != nil {
		return nil, err
	}
	
	// 验证发行者
	if v.issuer != "" && claims.Issuer != v.issuer {
		return nil, fmt.Errorf("invalid issuer: %s", claims.Issuer)
	}
	
	return claims, nil
}

// IntrospectionValidator 自检验证器（调用用户中心API）
type IntrospectionValidator struct {
	httpClient   *http.Client
	introspectURL string
	clientID     string
	clientSecret string
}

// NewIntrospectionValidator 创建自检验证器
func NewIntrospectionValidator(introspectURL, clientID, clientSecret string) *IntrospectionValidator {
	return &IntrospectionValidator{
		httpClient:   &http.Client{Timeout: 5 * time.Second},
		introspectURL: introspectURL,
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}

// Validate 实现Validator接口
func (v *IntrospectionValidator) Validate(ctx context.Context, tokenString string) (*StandardClaims, error) {
	// 调用用户中心的 introspection 端点
	req, err := http.NewRequestWithContext(ctx, "POST", v.introspectURL,
		strings.NewReader(fmt.Sprintf("token=%s", tokenString)))
	if err != nil {
		return nil, fmt.Errorf("failed to create introspect request: %w", err)
	}
	
	req.SetBasicAuth(v.clientID, v.clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("introspect request failed: %w", err)
	}
	defer resp.Body.Close()
	
	// TODO: 解析响应并构建StandardClaims
	// 这里简化处理，实际需要完整实现
	
	return &StandardClaims{}, nil
}

// GatewayTrustValidator Gateway信任验证器
type GatewayTrustValidator struct{}

// NewGatewayTrustValidator 创建Gateway信任验证器
func NewGatewayTrustValidator() *GatewayTrustValidator {
	return &GatewayTrustValidator{}
}

// Validate 实现Validator接口（从Header提取，无需验证Token）
func (v *GatewayTrustValidator) Validate(ctx context.Context, tokenString string) (*StandardClaims, error) {
	// Gateway模式下不验证Token，直接信任Gateway传递的信息
	// 实际实现需要从HTTP Context中获取Header
	return &StandardClaims{}, nil
}
```

---

#### 步骤4: 重构 AuthMiddleware

**文件**: `internal/middleware/auth.go` (重构)

```go
package middleware

import (
	"net/http"
	"strings"
	
	"github.com/gin-gonic/gin"
	
	"yunyez/internal/service/auth"
)

// AuthMiddleware 统一认证中间件（重构后）
func AuthMiddleware(cfg auth.AuthConfig) gin.HandlerFunc {
	// 创建验证器
	validator := createValidator(cfg)
	
	return func(c *gin.Context) {
		// 1. 提取Token
		tokenString := extractToken(c)
		
		// 2. 处理可选认证
		if tokenString == "" {
			if cfg.Optional {
				c.Next()
				return
			}
			respondAuthError(c, 40101, "缺少认证令牌")
			c.Abort()
			return
		}
		
		// 3. 验证Token
		claims, err := validator.Validate(c.Request.Context(), tokenString)
		if err != nil {
			respondAuthError(c, mapAuthError(err), err.Error())
			c.Abort()
			return
		}
		
		// 4. 注入标准化用户信息到Context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("roles", claims.Roles)
		c.Set("platform_type", claims.PlatformType)
		c.Set("claims", claims)
		c.Set("token_id", claims.TokenID)
		
		c.Next()
	}
}

// createValidator 根据配置创建合适的验证器
func createValidator(cfg auth.AuthConfig) auth.Validator {
	switch cfg.Mode {
	case auth.ModeLocal:
		return auth.NewLocalValidator(cfg.Secret, cfg.RedisClient, cfg.BlacklistPrefix)
		
	case auth.ModeMultiSecret:
		// 多Secret模式，需要根据Token内容选择正确的Secret
		// TODO: 实现多Secret验证器
		return auth.NewLocalValidator(cfg.Secret, cfg.RedisClient, cfg.BlacklistPrefix)
		
	case auth.ModeOIDC:
		return auth.NewOIDCValidator(cfg.PublicKey, cfg.OIDCIssuer)
		
	case auth.ModeIntrospection:
		return auth.NewIntrospectionValidator(
			cfg.IntrospectionURL,
			cfg.ClientID,
			cfg.ClientSecret,
		)
		
	case auth.ModeTrustGateway:
		return auth.NewGatewayTrustValidator()
		
	default:
		// 默认使用本地验证器
		return auth.NewLocalValidator(cfg.Secret, cfg.RedisClient, cfg.BlacklistPrefix)
	}
}

// extractToken 从请求中提取Token
func extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		// 尝试从查询参数获取（WebSocket等场景）
		return c.Query("token")
	}
	
	// 处理不同的头部格式
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	} else if strings.HasPrefix(authHeader, "Token ") {
		return strings.TrimPrefix(authHeader, "Token ")
	}
	
	return authHeader
}

// respondAuthError 统一认证错误响应
func respondAuthError(c *gin.Context, code int, message string) {
	c.JSON(http.StatusUnauthorized, gin.H{
		"code":    code,
		"message": message,
		"data":    nil,
	})
}

// mapAuthError 映射错误到标准错误码
func mapAuthError(err error) int {
	errMsg := err.Error()
	
	if strings.Contains(errMsg, "token is expired") {
		return 40103 // Token已过期
	}
	if strings.Contains(errMsg, "revoked") || strings.Contains(errMsg, "blacklist") {
		return 40105 // Token已吊销
	}
	if strings.Contains(errMsg, "invalid issuer") {
		return 40106 // 发行者无效
	}
	
	return 40102 // Token格式错误
}

// ===== 向后兼容的便捷函数 =====

// AuthMiddlewareSimple 向后兼容的简单版本
// 使用方式: middleware.AuthMiddlewareSimple(secret)
func AuthMiddlewareSimple(secret string) gin.HandlerFunc {
	cfg := auth.DefaultAuthConfig(secret)
	return AuthMiddleware(cfg)
}

// 保持旧函数名兼容（逐步废弃）
var AuthMiddlewareLegacy = AuthMiddlewareSimple
```

---

### 2.4 使用示例

#### 场景1: 当前项目（向后兼容）

```go
// 无需修改现有代码
jwtSecret := config.GetString("jwt.secret")
router.Use(middleware.AuthMiddlewareSimple(jwtSecret))
```

#### 场景2: 多端不同Secret

```go
cfg := auth.AuthConfig{
    Mode: auth.ModeMultiSecret,
    Secrets: map[string]string{
        "admin": adminSecret,
        "open":  openSecret,
        "device": deviceSecret,
    },
    RedisClient: redisClient,
}
router.Use(middleware.AuthMiddleware(cfg))
```

#### 场景3: OIDC集成（SSO）

```go
cfg := auth.AuthConfig{
    Mode: auth.ModeOIDC,
    PublicKey: oidcPublicKey,
    OIDCIssuer: "https://sso.yunyez.com",
    RedisClient: redisClient,
    CacheTTL: 5 * time.Minute,
}
router.Use(middleware.AuthMiddleware(cfg))
```

#### 场景4: 微服务信任Gateway

```go
cfg := auth.AuthConfig{
    Mode: auth.ModeTrustGateway,
    Optional: false,
}
router.Use(middleware.AuthMiddleware(cfg))
```

---

## 3. 其他 Middleware 调整

### 3.1 Rate Limit 增强

**目标**: 支持基于认证用户的限流

```go
// internal/middleware/rate_limit.go (增强版)

// AuthenticatedRateLimitMiddleware 基于认证用户的限流
func AuthenticatedRateLimitMiddleware(cfg RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 优先使用用户ID限流，其次使用IP
		key := ""
		if userID, exists := c.Get("user_id"); exists {
			key = fmt.Sprintf("user:%d", userID)
		} else {
			key = fmt.Sprintf("ip:%s", c.ClientIP())
		}
		
		// 执行限流检查
		// ... (复用现有逻辑)
	}
}
```

### 3.2 Security Headers 调整

**问题**: CSP策略可能阻止SSO登录页面

```go
// internal/middleware/security.go (调整)

// SecurityHeadersMiddleware 安全头部中间件（SSO适配）
func SecurityHeadersMiddleware(cfg SecurityConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 动态CSP配置，支持SSO域名
		csp := fmt.Sprintf(
			"default-src 'self'; script-src 'self' 'unsafe-inline' %s; connect-src 'self' %s; frame-src %s",
			cfg.SSODomain,
			cfg.SSODomain,
			cfg.SSODomain,
		)
		c.Header("Content-Security-Policy", csp)
		
		// ... 其他头部
	}
}

type SecurityConfig struct {
	SSODomain string // SSO服务器域名
}
```

---

## 4. 迁移检查清单

### 4.1 代码迁移

- [ ] 创建 `internal/service/auth/` 目录
- [ ] 迁移认证逻辑到 `auth.go`、`claims.go`、`validator.go`
- [ ] 重构 `middleware/auth.go` 使用新架构
- [ ] 更新 `middleware/init.go` 支持新配置
- [ ] 编写单元测试（覆盖率 > 80%）
- [ ] 更新集成测试

### 4.2 配置迁移

- [ ] 更新 `configs/dev/default.yaml` 添加认证模式配置
- [ ] 添加 OIDC 公钥配置项
- [ ] 添加用户中心 URL 配置项
- [ ] 更新环境变量文档

### 4.3 文档更新

- [ ] 更新 API 文档
- [ ] 更新中间件使用指南
- [ ] 编写迁移指南
- [ ] 更新部署文档

### 4.4 测试验证

- [ ] 本地模式测试（向后兼容）
- [ ] 多Secret模式测试
- [ ] OIDC模式测试（使用测试IdP）
- [ ] 性能基准测试
- [ ] 灰度发布验证

---

## 5. 性能优化建议

### 5.1 Token验证缓存

```go
// 使用本地缓存减少重复验证
type CachedValidator struct {
    validator auth.Validator
    cache     *ristretto.Cache  // 高性能本地缓存
}

func (c *CachedValidator) Validate(ctx context.Context, token string) (*auth.StandardClaims, error) {
    // 检查缓存
    if cached, found := c.cache.Get(token); found {
        return cached.(*auth.StandardClaims), nil
    }
    
    // 验证Token
    claims, err := c.validator.Validate(ctx, token)
    if err != nil {
        return nil, err
    }
    
    // 写入缓存（TTL = Token剩余有效期）
    c.cache.SetWithTTL(token, claims, time.Until(time.Unix(claims.ExpiresAt, 0)))
    
    return claims, nil
}
```

### 5.2 黑名单检查优化

```go
// 使用Bloom Filter减少Redis访问
type BloomBlacklistChecker struct {
    bloomFilter *bloom.BloomFilter
    redisClient *redis.Client
    prefix      string
}

func (b *BloomBlacklistChecker) IsBlacklisted(ctx context.Context, tokenID string) bool {
    // 先检查Bloom Filter（快速失败）
    if !b.bloomFilter.Test([]byte(tokenID)) {
        return false
    }
    
    // 再检查Redis（确认）
    key := b.prefix + tokenID
    exists, _ := b.redisClient.Exists(ctx, key).Result()
    return exists > 0
}
```

---

## 6. 安全加固建议

### 6.1 密钥管理

| 阶段 | 方案 | 说明 |
|------|------|------|
| 当前 | 配置文件存储 | 基础安全 |
| 过渡 | HashiCorp Vault | 动态密钥 |
| 最终 | KMS服务 | 云密钥管理 |

### 6.2 Token安全

- [ ] 启用 `jti` (JWT ID) 用于精确吊销
- [ ] 添加 `aud` (Audience) 限制Token使用范围
- [ ] 设置合理的 `exp` (过期时间)，不超过2小时
- [ ] 使用 `nbf` (Not Before) 防止Token提前使用

### 6.3 中间件链安全

```
请求 → CORS → 安全头部 → 限流 → 认证 → 业务逻辑
  ↓      ↓        ↓      ↓     ↓       ↓
信任   信任      信任   验证   验证    验证
```

**原则**: 认证中间件应在限流之后、业务逻辑之前执行。

---

## 附录

### A. 错误码对照表

| 错误码 | HTTP状态 | 说明 | 处理建议 |
|--------|----------|------|----------|
| 40101 | 401 | 缺少认证令牌 | 检查请求头 |
| 40102 | 401 | Token格式错误 | 检查Token格式 |
| 40103 | 401 | Token已过期 | 刷新Token |
| 40104 | 401 | Token签名无效 | 检查密钥配置 |
| 40105 | 401 | Token已吊销 | 重新登录 |
| 40106 | 401 | 发行者无效 | 检查OIDC配置 |
| 40301 | 403 | 权限不足 | 检查角色配置 |

### B. 参考资源

- [JWT最佳实践](https://tools.ietf.org/html/rfc8725)
- [OIDC核心规范](https://openid.net/specs/openid-connect-core-1_0.html)
- [Gin中间件指南](https://gin-gonic.com/docs/examples/custom-middleware/)
