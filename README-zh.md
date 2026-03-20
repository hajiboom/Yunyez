# Yunyez
"行过山野，记入云间。"
## 🎯 核心价值
> 3D 空间重建
- 输出可交互 3D 场景，而非图片 / 视频
> 空间音频同步记录
- 实时收录环境音、人声、现场语音
- 与 3D 空间对齐，还原现场氛围
> 旅行数字足迹
- GPS + 轨迹 + 3D 场景一体化存储
- 形成可回溯的立体旅行路线
> 多设备协同渲染（隐私安全模式）
- 同区域设备可共享已重建区域
- 避免重复计算，降低设备算力压力
- 隐私优先：不采集人脸、不包含私人空间、默认不共享
> 轻量化展示与分享
- 手机 / PC 端可进入 3D 场景漫游
- 适合个人回忆、旅行分享、场景存档

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
