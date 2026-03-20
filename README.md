# 云也子(Yunyez)

> **你安心看世界，我默默守着你。**

Yunyez 是一个为**独自旅行者**设计的个人安全守护设备。它不是一个玩具，也不是另一个打卡 App，而是一个能在你无法操作手机时（如遇险、昏迷、被胁迫），**自动取证、静默报警、持续求救**的小助手；

---

## 🎯 核心价值

### 1. **被动式安全守护**
- **遇险自动响应**：检测到高冲击坠落、异常语音或 SOS 按钮触发，自动录像+报警
- **昏迷求救**：长时间静止 + 偏离路线 → 自动发送位置 + 播放求救音
- **双电池冗余**：主备电池热切换，确保 48 小时以上续航，永不突然断电

### 2. **隐私优先设计**
- 摄像头默认物理遮蔽，仅紧急时开启且 LED 红灯常亮
- 所有音视频端到端加密，云端不解密，24 小时后自动删除
- 无历史存储、无回放功能，杜绝偷拍滥用可能

### 3. **轻量语音交互**
- 本地 ASR/NLU 支持离线问答（如"省钱推荐"）
- 通过蓝牙耳机输出语音，本体无扬声器 → 节电 + 隐私
- 多模态融合判断（语音+IMU+GPS），避免菜市场等高噪环境误触发

---

## 🛠️ 技术架构（个人版）

| 层级       | 技术选型 |
|------------|---------|
| 设备端     | 树莓派 Zero 2W / ESP32-S3 + GPS + IMU + 摄像头 |
| 通信协议   | MQTT（设备↔后端）、gRPC/HTTP（Go↔Python 模型） |
| 后端       | Go（业务逻辑） + Python（ASR/NLU/LLM/TTS 本地模型） |
| 数据库     | PostgreSQL（含 PostGIS） + Redis |
| 部署       | Docker Compose 一键启动，支持单机运行 |
| AI 服务    | HTTP/gRPC 双模式，支持本地部署 |

> 注：暂时不追求微服务、K8s 等复杂架构，聚焦功能闭环与真实体验。

---

## 📁 项目结构

```
Yunyez/
├── cmd/                      # 程序入口（待完善）
├── configs/                  # 配置文件
│   ├── config.yaml           # 基础配置
│   ├── device.yaml           # 设备配置
│   └── dev/                  # 开发环境配置
│       ├── database.yaml     # 数据库配置
│       ├── mqtt.yaml         # MQTT 配置
│       ├── default.yaml      # 通用配置
│       ├── ai.yaml           # AI 模型配置
│       └── rate_limit.yaml   # 限流配置
├── docker/                   # Docker 相关
│   ├── docker-compose.yml    # 基础设施 Compose 文件
│   ├── device/               # 设备端 Docker
│   ├── video/                # 视频服务
│   └── rtsp-server/          # RTSP 服务器
├── internal/                 # 核心代码
│   ├── app/                  # 应用层
│   ├── common/               # 公共模块（配置、工具、常量）
│   ├── controller/           # HTTP 控制器
│   ├── middleware/           # 中间件（认证、日志、CORS 等）
│   ├── model/                # 数据模型
│   ├── pkg/                  # 公共包
│   │   ├── agent/            # AI Agent（LLM、ASR、TTS、NLU）
│   │   ├── logger/           # 日志封装
│   │   ├── mqtt/             # MQTT 客户端与协议
│   │   ├── postgre/          # PostgreSQL 客户端
│   │   ├── redis/            # Redis 客户端
│   │   ├── transport/        # 传输层（TCP/UDP）
│   │   ├── rtsp/             # RTSP 协议解析
│   │   └── media/            # 媒体格式处理
│   ├── service/              # 业务服务层
│   ├── types/                # 类型定义
│   └── video/                # 视频流服务（RTSP Server）
├── sql/                      # SQL 脚本
│   ├── default.sql
│   ├── agent/                # Agent 相关表
│   └── device/               # 设备相关表
├── storage/                  # 存储目录
│   ├── logs/                 # 日志文件
│   └── tmp/audio/            # 临时音频文件
└── example/                  # 示例代码
    ├── mock/                 # 模拟设备（虚拟音视频采集）
    └── scripts/              # 示例脚本
```

---

## 🔧 环境配置

### 依赖服务

项目依赖以下基础设施服务，通过 Docker Compose 一键启动：

| 服务       | 端口   | 用户名   | 密码      | 说明           |
|------------|--------|----------|-----------|----------------|
| PostgreSQL | 5432   | postgres | root      | 主数据库       |
| Redis      | 6379   | -        | -         | 缓存服务       |
| EMQX       | 1883   | root     | root123   | MQTT Broker    |
| EMQX Dashboard | 18083 | root  | root123   | MQTT Web 管理  |

### 配置文件说明

- `configs/config.yaml` - 项目基础配置（环境、日志路径）
- `configs/device.yaml` - 设备相关配置
- `configs/dev/database.yaml` - 数据库连接配置
- `configs/dev/mqtt.yaml` - MQTT 连接配置
- `configs/dev/default.yaml` - HTTP 服务配置
- `configs/dev/ai.yaml` - AI 模型配置（LLM、ASR、TTS、NLU）

---

## 🚀 服务启动

### 1. 启动基础设施

```bash
# 一键启动（推荐）
./setup-infra.sh

# 或手动执行
docker compose -f ./docker/docker-compose.yml up -d
```

启动后访问 EMQX Dashboard：`http://localhost:18083`（用户：root，密码：root123）

### 2. 初始化数据库

```bash
# 连接 PostgreSQL 并执行 SQL 脚本
psql -h localhost -U postgres -d yunyez -f sql/default.sql
psql -h localhost -U postgres -d yunyez -f sql/device/device.sql
psql -h localhost -U postgres -d yunyez -f sql/agent/cost.sql
```

### 3. 启动 AI 服务

AI 服务（ASR、NLU、TTS）支持 **HTTP** 和 **gRPC** 两种调用方式，通过环境变量切换。

#### 开发环境（手动启动）

```bash
# HTTP 模式（默认）
cd ai && ./start.sh

# gRPC 模式
export YUNYEZ_AI_TRANSPORT_MODE=grpc
cd ai && ./start.sh
```

#### 生产环境（Docker 部署）

```bash
# HTTP 模式
docker compose -f ai/docker-compose.yml --profile http up -d

# gRPC 模式
docker compose -f ai/docker-compose.yml --profile grpc up -d
```

| 服务 | HTTP 端口 | gRPC 端口 |
|------|----------|----------|
| NLU  | 8001     | 50051    |
| ASR  | 8002     | 50052    |
| TTS  | 8003     | 50053    |

### 4. 运行后端服务

```bash
# 安装 Go 依赖
go mod tidy

# 启动服务
go run .
```

服务默认监听：`http://127.0.0.1:8080`

---

## 📋 环境变量

项目通过 `configs/config.yaml` 中的 `app.env` 字段指定环境，默认使用 `dev` 环境。

开发环境配置文件位于 `configs/dev/` 目录下。
