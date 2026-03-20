# 对话上下文管理设计文档

## 背景

当前语音对话流程中，每次请求都是独立的，缺乏对设备会话上下文的记忆能力。用户需要重复表达相同的信息，无法进行连续的多轮对话。

### 当前问题

1. **无会话记忆**：每次语音请求都是独立的，无法记住之前的对话内容
2. **无设备上下文**：不同设备的对话混在一起，无法区分
3. **无意图延续**：无法理解跨轮次的意图（如"打开空调"→"调到 26 度"）
4. **无用户偏好**：无法记忆用户的习惯和偏好设置

---

## 需求分析

### 核心需求

| 需求 | 描述 | 优先级 |
|------|------|--------|
| 设备会话隔离 | 每个设备 (clientID) 有独立的上下文存储 | P0 |
| 短期记忆 | 记住最近 N 轮对话内容（会话窗口） | P0 |
| 全局上下文 | 存储设备元数据、用户偏好等长期信息 | P0 |
| 上下文过期 | 自动清理过期上下文，避免内存泄漏 | P1 |
| 上下文压缩 | Token 超限时的智能压缩策略 | P1 |
| RAG 检索 | 基于向量检索的长期记忆增强 | P2 |

### 上下文分类

```
┌─────────────────────────────────────────────────────────┐
│                    上下文分层架构                        │
├─────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────┐   │
│  │           全局上下文 (Global Context)            │   │
│  │  - 设备元数据 (位置、时区、语言)                  │   │
│  │  - 用户偏好 (语速、音量、常用指令)                │   │
│  │  - 技能配置 (已启用的功能模块)                    │   │
│  │  - 长期记忆 (重要事件、习惯)                      │   │
│  │  存储：PostgreSQL + Redis Hash                  │   │
│  └─────────────────────────────────────────────────┘   │
│                          ↓                              │
│  ┌─────────────────────────────────────────────────┐   │
│  │          短期记忆 (Session Memory)               │   │
│  │  - 最近 N 轮对话历史 (滑动窗口)                   │   │
│  │  - 当前意图状态机                               │   │
│  │  - 临时槽位值 (如"温度=26")                      │   │
│  │  存储：Redis String (TTL 自动过期)               │   │
│  └─────────────────────────────────────────────────┘   │
│                          ↓                              │
│  ┌─────────────────────────────────────────────────┐   │
│  │          单轮上下文 (Turn Context)               │   │
│  │  - 当前请求的 ASR 文本                           │   │
│  │  - NLU 意图识别结果                              │   │
│  │  - LLM 流式响应片段                              │   │
│  │  存储：内存 (请求结束即释放)                      │   │
│  └─────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────┘
```

---

## 是否需要 RAG？

### RAG（检索增强生成）适用场景

| 场景 | 是否需要 RAG | 原因 |
|------|-------------|------|
| 设备操作指令 | ❌ 不需要 | 短期记忆足够，意图明确 |
| 多轮对话补全 | ❌ 不需要 | 依赖会话窗口内的上下文 |
| 用户偏好记忆 | ⚠️ 可选 | 数据量小，可直接存全局上下文 |
| 历史事件查询 | ✅ 需要 | 如"上周我说过..."，需向量检索 |
| 知识库问答 | ✅ 需要 | 如天气、新闻等外部知识 |
| 个性化推荐 | ✅ 需要 | 基于历史行为的相似性检索 |

### 本项目决策

**阶段一（当前）：不使用 RAG**

理由：
1. 核心场景是设备控制 + 闲聊，短期记忆足够
2. RAG 增加系统复杂度，需要向量数据库
3. 用户数据量小（单设备日均对话<100 轮），可直接全量检索

**阶段二（未来）：按需引入 RAG**

触发条件：
- 需要查询历史对话（"我昨天说过什么"）
- 需要外部知识库（天气、新闻、百科）
- 用户数据量增长（>1000 轮/设备）

---

## 架构设计

### 整体架构

