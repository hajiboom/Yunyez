# Yunyez AI 服务

本目录包含云也子 (Yunyez) 项目的 AI 核心组件，为独自旅行者提供智能语音交互能力。

## 组件概览

| 服务 | 功能 | 模型/技术 | 端口 |
|------|------|----------|------|
| **ASR** | 语音识别 | FunASR + Paraformer-large | 8002 (HTTP) / 50052 (gRPC) |
| **NLU** | 意图识别 | Sentence-BERT + LogisticRegression | 8001 (HTTP) / 50051 (gRPC) |
| **TTS** | 语音合成 | Edge TTS | 8003 (HTTP) / 50053 (gRPC) |
| **LLM** | 大语言模型 | 通义千问 (Qwen-Flash) | 云端 API |

---

## 模型详情

### 1. ASR (语音识别)

**模型**: `iic/speech_paraformer-large-vad-punc_asr_nat-zh-cn-16k-common-vocab8404-pytorch`

- **框架**: FunASR 1.1.8
- **采样率**: 16kHz
- **语言**: 中文
- **特性**: 
  - 支持 VAD (语音活动检测)
  - 自动标点
  - 去空格处理
- **部署**: 本地部署，支持 CUDA 加速

**配置**:
```yaml
asr:
  model: local
  protocol: http
  http_endpoint: "http://127.0.0.1:8002/asr"
```

---

### 2. NLU (意图识别)

**模型架构**:
- **Encoder**: `paraphrase-multilingual-MiniLM-L12-v2` (Sentence-BERT)
- **分类器**: LogisticRegression

**技术细节**:
- **向量维度**: 384 维
- **训练框架**: PyTorch + HuggingFace Transformers
- **分类框架**: scikit-learn
- **语言**: 多语言支持 (侧重中文)

**配置**:
```yaml
nlu:
  model: local
  protocol: http
  http_endpoint: "http://127.0.0.1:8001/nlu"
```

**训练数据**:
- 数据来源：`ai/nlu/data/train.csv`
- 训练命令：`python ai/nlu/src/train.py`
- 模型保存：`ai/nlu/model/`

---

### 3. TTS (语音合成)

**模型**: Edge TTS (Microsoft Edge 在线 TTS 服务)

**默认配置**:
- **音色**: `zh-CN-XiaoyiNeural` (小艺)
- **语速**: `+0%`
- **音调**: `-1Hz`
- **音量**: `+0%`
- **输出格式**: MP3

**配置**:
```yaml
tts:
  model: edge
  protocol: http
  edge:
    endpoint: "http://127.0.0.1:8003/tts"
    params:
      voice: zh-CN-XiaoyiNeural
      rate: "+0%"
      pitch: "-1Hz"
      volume: "+0%"
```

**扩展支持**:
- ChatTTS (本地部署，待启用)
- 阿里云 TTS (预留接口)

---

### 4. LLM (大语言模型)

**模型**: 通义千问 - Qwen-Flash

**配置**:
```yaml
agent:
  model: qwen

qwen:
  model: qwen-flash
  endpoint: https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions
  systemDesc: "你是一个智能语音助手，请用简洁的纯中文回答问题"
```

**计费**:
- 输入：¥0.001 / 1K tokens
- 输出：¥0.002 / 1K tokens

---

## 快速开始

### 环境要求

- **Python**: 3.10.11+
- **CUDA**: 12.1+ (可选，用于 ASR 加速)
- **Docker**: 20.10+ (容器化部署)

### 开发环境启动

```bash
# 进入 ai 目录
cd ai

# 启动所有服务 (HTTP 模式)
./start_all.sh

# 或使用 Docker Compose
docker-compose --profile http up --build
```

### 服务验证

访问 API 文档:
- ASR: http://127.0.0.1:8002/docs
- NLU: http://127.0.0.1:8001/docs
- TTS: http://127.0.0.1:8003/docs

---

## Docker 部署

### HTTP 模式

```bash
docker-compose --profile http up -d
```

### gRPC 模式

```bash
docker-compose --profile grpc up -d
```

### 服务端口

| 服务 | HTTP 端口 | gRPC 端口 |
|------|----------|----------|
| ASR | 8002 | 50052 |
| NLU | 8001 | 50051 |
| TTS | 8003 | 50053 |

---

## 目录结构

```
ai/
├── asr/                    # 语音识别服务
│   ├── src/
│   │   ├── config.py       # 模型配置
│   │   ├── engine.py       # 识别引擎
│   │   ├── server.py       # HTTP 服务
│   │   └── grpc_server.py  # gRPC 服务
│   ├── scripts/
│   │   ├── fine_tune.py    # 微调脚本
│   │   └── test.py         # 测试脚本
│   └── README.md           # ASR 详细文档
├── nlu/                    # 意图识别服务
│   ├── src/
│   │   ├── train.py        # 训练脚本
│   │   ├── server.py       # HTTP 服务
│   │   └── grpc_server.py  # gRPC 服务
│   ├── data/               # 训练数据
│   ├── model/              # 训练好的模型
│   └── README.md           # NLU 详细文档
├── tts/                    # 语音合成服务
│   ├── src/
│   │   ├── edgeTTS.py      # Edge TTS 实现
│   │   ├── server.py       # HTTP 服务
│   │   └── grpc_server.py  # gRPC 服务
│   └── README.md           # TTS 详细文档
├── docker-compose.yml      # Docker 编排配置
├── requirements-common.txt # 公共依赖
└── start.sh                # 启动脚本
```

---

## 配置文件

后端服务通过 `configs/dev/ai.yaml` 配置 AI 服务连接:

```yaml
# LLM 模型选择
agent:
  model: qwen

# NLU 语义识别服务
nlu:
  model: local        # 本地 NLU 模型
  protocol: http
  http_endpoint: "http://127.0.0.1:8001/nlu"

# ASR 语音识别服务
asr:
  model: local
  protocol: http
  http_endpoint: "http://127.0.0.1:8002/asr"

# TTS 语音合成服务
tts:
  model: edge
  protocol: http
  edge:
    endpoint: "http://127.0.0.1:8003/tts"
```

---

## 开发指南

### ASR 微调

```bash
python ai/asr/scripts/fine_tune.py
```

### NLU 训练

```bash
python ai/nlu/src/train.py
```

### 测试语音识别

```bash
python ai/asr/scripts/test.py
```

---

## 性能指标

| 指标 | 目标值 |
|------|--------|
| ASR 识别延迟 | < 500ms |
| NLU 推理延迟 | < 100ms |
| TTS 首包延迟 | < 300ms |
| 端到端响应 | < 1.5s |

---

## 相关文档

- [ASR 详细文档](asr/README.md)
- [NLU 详细文档](nlu/README.md)
- [实时流式语音架构](../docs/hard/realtime-voice-streaming.md)
- [通信协议规范](../docs/protocol.md)
