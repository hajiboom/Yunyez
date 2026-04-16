# 登录认证模块 PRD

| 版本 | 日期 | 作者 | 说明 |
|------|------|------|------|
| v1.0 | 2026-04-06 | Qwen | 初始版本 |

---

## 1. 文档概述

### 1.1 目的
定义 Yunyez 项目统一登录认证模块的产品需求，为管理平台、开放平台、设备端提供多端统一的认证体系。

### 1.2 范围
- 管理平台登录（用户名+密码）
- 开放平台认证（API Key / OAuth2）
- 用户与角色管理
- Token 生命周期管理（签发/刷新/吊销）
- 安全机制（限流/锁定/审计）

### 1.3 术语

| 术语 | 说明 |
|------|------|
| JWT | JSON Web Token，无状态认证令牌 |
| Access Token | 短期访问令牌（2小时） |
| Refresh Token | 长期刷新令牌（7天） |
| API Key | 开放平台客户端凭证 |
| Scope | 权限范围标识 |

---

## 2. 业务需求

### 2.1 背景
Yunyez 作为 3D 空间重建与旅行数字足迹系统，存在多端接入场景（后台管理、第三方应用、设备端），需统一认证体系避免重复建设。

### 2.2 目标
- 一套认证体系服务多端（admin / open / api）
- 支持高并发（1000 QPS）与高可用（99.9%）
- 完善的安全机制（密码加密、限流锁定、审计日志）
- 为后续 OAuth2、SSO 扩展预留能力

### 2.3 用户角色

| 角色 | 说明 | 登录方式 |
|------|------|----------|
| 超级管理员 | 系统最高权限，管理用户/角色 | 用户名+密码 |
| 运营人员 | 日常运营、内容管理 | 用户名+密码 |
| 开发者 | 开放平台 API 调用者 | API Key / OAuth2 |
| 只读用户 | 仅查看权限 | 用户名+密码 |

---

## 3. 功能需求

### 3.1 功能列表

| 功能 | 优先级 | 说明 |
|------|--------|------|
| 管理平台登录 | P0 | 用户名+密码登录，签发 JWT |
| Token 刷新 | P0 | refresh_token 换取新 access_token |
| 登出功能 | P0 | Token 加入 Redis 黑名单 |
| 登录失败限流 | P0 | 连续失败 5 次锁定 15 分钟 |
| 密码加密存储 | P0 | bcrypt 加密 |
| 获取当前用户信息 | P0 | 根据 Token 返回用户资料 |
| 修改密码 | P0 | 旧密码校验 + 新密码强度校验 |
| API Key 认证 | P0 | 开放平台服务间调用 |
| 用户/角色管理 | P1 | CRUD 操作（独立模块） |
| 登录日志查询 | P1 | 审计日志查看 |
| OAuth2 授权码模式 | P2 | 第三方应用授权 |
| IP 白名单 | P2 | 开放平台访问控制 |
| 二次验证 | P2 | 敏感操作短信/邮箱验证码 |

---

### 3.2 管理平台登录

#### 3.2.1 流程
```
用户输入账号密码 → 校验格式 → 查询用户 → 校验密码(bcrypt) → 检查锁定状态
  → 签发 JWT(access_token + refresh_token) → 记录登录日志 → 返回 Token
```

#### 3.2.2 规则
- 密码强度：至少 8 位，包含大小写字母+数字+特殊字符
- Access Token 有效期：2 小时
- Refresh Token 有效期：7 天
- "记住登录"选项：Refresh Token 延长至 30 天
- Token 载荷：`{user_id, username, role_codes, platform_type, exp, iat}`

#### 3.2.3 异常处理

| 异常 | 错误码 | 提示 |
|------|--------|------|
| 用户不存在 | 40001 | 用户名或密码错误 |
| 密码错误 | 40002 | 用户名或密码错误 |
| 账号已锁定 | 40003 | 账号已被锁定，请 15 分钟后重试 |
| 账号已禁用 | 40004 | 账号已被禁用，请联系管理员 |
| 参数校验失败 | 40005 | 请求参数错误 |

