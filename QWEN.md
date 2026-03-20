# Yunyez (云也子) 项目上下文文档

## 项目概述

**Yunyez (云也子)** 是一个为**独自旅行者**设计的个人安全守护设备系统。它能在用户无法操作手机时（如遇险、昏迷、被胁迫），**自动取证、静默报警、持续求救**。

> **核心理念**: 你安心看世界，我默默守着你。

### 核心价值

1. **被动式安全守护** - 遇险自动响应、昏迷求救、双电池冗余
2. **隐私优先设计** - 摄像头物理遮蔽、端到端加密、24 小时后自动删除
3. **轻量语音交互** - 本地 ASR/NLU、蓝牙耳机输出、多模态融合判断

---

## 技术栈

### 后端 (Go)

| 类别 | 技术 |
|------|------|
| **语言** | Go 1.24.8 |
| **Web 框架** | Gin |
| **数据库 ORM** | GORM + PostgreSQL (pgx) |
| **缓存** | Redis (go-redis v9) |
| **MQTT** | paho.mqtt.golang |
| **配置管理** | Viper |
| **日志** | Zap |
| **认证** | JWT (golang-jwt/jwt/v4) |
| **限流** | golang.org/x/time |

### AI 服务 (Python)

| 服务 | 功能 |
|------|------|
| **ASR** | 语音识别 (Automatic Speech Recognition) |
| **NLU** | 自然语言理解 (Natural Language Understanding) |
| **TTS** | 文本转语音 (Text-to-Speech) |
| **LLM** | 大语言模型 (通义千问) |

### 基础设施

| 服务 | 端口 | 凭证 |
|------|------|------|
| PostgreSQL | 5432 | postgres/root |
| Redis | 6379 | - |
| EMQX (MQTT) | 1883 | root/root123 |
| EMQX Dashboard | 18083 | root/root123 |

---

## 项目结构

```
Yunyez/
├── cmd/                      # 程序入口
├── configs/                  # 配置文件
│   ├── config.yaml           # 基础配置 (环境、日志)
│   ├── device.yaml           # 设备配置
│   └── dev/                  # 开发环境配置
│       ├── database.yaml     # 数据库配置
│       ├── mqtt.yaml         # MQTT 配置
│       ├── default.yaml      # HTTP 服务配置
│       ├── ai.yaml           # AI 模型配置
│       └── rate_limit.yaml   # 限流配置
├── docker/                   # Docker 部署
│   ├── docker-compose.yml    # 基础设施编排
│   ├── device/               # 设备端 Docker
│   ├── video/                # 视频服务
│   └── rtsp-server/          # RTSP 服务器
├── internal/                 # 核心业务代码
│   ├── app/                  # 应用层
│   │   └── device/           # 设备应用
│   ├── common/               # 公共模块
│   │   ├── config/           # 配置管理 (Viper 热加载)
│   │   ├── constant/         # 常量定义 (错误码、默认值)
│   │   ├── frequency/        # 频率控制
│   │   └── tools/            # 工具函数
│   ├── controller/           # HTTP 控制器
│   │   ├── deviceManage/     # 设备管理 API
│   │   └── voiceManage/      # 语音管理 API
│   ├── middleware/           # Gin 中间件
│   │   ├── auth.go           # JWT 认证
│   │   ├── cors.go           # CORS
│   │   ├── logger.go         # 日志
│   │   ├── rate_limit.go     # 限流
│   │   ├── recovery.go       # 恢复
│   │   └── security.go       # 安全
│   ├── model/                # GORM 数据模型
│   │   └── device/           # 设备模型
│   ├── service/              # 业务服务层
│   │   ├── device/           # 设备服务
│   │   └── voice/            # 语音服务
│   ├── types/                # 类型定义
│   │   ├── common/           # 通用类型
│   │   ├── device/           # 设备类型
│   │   └── voice/           # 语音类型
│   ├── pkg/                  # 公共包
│   │   ├── agent/            # AI Agent (LLM/ASR/TTS/NLU)
│   │   ├── logger/           # 日志封装
│   │   ├── mqtt/             # MQTT 客户端与协议
│   │   ├── postgre/          # PostgreSQL 客户端
│   │   ├── redis/            # Redis 客户端
│   │   ├── rtsp/             # RTSP 协议解析
│   │   ├── transport/        # 传输层 (TCP/UDP)
│   │   └── media/            # 媒体格式处理
│   └── video/                # 视频流服务 (RTSP Server)
│       ├── types/            # 视频类型
│       ├── interfaces/       # 接口定义
│       ├── rtsp_server.go    # RTSP 服务器
│       ├── rtsp_handler.go   # RTSP 请求处理
│       ├── stream_manager.go # 流管理
│       └── connection_manager.go # 连接管理
├── api/                      # API 相关
│   └── proto/                # Protobuf 定义 (待完善)
├── ai/                       # AI 服务 (Python)
│   ├── asr/                  # 语音识别
│   ├── nlu/                  # 自然语言理解
│   └── tts/                  # 文本转语音
├── sql/                      # SQL 脚本
│   ├── default.sql           # Schema 初始化
│   ├── device/               # 设备相关表
│   └── agent/                # Agent 计费相关
├── fronted/                  # 前端项目
│   ├── Yunyez/               # 主前端 (Vue/Vite)
│   └── admin/                # 管理后台
├── example/                  # 示例代码
│   ├── mock/                 # 模拟设备
│   │   ├── virtual_capture/  # 虚拟视频采集
│   │   ├── virtual_voice/    # 虚拟语音 (WAV 文件回放)
│   │   └── virtual_voice_realtime/  # 实时流式语音 (麦克风采集)
│   └── scripts/              # 示例脚本
├── docs/                     # 文档
│   ├── protocol.md           # 通信协议规范
│   ├── rtsp-server.md        # RTSP 服务文档
│   ├── error_records.md      # 错误码定义
│   ├── setup.md              # 部署指南
│   └── hard/                 # 硬件/嵌入式相关文档
│       └── realtime-voice-streaming.md  # 实时流式语音架构设计
├── storage/                  # 存储目录
│   ├── logs/                 # 日志文件
│   └── tmp/audio/            # 临时音频文件
└── test/                     # 测试代码
```

