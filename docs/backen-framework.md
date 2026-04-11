# Yunyez 后端架构规范文档

| 版本 | 日期 | 作者 | 说明 |
|------|------|------|------|
| v1.0 | 2026-04-07 | Qwen | 初始版本，定义后端架构规范与SSO集成方案 |

---

## 1. 架构概述

### 1.1 当前架构（Monolith）

```
Yunyez 主服务
├── HTTP Server (Gin)
│   ├── Admin 管理端路由
│   ├── Open 开放平台路由
│   └── Device 设备端路由
├── 业务逻辑层
│   ├── 3D 重建服务
│   ├── 音频处理服务
│   ├── 视频流服务
│   └── 认证服务（内置）
├── 数据层
│   ├── PostgreSQL
│   └── Redis
└── 消息层
    └── MQTT (EMQX)
```

**特点：**
- 单一可执行文件
- 所有功能耦合在同一进程
- 认证逻辑内置在 `internal/middleware/auth.go`
- 适合早期快速迭代

### 1.2 目标架构（Microservices）

```
┌─────────────────────────────────────────────────────────┐
│                    API Gateway (APISIX/Kong)             │
│  - 统一入口                                              │
│  - SSL 终止                                              │
│  - 限流/熔断                                             │
│  - OIDC 认证插件                                         │
└────────────────┬────────────────────────────────────────┘
                 │
    ┌────────────┼────────────┐
    │            │            │
    ▼            ▼            ▼
┌────────┐  ┌────────┐  ┌──────────┐
│用户中心 │  │主业务服务│ │其他微服务 │
│(User   │  │(Yunyez │  │(3D重建、  │
│Center) │  │Main)   │  │ 音频、视频)│
└────────┘  └────────┘  └──────────┘
    │            │            │
    └────────────┼────────────┘
                 ▼
        ┌────────────────┐
        │  基础设施层     │
        │  - PostgreSQL   │
        │  - Redis        │
        │  - MQTT         │
        └────────────────┘
```

**特点：**
- 用户中心独立部署，专注认证与用户管理
- 业务服务解耦，独立扩展
- API Gateway 统一认证，微服务信任 Gateway
- 支持多团队协作，独立开发/部署

---

## 2. 项目结构规范

### 2.1 当前结构（Monolith）

```
Yunyez/
├── cmd/                      # 程序入口（待完善）
├── internal/                 # 核心代码（私有）
│   ├── app/                  # 应用层
│   ├── common/               # 公共模块
│   ├── controller/           # HTTP 控制器
│   ├── middleware/           # Gin 中间件
│   ├── model/                # 数据模型
│   ├── service/              # 业务服务层
│   └── pkg/                  # 公共工具包
├── api/                      # 对外 API 定义（Protobuf/OpenAPI）
├── configs/                  # 配置文件
└── sql/                      # 数据库脚本
```

### 2.2 微服务化后结构

```
Yunyez/
├── services/                 # 微服务集合
│   ├── user-center/          # 用户中心（SSO）
│   │   ├── cmd/              # 入口
│   │   ├── internal/         # 内部实现
│   │   ├── api/              # API 定义
│   │   └── configs/          # 配置
│   ├── main-service/         # 主业务服务（当前 Yunyez）
│   │   ├── cmd/
│   │   ├── internal/
│   │   └── configs/
│   └── ...                   # 其他微服务
├── gateway/                  # Gateway 配置
│   └── apisix/               # APISIX 配置
├── shared/                   # 共享代码
│   ├── go-auth/              # 认证公共包
│   ├── go-middleware/        # 通用中间件
│   └── go-logger/            # 日志公共包
├── deploy/                   # 部署配置
│   ├── docker-compose.yml
│   └── k8s/                  # K8s 配置
└── docs/                     # 文档
```

**迁移原则：**
1. **渐进式迁移**：不破坏现有功能
2. **共享包提取**：认证、日志等公共代码提取到 `shared/`
3. **API 契约优先**：先定义 Protobuf/OpenAPI，再实现
4. **数据独立**：每个服务独立数据库，通过 API 通信

---

## 3. 认证架构设计

### 3.1 认证模式演进

#### 模式1：当前模式（Monolith Auth）