#### 3.2.4 登录失败锁定
- 连续失败 5 次 → 锁定账号 15 分钟
- Redis Key：`auth:login:fail:{username}`，TTL 15 分钟
- 锁定期间拒绝登录，返回 40003
- 成功登录后清除失败计数

---

### 3.3 Token 刷新

#### 3.3.1 流程
```
客户端携带 refresh_token → 校验 Token 有效性 → 检查黑名单
  → 签发新 access_token → 可选轮换 refresh_token → 返回新 Token
```

#### 3.3.2 规则
- refresh_token 必须在有效期内且未被吊销
- 支持 Token 轮换（可选配置）：旧 refresh_token 失效，签发新 refresh_token
- 刷新后旧 access_token 加入黑名单

#### 3.3.3 异常

| 异常 | 错误码 | 提示 |
|------|--------|------|
| Token 无效 | 40101 | 无效的刷新令牌 |
| Token 已过期 | 40102 | 刷新令牌已过期 |
| Token 已吊销 | 40103 | 刷新令牌已被吊销 |

---

### 3.4 登出功能

#### 3.4.1 流程
```
客户端携带 access_token/refresh_token → 加入 Redis 黑名单
  → 清除登录失败计数 → 记录登出日志 → 返回成功
```

#### 3.4.2 规则
- Redis Key：`auth:token:blacklist:{jti}`，TTL = Token 剩余有效期
- 登出后 Token 立即失效
- 中间件校验 Token 时检查黑名单

---

### 3.5 修改密码

#### 3.5.1 流程
```
输入旧密码 + 新密码 → 校验旧密码 → 校验新密码强度
  → 更新密码(bcrypt) → 吊销该用户所有 Token → 返回成功
```

#### 3.5.2 规则
- 新密码不能与旧密码相同
- 修改密码后，该用户所有已签发 Token 立即失效
- 记录密码修改日志

---

### 3.6 开放平台认证

#### 3.6.1 API Key 认证
- 请求头携带：`X-API-Key: {api_key}` + `X-API-Secret: {api_secret}`
- 服务端校验 Key/Secret 匹配 + 状态启用 + IP 白名单（可选）
- 签发短期 Access Token（1 小时）

#### 3.6.2 Client Credentials 模式
```
POST /api/v1/oauth/token
Body: {
  "grant_type": "client_credentials",
  "client_id": "{api_key}",
  "client_secret": "{api_secret}",
  "scope": "read write"
}
```

#### 3.6.3 Scope 控制

| Scope | 说明 |
|-------|------|
| read | 只读权限 |
| write | 写入权限 |
| admin | 管理权限 |
| device | 设备端权限 |

---

## 4. 数据模型

### 4.1 ER 图
```
roles (1) ────< user_roles >──── (N) users
users (1) ────< login_logs
users (1) ────< api_keys
```

### 4.2 表结构

#### 4.2.1 `auth.users` - 用户表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGSERIAL | PK | 用户 ID |
| username | VARCHAR(64) | UNIQUE, NOT NULL | 用户名 |
| password_hash | VARCHAR(255) | NOT NULL | bcrypt 密码哈希 |
| email | VARCHAR(128) | UNIQUE | 邮箱 |
| phone | VARCHAR(20) | UNIQUE | 手机号 |
| status | SMALLINT | NOT NULL, DEFAULT 1 | 1-启用 2-禁用 3-锁定 |
| platform_type | VARCHAR(16) | NOT NULL | admin/open/api |
| last_login_at | TIMESTAMPTZ | | 最后登录时间 |
| last_login_ip | INET | | 最后登录 IP |
| created_at | TIMESTAMPTZ | NOT NULL | 创建时间 |
| updated_at | TIMESTAMPTZ | NOT NULL | 更新时间 |

索引：`idx_users_username`, `idx_users_email`, `idx_users_phone`, `idx_users_status`

