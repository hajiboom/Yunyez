# Yunyez 环境搭建

## 服务器部署

### 1. 克隆项目

```bash
cd /root/app
# 替换正确的仓库地址
git clone https://github.com/hajiboom/Yunyez.git
cd Yunyez
chmod +x setup-infra.sh
./setup-infra.sh
```

### 2. 初始化数据库

```bash
psql -h localhost -U postgres -d yunyez -f sql/default.sql
psql -h localhost -U postgres -d yunyez -f sql/device/device.sql
psql -h localhost -U postgres -d yunyez -f sql/agent/cost.sql
```

### 3. 部署 AI 服务

AI 服务支持 **HTTP** 和 **gRPC** 两种模式，生产环境使用 Docker 部署。

#### HTTP 模式

```bash
docker compose -f ai/docker-compose.yml --profile http up -d
```

#### gRPC 模式

```bash
docker compose -f ai/docker-compose.yml --profile grpc up -d
```

| 服务 | HTTP 端口 | gRPC 端口 |
|------|----------|----------|
| NLU  | 8001     | 50051    |
| ASR  | 8002     | 50052    |
| TTS  | 8003     | 50053    |

> **注意**: ASR 服务需要 GPU 支持，确保服务器已安装 NVIDIA 驱动并配置 Docker GPU 支持。

### 4. 部署后端服务

```bash
# 编译 Go 服务
go build -o yunyez ./cmd/server

# 运行（后台）
nohup ./yunyez > yunyez.log 2>&1 &
```

### 5. 验证服务

```bash
# 检查基础设施
docker ps | grep yunyez

# 检查 AI 服务
docker compose -f ai/docker-compose.yml ps

# 检查后端服务
curl http://localhost:8080/health
```

---

## 开发环境

开发环境下可以手动启动 AI 服务：

```bash
# HTTP 模式（默认）
cd ai && ./start.sh

# gRPC 模式
export YUNYEZ_AI_TRANSPORT_MODE=grpc
cd ai && ./start.sh
```
