# AI 服务 RPC 改造方案

## 背景

当前 AI 服务（ASR、NLU、TTS）采用 HTTP 协议与 Go 后端通信。为提升性能、降低延迟、增强类型安全性，计划将通信协议改为 RPC。

## 改造目标

1. **性能提升**：减少 HTTP 协议的 overhead，降低请求延迟
2. **类型安全**：通过 Protobuf 定义接口，实现强类型约束
3. **双向通信**：支持流式传输（特别是 ASR 和 TTS 的实时场景）
4. **服务发现**：为未来服务扩展和负载均衡做准备

## 技术选型

### RPC 框架：gRPC

选择 gRPC 的原因：
- 基于 HTTP/2，支持双向流式传输
- 使用 Protobuf 作为 IDL，类型安全
- 多语言支持（Go、Python）
- 性能优异，二进制编码

### 协议定义：Protocol Buffers (Protobuf)

- 版本：protobuf v3
- 语言：Go + Python

---

## 架构设计

### 改造前架构

```
┌─────────────┐      HTTP/JSON      ┌──────────────┐
│   Go Backend│ ──────────────────► │  AI Service  │
│             │ ◄────────────────── │  (FastAPI)   │
└─────────────┘      HTTP/JSON      └──────────────┘
```

### 改造后架构

```
┌─────────────┐      gRPC/Protobuf  ┌──────────────┐
│   Go Backend│ ──────────────────► │  AI Service  │
│  (gRPC      │ ◄────────────────── │  (gRPC +     │
│   Client)   │      gRPC/Protobuf  │   FastAPI)   │
└─────────────┘                     └──────────────┘
```

---

## Protobuf 接口定义

### 文件结构

```
api/proto/
├── ai/
│   ├── asr.proto       # ASR 服务接口
│   ├── nlu.proto       # NLU 服务接口
│   └── tts.proto       # TTS 服务接口
└── common/
    └── types.proto     # 公共类型定义
```

### `api/proto/common/types.proto`

```protobuf
syntax = "proto3";

package common;

option go_package = "internal/pkg/types/pb";

// 错误响应
message Error {
  int32 code = 1;
  string message = 2;
}

// 健康检查响应
message HealthResponse {
  string status = 1;
  string version = 2;
}
```

### `api/proto/ai/asr.proto`

```protobuf
syntax = "proto3";

package ai;

option go_package = "internal/pkg/types/pb/ai";

import "common/types.proto";

// ASR 语音识别服务
service ASRService {
  // 流式语音识别（推荐用于实时场景）
  rpc StreamingRecognize(stream StreamingRecognizeRequest) returns (stream StreamingRecognizeResponse);
  
  // 单次语音识别
  rpc Recognize(RecognizeRequest) returns (RecognizeResponse);
  
  // 健康检查
  rpc Health(HealthRequest) returns (common.HealthResponse);
}

message HealthRequest {}

message StreamingRecognizeRequest {
  oneof request {
    StreamingRecognizeConfig config = 1;  // 配置信息（首包）
    bytes audio_content = 2;               // 音频数据（后续包）
  }
}

message StreamingRecognizeConfig {
  string encoding = 1;      // 音频编码：LINEAR16_PCM, OPUS, AAC
  int32 sample_rate = 2;    // 采样率：8000, 16000, 44100, 48000
  int32 num_channels = 3;   // 声道数：1=单声道，2=立体声
  string language_code = 4; // 语言：zh-CN, en-US
}

message StreamingRecognizeResponse {
  oneof result {
    StreamingRecognitionResult recognition_result = 1;  // 识别结果
    common.Error error = 2;                              // 错误信息
  }
}

message StreamingRecognitionResult {
  string transcript = 1;     // 识别文本
  float confidence = 2;      // 置信度 (0-1)
  bool is_final = 3;         // 是否为最终结果
  float stability = 4;       // 稳定性 (0-1)，仅中间结果有效
}

message RecognizeRequest {
  bytes audio_content = 1;       // 音频数据（完整）
  string encoding = 2;           // 音频编码
  int32 sample_rate = 3;         // 采样率
  int32 num_channels = 4;        // 声道数
  string language_code = 5;      // 语言
}

message RecognizeResponse {
  string transcript = 1;         // 识别文本
  float confidence = 2;          // 置信度
  common.Error error = 3;        // 错误信息
}
```