#### 4.2.2 `auth.roles` - 角色表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGSERIAL | PK | 角色 ID |
| code | VARCHAR(32) | UNIQUE, NOT NULL | 角色标识（super_admin/operator/developer/reader） |
| name | VARCHAR(64) | NOT NULL | 角色名称 |
| description | TEXT | | 描述 |
| permissions | JSONB | | 权限列表 |
| created_at | TIMESTAMPTZ | NOT NULL | 创建时间 |

#### 4.2.3 `auth.user_roles` - 用户角色关联表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGSERIAL | PK | ID |
| user_id | BIGINT | FK → users.id, NOT NULL | 用户 ID |
| role_id | BIGINT | FK → roles.id, NOT NULL | 角色 ID |
| created_at | TIMESTAMPTZ | NOT NULL | 创建时间 |

唯一索引：`uk_user_roles_user_role` (user_id, role_id)

#### 4.2.4 `auth.api_keys` - API 密钥表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGSERIAL | PK | ID |
| user_id | BIGINT | FK → users.id, NOT NULL | 关联用户 |
| api_key | VARCHAR(64) | UNIQUE, NOT NULL | API Key |
| api_secret_hash | VARCHAR(255) | NOT NULL | bcrypt 加密 Secret |
| name | VARCHAR(128) | NOT NULL | 密钥名称 |
| scopes | VARCHAR(255) | NOT NULL | 权限范围（逗号分隔） |
| ip_whitelist | JSONB | | IP 白名单 |
| status | SMALLINT | NOT NULL, DEFAULT 1 | 1-启用 2-禁用 |
| expires_at | TIMESTAMPTZ | | 过期时间 |
| last_used_at | TIMESTAMPTZ | | 最后使用时间 |
| created_at | TIMESTAMPTZ | NOT NULL | 创建时间 |

索引：`idx_api_keys_key`, `idx_api_keys_user`, `idx_api_keys_status`

#### 4.2.5 `auth.login_logs` - 登录日志表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGSERIAL | PK | ID |
| user_id | BIGINT | FK → users.id | 用户 ID |
| action | VARCHAR(16) | NOT NULL | login/logout/refresh/change_password |
| status | SMALLINT | NOT NULL | 1-成功 2-失败 |
| fail_reason | VARCHAR(255) | | 失败原因 |
| ip | INET | | 客户端 IP |
| user_agent | TEXT | | User-Agent |
| created_at | TIMESTAMPTZ | NOT NULL | 创建时间 |

索引：`idx_login_logs_user`, `idx_login_logs_created_at`, `idx_login_logs_action`

---

## 5. API 接口

### 5.1 统一响应格式

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

### 5.2 管理平台接口

#### 5.2.1 管理员登录

```
POST /api/v1/admin/login
```

**请求体**
```json
{
  "username": "admin",
  "password": "YourPassword123!",
  "remember": false
}
```

**响应**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "access_token": "eyJhbGci...",
    "refresh_token": "eyJhbGci...",
    "expires_in": 7200,
    "token_type": "Bearer"
  }
}
```

#### 5.2.2 管理员登出

```
POST /api/v1/admin/logout
Authorization: Bearer {access_token}
```

**响应**
```json
{
  "code": 0,
  "message": "登出成功"
}
```

#### 5.2.3 刷新 Token

```
POST /api/v1/admin/refresh
```

**请求体**
```json
{
  "refresh_token": "eyJhbGci..."
}
```

**响应**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "access_token": "eyJhbGci...",
    "refresh_token": "eyJhbGci...",
    "expires_in": 7200,
    "token_type": "Bearer"
  }
}
```

#### 5.2.4 获取当前用户信息

```
GET /api/v1/admin/profile
Authorization: Bearer {access_token}
```

