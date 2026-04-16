# Yunyez 认证系统实现文档

## 概述

本文档描述 Yunyez 项目的认证系统实现，包括架构设计、API 接口、使用方式和部署说明。

## 架构设计

### 核心组件

```
┌─────────────────────────────────────────────────────────────┐
│                      API Gateway / Router                    │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
        ┌────────────────────────────────────┐
        │      Auth Middleware (重构版)       │
        │  - JWT 验证 (HS256)                │
        │  - Token 黑名单检查 (Redis)         │
        │  - 角色权限验证                     │
        │  - 标准化 Claims 注入 Context       │
        └────────────────┬───────────────────┘
                         │
                         ▼
        ┌────────────────────────────────────┐
        │       Auth Controller              │
        │  - Login / Logout / Refresh        │
        │  - GetUserInfo                     │
        └────────────────┬───────────────────┘
                         │
                         ▼
        ┌────────────────────────────────────┐
        │        Auth Service                │
        │  - 密码验证 (bcrypt)                │
        │  - Token 签发/验证                  │
        │  - 登录日志记录                     │
        │  - 用户角色查询                     │
        └────────────────┬───────────────────┘
                         │
                         ▼
        ┌────────────────────────────────────┐
        │      Auth Public Package           │
        │  - JWTManager                      │
        │  - TokenBlacklist (Redis)          │
        │  - LoginAttemptManager (Redis)     │
        │  - Password Hash/Verify            │
        └────────────────────────────────────┘
```

### 目录结构

```
Yunyez/
├── sql/auth/
│   └── auth.sql                          # 数据库迁移脚本
├── internal/
│   ├── pkg/auth/                         # 认证公共包
│   │   ├── config.go                     # 认证配置结构
│   │   ├── claims.go                     # 标准化 Claims 定义
│   │   ├── jwt.go                        # JWT 签发/验证
│   │   ├── password.go                   # 密码加密/验证
│   │   ├── validator.go                  # Token 黑名单 & 登录失败管理
│   │   └── errors.go                     # 统一错误码
│   ├── model/auth/                       # 数据模型层
│   │   ├── user.go                       # 用户模型
│   │   ├── role.go                       # 角色模型
│   │   ├── user_role.go                  # 用户角色关联模型
│   │   └── login_log.go                  # 登录日志模型
│   ├── service/auth/                     # 业务服务层
│   │   └── auth_service.go               # 认证核心服务
│   ├── controller/auth/                  # 控制器层
│   │   └── login_controller.go           # 登录/登出控制器
│   ├── middleware/
│   │   ├── auth.go                       # 认证中间件 (重构版)
│   │   └── init.go                       # 中间件初始化 (已移除全局 Auth)
│   └── app/routes/
│       └── auth_routes.go                # 路由注册示例
```

## 数据库初始化

### 1. 执行迁移脚本

```bash
psql -h localhost -U postgres -d yunyez -f sql/auth/auth.sql
```

### 2. 创建的表

| 表名 | 说明 |
|------|------|
| `auth.users` | 用户表 |
| `auth.roles` | 角色表 (预置: super_admin, admin, operator, viewer) |
| `auth.user_roles` | 用户角色关联表 |
| `auth.login_logs` | 登录日志表 |
| `auth.api_keys` | API Keys 表 (开放平台使用) |
| `auth.token_blacklist` | Token 黑名单表 |

### 3. 默认数据

脚本会自动插入以下数据：

- **默认角色**: super_admin, admin, operator, viewer
- **默认管理员**: 用户名 `admin` (密码需要在生产环境中修改)

## API 接口

### 认证相关端点

| 方法 | 路径 | 认证 | 说明 |
|------|------|------|------|
| POST | `/api/auth/login` | ❌ | 用户登录 |
| POST | `/api/auth/logout` | ✅ | 用户登出 |
| POST | `/api/auth/refresh` | ❌ | 刷新 Token |
| GET | `/api/auth/userinfo` | ✅ | 获取当前用户信息 |

### 请求/响应示例

#### 1. 登录

**请求:**
```http
POST /api/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "admin123",
  "remember": false
}
```

**成功响应 (200):**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "token_type": "Bearer",
    "expires_in": 7200,
    "user": {
      "id": 1,
      "username": "admin",
      "nickname": "系统管理员",
      "email": "",
      "phone": "",
      "avatar": "",
      "roles": ["super_admin"]
    }
  }
}
```

**失败响应 (401):**
```json
{
  "code": 40004,
  "message": "用户名或密码错误"
}
```

#### 2. 刷新 Token

**请求:**
```http
POST /api/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**成功响应 (200):**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "token_type": "Bearer",
    "expires_in": 7200
  }
}
```

#### 3. 登出

**请求:**
```http
POST /api/auth/logout
Authorization: Bearer <access_token>
```

**成功响应 (200):**
```json
{
  "code": 0,
  "message": "success"
}
```

#### 4. 获取用户信息

**请求:**
```http
GET /api/auth/userinfo
Authorization: Bearer <access_token>
```

**成功响应 (200):**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "username": "admin",
    "nickname": "系统管理员",
    "email": "",
    "phone": "",
    "avatar": "",
    "roles": ["super_admin"]
  }
}
```