### `api/proto/ai/nlu.proto`

```protobuf
syntax = "proto3";

package ai;

option go_package = "internal/pkg/types/pb/ai";

import "common/types.proto";

// NLU 意图识别服务
service NLUService {
  // 意图识别
  rpc Predict(PredictRequest) returns (PredictResponse);
  
  // 情感识别
  rpc EmotionJudge(EmotionJudgeRequest) returns (EmotionJudgeResponse);
  
  // 健康检查
  rpc Health(HealthRequest) returns (common.HealthResponse);
}

message HealthRequest {}

message PredictRequest {
  string text = 1;           // 输入文本
}

message PredictResponse {
  string text = 1;           // 输入文本
  string intent = 2;         // 意图
  float confidence = 3;      // 置信度
  bool is_command = 4;       // 是否为命令意图
  common.Error error = 5;    // 错误信息
}

message EmotionJudgeRequest {
  string text = 1;           // 输入文本
}

message EmotionJudgeResponse {
  string text = 1;           // 输入文本
  string emotion = 2;        // 情感类型
  float confidence = 3;      // 置信度
  common.Error error = 4;    // 错误信息
}
```

### `api/proto/ai/tts.proto`

```protobuf
syntax = "proto3";

package ai;

option go_package = "internal/pkg/types/pb/ai";

import "common/types.proto";

// TTS 语音合成服务
service TTSService {
  // 流式语音合成（推荐用于实时场景）
  rpc StreamingSynthesize(stream StreamingSynthesizeRequest) returns (stream StreamingSynthesizeResponse);
  
  // 单次语音合成
  rpc Synthesize(SynthesizeRequest) returns (SynthesizeResponse);
  
  // 健康检查
  rpc Health(HealthRequest) returns (common.HealthResponse);
}

message HealthRequest {}

message StreamingSynthesizeRequest {
  oneof request {
    StreamingSynthesizeConfig config = 1;  // 配置信息（首包）
    string text = 2;                        // 文本内容（后续包，支持分块）
  }
}

message StreamingSynthesizeConfig {
  string voice = 1;         // 语音：zh-CN-XiaoxiaoNeural, en-US-JennyNeural
  float rate = 2;           // 语速：0.5-2.0
  float pitch = 3;          // 音调：0.5-2.0
  float volume = 4;         // 音量：0.0-1.0
  string output_format = 5; // 输出格式：AUDIO_16KHZ_16BIT_RAW_PCM, AUDIO_24KHZ_48KBITRATE_GZIP
}

message StreamingSynthesizeResponse {
  oneof result {
    bytes audio_content = 1;  // 音频数据
    common.Error error = 2;    // 错误信息
  }
}

message SynthesizeRequest {
  string text = 1;            // 输入文本
  string voice = 2;           // 语音
  float rate = 3;             // 语速
  float pitch = 4;            // 音调
  float volume = 5;           // 音量
  string output_format = 6;   // 输出格式
}

message SynthesizeResponse {
  bytes audio_content = 1;    // 音频数据
  int32 duration_ms = 2;      // 音频时长（毫秒）
  common.Error error = 3;     // 错误信息
}
```

---

## Go 端改造

### 1. 生成 Protobuf 代码

```bash
# 安装工具
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 生成 Go 代码
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       api/proto/ai/*.proto api/proto/common/*.proto
```

