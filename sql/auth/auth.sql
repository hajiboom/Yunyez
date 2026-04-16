-- ============================================================
-- Yunyez 认证系统数据库迁移脚本
-- Schema: auth
-- 创建时间: 2026-04-12
-- 说明: 管理平台用户认证相关表
-- ============================================================

-- 确保 auth schema 存在
CREATE SCHEMA IF NOT EXISTS auth;

-- ============================================================
-- 1. 角色表 (auth.roles)
-- ============================================================
CREATE TABLE auth.roles (
    id              BIGSERIAL       PRIMARY KEY,
    role_code       VARCHAR(64)     NOT NULL UNIQUE,              -- 角色代码 (如: admin, operator, viewer)
    role_name       VARCHAR(128)    NOT NULL,                     -- 角色名称
    description     VARCHAR(256),                                  -- 角色描述
    status          SMALLINT        NOT NULL DEFAULT 1,           -- 状态: 1-启用, 0-禁用
    created_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at      TIMESTAMP                                       -- 软删除标记
    
    CONSTRAINT chk_roles_status CHECK (status IN (0, 1))
);

-- 创建索引
CREATE INDEX idx_roles_status ON auth.roles(status);
CREATE INDEX idx_roles_deleted ON auth.roles(deleted_at);

-- 插入默认角色
INSERT INTO auth.roles (role_code, role_name, description) VALUES
    ('super_admin', '超级管理员', '拥有系统所有权限'),
    ('admin', '管理员', '拥有管理平台大部分权限'),
    ('operator', '操作员', '日常操作权限'),
    ('viewer', '观察者', '仅查看权限')
ON CONFLICT (role_code) DO NOTHING;

COMMENT ON TABLE auth.roles IS '角色表';
COMMENT ON COLUMN auth.roles.role_code IS '角色代码';
COMMENT ON COLUMN auth.roles.role_name IS '角色名称';
COMMENT ON COLUMN auth.roles.status IS '状态: 1-启用, 0-禁用';

-- ============================================================
-- 2. 用户表 (auth.users)
-- ============================================================
CREATE TABLE auth.users (
    id                  BIGSERIAL       PRIMARY KEY,
    username            VARCHAR(64)     NOT NULL UNIQUE,          -- 用户名 (登录用)
    password_hash       VARCHAR(128)    NOT NULL,                 -- bcrypt 密码哈希
    nickname            VARCHAR(128),                              -- 昵称
    email               VARCHAR(128)    UNIQUE,                   -- 邮箱
    phone               VARCHAR(32)     UNIQUE,                   -- 手机号
    avatar_url          VARCHAR(256),                              -- 头像 URL
    status              SMALLINT        NOT NULL DEFAULT 1,       -- 状态: 1-启用, 0-禁用, 2-锁定
    last_login_at       TIMESTAMP,                                 -- 最后登录时间
    last_login_ip       VARCHAR(64),                               -- 最后登录 IP
    failed_attempts     INT             NOT NULL DEFAULT 0,       -- 连续失败次数
    locked_at           TIMESTAMP,                                 -- 锁定时间
    created_at          TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at          TIMESTAMP,                                 -- 软删除标记
    
    CONSTRAINT chk_users_status CHECK (status IN (0, 1, 2))
);

-- 创建索引
CREATE INDEX idx_users_status ON auth.users(status);
CREATE INDEX idx_users_deleted ON auth.users(deleted_at);
CREATE INDEX idx_users_phone ON auth.users(phone);
CREATE INDEX idx_users_email ON auth.users(email);

-- 插入默认超级管理员 (密码: admin123, bcrypt cost=12)
-- 注意: 实际部署时应修改默认密码
INSERT INTO auth.users (username, password_hash, nickname, role_code) VALUES
    ('admin', '$2a$12$LJ3m4ys5Lq2zK4k/bqHA5eUa3qzK.yPqZ8c5n5jZqF5x5qF5x5qF5', '系统管理员', 'super_admin')
