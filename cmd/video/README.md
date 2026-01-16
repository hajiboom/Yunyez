# 视频流媒体服务

这是一个基于RTSP协议的视频流媒体服务器，用于处理视频流的传输、管理和分发。

## 功能特性

- 支持标准RTSP协议（RFC 2326）
- 实现完整的RTSP方法：OPTIONS, DESCRIBE, SETUP, PLAY, PAUSE, TEARDOWN
- 提供流媒体管理功能
- 支持多客户端连接
- 提供HTTP健康检查和状态查询接口

## 目录结构

```
cmd/video/
├── video.go          # RTSP服务器主程序
└── README.md         # 本文档
```

## 依赖项

- Go 1.24.8 或更高版本
- Linux 系统（支持TCP/UDP套接字操作）

## 安装和构建

1. 确保你已经克隆了项目并位于项目根目录：

```bash
cd /path/to/Yunyez
```

2. 构建视频服务器：

```bash
go build ./cmd/video/video.go
```

这将在当前目录生成一个名为 `video` 的可执行文件。

## 启动服务器

### 方法一：直接运行

```bash
./video
```

### 方法二：指定参数（如果需要扩展功能）

目前服务器默认在端口 8554 上监听RTSP请求，在端口 8080 上提供HTTP健康检查。

## 操作步骤

### 1. 启动服务器

运行以下命令启动RTSP服务器：

```bash
go run ./cmd/video/video.go
```

服务器启动后会显示类似以下信息：

```
Starting RTSP server on :8554...
RTSP Server Status:
- Uptime: 0s
- Total Streams: 1
- Active Streams: 0
- Current Sessions: 0
```

### 2. 使用RTSP客户端连接

服务器启动后，你可以使用任何RTSP客户端连接到服务器。默认情况下，服务器会提供一个名为 `mystream` 的示例流。

#### 使用示例C客户端（推荐测试方式）

1. 编译C客户端程序：

```bash
cd example/mock/virtual_capture
make
```

2. 运行C客户端连接到服务器：

```bash
./virtual_capture 127.0.0.1 mystream
```

这将执行完整的RTSP交互流程：
- 发送OPTIONS请求以获取服务器支持的方法
- 发送DESCRIBE请求以获取流的SDP描述
- 发送SETUP请求以建立传输会话
- 发送PLAY请求以开始播放流
- 播放5秒后发送TEARDOWN请求结束会话

#### 使用VLC或其他RTSP客户端

你也可以使用VLC媒体播放器或其他RTSP客户端连接到服务器：

```
rtsp://127.0.0.1:8554/mystream
```

### 3. 查看服务器状态

服务器同时提供HTTP接口用于健康检查和状态查询：

- 健康检查：`http://127.0.0.1:8080/health`
- 服务器状态：`http://127.0.0.1:8080/status`
- 流列表：`http://127.0.0.1:8080/streams`
- 客户端列表：`http://127.0.0.1:8080/clients`

### 4. 停止服务器

使用 `Ctrl+C` 组合键优雅地停止服务器。

## 配置

目前服务器使用默认配置，包括：

- RTSP端口：8554
- HTTP端口：8080（用于健康检查和状态查询）
- 示例流ID：mystream
- 示例流名称：Sample Stream

## 故障排除

### 常见问题

1. **端口被占用**：如果遇到端口被占用错误，请检查是否有其他服务正在使用8554或8080端口。

2. **防火墙阻止**：确保防火墙允许8554（RTSP）和8080（HTTP）端口的流量。

3. **权限问题**：如果在某些系统上遇到权限问题，请确保有足够的权限绑定到所需的端口。

### 日志输出

服务器会在控制台输出日志信息，包括：
- 服务器启动和停止信息
- 新的客户端连接
- RTSP请求处理信息
- 错误信息（如果有）

## API 接口

### RTSP 接口

- `OPTIONS` - 查询服务器支持的方法
- `DESCRIBE` - 获取流的SDP描述
- `SETUP` - 设置传输参数并建立会话
- `PLAY` - 开始播放流
- `PAUSE` - 暂停播放
- `TEARDOWN` - 终止会话

### HTTP 接口

- `GET /health` - 返回服务器健康状态
- `GET /status` - 返回服务器详细状态
- `GET /streams` - 返回所有流的信息
- `GET /clients` - 返回所有连接的客户端信息

## 扩展性

此服务器设计为模块化，可以轻松扩展：

- 通过 `StreamManager` 接口添加更多流管理功能
- 通过 `ConnectionManager` 接口增强客户端管理功能
- 添加对更多媒体格式的支持
- 实现更复杂的认证和授权机制