**响应**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "username": "admin",
    "email": "admin@yunyez.com",
    "roles": ["super_admin"],
    "platform_type": "admin",
    "last_login_at": "2026-04-06T10:00:00Z"
  }
}
```

#### 5.2.5 修改密码

```
POST /api/v1/admin/change-password
Authorization: Bearer {access_token}
```

**请求体**
```json
{
  "old_password": "OldPassword123!",
  "new_password": "NewPassword456!"
}
```

**响应**
```json
{
  "code": 0,
  "message": "密码修改成功，请重新登录"
}
```

---

### 5.3 开放平台接口

#### 5.3.1 获取访问令牌

```
POST /api/v1/oauth/token
```

**请求体（Client Credentials）**
```json
{
  "grant_type": "client_credentials",
  "client_id": "ak_xxxxxx",
  "client_secret": "sk_xxxxxx",
  "scope": "read write"
}
```

**响应**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "access_token": "eyJhbGci...",
    "token_type": "Bearer",
    "expires_in": 3600,
    "scope": "read write"
  }
}
```

#### 5.3.2 撤销令牌

```
POST /api/v1/oauth/revoke
```

**请求体**
```json
{
  "token": "eyJhbGci..."
}
```

#### 5.3.3 令牌自检

```
GET /api/v1/oauth/introspect?token={access_token}
```

**响应**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "active": true,
    "user_id": 1,
    "scope": "read write",
    "exp": 1712390400,
    "client_id": "ak_xxxxxx"
  }
}
```

---

### 5.4 错误码规范

| 错误码 | HTTP 状态 | 说明 |
|--------|-----------|------|
| 0 | 200 | 成功 |
| 40001 | 400 | 用户名或密码错误 |
| 40002 | 400 | 密码错误 |
| 40003 | 403 | 账号已锁定 |
| 40004 | 403 | 账号已禁用 |
| 40005 | 400 | 参数校验失败 |
| 40006 | 400 | 密码强度不足 |
| 40101 | 401 | 无效 Token |
| 40102 | 401 | Token 已过期 |
| 40103 | 401 | Token 已吊销 |
| 40301 | 403 | 权限不足 |
| 42901 | 429 | 请求过于频繁 |
| 50000 | 500 | 服务器内部错误 |

---

## 6. 安全设计

### 6.1 密码安全
- 存储：bcrypt（cost=12）
- 传输：HTTPS 加密通道
- 强度：≥8 位，包含大小写字母+数字+特殊字符
- 禁止使用常见弱密码（弱密码字典校验）

### 6.2 Token 安全
- Access Token 短期有效（2 小时）
- Refresh Token 长期有效（7 天），支持吊销
- Token 黑名单存储于 Redis，TTL = 剩余有效期
- JWT 签名算法：HS256（后续可升级 RS256）

### 6.3 登录防护
- 连续失败 5 次 → 锁定 15 分钟
- Redis 限流：单 IP 每分钟最多 30 次登录请求
- 登录日志记录 IP、User-Agent、时间戳

### 6.4 API Key 安全
- API Secret 加密存储（bcrypt）
- 支持 IP 白名单限制
- 支持过期时间设置
- 使用后立即更新时间戳

### 6.5 审计日志
- 所有登录/登出/修改密码操作记录日志
- 日志保留 180 天
- 异常登录（异地 IP、频繁失败）告警

---

## 7. 非功能需求

| 指标 | 要求 |
|------|------|
| 响应时间 | 登录接口 < 200ms（P95） |
| 可用性 | 99.9% |
| 并发支持 | 1000 QPS |
| 数据存储 | PostgreSQL 持久化，Redis 缓存 |
| 日志 | 所有认证操作记录审计日志 |
| 监控 | 登录成功率、失败率、锁定次数监控 |

### 7.1 Redis Key 设计

| Key | 类型 | TTL | 说明 |
|-----|------|-----|------|
| `auth:login:fail:{username}` | String | 15min | 登录失败计数 |
| `auth:lock:{username}` | String | 15min | 账号锁定标记 |
| `auth:token:blacklist:{jti}` | String | Token 剩余时间 | Token 黑名单 |
| `auth:ratelimit:ip:{ip}` | String | 1min | IP 限流计数 |
| `auth:session:{user_id}` | Hash | 7天 | 用户会话信息 |

---

## 8. 后续规划

### 8.1 Phase 2（P1）
- 用户/角色管理 CRUD 界面
- 登录日志查询与导出
- 批量导入/导出用户
- 密码找回（邮箱/短信验证码）

### 8.2 Phase 3（P2）- SSO 单点登录架构

#### 8.2.1 架构演进方案

**当前架构（Monolith Auth）**
```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│  Admin端    │────▶│  Yunyez主服务 │────▶│  PostgreSQL │
│  Open平台   │────▶│  (内置Auth)   │────▶│   Redis     │
│  设备端     │────▶│               │     └─────────────┘
└─────────────┘     └──────────────┘
```

**目标架构（独立用户中心 - User Center）**
```
┌─────────────┐     ┌──────────────┐     ┌──────────────┐
│  Admin端    │────▶│  API Gateway │     │ 用户中心      │
│  Open平台   │────▶│              │────▶│ (User Center)│
│  设备端     │────▶│              │     │  - SSO Server │
└─────────────┘     └──────────────┘     │  - Auth Svc   │
                                          │  - OAuth2/OIDC│