```
┌──────────────────────────────────────────────────────────────┐
│                      Go Backend                              │
│  ┌────────────────────────────────────────────────────────┐ │
│  │              ChatPipeline (chat.go)                    │ │
│  │  ASR → NLU → Context Manager → LLM → TTS              │ │
│  └────────────────────────────────────────────────────────┘ │
│                          ↓                                   │
│  ┌────────────────────────────────────────────────────────┐ │
│  │           Context Manager (新增模块)                    │ │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐    │ │
│  │  │ TurnContext │  │SessionMgr   │  │ GlobalStore │    │ │
│  │  │ (内存)      │  │ (Redis)     │  │ (PostgreSQL)│    │ │
│  │  └─────────────┘  └─────────────┘  └─────────────┘    │ │
│  └────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────┘
                          ↓
┌──────────────────────────────────────────────────────────────┐
│                      Storage Layer                           │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│  │   Memory    │  │    Redis    │  │ PostgreSQL  │          │
│  │ (临时变量)  │  │  (会话缓存)  │  │  (持久化)    │          │
│  │ TTL: 请求级  │  │ TTL: 30min  │  │ TTL: 永久   │          │
│  └─────────────┘  └─────────────┘  └─────────────┘          │
└──────────────────────────────────────────────────────────────┘
```

### 数据模型设计

#### 1. 单轮上下文 (Turn Context)

**存储位置**：内存（Go struct）
**生命周期**：单次请求

```go
// TurnContext 单轮对话上下文
type TurnContext struct {
    TraceID       string                 // 请求追踪 ID
    ClientID      string                 // 设备序列号
    Timestamp     time.Time              // 请求时间戳
    
    // 输入
    ASRText       string                 // ASR 识别文本
    ASRConfidence float32                // ASR 置信度
    
    // NLU
    Intent        string                 // 意图
    IntentConf    float32                // 意图置信度
    Entities      map[string]interface{} // 槽位值
    
    // 输出
    LLMResponse   string                 // LLM 完整响应
    TTSAudio      []byte                 // TTS 音频数据
    
    // 元数据
    Duration      time.Duration          // 处理耗时
    TokenUsage    *metering.Usage        // Token 消耗
}
```

#### 2. 会话记忆 (Session Memory)

**存储位置**：Redis String (JSON 序列化)
**生命周期**：30 分钟（可配置，无活动自动过期）

```go
// SessionMemory 会话记忆
type SessionMemory struct {
    ClientID    string          // 设备序列号
    CreatedAt   time.Time       // 创建时间
    UpdatedAt   time.Time       // 最后更新时间
    TTL         time.Duration   // 剩余存活时间
    
    // 对话历史（滑动窗口）
    Messages    []Message       // 最近 N 轮对话
    
    // 意图状态机
    CurrentIntent string        // 当前意图
    IntentState   IntentState   // 意图状态（如待确认、执行中）
    
    // 临时槽位（跨轮次补全）
    Slots       map[string]interface{} // 如 {"temperature": 26, "device": "ac"}
    
    // 元数据
    TurnCount   int             // 对话轮次
}

// Message 对话消息
type Message struct {
    Role      string    `json:"role"`      // user/assistant/system
    Content   string    `json:"content"`   // 消息内容
    Timestamp time.Time `json:"timestamp"` // 时间戳
    Intent    string    `json:"intent"`    // 意图（可选）
}

// IntentState 意图状态
type IntentState string
const (
    IntentStatePending    IntentState = "pending"    // 待确认
    IntentStateExecuting  IntentState = "executing"  // 执行中
    IntentStateCompleted  IntentState = "completed"  // 已完成
    IntentStateCancelled  IntentState = "cancelled"  // 已取消
)
```

**Redis Key 设计**：
```
yunyez:session:{clientID}  →  SessionMemory (JSON)
TTL: 30 分钟（每次访问自动续期）
```

#### 3. 全局上下文 (Global Context)

**存储位置**：PostgreSQL + Redis 缓存
**生命周期**：永久（除非用户删除）

**数据库表设计**：