ON CONFLICT (username) DO NOTHING;

COMMENT ON TABLE auth.users IS '用户表';
COMMENT ON COLUMN auth.users.username IS '用户名';
COMMENT ON COLUMN auth.users.password_hash IS 'bcrypt 密码哈希';
COMMENT ON COLUMN auth.users.status IS '状态: 1-启用, 0-禁用, 2-锁定';
COMMENT ON COLUMN auth.users.failed_attempts IS '连续失败次数';
COMMENT ON COLUMN auth.users.locked_at IS '锁定时间';

-- ============================================================
-- 3. 用户角色关联表 (auth.user_roles)
-- ============================================================
CREATE TABLE auth.user_roles (
    id              BIGSERIAL       PRIMARY KEY,
    user_id         BIGINT          NOT NULL,                   -- 用户ID
    role_id         BIGINT          NOT NULL,                   -- 角色ID
    created_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by      BIGINT,                                      -- 创建人ID
    
    -- 外键约束
    CONSTRAINT fk_user_roles_user 
        FOREIGN KEY (user_id) REFERENCES auth.users(id) 
        ON DELETE CASCADE,
    CONSTRAINT fk_user_roles_role 
        FOREIGN KEY (role_id) REFERENCES auth.roles(id) 
        ON DELETE CASCADE,
    
    -- 唯一约束: 一个用户不能重复拥有同一角色
    CONSTRAINT uk_user_roles_user_role UNIQUE (user_id, role_id)
);

-- 创建索引
CREATE INDEX idx_user_roles_user_id ON auth.user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON auth.user_roles(role_id);

COMMENT ON TABLE auth.user_roles IS '用户角色关联表';
COMMENT ON COLUMN auth.user_roles.user_id IS '用户ID';
COMMENT ON COLUMN auth.user_roles.role_id IS '角色ID';
COMMENT ON COLUMN auth.user_roles.created_by IS '创建人ID';

-- ============================================================
-- 4. 登录日志表 (auth.login_logs)
-- ============================================================
CREATE TABLE auth.login_logs (
    id              BIGSERIAL       PRIMARY KEY,
    user_id         BIGINT,                                      -- 用户ID (登录成功时有值)
    username        VARCHAR(64)     NOT NULL,                   -- 用户名 (登录时输入)
    login_type      VARCHAR(32)     NOT NULL DEFAULT 'password',-- 登录类型: password, refresh_token, api_key
    status          SMALLINT        NOT NULL,                   -- 状态: 1-成功, 0-失败
    failure_reason  VARCHAR(128),                                -- 失败原因
    ip_address      VARCHAR(64),                                 -- 登录 IP
    user_agent      VARCHAR(256),                                -- User-Agent
    device_info     VARCHAR(256),                                -- 设备信息
    created_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP
    
    CONSTRAINT chk_login_logs_status CHECK (status IN (0, 1))
);

-- 创建索引 (按月分区可优化，这里先建普通索引)
CREATE INDEX idx_login_logs_user_id ON auth.login_logs(user_id);
CREATE INDEX idx_login_logs_username ON auth.login_logs(username);
CREATE INDEX idx_login_logs_created_at ON auth.login_logs(created_at DESC);
CREATE INDEX idx_login_logs_status ON auth.login_logs(status);

COMMENT ON TABLE auth.login_logs IS '登录日志表';
COMMENT ON COLUMN auth.login_logs.login_type IS '登录类型: password, refresh_token, api_key';
COMMENT ON COLUMN auth.login_logs.status IS '状态: 1-成功, 0-失败';
COMMENT ON COLUMN auth.login_logs.failure_reason IS '失败原因';