```go
// internal/middleware/auth.go
func AuthMiddleware(secret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        token, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            return []byte(secret), nil
        })
        // ... 验证逻辑
    }
}
```

**特点：**
- 单 Secret，对称加密（HS256）
- 认证逻辑内置在服务内部
- Redis 黑名单检查本地执行

**局限：**
- 无法跨服务共享认证
- Secret 轮换困难
- 不支持外部 IdP（身份提供商）

---

#### 模式2：过渡期（多Secret支持）

```go
// internal/middleware/auth.go (改造后)
type AuthConfig struct {
    Secrets    map[string]string  // 多Secret支持
    PublicKey  *rsa.PublicKey     // OIDC公钥（预留）
    RedisClient *redis.Client     // 黑名单检查
}

func AuthMiddleware(cfg AuthConfig) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. 提取 Token
        // 2. 尝试多个 Secret 验证
        // 3. 检查 Redis 黑名单
        // 4. 注入用户信息到 Context
    }
}
```

**适用场景：**
- 多个子系统共存期
- 不同端使用不同 Secret
- 为 SSO 迁移做准备

---

#### 模式3：SSO 集成（OIDC 标准）

```
用户请求 → API Gateway (OIDC插件) → 验证JWT → 注入用户信息到Header
                                              ↓
                                    微服务 (信任Gateway)
                                    - 从 Header 提取 X-User-Id
                                    - 从 Header 提取 X-User-Role
                                    - 可选：二次验证JWT签名
```

**Gateway 配置示例（APISIX）：**
```yaml
routes:
  - uri: /api/v1/*
    plugins:
      openid-connect:
        client_id: "yunyez-gateway"
        client_secret: "xxx"
        discovery: "https://sso.yunyez.com/.well-known/openid-configuration"
        scope: "openid profile email"
        set_access_token_header: true
        access_token_in_authorization_header: true
    upstream:
      nodes:
        "main-service:8080": 1
```

**微服务中间件（简化版）：**
```go
// shared/go-middleware/auth.go
func TrustGatewayMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetHeader("X-User-Id")
        userRole := c.GetHeader("X-User-Role")
        
        if userID == "" {
            c.JSON(401, gin.H{"error": "未认证"})
            c.Abort()
            return
        }
        
        // 注入到 Context
        c.Set("user_id", userID)
        c.Set("user_role", userRole)
        c.Next()
    }
}
```

---

### 3.2 Token 格式规范

#### Access Token（JWT）标准Claims

```json
{
  "iss": "https://sso.yunyez.com",           // 签发者（用户中心）
  "sub": "user_12345",                        // 主题（用户ID）
  "aud": ["yunyez-main", "yunyez-3d"],        // 受众（允许的服务）
  "exp": 1712390400,                          // 过期时间
  "iat": 1712383200,                          // 签发时间
  "jti": "uuid-xxxx-xxxx",                    // Token唯一ID（用于吊销）
  
  // 自定义Claims
  "user_id": 12345,
  "username": "admin",
  "roles": ["super_admin", "operator"],
  "platform_type": "admin",
  "tenant_id": "yunyez"                       // 多租户支持（预留）
}
```

#### Token 签发流程

```
用户登录 → 用户中心验证 → 生成JWT (RS256签名)
                    ↓
          返回 {access_token, refresh_token, expires_in}
                    ↓
          客户端存储（HttpOnly Cookie 或 Memory）
```

---

### 3.3 服务间认证（高安全场景）

**方案A：mTLS（双向TLS认证）**

```
服务A ──(mTLS)──▶ 服务B
   │                │
   └── 证书验证 ────┘
```

- 适用于内部服务通信
- Kubernetes Service Mesh (Istio) 自动管理证书
- 零信任架构基础

**方案B：服务令牌（Service Token）**

```
服务A 请求用户中心 → 获取 Service Token → 调用服务B (携带Service Token)
```

```go
// 服务间调用示例
func CallServiceB(ctx context.Context, targetURL string) (*Response, error) {
    // 1. 获取服务令牌
    token, _ := tokenManager.GetServiceToken("service-b")
    
    // 2. 发起请求
    req, _ := http.NewRequest("GET", targetURL, nil)
    req.Header.Set("Authorization", "Bearer "+token)
    
    // 3. 执行请求
    return httpClient.Do(req)
}
```