```sql
-- 设备全局上下文表
CREATE TABLE IF NOT EXISTS device_contexts (
    id SERIAL PRIMARY KEY,
    client_id VARCHAR(64) UNIQUE NOT NULL,  -- 设备序列号
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    -- 设备元数据
    metadata JSONB NOT NULL DEFAULT '{}',   -- 位置、时区、语言等
    
    -- 用户偏好
    preferences JSONB NOT NULL DEFAULT '{}', -- 语速、音量、常用指令等
    
    -- 技能配置
    skills JSONB NOT NULL DEFAULT '[]',      -- 已启用的技能列表
    
    -- 长期记忆（重要事件）
    long_term_memories JSONB NOT NULL DEFAULT '[]', -- 关键事件记录
    
    -- 统计信息
    stats JSONB NOT NULL DEFAULT '{}',      -- 对话轮次、活跃时间等
    
    deleted_at TIMESTAMPTZ DEFAULT NULL
);

CREATE INDEX idx_device_contexts_client_id ON device_contexts(client_id);
CREATE INDEX idx_device_contexts_deleted_at ON device_contexts(deleted_at) WHERE deleted_at IS NULL;

-- 对话历史归档表（可选，用于审计或长期分析）
CREATE TABLE IF NOT EXISTS conversation_archives (
    id SERIAL PRIMARY KEY,
    client_id VARCHAR(64) NOT NULL,
    session_id VARCHAR(64) NOT NULL,      -- 会话 ID
    turn_index INTEGER NOT NULL,           -- 轮次索引
    
    -- 对话内容
    user_text TEXT NOT NULL,
    assistant_text TEXT NOT NULL,
    intent VARCHAR(64),
    entities JSONB DEFAULT '{}',
    
    -- 元数据
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    duration_ms INTEGER DEFAULT 0,
    token_usage JSONB DEFAULT '{}',
    
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_conversation_archives_client_id ON conversation_archives(client_id);
CREATE INDEX idx_conversation_archives_session_id ON conversation_archives(session_id);
CREATE INDEX idx_conversation_archives_timestamp ON conversation_archives(timestamp);
```

---

## 接口设计

### Context Manager 接口

**文件**：`internal/service/context/manager.go`

```go
// Package context 对话上下文管理
package context

import (
    "context"
    "time"
)

// Manager 上下文管理器接口
type Manager interface {
    // 获取/创建会话
    GetSession(ctx context.Context, clientID string) (*SessionMemory, error)
    UpdateSession(ctx context.Context, session *SessionMemory) error
    DeleteSession(ctx context.Context, clientID string) error
    
    // 添加对话消息
    AddMessage(ctx context.Context, clientID string, msg Message) error
    
    // 获取对话历史（用于 LLM 上下文）
    GetMessages(ctx context.Context, clientID string, limit int) ([]Message, error)
    
    // 意图状态管理
    SetIntentState(ctx context.Context, clientID string, intent string, state IntentState) error
    GetIntentState(ctx context.Context, clientID string) (string, IntentState, error)
    
    // 槽位管理
    SetSlot(ctx context.Context, clientID string, key string, value interface{}) error
    GetSlot(ctx context.Context, clientID string, key string) (interface{}, error)
    ClearSlots(ctx context.Context, clientID string) error
    
    // 全局上下文
    GetGlobalContext(ctx context.Context, clientID string) (*GlobalContext, error)
    UpdateGlobalContext(ctx context.Context, gc *GlobalContext) error
    
    // 长期记忆（未来 RAG 预留）
    AddMemory(ctx context.Context, clientID string, memory *LongTermMemory) error
    SearchMemories(ctx context.Context, clientID string, query string, limit int) ([]*LongTermMemory, error)
}

// GlobalContext 全局上下文
type GlobalContext struct {
    ClientID    string                 `json:"client_id"`
    Metadata    map[string]interface{} `json:"metadata"`     // 设备元数据
    Preferences map[string]interface{} `json:"preferences"`  // 用户偏好
    Skills      []string               `json:"skills"`       // 技能列表
    Memories    []LongTermMemory       `json:"memories"`     // 长期记忆
    Stats       map[string]interface{} `json:"stats"`        // 统计信息
}

// LongTermMemory 长期记忆（用于未来 RAG）
type LongTermMemory struct {
    ID        string                 `json:"id"`
    Type      string                 `json:"type"`      // event/fact/preference
    Content   string                 `json:"content"`
    Embedding []float32              `json:"embedding"` // 向量嵌入（未来）
    Metadata  map[string]interface{} `json:"metadata"`
    CreatedAt time.Time              `json:"created_at"`
}
```

