# 轻量级RTSP服务模块

这是一个为Yunyez项目实现的轻量级RTSP服务模块，用于在设备端接收本地摄像头推流，提供健康检查与状态查询接口，并为未来扩展为WebRTC网关的上游源做好准备。

## 功能特性

- **RTSP协议支持**：实现了基本的RTSP方法（OPTIONS, DESCRIBE, SETUP, PLAY, PAUSE, TEARDOWN）
- **流管理**：支持多个视频流的管理
- **连接管理**：管理客户端连接和会话
- **健康检查**：提供HTTP健康检查接口
- **状态查询**：提供详细的状态信息查询接口
- **SDP生成**：自动生成SDP描述信息
- **轻量级设计**：专为资源受限设备设计

## 视频帧数据传输流程

视频帧数据从C端(设备端)到服务端的传输流程如下：

### 1. 设备端数据采集与编码

设备端首先通过摄像头采集原始视频帧数据，然后使用H.264或H.265编码器对视频帧进行压缩编码，生成编码后的视频帧数据。

### 2. RTP封装

编码后的视频帧数据被封装进RTP（Real-time Transport Protocol）数据包中：

- **RTP Header**: 包含版本号、负载类型、序列号、时间戳等信息
- **Payload**: 编码后的视频帧数据（如H.264 NALU单元）
- **RTP Packet**: 最终形成完整的RTP数据包

### 3. 传输层封装

RTP数据包进一步封装到传输层协议中，通常使用UDP或TCP：

- **UDP封装**: RTP数据包直接封装到UDP数据报中，再封装到IP数据包
- **TCP封装**: 在RTSP中可通过交互式传输，RTP数据包通过已建立的TCP连接传输

### 4. RTSP信令协商

在实际传输视频数据之前，设备端和服务端通过RTSP信令协议进行协商：

- **OPTIONS**: 查询服务端支持的方法
- **DESCRIBE**: 获取流的SDP描述信息
- **SETUP**: 设置传输参数，协商传输方式（UDP/TCP）和端口
- **PLAY**: 开始播放/传输视频流

### 5. 数据传输

设备端按照协商好的传输方式，将RTP数据包发送到服务端：

- **UDP模式**: 直接向服务端指定端口发送RTP数据包
- **TCP模式**: 通过RTSP控制连接发送RTP数据

### 6. 服务端接收与处理

服务端接收RTP数据包并进行以下处理：

- 解析RTP头部信息
- 提取视频帧数据
- 按时间戳排序（处理网络抖动）
- 存储或转发视频数据

### 示例数据包格式

#### RTSP SETUP请求示例
```
SETUP rtsp://192.168.1.100:8554/cam1/track1 RTSP/1.0
CSeq: 3
Transport: RTP/AVP;unicast;client_port=8000-8001
Session: 12345678
```

#### RTSP SETUP响应示例
```
RTSP/1.0 200 OK
CSeq: 3
Transport: RTP/AVP;unicast;client_port=8000-8001;server_port=8002-8003;ssrc=12345678
Session: 12345678;timeout=60
```

#### RTP数据包头部示例
```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|V=2|P|X|  CC   |M|     PT      |       sequence number         |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                           timestamp                         |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|           synchronization source (SSRC) identifier          |
+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+
|            contributing source (CSRC) identifiers           |
|                             ....                            |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                   Payload (Encoded Video Frame)             |
|                             ....                            |
```

其中：
- V=2: RTP版本号
- P: 填充位
- X: 扩展位
- CC: CSRC计数
- M: 标记位
- PT: 有效载荷类型（如H.264为96-127动态分配）
- Sequence Number: 序列号，用于排序
- Timestamp: 时间戳，用于同步
- SSRC: 同步源标识符
- Payload: 编码后的视频帧数据

## 架构设计