-- ============================================================
-- 5. API Keys 表 (auth.api_keys) - 开放平台使用
-- ============================================================
CREATE TABLE auth.api_keys (
    id                  BIGSERIAL       PRIMARY KEY,
    key_id              VARCHAR(64)     NOT NULL UNIQUE,        -- API Key ID (公开)
    key_secret          VARCHAR(128)    NOT NULL,               -- API Secret (加密存储)
    user_id             BIGINT          NOT NULL,               -- 关联用户ID
    name                VARCHAR(128)    NOT NULL,               -- Key 名称
    description         VARCHAR(256),                            -- 描述
    permissions         JSONB,                                   -- 权限列表
    ip_whitelist        JSONB,                                   -- IP 白名单
    status              SMALLINT        NOT NULL DEFAULT 1,     -- 状态: 1-启用, 0-禁用
    last_used_at        TIMESTAMP,                               -- 最后使用时间
    expires_at          TIMESTAMP,                               -- 过期时间
    created_at          TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at          TIMESTAMP,                               -- 软删除标记
    
    -- 外键约束
    CONSTRAINT fk_api_keys_user 
        FOREIGN KEY (user_id) REFERENCES auth.users(id) 
        ON DELETE CASCADE,
    
    CONSTRAINT chk_api_keys_status CHECK (status IN (0, 1))
);

-- 创建索引
CREATE INDEX idx_api_keys_user_id ON auth.api_keys(user_id);
CREATE INDEX idx_api_keys_status ON auth.api_keys(status);
CREATE INDEX idx_api_keys_deleted ON auth.api_keys(deleted_at);

COMMENT ON TABLE auth.api_keys IS 'API Keys 表';
COMMENT ON COLUMN auth.api_keys.key_id IS 'API Key ID (公开)';
COMMENT ON COLUMN auth.api_keys.key_secret IS 'API Secret (加密存储)';
COMMENT ON COLUMN auth.api_keys.permissions IS '权限列表 (JSON)';
COMMENT ON COLUMN auth.api_keys.ip_whitelist IS 'IP 白名单 (JSON)';

-- ============================================================
-- 6. Token 黑名单表 (auth.token_blacklist) - 用于登出时吊销 Token
-- ============================================================
CREATE TABLE auth.token_blacklist (
    id              BIGSERIAL       PRIMARY KEY,
    jti             VARCHAR(64)     NOT NULL UNIQUE,            -- JWT ID
    user_id         BIGINT          NOT NULL,                   -- 用户ID
    token_type      VARCHAR(32)     NOT NULL,                   -- Token 类型: access, refresh
    expires_at      TIMESTAMP       NOT NULL,                   -- 过期时间
    created_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by      VARCHAR(32),                                 -- 吊销原因
    
    CONSTRAINT fk_token_blacklist_user 
        FOREIGN KEY (user_id) REFERENCES auth.users(id) 
        ON DELETE CASCADE
);

-- 创建索引
CREATE INDEX idx_token_blacklist_jti ON auth.token_blacklist(jti);
CREATE INDEX idx_token_blacklist_expires ON auth.token_blacklist(expires_at);

COMMENT ON TABLE auth.token_blacklist IS 'Token 黑名单表';
COMMENT ON COLUMN auth.token_blacklist.jti IS 'JWT ID';
COMMENT ON COLUMN auth.token_blacklist.token_type IS 'Token 类型: access, refresh';

-- ============================================================
-- 7. 自动更新 updated_at 的触发器
-- ============================================================
CREATE OR REPLACE FUNCTION auth.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 为需要的表添加触发器
CREATE TRIGGER update_roles_updated_at 
    BEFORE UPDATE ON auth.roles
    FOR EACH ROW EXECUTE FUNCTION auth.update_updatedat_column();

CREATE TRIGGER update_users_updated_at 
    BEFORE UPDATE ON auth.users
    FOR EACH ROW EXECUTE FUNCTION auth.update_updatedat_column();

CREATE TRIGGER update_api_keys_updated_at 
    BEFORE UPDATE ON auth.api_keys
    FOR EACH ROW EXECUTE FUNCTION auth.update_updatedat_column();

-- ============================================================
-- 完成提示
-- ============================================================
DO $$
BEGIN
    RAISE NOTICE 'Auth schema tables created successfully';
END $$;