---

## 4. Middleware 改造指南

### 4.1 当前问题清单

| 问题 | 文件 | 影响 |
|------|------|------|
| 硬编码单Secret | `middleware/auth.go` | 无法支持多服务 |
| 无公钥验证 | `middleware/auth.go` | 无法验证OIDC Token |
| 黑名单检查耦合 | `middleware/auth.go` | 依赖内部Redis实例 |
| 无标准化用户信息提取 | `middleware/auth.go` | Claims结构不统一 |
| 错误码不规范 | `middleware/auth.go` | 与业务错误码不统一 |

### 4.2 改造步骤

#### 阶段一：内部抽象（1周）

**1. 创建认证配置结构**

```go
// internal/service/auth/config.go
type AuthConfig struct {
    // JWT验证模式
    Mode AuthMode // local, oidc, introspection
    
    // 本地模式配置
    Secrets map[string]string
    
    // OIDC模式配置
    OIDCPublicKey     *rsa.PublicKey
    OIDCIssuer        string
    
    // Introspection模式配置
    IntrospectionURL  string
    ClientID          string
    ClientSecret      string
    
    // 通用配置
    RedisClient       *redis.Client
    BlacklistPrefix   string
    CacheTTL          time.Duration
}
```

**2. 重构 AuthMiddleware**

```go
// internal/middleware/auth.go (重构后)
func AuthMiddleware(cfg AuthConfig) gin.HandlerFunc {
    validator := auth.NewValidator(cfg)
    
    return func(c *gin.Context) {
        tokenString := extractToken(c)
        if tokenString == "" {
            respondError(c, 40101, "缺少认证令牌")
            c.Abort()
            return
        }
        
        // 验证Token（支持多种模式）
        claims, err := validator.Validate(c.Request.Context(), tokenString)
        if err != nil {
            respondError(c, mapAuthError(err), err.Error())
            c.Abort()
            return
        }
        
        // 标准化用户信息
        c.Set("user_id", claims.UserID)
        c.Set("username", claims.Username)
        c.Set("roles", claims.Roles)
        c.Set("claims", claims.Raw)
        
        c.Next()
    }
}
```

**3. 创建验证器接口**

```go
// internal/service/auth/validator.go
type Validator interface {
    Validate(ctx context.Context, token string) (*Claims, error)
}

// 本地验证器（当前逻辑）
type LocalValidator struct {
    secrets map[string]string
    redis   *redis.Client
}

// OIDC验证器（预留）
type OIDCValidator struct {
    publicKey *rsa.PublicKey
    issuer    string
}

// Introspection验证器（微服务化）
type IntrospectionValidator struct {
    httpClient   *http.Client
    introspectURL string
    clientID     string
    clientSecret string
    cache        *ristretto.Cache
}
```

---

#### 阶段二：共享包提取（1周）

**提取公共代码到 `shared/go-auth/`**

```
shared/go-auth/
├── middleware.go       # Gin/Echo/其他框架中间件
├── validator.go        # Token验证器接口
├── claims.go           # 标准化Claims结构
├── config.go           # 认证配置
└── errors.go           # 统一错误码
```

**使用方式：**

```go
import "yunyez/shared/go-auth"

// 在主服务中使用
authMiddleware := go-auth.NewGinMiddleware(go-auth.Config{
    Mode: go-auth.ModeLocal,
    Secrets: map[string]string{"admin": secret},
})

// 在微服务中使用（信任Gateway）
authMiddleware := go-auth.NewTrustGatewayMiddleware()
```

---

#### 阶段三：OIDC集成（2周）

**集成 OIDC Provider（以 Casdoor 为例）**

```go
// internal/service/auth/oidc_casdoor.go
type CasdoorClient struct {
    endpoint       string
    clientID       string
    clientSecret   string
    certificate    *x509.Certificate
}

func (c *CasdoorClient) GetPublicKey(ctx context.Context) (*rsa.PublicKey, error) {
    // 从 Casdoor 获取公钥
    // GET /api/get-certificate
}

func (c *CasdoorClient) IntrospectToken(ctx context.Context, token string) (*TokenInfo, error) {
    // 调用 Token Introspection 端点
    // POST /api/login/oauth/introspect
}
```

---