---

## 核心模块说明

### 1. 配置管理 (`internal/common/config`)

- 使用 **Viper** 实现配置热加载
- 支持多环境配置 (`dev`, `pre`, `test`)
- 配置文件合并策略：基础配置 + 公共配置 + 环境特定配置
- 监听配置文件变化自动重载

### 2. MQTT 通信 (`internal/pkg/mqtt`)

**主题设计规范**:
- 设备上行：`[vendor]/[device_type]/[device_sn]/[command]/server`
- 服务端下行：`[vendor]/[device_type]/[device_sn]/[command]/client`

**音频传输头部** (96 位):
```
| Ver(4) | AudioFormat(8) | SampleRate(16) | Ch(2) | F(2) |
| FrameSeq(16) | Timestamp(16) | PayloadLen(16) | CRC16(16) |
```

### 3. 语音处理流水线 (`internal/service/voice/handler`)

```
语音输入 → ASR → NLU → LLM Chat → TTS → 音频输出
                ↓
          (非闲聊意图 → SpecialAction)
```

**处理流程**:
1. **ASR**: 语音转文字 (调用本地/云端 ASR 服务)
2. **NLU**: 意图识别 (判断是否为闲聊或特定命令)
3. **LLM**: 智能对话 (使用通义千问或本地模型)
4. **TTS**: 文字转语音 (Edge TTS 或本地 TTS)
5. **计费**: 记录 Token 使用量到数据库

### 3.1 实时流式语音传输 (`example/mock/virtual_voice_realtime`)

为开发测试环境提供的实时语音交互模块，支持本地麦克风采集和流式传输。

**核心模块**:
| 模块 | 职责 | 技术选型 |
|------|------|----------|
| **音频采集** | 麦克风 PCM 采集 | PortAudio (gordonklaus/portaudio) |
| **预处理** | 重采样、VAD、分帧 | 线性插值/能量阈值法 |
| **MQTT 传输** | 协议封装、流式发送 | 自定义 12 字节协议头 |
| **音频播放** | TTS 响应播放 | PortAudio + 环形缓冲 |
| **会话管理** | 状态机控制 | FSM (IDLE→LISTENING→PROCESSING→SPEAKING) |

**音频协议参数**:
- 采样率：16000 Hz
- 位深：16-bit PCM
- 声道：单声道
- 帧时长：20ms/帧 (640 字节)
- QoS：1 (至少一次)