┌─────────────┐     ┌──────────────┐     └──────────────┘
│  3D重建服务 │────▶│  API Gateway │           │
│  音频服务   │────▶│              │───────────┤
│  视频服务   │────▶│              │           ▼
└─────────────┘     └──────────────┘     ┌──────────────┐
                                          │  PostgreSQL  │
                                          │   Redis      │
                                          └──────────────┘
```

#### 8.2.2 SSO 技术方案选型

| 方案 | 适用场景 | 复杂度 | 推荐度 |
|------|---------|--------|--------|
| **OIDC (OpenID Connect)** | 现代Web应用、移动端 | 中 | ⭐⭐⭐⭐⭐ |
| **SAML 2.0** | 企业级、传统系统 | 高 | ⭐⭐⭐ |
| **CAS** | 简单Web SSO | 低 | ⭐⭐⭐ |
| **自定义 Token 共享** | 同域系统 | 低 | ⭐⭐ |

**推荐方案：OIDC (基于 OAuth 2.0)**
- 标准化程度高，业界最佳实践
- 支持现代认证场景（MFA、生物识别等）
- 与现有 OAuth2 架构平滑演进
- 良好的跨语言/跨平台支持

#### 8.2.3 微服务认证流程

**方案A：API Gateway 统一认证（推荐）**

```
用户请求 → API Gateway (验证JWT) → 路由到微服务
                                  ↓
                          微服务信任 Gateway
                          - 从 Header 提取用户信息
                          - X-User-Id: {user_id}
                          - X-User-Role: {role}
                          - 可选：二次验证 JWT
```

**流程说明：**
1. 用户首次访问 → Gateway 重定向到 SSO Server
2. SSO Server 完成认证 → 签发 JWT + ID Token
3. Gateway 验证 Token → 注入用户信息到请求头 → 转发到微服务
4. 微服务信任 Gateway 传递的用户信息（内部网络隔离保证安全）

**方案B：服务间 Token 传递（适合高安全场景）**

```
用户请求 → API Gateway (验证JWT) → 微服务A (验证JWT) → 微服务B (验证JWT)
```

**流程说明：**
1. Gateway 验证 JWT 合法性
2. 微服务独立验证 JWT（共享公钥或Secret）
3. 每个服务独立鉴权，零信任架构

#### 8.2.4 用户中心（User Center）模块设计

```
user-center/
├── sso-server/          # SSO 认证服务器
│   ├── login/           # 统一登录页面
│   ├── mfa/             # 多因素认证
│   └── session/         # 会话管理
├── auth-service/        # 认证服务
│   ├── jwt/             # JWT 签发/验证
│   ├── oauth2/          # OAuth2 授权
│   └── oidc/            # OIDC 协议实现
├── user-service/        # 用户管理服务
│   ├── crud/            # 用户CRUD
│   ├── role/            # 角色权限
│   └── profile/         # 用户资料
└── audit-service/       # 审计日志服务
    ├── login-log/       # 登录日志
    └── operation-log/   # 操作日志