### 4.3 错误码映射表

| 场景 | 错误码 | HTTP状态 | 说明 |
|------|--------|----------|------|
| 缺少Token | 40101 | 401 | 缺少认证令牌 |
| Token格式错误 | 40102 | 401 | Token格式无效 |
| Token过期 | 40103 | 401 | Token已过期 |
| Token签名无效 | 40104 | 401 | Token签名验证失败 |
| Token已吊销 | 40105 | 401 | Token已被吊销 |
| 用户不存在 | 40106 | 401 | 用户不存在或已禁用 |
| Introspection失败 | 40107 | 401 | Token自检失败 |
| 权限不足 | 40301 | 403 | 无权访问该资源 |

---

## 5. 用户中心架构设计

### 5.1 模块划分

```
user-center/
├── cmd/
│   └── server/
│       └── main.go           # 服务入口
├── internal/
│   ├── sso/                  # SSO服务器
│   │   ├── handler.go        # HTTP处理器
│   │   ├── login.html        # 登录页面
│   │   └── session.go        # 会话管理
│   ├── auth/                 # 认证服务
│   │   ├── jwt.go            # JWT签发/验证
│   │   ├── oauth2.go         # OAuth2协议
│   │   ├── oidc.go           # OIDC协议
│   │   └── mfa.go            # 多因素认证
│   ├── user/                 # 用户服务
│   │   ├── crud.go           # 用户CRUD
│   │   ├── role.go           # 角色管理
│   │   └── profile.go        # 用户资料
│   └── audit/                # 审计服务
│       ├── login_log.go      # 登录日志
│       └── operation_log.go  # 操作日志
├── api/
│   ├── proto/                # Protobuf定义
│   └── openapi/              # OpenAPI规范
├── configs/
│   └── config.yaml           # 配置文件
└── sql/
    └── init.sql              # 数据库初始化
```

### 5.2 核心API端点

#### 认证相关

```
# OIDC标准端点
GET  /.well-known/openid-configuration    # 发现文档
GET  /oauth2/authorize                    # 授权端点
POST /oauth2/token                        # Token端点
GET  /oauth2/userinfo                     # 用户信息
GET  /oauth2/introspect                   # Token自检
POST /oauth2/revoke                       # Token吊销
GET  /.well-known/jwks.json               # 公钥端点

# 管理端点
POST /api/v1/users                        # 创建用户
GET  /api/v1/users                        # 查询用户列表
GET  /api/v1/users/{id}                   # 获取用户详情
PUT  /api/v1/users/{id}                   # 更新用户
DELETE /api/v1/users/{id}                 # 删除用户

POST /api/v1/roles                        # 创建角色
GET  /api/v1/roles                        # 查询角色列表
POST /api/v1/users/{id}/roles             # 分配角色

GET  /api/v1/audit/login-logs             # 登录日志查询
GET  /api/v1/audit/operation-logs         # 操作日志查询
```

### 5.3 数据库设计

用户中心独立数据库，与主业务库隔离：

```sql
-- 用户中心数据库
CREATE SCHEMA user_center;

-- 用户表
CREATE TABLE user_center.users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(64) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email VARCHAR(128) UNIQUE,
    phone VARCHAR(20) UNIQUE,
    status SMALLINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

-- 角色表
CREATE TABLE user_center.roles (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(32) UNIQUE NOT NULL,
    name VARCHAR(64) NOT NULL,
    permissions JSONB,
    created_at TIMESTAMPTZ NOT NULL
);

-- 用户角色关联表
CREATE TABLE user_center.user_roles (
    user_id BIGINT REFERENCES user_center.users(id),
    role_id BIGINT REFERENCES user_center.roles(id),
    PRIMARY KEY (user_id, role_id)
);

-- 登录日志表
CREATE TABLE user_center.login_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES user_center.users(id),
    action VARCHAR(16) NOT NULL,
    status SMALLINT NOT NULL,
    ip INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL
);

-- OAuth2客户端表
CREATE TABLE user_center.oauth2_clients (
    id BIGSERIAL PRIMARY KEY,
    client_id VARCHAR(64) UNIQUE NOT NULL,
    client_secret_hash VARCHAR(255) NOT NULL,
    redirect_uris JSONB NOT NULL,
    scopes VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL
);
```