### 实现示例

**文件**：`internal/service/context/redis_manager.go`

```go
// RedisManager 基于 Redis 的会话管理器
type RedisManager struct {
    redisClient *redis.Client
    sessionTTL  time.Duration // 会话过期时间（默认 30 分钟）
    maxTurns    int           // 最大对话轮次（滑动窗口大小）
}

func NewRedisManager(redisClient *redis.Client, opts ...Option) *RedisManager {
    rm := &RedisManager{
        redisClient: redisClient,
        sessionTTL:  30 * time.Minute,
        maxTurns:    10, // 默认保留最近 10 轮对话
    }
    for _, opt := range opts {
        opt(rm)
    }
    return rm
}

// GetSession 获取会话（不存在则创建）
func (m *RedisManager) GetSession(ctx context.Context, clientID string) (*SessionMemory, error) {
    key := fmt.Sprintf("yunyez:session:%s", clientID)
    
    data, err := m.redisClient.Get(ctx, key).Bytes()
    if err == redis.Nil {
        // 创建新会话
        session := &SessionMemory{
            ClientID:  clientID,
            CreatedAt: time.Now(),
            UpdatedAt: time.Now(),
            TTL:       m.sessionTTL,
            Messages:  make([]Message, 0),
            Slots:     make(map[string]interface{}),
        }
        return session, nil
    }
    if err != nil {
        return nil, fmt.Errorf("get session: %w", err)
    }
    
    var session SessionMemory
    if err := json.Unmarshal(data, &session); err != nil {
        return nil, fmt.Errorf("unmarshal session: %w", err)
    }
    
    session.TTL = m.sessionTTL // 重置 TTL
    return &session, nil
}

// UpdateSession 更新会话（带 TTL 续期）
func (m *RedisManager) UpdateSession(ctx context.Context, session *SessionMemory) error {
    key := fmt.Sprintf("yunyez:session:%s", session.ClientID)
    session.UpdatedAt = time.Now()
    
    // 滑动窗口：只保留最近 N 轮
    if len(session.Messages) > m.maxTurns {
        session.Messages = session.Messages[len(session.Messages)-m.maxTurns:]
    }
    
    data, err := json.Marshal(session)
    if err != nil {
        return fmt.Errorf("marshal session: %w", err)
    }
    
    return m.redisClient.Set(ctx, key, data, m.sessionTTL).Err()
}

// AddMessage 添加对话消息
func (m *RedisManager) AddMessage(ctx context.Context, clientID string, msg Message) error {
    session, err := m.GetSession(ctx, clientID)
    if err != nil {
        return err
    }
    
    session.Messages = append(session.Messages, msg)
    session.TurnCount++
    
    return m.UpdateSession(ctx, session)
}

// GetMessages 获取对话历史（用于 LLM 上下文）
func (m *RedisManager) GetMessages(ctx context.Context, clientID string, limit int) ([]Message, error) {
    session, err := m.GetSession(ctx, clientID)
    if err != nil {
        return nil, err
    }
    
    if limit <= 0 || limit > len(session.Messages) {
        return session.Messages, nil
    }
    
    return session.Messages[len(session.Messages)-limit:], nil
}
```

---

## 集成到 ChatPipeline

### 改造后的流程