生成文件结构：
```
internal/pkg/types/pb/
├── common/
│   └── types.pb.go
└── ai/
    ├── asr.pb.go
    ├── asr_grpc.pb.go
    ├── nlu.pb.go
    ├── nlu_grpc.pb.go
    ├── tts.pb.go
    └── tts_grpc.pb.go
```

### 2. 改造 ASR 客户端

**文件：`internal/pkg/agent/asr/grpc_client.go`**

```go
// Package asr 语音识别服务
package asr

import (
    "context"
    "fmt"
    "io"
    
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    pb "yunyez/internal/pkg/types/pb/ai"
)

// GRPCClient gRPC ASR 客户端
type GRPCClient struct {
    conn   *grpc.ClientConn
    client pb.ASRServiceClient
}

// NewGRPCClient 创建 gRPC 客户端
func NewGRPCClient(endpoint string) (*GRPCClient, error) {
    conn, err := grpc.NewClient(endpoint,
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    if err != nil {
        return nil, fmt.Errorf("create gRPC connection: %w", err)
    }
    
    return &GRPCClient{
        conn:   conn,
        client: pb.NewASRServiceClient(conn),
    }, nil
}

// Close 关闭连接
func (c *GRPCClient) Close() error {
    return c.conn.Close()
}

// Transfer 语音识别（单次）
func (c *GRPCClient) Transfer(ctx context.Context, data []byte) (string, error) {
    resp, err := c.client.Recognize(ctx, &pb.RecognizeRequest{
        AudioContent: data,
        Encoding:     "LINEAR16_PCM",
        SampleRate:   16000,
        NumChannels:  1,
        LanguageCode: "zh-CN",
    })
    if err != nil {
        return "", fmt.Errorf("call ASR service: %w", err)
    }
    
    if resp.Error != nil {
        return "", fmt.Errorf("ASR error: %s", resp.Error.Message)
    }
    
    return resp.Transcript, nil
}

// StreamingTransfer 流式语音识别
func (c *GRPCClient) StreamingTransfer(ctx context.Context) (StreamingASRStream, error) {
    stream, err := c.client.StreamingRecognize(ctx)
    if err != nil {
        return nil, fmt.Errorf("create streaming ASR: %w", err)
    }
    
    // 发送配置
    err = stream.Send(&pb.StreamingRecognizeRequest{
        Request: &pb.StreamingRecognizeRequest_Config{
            Config: &pb.StreamingRecognizeConfig{
                Encoding:     "LINEAR16_PCM",
                SampleRate:   16000,
                NumChannels:  1,
                LanguageCode: "zh-CN",
            },
        },
    })
    if err != nil {
        return nil, fmt.Errorf("send config: %w", err)
    }
    
    return &grpcASRStream{stream: stream}, nil
}

// StreamingASRStream 流式 ASR 接口
type StreamingASRStream interface {
    Send(audioData []byte) error
    Recv() (transcript string, confidence float32, isFinal bool, err error)
    Close() error
}

type grpcASRStream struct {
    stream pb.ASRService_StreamingRecognizeClient
}

func (s *grpcASRStream) Send(audioData []byte) error {
    return s.stream.Send(&pb.StreamingRecognizeRequest{
        Request: &pb.StreamingRecognizeRequest_AudioContent{
            AudioContent: audioData,
        },
    })
}

func (s *grpcASRStream) Recv() (string, float32, bool, error) {
    resp, err := s.stream.Recv()
    if err != nil {
        if err == io.EOF {
            return "", 0, false, io.EOF
        }
        return "", 0, false, err
    }
    
    if resp.GetError() != nil {
        return "", 0, false, fmt.Errorf("ASR error: %s", resp.GetError().Message)
    }
    
    result := resp.GetRecognitionResult()
    return result.Transcript, result.Confidence, result.IsFinal, nil
}

func (s *grpcASRStream) Close() error {
    return s.stream.CloseSend()
}
```

### 3. 改造 NLU 客户端