### 5.4 部署架构

#### 开发环境（Docker Compose）

```yaml
# docker-compose.user-center.yml
version: '3.8'

services:
  user-center:
    build: ./user-center
    ports:
      - "8081:8080"
    environment:
      - DB_HOST=postgres-user
      - REDIS_HOST=redis-user
    depends_on:
      - postgres-user
      - redis-user

  postgres-user:
    image: postgres:15
    environment:
      POSTGRES_DB: user_center
      POSTGRES_PASSWORD: root

  redis-user:
    image: redis:7-alpine
```

#### 生产环境（Kubernetes）

```yaml
# k8s/user-center-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: user-center
spec:
  replicas: 3
  selector:
    matchLabels:
      app: user-center
  template:
    metadata:
      labels:
        app: user-center
    spec:
      containers:
      - name: user-center
        image: yunyez/user-center:latest
        ports:
        - containerPort: 8080
        env:
        - name: DB_HOST
          valueFrom:
            secretKeyRef:
              name: db-credentials
              key: host
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
---
apiVersion: v1
kind: Service
metadata:
  name: user-center
spec:
  selector:
    app: user-center
  ports:
  - port: 80
    targetPort: 8080
  type: ClusterIP
```

---

## 6. API Gateway 集成

### 6.1 为什么需要 API Gateway？

| 功能 | 无Gateway | 有Gateway |
|------|-----------|-----------|
| 认证 | 每个服务独立实现 | Gateway统一处理 |
| 限流 | 每个服务独立配置 | 全局统一限流 |
| 路由 | 客户端直连服务 | Gateway动态路由 |
| 监控 | 分散在各服务 | 集中采集 |
| SSL | 每个服务独立配置 | Gateway统一终止 |
| 版本管理 | 客户端适配多版本 | Gateway路由版本 |

### 6.2 Gateway 选型

| 方案 | 语言 | 性能 | 生态 | 学习曲线 | 推荐场景 |
|------|------|------|------|----------|----------|
| **APISIX** | Lua | 高 | 丰富 | 中 | 云原生、K8s |
| **Kong** | Lua | 高 | 丰富 | 中 | 企业级 |
| **Traefik** | Go | 中 | 中 | 低 | 简单场景 |
| **Envoy** | C++ | 极高 | 丰富 | 高 | Service Mesh |

**推荐：APISIX**
- 性能优异（单节点 20k+ QPS）
- 丰富的插件生态（OIDC、限流、熔断等）
- 云原生支持（K8s Ingress）
- 国内社区活跃

### 6.3 APISIX 配置示例

#### 路由配置

```yaml
# apisix/routes.yaml
routes:
  # 用户中心路由（直接暴露）
  - uri: /user-center/*
    plugins:
      limit-count:
        count: 100
        time_window: 60
    upstream:
      nodes:
        "user-center:8080": 1
      type: roundrobin

  # 主业务路由（需要OIDC认证）
  - uri: /api/v1/*
    plugins:
      openid-connect:
        client_id: "yunyez-main"
        client_secret: "xxx"
        discovery: "http://user-center/.well-known/openid-configuration"
        set_access_token_header: true
        header_upstream: true
      limit-count:
        count: 1000
        time_window: 60
    upstream:
      nodes:
        "main-service:8080": 1
      type: roundrobin

  # 设备端路由（使用API Key认证）
  - uri: /device/*
    plugins:
      key-auth:
        header: X-API-Key
      limit-count:
        count: 5000
        time_window: 60
    upstream:
      nodes:
        "device-service:8080": 1
      type: roundrobin
```

#### 全局插件

```yaml
# apisix/plugins.yaml
plugins:
  - prometheus        # 监控指标
  - zipkin            # 链路追踪
  - fault-injection   # 故障注入（测试）
  - response-rewrite  # 响应重写
  - cors              # 跨域处理
```

---

## 7. 迁移路线图

### 7.1 三阶段迁移

```
阶段一（当前）          阶段二（1-2月）        阶段三（3-4月）
─────────────      ───────────────      ───────────────
Monolith Auth  →   内部解耦 + 抽象    →   用户中心独立部署
                      ↓                      ↓
                 共享包提取              API Gateway集成
                      ↓                      ↓
                 支持多Secret           灰度迁移流量
```