```

#### 8.2.5 当前项目改造步骤

**阶段一：内部解耦（2周）**
1. 提取 `internal/middleware/auth.go` 为独立认证包
2. 抽象 JWT 验证接口，支持多Secret/公钥
3. 添加 Token  introspection（自检）端点
4. 创建 `internal/service/auth/` 独立认证服务层

**阶段二：外部化用户中心（4周）**
1. 新建 `user-center` 独立服务（可放在 `api/user-center/`）
2. 迁移用户表、角色表、登录逻辑到用户中心
3. 实现 OIDC Provider（或使用 Keycloak/Casdoor）
4. 主服务通过 HTTP/gRPC 调用用户中心验证 Token

**阶段三：Gateway 集成（2周）**
1. 部署 API Gateway（推荐：APISIX/Kong/Traefik）
2. 配置 Gateway OIDC 插件
3. 微服务信任 Gateway 传递的用户信息
4. 灰度迁移，逐步切换流量

#### 8.2.6 Middleware 改造建议

**当前架构问题：**
```go
// 当前：硬编码单个 Secret
AuthMiddleware(jwtSecret)
```

**改造后：支持多租户/多服务验证**
```go
// 方案A：多Secret支持（过渡期）
AuthMiddleware(AuthConfig{
    Secrets: map[string]string{
        "admin": adminSecret,
        "open":  openSecret,
    },
    PublicKey: oidcPublicKey,  // SSO公钥
    BlacklistChecker: redisClient,
})

// 方案B：HTTP Introspection（微服务化）
AuthMiddleware(AuthConfig{
    IntrospectionURL: "http://user-center/oauth2/introspect",
    ClientID:         "yunyez-main-service",
    ClientSecret:     introspectionSecret,
    CacheTTL:         5 * time.Minute,  // 本地缓存减少RPC
})
```

#### 8.2.7 技术选型对比

| 组件 | 自研 | Keycloak | Casdoor | Authing（云） |
|------|------|----------|---------|--------------|
| **开发成本** | 高 | 低 | 中 | 无 |
| **运维成本** | 中 | 高 | 中 | 无 |
| **定制化** | 完全 | 受限 | 中等 | 受限 |
| **OIDC支持** | 需开发 | 完整 | 完整 | 完整 |
| **推荐场景** | 深度定制 | 企业级 | 中小型 | 快速上线 |

**推荐：中期使用 Casdoor（开源、轻量、Go编写），长期可评估自研**

### 8.3 Phase 3（其他P2功能）
- OAuth2 授权码模式完整实现
- 二次验证（TOTP/短信/邮箱）
- 设备端免密登录
- 异地登录告警

### 8.4 技术优化
- JWT 签名算法升级 RS256（非对称加密）
- Redis Cluster 支持高可用
- 数据库读写分离
- Token 自动续期（滑动窗口）

---

## 9. SSO 集成检查清单

### 9.1 代码层面
- [ ] 抽象 JWT 验证逻辑，支持多Secret/公钥
- [ ] 添加 Token introspection 客户端
- [ ] 中间件支持可配置认证源（本地/远程）
- [ ] 用户信息从 Token Claims 标准化提取
- [ ] 错误码统一，便于 Gateway 处理

### 9.2 架构层面
- [ ] 评估 API Gateway 方案（APISIX/Kong）
- [ ] 设计服务间信任链（mTLS/Header信任）
- [ ] 规划网络隔离（用户中心独立部署）
- [ ] 设计灰度迁移方案

### 9.3 安全层面
- [ ] 评估 RS256 非对称加密迁移
- [ ] 设计公钥轮换机制
- [ ] 审计日志集中化
- [ ] 敏感操作二次验证

---

## 附录

### A. 技术栈引用
- JWT 中间件：`internal/middleware/auth.go`
- Redis 客户端：`internal/pkg/redis/`
- PostgreSQL 客户端：`internal/pkg/postgre/`
- 日志：`internal/pkg/logger/`

### B. 参考文档
- RFC 7519 - JSON Web Token (JWT)
- RFC 6749 - OAuth 2.0 Authorization Framework
- bcrypt 密码加密标准