**文件：`internal/pkg/agent/nlu/grpc_client.go`**

```go
// Package nlu NLU 意图识别客户端
package nlu

import (
    "context"
    "fmt"
    
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    pb "yunyez/internal/pkg/types/pb/ai"
)

// GRPCClient gRPC NLU 客户端
type GRPCClient struct {
    conn   *grpc.ClientConn
    client pb.NLUServiceClient
}

// NewGRPCClient 创建 gRPC 客户端
func NewGRPCClient(endpoint string) (*GRPCClient, error) {
    conn, err := grpc.NewClient(endpoint,
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    if err != nil {
        return nil, fmt.Errorf("create gRPC connection: %w", err)
    }
    
    return &GRPCClient{
        conn:   conn,
        client: pb.NewNLUServiceClient(conn),
    }, nil
}

// Close 关闭连接
func (c *GRPCClient) Close() error {
    return c.conn.Close()
}

// Predict 意图识别
func (c *GRPCClient) Predict(ctx context.Context, text string) (*Intent, error) {
    resp, err := c.client.Predict(ctx, &pb.PredictRequest{
        Text: text,
    })
    if err != nil {
        return nil, fmt.Errorf("call NLU service: %w", err)
    }
    
    if resp.Error != nil {
        return nil, fmt.Errorf("NLU error: %s", resp.Error.Message)
    }
    
    return &Intent{
        Text:       resp.Text,
        Intent:     resp.Intent,
        Confidence: resp.Confidence,
        IsCommand:  resp.IsCommand,
    }, nil
}

// Health 健康检查
func (c *GRPCClient) Health(ctx context.Context) error {
    resp, err := c.client.Health(ctx, &pb.HealthRequest{})
    if err != nil {
        return fmt.Errorf("health check failed: %w", err)
    }
    
    if resp.Status != "ok" {
        return fmt.Errorf("unhealthy status: %s", resp.Status)
    }
    
    return nil
}

// Intent NLU 意图识别结果
type Intent struct {
    Text       string
    Intent     string
    Confidence float32
    IsCommand  bool
}
```

### 4. 改造 TTS 客户端

**文件：`internal/pkg/agent/tts/grpc_client.go`**