```go
// ChatPipeline 改造后
func ChatPipeline(ctx context.Context, clientID string, message []byte) error {
    // 1. ASR
    text, err := asrClient.Transfer(ctx, message)
    if err != nil {
        return err
    }
    
    // 2. 获取会话上下文
    session, err := contextManager.GetSession(ctx, clientID)
    if err != nil {
        logger.Error(ctx, "get session failed", map[string]any{"error": err})
        return err
    }
    
    // 3. NLU（可结合历史上下文增强意图识别）
    intent, err := nluClient.Predict(&nlu.Input{Text: text})
    if err != nil {
        return err
    }
    
    // 4. 非闲聊意图：更新槽位
    if intent.Intent != constant.IntentChitChat {
        // 更新槽位
        session.Slots[intent.Intent] = extractEntities(intent)
        session.CurrentIntent = intent.Intent
        
        err := SpecialAction(ctx, clientID, intent, session)
        if err != nil {
            return err
        }
        
        // 保存会话
        contextManager.UpdateSession(ctx, session)
        return nil
    }
    
    // 5. 闲聊：构建带上下文的 LLM 请求
    messages, err := contextManager.GetMessages(ctx, clientID, 5) // 最近 5 轮
    if err != nil {
        return err
    }
    
    // 添加用户消息
    messages = append(messages, Message{
        Role:      "user",
        Content:   text,
        Timestamp: time.Now(),
    })
    
    // 6. LLM（传入上下文）
    replyChan, usageChan, err := agentStrategy.Model.ChatWithContext(ctx, clientID, messages)
    if err != nil {
        return err
    }
    
    // 7. TTS + 发布
    var fullResponse strings.Builder
    for sentence := range textBuffer.Output() {
        audio, err := ttsClient.Synthesize(ctx, sentence)
        if err != nil {
            continue
        }
        Publish(ctx, clientID, audio)
        fullResponse.WriteString(sentence)
    }
    
    // 8. 保存对话历史
    contextManager.AddMessage(ctx, clientID, Message{
        Role:      "user",
        Content:   text,
        Timestamp: time.Now(),
    })
    contextManager.AddMessage(ctx, clientID, Message{
        Role:      "assistant",
        Content:   fullResponse.String(),
        Timestamp: time.Now(),
    })
    
    return nil
}
```

---

## 配置设计

**文件**：`configs/dev/context.yaml`

```yaml
# 上下文管理配置
context:
  # 会话配置
  session:
    ttl: 30m              # 会话过期时间
    max_turns: 10         # 最大对话轮次（滑动窗口）
    key_prefix: "yunyez:session:"
  
  # 全局上下文配置
  global:
    cache_ttl: 5m         # Redis 缓存过期时间
    key_prefix: "yunyez:global:"
  
  # LLM 上下文配置
  llm:
    max_context_turns: 5  # 传给 LLM 的历史轮次
    max_tokens: 2000      # 上下文 Token 上限
    compression_enabled: true  # 启用上下文压缩
  
  # 长期记忆配置（未来 RAG）
  memory:
    enabled: false        # 暂时关闭
    vector_store: "pgvector"  # 向量存储类型
    embedding_model: "bge-small-zh"
    search_top_k: 3
```

---

## 数据库迁移

**文件**：`sql/context/context.sql`

```sql
-- 上下文管理 Schema
CREATE SCHEMA IF NOT EXISTS context;

-- 设备全局上下文表
CREATE TABLE IF NOT EXISTS context.device_contexts (
    id SERIAL PRIMARY KEY,
    client_id VARCHAR(64) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    metadata JSONB NOT NULL DEFAULT '{}',
    preferences JSONB NOT NULL DEFAULT '{}',
    skills JSONB NOT NULL DEFAULT '[]',
    long_term_memories JSONB NOT NULL DEFAULT '[]',
    stats JSONB NOT NULL DEFAULT '{}',
    
    deleted_at TIMESTAMPTZ DEFAULT NULL
);

CREATE INDEX idx_device_contexts_client_id ON context.device_contexts(client_id);
CREATE INDEX idx_device_contexts_deleted_at ON context.device_contexts(deleted_at) WHERE deleted_at IS NULL;

-- 对话历史归档表
CREATE TABLE IF NOT EXISTS context.conversation_archives (
    id SERIAL PRIMARY KEY,
    client_id VARCHAR(64) NOT NULL,
    session_id VARCHAR(64) NOT NULL,
    turn_index INTEGER NOT NULL,
    
    user_text TEXT NOT NULL,
    assistant_text TEXT NOT NULL,
    intent VARCHAR(64),
    entities JSONB DEFAULT '{}',
    
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    duration_ms INTEGER DEFAULT 0,
    token_usage JSONB DEFAULT '{}',
    
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_conversation_archives_client_id ON context.conversation_archives(client_id);
CREATE INDEX idx_conversation_archives_session_id ON context.conversation_archives(session_id);
CREATE INDEX idx_conversation_archives_timestamp ON context.conversation_archives(timestamp);

-- 初始化触发器：自动更新 updated_at
CREATE OR REPLACE FUNCTION context.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_device_contexts_updated_at
    BEFORE UPDATE ON context.device_contexts
    FOR EACH ROW
    EXECUTE FUNCTION context.update_updated_at_column();
```

