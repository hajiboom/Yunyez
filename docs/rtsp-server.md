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