```go
// Package tts TTS 语音合成服务
package tts

import (
    "context"
    "fmt"
    "io"
    
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    pb "yunyez/internal/pkg/types/pb/ai"
)

// GRPCClient gRPC TTS 客户端
type GRPCClient struct {
    conn   *grpc.ClientConn
    client pb.TTSServiceClient
}

// NewGRPCClient 创建 gRPC 客户端
func NewGRPCClient(endpoint string) (*GRPCClient, error) {
    conn, err := grpc.NewClient(endpoint,
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    if err != nil {
        return nil, fmt.Errorf("create gRPC connection: %w", err)
    }
    
    return &GRPCClient{
        conn:   conn,
        client: pb.NewTTSServiceClient(conn),
    }, nil
}

// Close 关闭连接
func (c *GRPCClient) Close() error {
    return c.conn.Close()
}

// Synthesize 语音合成（单次）
func (c *GRPCClient) Synthesize(ctx context.Context, text string) ([]byte, error) {
    resp, err := c.client.Synthesize(ctx, &pb.SynthesizeRequest{
        Text:         text,
        Voice:        "zh-CN-XiaoxiaoNeural",
        Rate:         1.0,
        Pitch:        1.0,
        Volume:       1.0,
        OutputFormat: "AUDIO_16KHZ_16BIT_RAW_PCM",
    })
    if err != nil {
        return nil, fmt.Errorf("call TTS service: %w", err)
    }
    
    if resp.Error != nil {
        return nil, fmt.Errorf("TTS error: %s", resp.Error.Message)
    }
    
    return resp.AudioContent, nil
}

// StreamingSynthesize 流式语音合成
func (c *GRPCClient) StreamingSynthesize(ctx context.Context) (StreamingTTSSStream, error) {
    stream, err := c.client.StreamingSynthesize(ctx)
    if err != nil {
        return nil, fmt.Errorf("create streaming TTS: %w", err)
    }
    
    // 发送配置
    err = stream.Send(&pb.StreamingSynthesizeRequest{
        Request: &pb.StreamingSynthesizeRequest_Config{
            Config: &pb.StreamingSynthesizeConfig{
                Voice:        "zh-CN-XiaoxiaoNeural",
                Rate:         1.0,
                Pitch:        1.0,
                Volume:       1.0,
                OutputFormat: "AUDIO_16KHZ_16BIT_RAW_PCM",
            },
        },
    })
    if err != nil {
        return nil, fmt.Errorf("send config: %w", err)
    }
    
    return &grpcTTSStream{stream: stream}, nil
}

// StreamingTTSSStream 流式 TTS 接口
type StreamingTTSSStream interface {
    Send(text string) error
    Recv() (audioData []byte, err error)
    Close() error
}

type grpcTTSStream struct {
    stream pb.TTSService_StreamingSynthesizeClient
}

func (s *grpcTTSStream) Send(text string) error {
    return s.stream.Send(&pb.StreamingSynthesizeRequest{
        Request: &pb.StreamingSynthesizeRequest_Text{
            Text: text,
        },
    })
}

func (s *grpcTTSStream) Recv() ([]byte, error) {
    resp, err := s.stream.Recv()
    if err != nil {
        if err == io.EOF {
            return nil, io.EOF
        }
        return nil, err
    }
    
    if resp.GetError() != nil {
        return nil, fmt.Errorf("TTS error: %s", resp.GetError().Message)
    }
    
    return resp.GetAudioContent(), nil
}

func (s *grpcTTSStream) Close() error {
    return s.stream.CloseSend()
}
```

---

## Python 端改造

### 1. 安装依赖

**文件：`ai/asr/requirements.txt`**

```txt
fastapi==0.104.1
uvicorn[standard]==0.24.0
grpcio==1.59.3
grpcio-tools==1.59.3
# 其他原有依赖...
```

### 2. 生成 Python Protobuf 代码

```bash
# 在 ai 目录下
python -m grpc_tools.protoc \
    -I../api/proto \
    --python_out=asr/src/pb \
    --grpc_python_out=asr/src/pb \
    ../api/proto/ai/asr.proto ../api/proto/common/types.proto
```

### 3. 改造 ASR 服务（兼容 HTTP + gRPC）

**文件：`ai/asr/src/grpc_server.py`**