**性能目标**:
- 端到端延迟：< 500ms
- 网络带宽：~64kbps
- CPU 占用：< 5%

**状态机**:
```
IDLE ──(语音活动)──> LISTENING ──(ASR 完成)──> PROCESSING ──(LLM 响应)──> SPEAKING ──(播放完成)──> IDLE
                            │                                                        │
                            └──────────────────(静音超时)────────────────────────────┘
```

详见：`docs/hard/realtime-voice-streaming.md`

### 4. RTSP 视频服务 (`internal/video`)

- 轻量级 RTSP 服务器实现
- 支持 H.264/H.265 编码
- 提供健康检查接口 (`/health`, `/status`)
- 流管理、连接管理、SDP 生成

### 5. 设备管理 (`internal/service/device`)

- 设备注册、激活、注销
- 设备状态管理 (在线/离线/激活)
- 设备网络信息管理
- 支持 GORM 事务操作

### 6. AI Agent (`internal/pkg/agent`)

- **LLM**: 支持通义千问和本地模型策略
- **ASR**: 语音识别客户端
- **NLU**: 意图识别客户端
- **TTS**: 语音合成服务
- **Metering**: Token 计费服务

---

## 数据库 Schema

```sql
-- Schema 划分
CREATE SCHEMA auth;    -- 用户与账户系统
CREATE SCHEMA iot;     -- 设备与 IoT 通信
CREATE SCHEMA media;   -- 媒体存储元数据
CREATE SCHEMA logging; -- 日志与审计
CREATE SCHEMA public;  -- 公共模式
```

---

## API 接口

### 设备管理

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/device/fetch` | 获取设备列表 |
| GET | `/device/detail/{sn}` | 获取设备详情 |

### 语音管理

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | `/voice/upload` | 上传语音数据 |

### RTSP 服务

| 路径 | 描述 |
|------|------|
| `/health` | 健康检查 |
| `/status` | 服务器状态 |
| `/streams` | 流列表 |
| `/clients` | 客户端列表 |

---

## 错误码体系

| 范围 | 类别 |
|------|------|
| 1000-1999 | 通用错误 |
| 2000-2999 | 数据库错误 |
| 3000-3999 | 设备错误 |
| 4000-4999 | 用户错误 |
| 5000-5999 | 语音错误 |

---

## 开发指南

### 启动基础设施

```bash
./setup-infra.sh
# 或
docker compose -f ./docker/docker-compose.yml up -d
```

### 初始化数据库

```bash
psql -h localhost -U postgres -d yunyez -f sql/default.sql
psql -h localhost -U postgres -d yunyez -f sql/device/device.sql
psql -h localhost -U postgres -d yunyez -f sql/agent/cost.sql
```

### 启动后端服务

```bash
go mod tidy
go run .
# 服务监听：http://127.0.0.1:8080
```

### 启动 AI 服务

```bash
cd ai
docker-compose up --build
# ASR: http://127.0.0.1:8002
# NLU: http://127.0.0.1:8001
# TTS: http://127.0.0.1:8003
```

---

## 关键设计决策

1. **单体优先**: 暂不使用微服务/K8s，聚焦功能闭环
2. **隐私优先**: 摄像头物理遮蔽、端到端加密、无历史存储
3. **双电池冗余**: 主备电池热切换，确保 48 小时续航
4. **本地 AI**: 支持离线 ASR/NLU，降低云端依赖
5. **轻量协议**: 自定义 MQTT 音频传输协议，适配低带宽场景

---

## 相关文件索引

| 文件 | 描述 |
|------|------|
| `README.md` | 项目总览 |
| `docs/protocol.md` | 通信协议规范 |
| `docs/rtsp-server.md` | RTSP 服务文档 |
| `docs/error_records.md` | 错误码定义 |
| `docs/hard/realtime-voice-streaming.md` | 实时流式语音架构设计 |
| `configs/config.yaml` | 基础配置 |
| `docker/docker-compose.yml` | 基础设施编排 |
| `internal/service/voice/handler/chat.go` | 语音处理核心逻辑 |
| `internal/pkg/mqtt/protocol/voice/voice.go` | MQTT 音频协议头 |
| `internal/pkg/mqtt/protocol/voice/message.go` | MQTT 音频消息构建 |
| `internal/pkg/mqtt/core/client.go` | MQTT 客户端封装 |
