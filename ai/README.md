# Yunyez AI Services

This directory contains the AI core components for the Yunyez project, providing intelligent voice interaction capabilities for solo travelers.

## Components Overview

| Service | Function | Model/Technology | Ports |
|---------|----------|------------------|-------|
| **ASR** | Speech Recognition | FunASR + Paraformer-large | 8002 (HTTP) / 50052 (gRPC) |
| **NLU** | Intent Recognition | Sentence-BERT + LogisticRegression | 8001 (HTTP) / 50051 (gRPC) |
| **TTS** | Speech Synthesis | Edge TTS | 8003 (HTTP) / 50053 (gRPC) |
| **LLM** | Large Language Model | Qwen-Flash (Aliyun) | Cloud API |

---

## Model Details

### 1. ASR (Automatic Speech Recognition)

**Model**: `iic/speech_paraformer-large-vad-punc_asr_nat-zh-cn-16k-common-vocab8404-pytorch`

- **Framework**: FunASR 1.1.8
- **Sample Rate**: 16kHz
- **Language**: Chinese
- **Features**: 
  - VAD (Voice Activity Detection)
  - Auto-punctuation
  - Space removal
- **Deployment**: Local deployment with CUDA acceleration support

**Configuration**:
```yaml
asr:
  model: local
  protocol: http
  http_endpoint: "http://127.0.0.1:8002/asr"
```

---

### 2. NLU (Natural Language Understanding)

**Model Architecture**:
- **Encoder**: `paraphrase-multilingual-MiniLM-L12-v2` (Sentence-BERT)
- **Classifier**: LogisticRegression

**Technical Details**:
- **Embedding Dimension**: 384
- **Training Framework**: PyTorch + HuggingFace Transformers
- **Classification Framework**: scikit-learn
- **Language**: Multi-language support (Chinese focused)

**Configuration**:
```yaml
nlu:
  model: local
  protocol: http
  http_endpoint: "http://127.0.0.1:8001/nlu"
```

**Training Data**:
- Data source: `ai/nlu/data/train.csv`
- Training command: `python ai/nlu/src/train.py`
- Model output: `ai/nlu/model/`

---

### 3. TTS (Text-to-Speech)

**Model**: Edge TTS (Microsoft Edge Online TTS Service)

**Default Configuration**:
- **Voice**: `zh-CN-XiaoyiNeural` (Xiaoyi)
- **Rate**: `+0%`
- **Pitch**: `-1Hz`
- **Volume**: `+0%`
- **Output Format**: MP3

**Configuration**:
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

**Extended Support**:
- ChatTTS (Local deployment, pending activation)
- Aliyun TTS (Reserved interface)

---

### 4. LLM (Large Language Model)

**Model**: Qwen-Flash (Alibaba Cloud)

**Configuration**:
```yaml
agent:
  model: qwen

qwen:
  model: qwen-flash
  endpoint: https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions
  systemDesc: "You are an intelligent voice assistant. Please answer in concise, pure Chinese."
```

**Pricing**:
- Input: ¥0.001 / 1K tokens
- Output: ¥0.002 / 1K tokens

---

## Quick Start

### Prerequisites

- **Python**: 3.10.11+
- **CUDA**: 12.1+ (Optional, for ASR acceleration)
- **Docker**: 20.10+ (Containerized deployment)

### Development Setup

```bash
# Navigate to ai directory
cd ai

# Start all services (HTTP mode)
./start_all.sh

# Or use Docker Compose
docker-compose --profile http up --build
```

### Service Verification

Access API documentation:
- ASR: http://127.0.0.1:8002/docs
- NLU: http://127.0.0.1:8001/docs
- TTS: http://127.0.0.1:8003/docs

---

## Docker Deployment

### HTTP Mode

```bash
docker-compose --profile http up -d
```

### gRPC Mode

```bash
docker-compose --profile grpc up -d
```

### Service Ports

| Service | HTTP Port | gRPC Port |
|---------|-----------|-----------|
| ASR | 8002 | 50052 |
| NLU | 8001 | 50051 |
| TTS | 8003 | 50053 |

---

## Directory Structure

```
ai/
├── asr/                    # Speech Recognition Service
│   ├── src/
│   │   ├── config.py       # Model configuration
│   │   ├── engine.py       # Recognition engine
│   │   ├── server.py       # HTTP server
│   │   └── grpc_server.py  # gRPC server
│   ├── scripts/
│   │   ├── fine_tune.py    # Fine-tuning script
│   │   └── test.py         # Test script
│   └── README.md           # ASR detailed documentation
├── nlu/                    # Intent Recognition Service
│   ├── src/
│   │   ├── train.py        # Training script
│   │   ├── server.py       # HTTP server
│   │   └── grpc_server.py  # gRPC server
│   ├── data/               # Training data
│   ├── model/              # Trained models
│   └── README.md           # NLU detailed documentation
├── tts/                    # Speech Synthesis Service
│   ├── src/
│   │   ├── edgeTTS.py      # Edge TTS implementation
│   │   ├── server.py       # HTTP server
│   │   └── grpc_server.py  # gRPC server
│   └── README.md           # TTS detailed documentation
├── docker-compose.yml      # Docker orchestration configuration
├── requirements-common.txt # Common dependencies
└── start.sh                # Startup script
```

---

## Configuration Files

Backend services configure AI service connections via `configs/dev/ai.yaml`:

```yaml
# LLM model selection
agent:
  model: qwen

# NLU semantic recognition service
nlu:
  model: local        # Local NLU model
  protocol: http
  http_endpoint: "http://127.0.0.1:8001/nlu"

# ASR speech recognition service
asr:
  model: local
  protocol: http
  http_endpoint: "http://127.0.0.1:8002/asr"

# TTS speech synthesis service
tts:
  model: edge
  protocol: http
  edge:
    endpoint: "http://127.0.0.1:8003/tts"
```

---

## Development Guide

### ASR Fine-tuning

```bash
python ai/asr/scripts/fine_tune.py
```

### NLU Training

```bash
python ai/nlu/src/train.py
```

### Test Speech Recognition

```bash
python ai/asr/scripts/test.py
```

---

## Performance Metrics

| Metric | Target |
|--------|--------|
| ASR Recognition Latency | < 500ms |
| NLU Inference Latency | < 100ms |
| TTS First Packet Latency | < 300ms |
| End-to-End Response | < 1.5s |

---

## Related Documentation

- [ASR Detailed Documentation](asr/README.md)
- [NLU Detailed Documentation](nlu/README.md)
- [Realtime Voice Streaming Architecture](../docs/hard/realtime-voice-streaming.md)
- [Communication Protocol Specification](../docs/protocol.md)