```
┌─────────────────────────────────────────────────────────────┐
│                    RTSP Server                              │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌──────────────┐ │
│  │   RTSP Handler  │  │   Media Store   │  │ Health Check │ │
│  │                 │  │                 │  │              │ │
│  │ - OPTIONS       │  │ - Stream Info   │  │ - Status     │ │
│  │ - DESCRIBE      │  │ - Client List   │  │ - Metrics    │ │
│  │ - SETUP         │  │ - Stats         │  │              │ │
│  │ - PLAY          │  │                 │  │              │ │
│  │ - PAUSE         │  │                 │  │              │ │
│  │ - TEARDOWN      │  │                 │  │              │ │
│  └─────────────────┘  └─────────────────┘  └──────────────┘ │
└─────────────────────────────────────────────────────────────┘
                                │
                   ┌─────────────────────────┐
                   │   Internal Packages     │
                   │                         │
                   │ - RTSP Protocol Parser  │
                   │ - Connection Manager    │
                   │ - SDP Generator         │
                   │ - Transport Handler     │
                   └─────────────────────────┘
```

## 模块结构

```
internal/
├── pkg/
│   ├── rtsp/
│   │   ├── constants.go    # RTSP常量定义
│   │   ├── message.go      # RTSP消息结构
│   │   ├── parser.go       # RTSP协议解析器
│   │   └── utils.go        # RTSP工具函数
│   ├── transport/
│   │   ├── interface.go    # 传输接口定义
│   │   ├── tcp.go          # TCP传输处理
│   │   └── udp.go          # UDP传输处理
│   └── media/
│       └── sdp/
│           ├── types.go    # SDP类型定义
│           └── sdp.go      # SDP解析和生成
└── video/
    ├── types/
    │   ├── stream.go       # 流相关类型定义
    │   ├── client.go       # 客户端相关类型定义
    │   └── response.go     # 响应类型定义
    ├── interfaces/
    │   ├── server.go       # 服务器接口定义
    │   └── manager.go      # 管理器接口定义
    ├── rtsp_server.go      # RTSP服务器主结构
    ├── rtsp_handler.go     # RTSP请求处理逻辑
    ├── stream_manager.go   # 流管理器
    ├── connection_manager.go # 连接管理器
    └── sdp_generator.go    # SDP生成器
```

## 使用方法

### 启动RTSP服务器

```go
package main

import (
    "context"
    "log"
    "time"

    "yunyez/internal/video"
    "yunyez/internal/video/types"
)

func main() {
    // 创建流管理器
    streamManager := video.NewStreamManager()

    // 创建连接管理器
    connectionManager := video.NewConnectionManager(streamManager)

    // 创建RTSP服务器
    rtspServer := video.NewRTSPServer(":8554", streamManager, connectionManager)

    // 添加示例流
    sampleStream := &types.Stream{
        ID:          "cam1",
        Name:        "Camera 1",
        State:       types.StreamActive,
        MediaType:   types.VideoMediaType,
        MediaFormat: types.H264MediaFormat,
        Resolution:  "1920x1080",
        Bitrate:     2048,
        Framerate:   30.0,
        CreatedAt:   time.Now(),
        LastActivity: time.Now(),
        Source:      "rtsp://localhost:8554/cam1",
    }

    if err := streamManager.AddStream(sampleStream); err != nil {
        log.Fatalf("Failed to add sample stream: %v", err)
    }

    // 启动RTSP服务器
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    if err := rtspServer.Start(ctx); err != nil {
        log.Fatalf("Failed to start RTSP server: %v", err)
    }

    // 服务器运行中...
    
    // 关闭服务器
    if err := rtspServer.Stop(ctx); err != nil {
        log.Printf("Error stopping RTSP server: %v", err)
    }
}
```

### 健康检查接口

- `/health` - 基本健康状态
- `/status` - 详细服务器状态
- `/streams` - 当前活动流列表
- `/clients` - 当前连接客户端列表

## 扩展性设计

该模块设计时考虑了未来的扩展性：

- **WebRTC网关扩展**：通过接口抽象，便于将来扩展为WebRTC网关的上游源
- **模块化设计**：各组件松耦合，易于替换和扩展
- **标准化实现**：使用Go标准库实现，避免外部依赖

## 依赖关系

- Go 1.24.8+
- 标准库（net, http, sync, context等）
- 无外部RTSP库依赖

## 测试

运行测试：

```bash
go test ./internal/video/...
```

## 许可证

MIT License