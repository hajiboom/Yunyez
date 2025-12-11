# QWEN.md - Yunyez 项目文档

## 项目概述

Yunyez（云也子）是一个电子宠物解决方案，AI陪伴机器人，主要提供情绪价值。该项目主要使用场景包括：

- 家庭宠物陪伴
- 远程宠物监控
- 旅行搭子陪伴 + 旅行记录打卡

参考产品：小智AI

旨在打通"智能设备-后端服务-数据可视化"全链路。该项目提供一个AI陪伴机器人，以提供情绪价值为核心，支持旅行搭子陪伴和旅行记录打卡等使用场景。

### 核心特性

- **设备端**: 通过ESP32开发板或模拟器，支持MQTT协议上报定位、照片等打卡数据，支持自动/语音打卡
- **后端服务**: 基于Go开发，使用gRPC优化服务间通信，HTTP提供前端接口，支持配置化规则扩展
- **数据存储**: PostgreSQL存储核心结构化数据，MongoDB存储非结构化附件，Redis做缓存
- **前端看板**: 可视化展示电子宠物状态、旅行陪伴打卡记录、设备配置，支持实时推送

### 架构设计

项目采用微服务架构设计，主要包含以下模块：

1. **设备打卡模块**:
   - 设备通过MQTT上报打卡数据（定位、照片、设备ID），支持Proto/JSON双格式
   - 配置化打卡规则（景点范围、设备权限），无需改代码新增景点/设备
   - 打卡有效性校验（定位与景点距离≤100米）
   - 批量打卡支持（gRPC流式通信，适配设备批量上报）

2. **后端服务模块**:
   - 服务拆分：设备服务（MQTT消息处理、设备管理）、打卡服务（打卡逻辑、统计）
   - 双通信模式：gRPC（服务间高效通信）+ HTTP（前端/第三方调用）
   - 实时推送：WebSocket推送打卡通知、设备在线状态
   - 数据分层：Repository层封装数据库操作，Service层处理业务逻辑

3. **数据存储模块**:
   - 结构化数据（用户、景点、打卡核心信息）→ PostgreSQL
   - 非结构化数据（打卡照片、设备日志）→ MongoDB
   - 热点数据（设备在线状态、打卡缓存）→ Redis

## 环境要求

- OS: Ubuntu 22.04
- Go: 1.24.8
- Node: 20.19.0
- Redis
- Docker容器（包含EMQX、PostgreSQL）

## 技术栈

### 核心技术
| 分类         | 技术选型                                                                 |
|--------------|--------------------------------------------------------------------------|
| 后端语言     | Go 1.24.8                                                                 |
| 通信协议     | gRPC（服务间）、HTTP/JSON（前后端）、MQTT 3.1.1（设备-后端）、WebSocket（实时推送） |
| 数据序列化   | Protocol Buffers（Proto3）、JSON                                          |
| 数据库       | PostgreSQL 16（结构化核心数据）、MongoDB 6（非结构化数据）、Redis 7（缓存） |
| 中间件       | EMQX（MQTT Broker）、Docker（容器化）、K8s（集群部署）、Minikube（本地测试） |
| Web框架      | Gin（HTTP接口）                                                          |
| ORM/客户端   | GORM（PostgreSQL）、mongo-driver（MongoDB）、go-redis（Redis）、paho.mqtt.golang（MQTT） |

### 工程工具
- 版本控制：Git
- 文档：DrawIO + feishu
- 调试工具：Postman（HTTP）、MQTTX（MQTT）、Goland Debug
- 部署工具：Docker Compose（本地一键启动）、K8s（集群部署）

## 项目结构