```python
"""ASR gRPC 服务"""
import grpc
from concurrent import futures
import logging

from .pb import asr_pb2
from .pb import asr_pb2_grpc
from .engine import asr_engine
import tempfile
import os

logger = logging.getLogger(__name__)


class ASRServicer(asr_pb2_grpc.ASRServiceServicer):
    """ASR gRPC 服务实现"""
    
    def Recognize(self, request, context):
        """单次语音识别"""
        try:
            # 保存临时文件
            with tempfile.NamedTemporaryFile(delete=False, suffix=".pcm") as tmp:
                tmp.write(request.audio_content)
                tmp_path = tmp.name
            
            try:
                text = asr_engine.transcribe(tmp_path)
                return asr_pb2.RecognizeResponse(
                    transcript=text,
                    confidence=0.95  # 假设置信度
                )
            finally:
                os.unlink(tmp_path)
                
        except Exception as e:
            logger.error(f"ASR recognize error: {e}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"ASR failed: {str(e)}")
            return asr_pb2.RecognizeResponse()
    
    def StreamingRecognize(self, request_iterator, context):
        """流式语音识别"""
        config = None
        audio_chunks = []
        
        try:
            for request in request_iterator:
                if request.HasField('config'):
                    config = request.config
                    logger.info(f"Streaming ASR config: {config.encoding}, {config.sample_rate}Hz")
                elif request.audio_content:
                    audio_chunks.append(request.audio_content)
            
            # 合并音频数据
            audio_data = b''.join(audio_chunks)
            
            # 保存临时文件
            with tempfile.NamedTemporaryFile(delete=False, suffix=".pcm") as tmp:
                tmp.write(audio_data)
                tmp_path = tmp.name
            
            try:
                text = asr_engine.transcribe(tmp_path)
                
                # 发送最终结果
                yield asr_pb2.StreamingRecognizeResponse(
                    recognition_result=asr_pb2.StreamingRecognitionResult(
                        transcript=text,
                        confidence=0.95,
                        is_final=True,
                        stability=1.0
                    )
                )
            finally:
                os.unlink(tmp_path)
                
        except Exception as e:
            logger.error(f"Streaming ASR error: {e}")
            yield asr_pb2.StreamingRecognizeResponse(
                error=asr_pb2.Error(code=500, message=f"ASR failed: {str(e)}")
            )
    
    def Health(self, request, context):
        """健康检查"""
        return asr_pb2.HealthResponse(status="ok", version="1.0.0")


def serve(port=50051):
    """启动 gRPC 服务"""
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    asr_pb2_grpc.add_ASRServiceServicer_to_server(ASRServicer(), server)
    server.add_insecure_port(f'[::]:{port}')
    server.start()
    logger.info(f"ASR gRPC server started on port {port}")
    server.wait_for_termination()


if __name__ == '__main__':
    serve()
```

**文件：`ai/asr/src/server.py`（改造 main.py 兼容双协议）**

```python
"""ASR 服务入口 - 同时支持 HTTP 和 gRPC"""
import asyncio
import threading
from fastapi import FastAPI, Request
import uvicorn

from .grpc_server import serve as serve_grpc

app = FastAPI(title="Yunyez ASR Service")

# ... 保留原有 HTTP 接口 ...

@app.post("/asr")
async def transcribe(request: Request):
    # 保留原有 HTTP 实现
    pass

def run_grpc_server(port=50051):
    """在独立线程中运行 gRPC 服务器"""
    serve_grpc(port)

def main():
    """启动 HTTP + gRPC 双协议服务"""
    # 启动 gRPC 服务器（独立线程）
    grpc_thread = threading.Thread(target=run_grpc_server, daemon=True)
    grpc_thread.start()
    
    # 启动 HTTP 服务器
    uvicorn.run(app, host="0.0.0.0", port=8002)

if __name__ == "__main__":
    main()
```

### 4. 改造 NLU 服务

**文件：`ai/nlu/src/grpc_server.py`**

```python
"""NLU gRPC 服务"""
import grpc
from concurrent import futures
import logging
import numpy as np

from .pb import nlu_pb2
from .pb import nlu_pb2_grpc
from .engine import nlu_engine  # 假设的 NLU 引擎

logger = logging.getLogger(__name__)


class NLUServicer(nlu_pb2_grpc.NLUServiceServicer):
    """NLU gRPC 服务实现"""
    
    def Predict(self, request, context):
        """意图识别"""
        try:
            text = request.text
            emb = nlu_engine.encoder.encode([text])
            intent = nlu_engine.classifier.predict(emb)[0]
            confidence = float(np.max(nlu_engine.classifier.predict_proba(emb)))
            is_command = intent in nlu_engine.COMMAND_INTENTS
            
            return nlu_pb2.PredictResponse(
                text=text,
                intent=intent,
                confidence=confidence,
                is_command=is_command
            )
            
        except Exception as e:
            logger.error(f"NLU predict error: {e}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"NLU failed: {str(e)}")
            return nlu_pb2.PredictResponse()
    
    def EmotionJudge(self, request, context):
        """情感识别"""
        # TODO: 实现情感识别
        return nlu_pb2.EmotionJudgeResponse(
            text=request.text,
            emotion="neutral",
            confidence=0.5
        )
    
    def Health(self, request, context):
        """健康检查"""
        return nlu_pb2.HealthResponse(status="ok", version="1.0.0")


def serve(port=50052):
    """启动 gRPC 服务"""
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    nlu_pb2_grpc.add_NLUServiceServicer_to_server(NLUServicer(), server)
    server.add_insecure_port(f'[::]:{port}')
    server.start()
    logger.info(f"NLU gRPC server started on port {port}")
    server.wait_for_termination()
```