## 错误码

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 40001 | Token 无效 |
| 40002 | Token 已过期 |
| 40003 | Token 已被吊销 (在黑名单中) |
| 40004 | 用户名或密码错误 |
| 40005 | 账户已被禁用 |
| 40006 | 账户已被锁定 |
| 40007 | 用户不存在 |
| 40008 | 缺少 Authorization 头 |
| 40009 | Token 格式错误 |
| 50001 | 内部服务错误 |

## 使用方式

### 1. 在路由中注册认证

参考 `internal/app/routes/auth_routes.go`:

```go
import (
    "yunyez/internal/app/routes"
    "yunyez/internal/pkg/auth"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB, redisClient *redis.Client) {
    // 注册认证路由
    authDeps := routes.AuthDependencies{
        DB:          db,
        RedisClient: redisClient,
        AuthConfig:  routes.DefaultAuthConfig(),
    }
    routes.SetupAuthRoutes(r, authDeps)
    
    // 其他路由...
}
```

### 2. 使用认证中间件保护路由

```go
import (
    "yunyez/internal/middleware"
    "yunyez/internal/pkg/auth"
)

// 创建中间件配置
authCfg := middleware.AuthMiddlewareConfig{
    JWTManager:  jwtManager,
    Blacklist:   blacklist,
    RedisClient: redisClient,
    RequiredRoles: []string{"admin", "super_admin"}, // 可选：限制角色
}

// 保护路由
adminGroup := r.Group("/admin")
adminGroup.Use(middleware.AuthMiddleware(authCfg))
{
    adminGroup.GET("/users", listUsersHandler)
    adminGroup.POST("/users", createUserHandler)
}
```

### 3. 在处理器中获取用户信息

```go
func someHandler(c *gin.Context) {
    // 从 Context 中获取用户信息 (由 AuthMiddleware 注入)
    userID := c.GetInt64("user_id")
    username := c.GetString("username")
    roleCodes := c.GetStringSlice("role_codes")
    
    // 或者获取完整的 Claims 对象
    claims := c.MustGet("claims").(*auth.StandardClaims)
    
    // 业务逻辑...
}
```

## Token 设计

### Access Token
- **有效期**: 2 小时 (7200 秒)
- **用途**: 接口认证
- **载荷**: user_id, username, role_codes, platform_type, exp, iat, jti

### Refresh Token
- **有效期**: 7 天 (604800 秒)，"记住登录" 时 30 天
- **用途**: 刷新 Access Token
- **载荷**: 与 Access Token 相同

### Token 黑名单
- 登出或修改密码时，将 Token 加入 Redis 黑名单
- Key 格式: `auth:token:blacklist:{jti}`
- TTL: Token 剩余有效期
- 中间件在验证 Token 后会检查黑名单

## 安全机制

### 1. 密码加密
- 使用 bcrypt (cost=12) 加密存储
- 验证时使用恒定时间比较，防止时序攻击

### 2. Token 吊销
- 登出时立即将 Token 加入黑名单
- 中间件每次请求都会检查黑名单

### 3. 登录失败锁定 (Phase 2)
- 连续 5 次失败锁定账户 15 分钟
- Redis Key: `auth:lock:{username}`

### 4. IP 限流 (Phase 2)
- 单 IP 每分钟最多 30 次登录请求

## 配置说明

### 开发环境配置

```yaml
# configs/dev/auth.yaml (示例)
auth:
  mode: local
  jwt:
    access_secret: "your-secret-key-change-in-production"
    refresh_secret: "your-refresh-secret-key"
    access_expire: 7200        # 2 小时
    refresh_expire: 604800     # 7 天
    remember_expire: 2592000   # 30 天
    issuer: "yunyez-auth"
  redis:
    enabled: true
    key_prefix: "auth:token:blacklist:"
  login_safety:
    max_attempts: 5
    lock_duration: 900    # 15 分钟
    rate_limit: 30        # 每分钟 30 次
```

### 生产环境注意事项

1. **修改默认密钥**: 务必更换 `access_secret` 和 `refresh_secret`
2. **修改默认管理员密码**: 登录 `admin` 账户后立即修改密码
3. **启用 HTTPS**: 生产环境必须使用 HTTPS 传输
4. **启用 Redis**: Token 黑名单和登录锁定依赖 Redis
5. **日志审计**: 定期检查 `auth.login_logs` 表，发现异常登录

## 演进路线

### Phase 1 - 基础认证功能 (已完成 ✅)

**完成日期**: 2026-04-12  
**测试报告**: [TEST_REPORT.md](./TEST_REPORT.md)  
**测试通过率**: 100% (7/7)