---

## 实施计划

### 阶段一：基础上下文管理（Week 1-2）

| 任务 | 描述 | 优先级 |
|------|------|--------|
| ✅ 编写 Protobuf 接口 | 定义上下文管理接口 | P0 |
| ⬜ 实现 Redis Manager | 会话记忆管理 | P0 |
| ⬜ 实现 PostgreSQL Store | 全局上下文持久化 | P0 |
| ⬜ 集成到 ChatPipeline | 改造现有对话流程 | P0 |
| ⬜ 编写单元测试 | 覆盖核心功能 | P0 |

### 阶段二：上下文优化（Week 3-4）

| 任务 | 描述 | 优先级 |
|------|------|--------|
| ⬜ 上下文压缩 | Token 超限时的智能截断 | P1 |
| ⬜ 意图状态机 | 多轮意图补全（如"打开空调"→"26 度"） | P1 |
| ⬜ 用户偏好学习 | 基于历史行为自动调整参数 | P2 |
| ⬜ 管理后台 API | 查看/删除用户上下文 | P2 |

### 阶段三：RAG 增强（未来）

| 任务 | 描述 | 触发条件 |
|------|------|----------|
| ⬜ 向量数据库 | 引入 pgvector 或 Milvus | 需要历史查询 |
| ⬜ 嵌入模型 | 部署 bge-small-zh 等 | 需要语义检索 |
| ⬜ 检索增强 | LLM 调用前检索相关记忆 | 需要个性化回答 |

---

## 性能与成本

### Redis 内存占用估算

| 字段 | 大小估算 | 说明 |
|------|---------|------|
| Messages (10 轮) | ~5KB | 每轮约 500 字节 |
| Slots | ~1KB | 槽位键值对 |
| 元数据 | ~0.5KB | 时间戳、计数等 |
| **单设备总计** | **~6.5KB** | - |
| 1 万设备 | **~65MB** | 完全可接受 |

### PostgreSQL 存储估算

| 表 | 单条大小 | 日均增长 | 年存储量 (1 万设备) |
|----|---------|---------|-------------------|
| device_contexts | ~2KB | 0 | ~20MB (静态) |
| conversation_archives | ~1KB | 10 条/设备 | ~36GB/年 |

### 成本优化策略

1. **会话 TTL**：30 分钟无活动自动过期，减少 Redis 占用
2. **滑动窗口**：只保留最近 10 轮，避免无限增长
3. **归档策略**：对话历史定期归档到冷存储
4. **压缩算法**：长文本使用摘要压缩（未来）

---

## 相关文件索引

| 文件 | 描述 | 状态 |
|------|------|------|
| `internal/service/context/manager.go` | 上下文管理器接口 | 待创建 |
| `internal/service/context/redis_manager.go` | Redis 实现 | 待创建 |
| `internal/service/context/postgres_store.go` | PostgreSQL 实现 | 待创建 |
| `internal/types/context.go` | 上下文类型定义 | 待创建 |
| `configs/dev/context.yaml` | 上下文配置 | 待创建 |
| `sql/context/context.sql` | 数据库迁移 | 待创建 |
| `docs/rpc-migration.md` | RPC 改造文档 | ✅ 已完成 |

---

## 总结

**核心决策**：
1. **暂不引入 RAG**：当前场景短期记忆足够，RAG 增加不必要的复杂度
2. **三层上下文架构**：单轮（内存）+ 会话（Redis）+ 全局（PostgreSQL）
3. **滑动窗口设计**：自动保留最近 N 轮对话，平衡效果与成本
4. **TTL 自动过期**：会话 30 分钟无活动自动清理，避免内存泄漏

**下一步**：
1. 实现 Context Manager 基础接口
2. 集成到 ChatPipeline
3. 测试验证多轮对话效果