### 5. 改造 TTS 服务

**文件：`ai/tts/src/grpc_server.py`**

```python
"""TTS gRPC 服务"""
import grpc
from concurrent import futures
import logging

from .pb import tts_pb2
from .pb import tts_pb2_grpc
from .edgeTTS import synthesize_audio  # 假设的合成函数

logger = logging.getLogger(__name__)


class TTSServicer(tts_pb2_grpc.TTSServiceServicer):
    """TTS gRPC 服务实现"""
    
    def Synthesize(self, request, context):
        """单次语音合成"""
        try:
            audio_data = synthesize_audio(
                text=request.text,
                voice=request.voice,
                rate=request.rate,
                pitch=request.pitch,
                volume=request.volume
            )
            
            return tts_pb2.SynthesizeResponse(
                audio_content=audio_data,
                duration_ms=len(audio_data) // 32  # 假设 16kHz 16bit
            )
            
        except Exception as e:
            logger.error(f"TTS synthesize error: {e}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"TTS failed: {str(e)}")
            return tts_pb2.SynthesizeResponse()
    
    def StreamingSynthesize(self, request_iterator, context):
        """流式语音合成"""
        config = None
        
        try:
            text_chunks = []
            
            for request in request_iterator:
                if request.HasField('config'):
                    config = request.config
                    logger.info(f"Streaming TTS config: {config.voice}")
                elif request.text:
                    text_chunks.append(request.text)
            
            # 合并文本
            full_text = ''.join(text_chunks)
            
            # 合成音频
            audio_data = synthesize_audio(
                text=full_text,
                voice=config.voice if config else "zh-CN-XiaoxiaoNeural",
                rate=config.rate if config else 1.0,
                pitch=config.pitch if config else 1.0,
                volume=config.volume if config else 1.0
            )
            
            # 发送音频数据
            yield tts_pb2.StreamingSynthesizeResponse(
                audio_content=audio_data
            )
            
        except Exception as e:
            logger.error(f"Streaming TTS error: {e}")
            yield tts_pb2.StreamingSynthesizeResponse(
                error=tts_pb2.Error(code=500, message=f"TTS failed: {str(e)}")
            )
    
    def Health(self, request, context):
        """健康检查"""
        return tts_pb2.HealthResponse(status="ok", version="1.0.0")


def serve(port=50053):
    """启动 gRPC 服务"""
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    tts_pb2_grpc.add_TTSServiceServicer_to_server(TTSServicer(), server)
    server.add_insecure_port(f'[::]:{port}')
    server.start()
    logger.info(f"TTS gRPC server started on port {port}")
    server.wait_for_termination()
```

---

## 配置更新

### Go 端配置

**文件：`configs/dev/ai.yaml`**

```yaml
ai:
  asr:
    # HTTP 模式（向后兼容）
    http_endpoint: "http://127.0.0.1:8002/asr"
    # gRPC 模式（新）
    grpc_endpoint: "127.0.0.1:50051"
    protocol: "grpc"  # 可选：http, grpc
    model: "local"
    
  nlu:
    http_endpoint: "http://127.0.0.1:8001/nlu"
    grpc_endpoint: "127.0.0.1:50052"
    protocol: "grpc"
    
  tts:
    http_endpoint: "http://127.0.0.1:8003/tts"
    grpc_endpoint: "127.0.0.1:50053"
    protocol: "grpc"
    voice: "zh-CN-XiaoxiaoNeural"
    rate: 1.0
    pitch: 1.0
    volume: 1.0
```