```
.
├── api
│   └── proto # protobuf 定义
├── cmd # 服务启动入口
│   ├── device # 设备服务
│   │   └── device.go
│   └── web # web 服务
├── configs # 项目配置文件
│   ├── config.yaml # 全局配置
│   ├── dev # 开发环境配置
│   │   ├── database.yaml
│   │   └── mqtt.yaml
│   ├── device.yaml # 设备配置（公共）
│   ├── pre # 预发布环境配置
│   └── test # 测试环境配置
├── docs # 项目文档
│   ├── checkList # 代码审查要点
│   │   └── check.md
│   ├── error_records.md # 错误码对照表
│   ├── protocol.md # 自定义协议文档
│   └── sql # 数据库 SQL 脚本
│       ├── default.sql # 默认数据库脚本
│       └── device # 设备数据库脚本
│           └── device.sql
├── example # 示例代码
│   └── mock # mock代码
│       ├── virtual_capture # mock摄像
│       └── virtual_voice # mock语音
│           ├── device_voice
│           ├── device_voice.c # 设备语音处理函数（设备端）
│           └── voice_proto.h # 设备语音处理头文件
├── fronted # 前端代码
│   ├── admin # 管理平台前端目录
│   └── v1 # 其余前端目录 v1
│       └── welcome.html
├── internal # 项目内部代码
│   ├── app # 各服务主函数
│   │   └── device
│   │       ├── app.go
│   │       └── http.go
│   ├── common # 项目通用代码
│   │   ├── config # 配置相关代码
│   │   │   └── config.go
│   │   ├── constant # 常量相关代码
│   │   │   ├── default.go
│   │   │   └── error.go
│   │   ├── frequency # 频率相关代码
│   │   └── tools # 工具相关代码
│   │       ├── context.go
│   │       ├── file.go
│   │       └── file_test.go
│   ├── controller # 控制器相关代码
│   │   ├── deviceManage # 设备管理控制器
│   │   │   ├── delete.go
│   │   │   ├── fetch.go
│   │   │   └── update.go
│   │   └── voiceManage # 语音管理控制器
│   ├── middleware # 中间件相关代码
│   ├── model # 数据库模型相关代码
│   │   ├── device # 设备相关数据库模型
│   │   │   └── device.go
│   │   └── image # 图像相关数据库模型
│   │       └── imageMessage.go
│   ├── pkg # 项目中间件通用代码
│   │   ├── http # http 相关代码
│   │   │   ├── common
│   │   │   └── middleware
│   │   │       ├── authMiddleware.go
│   │   │       └── frequencyMiddleware.go
│   │   ├── logger # 日志相关代码
│   │   │   ├── default.go
│   │   │   └── gorm.go
│   │   ├── mqtt # mqtt 相关代码
│   │   │   ├── connect.go
│   │   │   ├── constant
│   │   │   │   └── constant.go
│   │   │   ├── core
│   │   │   │   ├── client.go
│   │   │   │   ├── mqtt.go
│   │   │   │   └── topic.go
│   │   │   ├── handler
│   │   │   │   └── forward.go
│   │   │   ├── middleware
│   │   │   │   ├── device_handler.go
│   │   │   │   └── middleware.go
│   │   │   └── protocol
│   │   │       └── voice
│   │   │           ├── message.go
│   │   │           ├── voice.go
│   │   │           └── voice_test.go
│   │   ├── postgre # postgresql 相关代码
│   │   │   └── db.go
│   │   ├── redis # redis 相关代码
│   │   └── websocket # websocket 相关代码
│   ├── service # 服务相关代码
│   │   ├── device # 设备相关服务
│   │   │   ├── device.go # 设备服务
│   │   │   └── device_test.go
│   │   └── voice # 语音相关服务
│   └── types # 数据类型相关代码
│       ├── common # 通用数据类型
│       │   └── default.go
│       └── device # 设备相关数据类型
│           ├── device.go
│           └── dto.go
├── pkg # 外部依赖代码
├── storage # 项目存储目录
│   ├── logs # 日志存储目录
│   └── tmp # 临时文件存储目录
└── test # 集成测试目录
```

## 配置文件

项目使用Viper进行配置管理，支持多环境配置：

- `configs/config.yaml`: 主配置文件，包含应用名、环境、调试模式等
- `configs/dev/`: 开发环境配置
  - `database.yaml`: 数据库配置
  - `mqtt.yaml`: MQTT配置
- `configs/pre/`: 预发布环境配置
- `configs/test/`: 测试环境配置

## 构建和运行

### 环境要求
- Go 1.24.8
- Node 20.19.0
- Redis
- Docker

### 本地运行步骤

1. 安装依赖：
```bash
go mod tidy
```

2. 启动必要的服务：
```bash
# 启动PostgreSQL, MongoDB, Redis, EMQX via Docker
docker-compose up -d
```

3. 启动设备服务：
```bash
go run cmd/device/device.go
```

4. 项目使用了以下主要配置：
- HTTP服务端口：8080
- MQTT地址：tcp://127.0.0.1:1883
- 数据库配置在configs/dev/database.yaml中

## 开发约定

1. **项目采用微服务架构**，每个服务都有自己的数据库
2. **项目采用RESTful API设计**，所有接口都返回JSON格式数据
3. **项目采用MQTT/UDP等协议**进行设备与服务端通信
4. **前端代码**有其他小伙伴在写 ./fronted/admin 目录下，测试时的前端代码只需要关注 ./fronted/v1 目录下的代码即可
5. **日志记录**使用Zap框架，按服务分类输出，便于问题排查
6. **AI助手行为**：在用户给出一个开发需求的时候，要列举出技术规格说明(spec)给用户评审

## 前端代码结构

- `./fronted/admin`: 管理平台前端代码
- `./fronted/v1`: 其他前端代码（用于测试）

## 核心业务流程

1. **设备注册**：设备通过序列号(SN)在系统中注册
2. **设备通信**：设备通过MQTT协议与后端通信
3. **数据存储**：设备上报的数据存储到相应的数据库
4. **数据展示**：前端看板展示设备状态和打卡记录
5. **实时推送**：通过WebSocket实现数据的实时推送

## 重要数据模型

### 设备基础信息模型 (BaseDevice)
- 包含设备序列号(SN)、IMEI、ICCID、设备类型、厂商信息、版本信息等
- 包含设备状态、激活时间、创建时间等时间信息
- 支持软删除

### 设备网络信息模型 (DeviceNetwork)
- 包含网络类型、MAC地址、IP地址、端口、信号强度等
- 包含连接状态、最后连接/断开时间等

## 重要文档

项目包含以下重要文档：

- `docs/checkList/check.md`: 代码审查要点
- `docs/error_records.md`: 错误码对照表
- `docs/protocol.md`: 自定义协议文档
- `docs/sql/default.sql`: 默认数据库脚本
- `docs/sql/device/device.sql`: 设备数据库脚本

## Mock代码和示例

项目包含用于测试和开发的mock代码：

- `example/mock/virtual_capture`: mock摄像相关代码
- `example/mock/virtual_voice`: mock语音相关代码
  - `device_voice.c`: 设备语音处理函数（设备端）
  - `voice_proto.h`: 设备语音处理头文件

## 测试

项目包含单元测试和集成测试：
- 服务层测试：`internal/service/device/device_test.go`
- 协议测试：`internal/pkg/mqtt/protocol/voice/voice_test.go`
- 工具函数测试：`internal/common/tools/file_test.go`