- [x] 数据库表结构 (users, roles, user_roles, login_logs)
- [x] JWT 认证 (HS256)
- [x] 登录/登出/刷新 Token
- [x] Token 黑名单 (Redis)
- [x] 标准化 Claims (user_id, username, role_codes)
- [x] 统一错误码
- [x] 审计日志 (login_logs)
- [x] bcrypt 密码加密 (cost=12)
- [x] 用户角色关联
- [x] 中间件重构 (支持多模式、黑名单、角色检查)

**已实现的 API 端点**:
| 端点 | 状态 | 测试 |
|------|------|------|
| POST /api/auth/login | ✅ 已实现 | ✅ 已测试 |
| POST /api/auth/logout | ✅ 已实现 | ✅ 已测试 |
| POST /api/auth/refresh | ✅ 已实现 | ✅ 已测试 |
| GET /api/auth/userinfo | ✅ 已实现 | ✅ 已测试 |

**已测试场景**:
- ✅ 登录成功
- ✅ 密码错误处理
- ✅ 用户不存在处理
- ✅ Token 刷新
- ✅ 登出和 Token 黑名单
- ✅ 获取用户信息

### Phase 2 - 安全增强 (待实现)

**优先级**: 高  
**预计工作量**: 中等

- [ ] 登录失败锁定实现 (代码已预留接口)
  - [ ] Redis 失败计数
  - [ ] 账户锁定机制 (5 次失败锁定 15 分钟)
  - [ ] 锁定状态查询
- [ ] IP 限流实现 (代码已预留配置)
  - [ ] Redis 滑动窗口
  - [ ] 单 IP 每分钟 30 次登录限制
- [ ] 修改密码功能
  - [ ] 验证旧密码
  - [ ] 更新密码
  - [ ] 吊销所有旧 Token
- [ ] 用户管理 CRUD
  - [ ] 创建用户
  - [ ] 更新用户信息
  - [ ] 禁用/启用用户
  - [ ] 删除用户 (软删除)
- [ ] API Key 管理
  - [ ] 生成 API Key
  - [ ] API Key 验证
  - [ ] 权限控制
  - [ ] IP 白名单

### Phase 3 - SSO/OIDC (规划中)

**优先级**: 低  
**预计工作量**: 大

- [ ] 升级 RS256 非对称加密
- [ ] 集成 OIDC Provider (Casdoor/Keycloak)
- [ ] Token Introspection 端点
- [ ] 微服务统一认证
- [ ] 独立用户中心服务

## 测试

### 测试报告

详细的测试报告请参阅: [TEST_REPORT.md](./TEST_REPORT.md)

### 测试统计

| 指标 | 值 |
|------|-----|
| 总测试用例数 | 7 |
| 通过 | 7 ✅ |
| 失败 | 0 |
| 通过率 | 100% |
| 总执行时间 | ~1.5 秒 |

### 运行测试

```bash
# 运行所有认证集成测试
go test -v ./tests/...

# 运行单个测试
go test -v ./tests/... -run TestLogin_Success
go test -v ./tests/... -run TestRefreshToken
go test -v ./tests/... -run TestLogout
```

### 测试结果摘要

```
✅ TestLogin_Success - 登录成功，返回有效 Token
✅ TestLogin_InvalidPassword - 密码错误正确处理
✅ TestLogin_UserNotFound - 用户不存在正确处理
✅ TestRefreshToken - Token 刷新成功
✅ TestLogout - 登出成功，Token 加入黑名单
✅ TestLogout_TokenBlacklisted - Token 黑名单机制有效
✅ TestGetUserInfo - 获取用户信息成功
```

### 手动测试

```bash
# 1. 登录
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}'

# 2. 使用返回的 access_token 获取用户信息
curl -X GET http://localhost:8080/api/auth/userinfo \
  -H "Authorization: Bearer <access_token>"

# 3. 刷新 Token
curl -X POST http://localhost:8080/api/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "<refresh_token>"}'

# 4. 登出
curl -X POST http://localhost:8080/api/auth/logout \
  -H "Authorization: Bearer <access_token>"
```

## 常见问题

### Q: 为什么移除了全局 AuthMiddleware?

A: 认证应该按路由选择性应用。登录、健康检查等公开端点不需要认证。新的中间件配置允许你灵活地为不同路由组应用不同的认证策略。

### Q: 如何在微服务间共享认证?

A: 当前实现是单体内的。未来演进到 Phase 3 时，会提取独立的用户中心服务，通过 gRPC 或 HTTP 提供 Token 验证接口。

### Q: 如何支持多端不同密钥?

A: 配置中的 `Mode` 字段预留了 `multi_secret` 模式。后续可以根据平台类型 (admin, open, device) 使用不同的 JWT Secret。

### Q: Redis 不可用时会怎样?

A: Token 黑名单检查会跳过，不影响正常认证。但登出后 Token 无法立即失效。建议生产环境必须启用 Redis。

## 参考文档

- [docs/admin/management/login.md](../../docs/admin/management/login.md) - 登录认证模块 PRD
- [docs/middleware-sso-adaptation.md](../../docs/middleware-sso-adaptation.md) - Middleware SSO 适配指南
- [docs/backen-framework.md](../../docs/backen-framework.md) - 后端架构规范