### Docker Compose 更新

**文件：`ai/docker-compose.yml`**

```yaml
version: '3.8'

services:
  asr:
    build:
      context: ./asr
      dockerfile: Dockerfile
    ports:
      - "8002:8002"  # HTTP
      - "50051:50051"  # gRPC
    volumes:
      - ./asr:/app
    command: python -m src.server
    networks:
      - yunyez-network

  nlu:
    build:
      context: ./nlu
      dockerfile: Dockerfile
    ports:
      - "8001:8001"  # HTTP
      - "50052:50052"  # gRPC
    volumes:
      - ./nlu:/app
    command: python -m src.server
    networks:
      - yunyez-network

  tts:
    build:
      context: ./tts
      dockerfile: Dockerfile
    ports:
      - "8003:8003"  # HTTP
      - "50053:50053"  # gRPC
    volumes:
      - ./tts:/app
    command: python -m src.server
    networks:
      - yunyez-network

networks:
  yunyez-network:
    driver: bridge
```

---

## 迁移计划

### 阶段一：准备阶段（Week 1）

1. ✅ 编写 Protobuf 接口定义
2. ⬜ 生成 Go 和 Python 的 Protobuf 代码
3. ⬜ 更新项目依赖（go.mod, requirements.txt）

### 阶段二：双协议兼容（Week 2-3）

1. ⬜ Python 端实现 gRPC 服务（保留 HTTP）
2. ⬜ Go 端实现 gRPC 客户端（保留 HTTP）
3. ⬜ 编写单元测试和集成测试
4. ⬜ 更新 Docker 配置

### 阶段三：切换验证（Week 4）

1. ⬜ 配置默认使用 gRPC 协议
2. ⬜ 性能测试和对比
3. ⬜ 灰度发布验证

### 阶段四：清理 HTTP（可选）

1. ⬜ 移除 HTTP 兼容代码
2. ⬜ 更新文档
3. ⬜ 清理依赖

---

## 性能对比预期

| 指标 | HTTP/JSON | gRPC/Protobuf | 提升 |
|------|-----------|---------------|------|
| 请求延迟 | ~50ms | ~20ms | 60% ↓ |
| 序列化大小 | ~500 bytes | ~200 bytes | 60% ↓ |
| 流式支持 | 有限 | 原生支持 | - |
| 类型安全 | 弱 | 强 | - |
| 多语言支持 | 是 | 是 | - |

---

## 注意事项

1. **向后兼容**：改造期间保留 HTTP 接口，确保平滑迁移
2. **错误处理**：gRPC 有独立的错误码体系，需要映射转换
3. **连接管理**：gRPC 需要管理连接生命周期，避免泄漏
4. **超时控制**：设置合理的 gRPC 超时时间
5. **负载均衡**：未来可接入 gRPC 负载均衡器（如 gRPC-LB）

---

## 相关文件索引

| 文件 | 描述 |
|------|------|
| `api/proto/ai/asr.proto` | ASR Protobuf 定义 |
| `api/proto/ai/nlu.proto` | NLU Protobuf 定义 |
| `api/proto/ai/tts.proto` | TTS Protobuf 定义 |
| `internal/pkg/agent/asr/grpc_client.go` | Go ASR gRPC 客户端 |
| `internal/pkg/agent/nlu/grpc_client.go` | Go NLU gRPC 客户端 |
| `internal/pkg/agent/tts/grpc_client.go` | Go TTS gRPC 客户端 |
| `ai/asr/src/grpc_server.py` | Python ASR gRPC 服务 |
| `ai/nlu/src/grpc_server.py` | Python NLU gRPC 服务 |
| `ai/tts/src/grpc_server.py` | Python TTS gRPC 服务 |
