# Yunyez 认证系统文档

## 文档索引

| 文档 | 说明 |
|------|------|
| [IMPLEMENTATION.md](./IMPLEMENTATION.md) | 实现文档，包含架构设计、API 接口、使用方式、进度状态 |
| [TEST_REPORT.md](./TEST_REPORT.md) | 测试报告，包含测试步骤、测试结果、问题与修复 |

## 快速开始

### 1. 数据库初始化

```bash
# 执行数据库迁移
psql -h localhost -U postgres -d yunyez -f sql/auth/auth.sql
```

### 2. 运行测试

```bash
# 运行所有认证测试
go test -v ./tests/...
```

### 3. 集成到项目

参考 [IMPLEMENTATION.md](./IMPLEMENTATION.md) 中的"使用方式"章节。

## 当前进度

### ✅ Phase 1 - 基础认证功能 (已完成)

- 数据库表结构
- JWT 认证 (HS256)
- 登录/登出/刷新 Token
- Token 黑名单
- 标准化 Claims
- 统一错误码
- 审计日志

**测试通过率**: 100% (7/7)

### 📋 Phase 2 - 安全增强 (待实现)

- 登录失败锁定
- IP 限流
- 修改密码
- 用户管理 CRUD
- API Key 管理

### 🔮 Phase 3 - SSO/OIDC (规划中)

- RS256 非对称加密
- OIDC Provider 集成
- 微服务统一认证

## API 端点

| 方法 | 路径 | 认证 | 说明 |
|------|------|------|------|
| POST | /api/auth/login | ❌ | 用户登录 |
| POST | /api/auth/logout | ✅ | 用户登出 |
| POST | /api/auth/refresh | ❌ | 刷新 Token |
| GET | /api/auth/userinfo | ✅ | 获取用户信息 |

## 核心组件

```
internal/
├── pkg/auth/              # 认证公共包
│   ├── config.go          # 配置
│   ├── claims.go          # Claims 定义
│   ├── jwt.go             # JWT 管理
│   ├── password.go        # 密码加密
│   ├── validator.go       # 黑名单/锁定管理
│   └── errors.go          # 错误码
├── model/auth/            # 数据模型
│   ├── user.go
│   ├── role.go
│   ├── user_role.go
│   └── login_log.go
├── service/auth/          # 业务服务
│   └── auth_service.go
├── controller/auth/       # HTTP 控制器
│   └── login_controller.go
├── middleware/
│   └── auth.go            # 认证中间件 (重构版)
└── app/routes/
    └── auth_routes.go     # 路由注册
```

## 技术栈

- **JWT**: github.com/golang-jwt/jwt/v4 (HS256)
- **密码加密**: golang.org/x/crypto/bcrypt (cost=12)
- **缓存**: github.com/redis/go-redis/v9
- **ORM**: gorm.io/gorm + gorm.io/driver/postgres
- **Web 框架**: github.com/gin-gonic/gin

## 相关链接

- [登录认证模块 PRD](../admin/management/login.md)
- [Middleware SSO 适配指南](../middleware-sso-adaptation.md)
- [后端架构规范](../backen-framework.md)