### 7.2 详细里程碑

| 阶段 | 任务 | 预计时间 | 负责人 | 验收标准 |
|------|------|----------|--------|----------|
| **阶段一** | 抽象AuthConfig | 2天 | 后端 | 代码Review通过 |
| | 创建Validator接口 | 2天 | 后端 | 单元测试覆盖 |
| | 重构AuthMiddleware | 3天 | 后端 | 集成测试通过 |
| | 提取shared/go-auth | 3天 | 后端 | 多服务可用 |
| **阶段二** | 部署Casdoor | 2天 | 运维 | 健康检查通过 |
| | 迁移用户数据 | 2天 | 后端 | 数据一致性验证 |
| | 实现OIDC客户端 | 5天 | 后端 | 认证流程走通 |
| | 灰度验证 | 3天 | 测试 | 无回归Bug |
| **阶段三** | 部署APISIX | 3天 | 运维 | 性能测试达标 |
| | 配置OIDC插件 | 2天 | 后端 | 端到端测试通过 |
| | 微服务改造 | 5天 | 后端 | 所有服务适配 |
| | 流量切换 | 3天 | 运维 | 100%流量迁移 |

### 7.3 风险控制

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| 数据迁移丢失 | 高 | 低 | 双写验证、数据对比脚本 |
| 认证服务不可用 | 高 | 低 | 多副本部署、健康检查 |
| 性能下降 | 中 | 中 | 性能基准测试、缓存优化 |
| 兼容性问题 | 中 | 中 | 灰度发布、快速回滚预案 |

---

## 8. 最佳实践

### 8.1 认证安全

1. **永远使用HTTPS**：Token在传输过程中必须加密
2. **Token短期有效**：Access Token ≤ 2小时
3. **Refresh Token轮换**：每次刷新签发新Token
4. **公钥定期轮换**：RS256签名每90天轮换
5. **审计日志集中化**：所有认证操作记录到ELK

### 8.2 微服务通信

1. **内部网络隔离**：用户中心部署在独立VPC
2. **mTLS加密**：服务间通信使用双向TLS
3. **最小权限原则**：服务只申请必要权限
4. **超时与重试**：所有RPC调用设置超时
5. **熔断降级**：认证服务故障时优雅降级

### 8.3 代码规范

1. **错误处理统一**：使用标准化错误码
2. **日志结构化**：JSON格式，包含TraceID
3. **配置外部化**：使用ConfigMap/Secrets
4. **健康检查端点**：所有服务暴露 `/health`
5. **优雅关闭**：处理完现有请求再退出

---

## 9. 监控与告警

### 9.1 关键指标

| 指标 | 说明 | 告警阈值 |
|------|------|----------|
| 认证成功率 | 登录成功/总请求 | < 95% |
| Token验证延迟 | P99延迟 | > 500ms |
| 黑名单检查失败率 | Redis不可用比例 | > 1% |
| 用户中心可用性 | 健康检查成功率 | < 99.9% |
| Gateway吞吐量 | 每秒请求数 | 突增/突降 > 50% |

### 9.2 Prometheus 指标

```go
// 用户中心暴露的指标
var (
    authSuccessTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{Name: "auth_success_total"},
        []string{"method"},
    )
    authFailureTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{Name: "auth_failure_total"},
        []string{"method", "reason"},
    )
    tokenValidationDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{Name: "token_validation_duration_seconds"},
        []string{"mode"},
    )
)
```

---

## 附录

### A. 参考资源

- [OIDC 规范](https://openid.net/connect/)
- [OAuth 2.0 RFC 6749](https://tools.ietf.org/html/rfc6749)
- [JWT RFC 7519](https://tools.ietf.org/html/rfc7519)
- [APISIX 文档](https://apisix.apache.org/docs/)
- [Casdoor 文档](https://casdoor.org/)

### B. 相关文档

- [登录认证模块 PRD](../admin/management/login.md)
- [RPC迁移文档](../rpc-migration.md)
- [RTSP服务器文档](../rtsp-server.md)

### C. 变更记录

| 版本 | 日期 | 变更内容 |
|------|------|----------|
| v1.0 | 2026-04-07 | 初始版本，定义架构规范与SSO集成方案 |
