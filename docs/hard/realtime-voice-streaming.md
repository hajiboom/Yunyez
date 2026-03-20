# 实时流式语音传输架构设计

## 1. 概述

本文档描述基于本地麦克风的实时流式语音传输系统架构，用于在开发环境中替代文件回放测试，实现真实的语音交互体验。

### 1.1 背景

当前 `example/mock/virtual_voice` 通过回放预录制的 WAV 文件模拟设备端语音上传，存在以下局限：

- **非实时性**: 无法测试真实场景下的语音打断、实时响应
- **固定内容**: 每次测试内容相同，无法验证 ASR 对不同语音的识别效果
- **缺少流式特性**: 文件是一次性发送，而非分帧流式传输

### 1.2 目标

在 `example/mock/virtual_voice_realtime` 中实现：

1. **实时采集**: 调用本地电脑麦克风，实时采集音频数据（使用 C + PortAudio）
2. **流式传输**: 按照 Yunyez 自定义 MQTT 音频协议，分帧发送
3. **全双工交互**: 支持边说边听，服务端响应音频本地播放
4. **低延迟**: 端到端延迟 < 500ms（网络良好情况下）

### 1.3 为什么使用 C 语言

`example/mock` 目录用于模拟真实硬件设备行为。在嵌入式/硬件设备场景中，底层音频采集通常由 C/C++ 实现：

- **更贴近硬件**: C 语言直接调用 ALSA/PulseAudio/CoreAudio，无中间层
- **性能更优**: 无 GC 暂停、无 CGO 调用开销，延迟更低
- **跨平台一致**: 真实设备端通常运行嵌入式 Linux，C 方案可直接复用
- **资源占用低**: 内存占用更小，适合资源受限场景

---

## 2. 系统架构

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           本地开发环境 (Local PC)                            │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                    C 语言音频采集程序 (独立进程)                      │   │
│  │  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐              │   │
│  │  │  麦克风采集  │───>│  音频预处理  │───>│  MQTT 发送   │──────────────┼──>│
│  │  │  (C +       │    │  (重采样/   │    │  (分帧/封包)│              │   │
│  │  │   PortAudio)│    │   编码/分帧) │    │             │              │   │
│  │  └─────────────┘    └─────────────┘    └─────────────┘              │   │
│  │                                                                     │   │
│  │  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐              │   │
│  │  │  音频播放   │<───│  TTS 响应    │<───│  MQTT 接收   │<─────────────┼──<│
│  │  │  (C +       │    │  (音频播放)  │    │  (解包/重组)│              │   │
│  │  │   PortAudio)│    │             │    │             │              │   │
│  │  └─────────────┘    └─────────────┘    └─────────────┘              │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                     │                                       │
│                                     │ stdout/stdin 或 Unix Socket           │
│                                     ▼                                       │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                    Go 控制程序 (可选)                                 │   │
│  │  ┌─────────────┐    ┌─────────────┐                                │   │
│  │  │  会话管理   │───>│  状态监控   │                                │   │
│  │  │  (FSM)      │    │  /日志      │                                │   │
│  │  └─────────────┘    └─────────────┘                                │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                  │          │
│                                                                  ▼          │
│                              网络传输 (TCP/TLS)                            │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    │ MQTT over TCP
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         Yunyez 后端服务 (Go)                                 │
│                                                                             │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐  │
│  │  MQTT Broker│───>│  语音服务   │───>│  ASR 服务    │───>│  NLU 服务    │  │
│  │  (EMQX)     │    │  (消息路由) │    │  (语音转文字)│    │  (意图识别) │  │
│  └─────────────┘    └─────────────┘    └─────────────┘    └──────┬──────┘  │
│                                                                  │          │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌──────▼──────┐  │
│  │  MQTT Broker│<───│  TTS 服务    │<───│  LLM Chat   │<───│  LLM 服务    │  │
│  │  (EMQX)     │    │  (文字转语音)│    │  (对话生成) │    │  (通义千问) │  │
│  └─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘  │
└─────────────────────────────────────────────────────────────────────────────┘
```

**架构说明**:

1. **C 语言音频程序** (`voice_capture.c`): 独立进程，负责音频采集、预处理、MQTT 通信
2. **Go 控制程序** (可选): 用于会话管理、状态监控，可通过 stdin/stdout 或 Unix Socket 与 C 程序通信
3. **简化方案**: 初期可仅使用 C 程序，Go 控制程序作为后续扩展

---

## 3. 核心模块设计

### 3.1 音频采集模块 (Audio Capture)

**职责**: 从本地麦克风采集原始音频数据

**技术选型**: 使用 **C 语言 + PortAudio** 实现

| 方案 | 优点 | 缺点 | 推荐度 |
|------|------|------|--------|
| **C + PortAudio** | 跨平台、成熟稳定、无 CGO 开销、延迟最低 | 需要编译 | ⭐⭐⭐⭐⭐ |
| Go + PortAudio | Go 语法友好 | CGO 调用开销、GC 暂停 | ⭐⭐⭐ |
| C + ALSA (Linux only) | 原生 Linux 支持、无额外依赖 | 不跨平台 | ⭐⭐⭐⭐ |

**采集参数**:
```c
typedef struct {
    int sample_rate;      // 采样率：16000 Hz (与后端 ASR 对齐)
    int channel_count;    // 通道数：1 (单声道)
    int bits_per_sample;  // 位深：16 (16-bit PCM)
    int frame_duration;   // 每帧时长：20ms (640 字节@16k/16bit/mono)
    int buffer_frames;    // 缓冲区大小：4 帧 (80ms 防抖动)
} capture_config_t;
```

**数据流**:
```
麦克风 → PortAudio 回调 → 环形缓冲区 → 音频预处理 → MQTT 发送
```

**PortAudio 初始化示例**:
```c
#include <portaudio.h>

#define SAMPLE_RATE      (16000)
#define FRAMES_PER_BUFFER (320)  // 20ms @ 16kHz
#define NUM_CHANNELS     (1)

static PaStream *g_stream = NULL;

static int audio_callback(const void *inputBuffer, void *outputBuffer,
                          unsigned long framesPerBuffer,
                          const PaStreamCallbackTimeInfo* timeInfo,
                          PaStreamCallbackFlags statusFlags,
                          void *userData) {
    // 将 inputBuffer 中的 PCM 数据写入环形缓冲区
    ring_buffer_write((int16_t*)inputBuffer, framesPerBuffer * NUM_CHANNELS);
    return paContinue;
}

int audio_capture_init() {
    PaError err = Pa_Initialize();
    if (err != paNoError) return -1;

    err = Pa_OpenDefaultStream(&g_stream,
                               NUM_CHANNELS,   // input channels
                               0,              // output channels
                               paInt16,        // sample format
                               SAMPLE_RATE,
                               FRAMES_PER_BUFFER,
                               audio_callback,
                               NULL);
    if (err != paNoError) return -1;

    err = Pa_StartStream(g_stream);
    if (err != paNoError) return -1;

    return 0;
}
```

---

### 3.2 音频预处理模块 (Audio Preprocessing)

**职责**: 对原始音频进行重采样、编码、分帧

**处理流程**:
```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  原始 PCM   │───>│  重采样     │───>│  VAD 检测    │───>│  分帧/打包  │
│  (48k/44.1k)│    │  (→16k)     │    │  (静音过滤) │    │  (20ms/帧)  │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
```

**关键处理**:

1. **重采样 (Resampling)**
   - 输入：麦克风原始采样率 (通常 44.1k/48k)
   - 输出：16000 Hz (与后端 ASR 对齐)
   - 算法：线性插值 或 libsamplerate (C 库)

2. **VAD (Voice Activity Detection)**
   - 目的：过滤静音，减少无效传输
   - 实现：能量阈值法 或 WebRTC VAD (C 接口)
   - 策略：连续 3 帧低于阈值 → 判定静音

3. **分帧 (Framing)**
   - 帧大小：20ms × 16000Hz × 16bit × 1ch = 640 字节
   - 帧类型：
     - `VOICE_FRAME_FRAGMENT` (2): 中间帧
     - `VOICE_FRAME_LAST` (3): 最后一帧

**C 语言实现示例**:
```c
#include <stdint.h>
#include <string.h>

#define SAMPLE_RATE       (16000)
#define FRAME_DURATION_MS (20)
#define FRAME_SIZE        (SAMPLE_RATE * FRAME_DURATION_MS / 1000 * 1 * 2)  // 640 bytes

typedef enum {
    VOICE_FRAME_UNKNOWN = 0,
    VOICE_FRAME_FIRST = 1,
    VOICE_FRAME_FRAGMENT = 2,
    VOICE_FRAME_LAST = 3,
} voice_frame_type_t;

// 简单的能量阈值 VAD
int vad_detect(int16_t *samples, int num_samples, int16_t threshold) {
    int32_t energy = 0;
    for (int i = 0; i < num_samples; i++) {
        energy += abs(samples[i]);
    }
    energy /= num_samples;
    return (energy > threshold) ? 1 : 0;  // 1=有声，0=静音
}

// 线性插值重采样 (48k → 16k)
int resample_48k_to_16k(int16_t *dst, const int16_t *src, int src_len) {
    // 3:1 降采样，简单抽取
    int dst_len = src_len / 3;
    for (int i = 0; i < dst_len; i++) {
        dst[i] = src[i * 3];
    }
    return dst_len;
}
```

---

### 3.3 MQTT 传输模块 (MQTT Transport)

**职责**: 按照 Yunyez 自定义协议封装音频帧并发送

**协议头结构** (12 字节):
```
┌───────────────────────────────────────────────────────────────────────────┐
│  96 bits = 12 bytes                                                       │
│                                                                           │
│  Byte 0:  [Version:4][AudioFormat high 4]                                 │
│  Byte 1:  [AudioFormat low 4][SampleRate high 4]                          │
│  Byte 2:  [SampleRate mid 8]                                              │
│  Byte 3:  [SampleRate low 4][Ch:2][F:2]                                   │
│  Byte 4-5: [FrameSeq:16]                                                  │
│  Byte 6-7: [Timestamp:16]                                                 │
│  Byte 8-9: [PayloadLen:16]                                                │
│  Byte 10-11: [CRC16:16]                                                   │
└───────────────────────────────────────────────────────────────────────────┘
```

**C 语言实现示例**:
```c
#include <paho.mqtt.client.h>
#include <arpa/inet.h>

#define VOICE_VERSION        (1)
#define VOICE_AUDIO_FORMAT_WAV (1)
#define VOICE_SAMPLE_RATE    (16000)
#define VOICE_CH             (1)

typedef struct {
    uint8_t  version_audio_fmt;    // [Version:4][AudioFormat high 4]
    uint8_t  audio_fmt_sample_hi;  // [AudioFormat low 4][SampleRate high 4]
    uint8_t  sample_mid;           // [SampleRate mid 8]
    uint8_t  sample_lo_ch_f;       // [SampleRate low 4][Ch:2][F:2]
    uint16_t frame_seq;            // [FrameSeq:16]
    uint16_t timestamp;            // [Timestamp:16]
    uint16_t payload_len;          // [PayloadLen:16]
    uint16_t crc16;                // [CRC16:16]
} __attribute__((packed)) voice_header_t;

// 构建协议头
void build_voice_header(voice_header_t *hdr, uint16_t frame_seq, 
                        uint16_t payload_len, uint8_t frame_type) {
    hdr->version_audio_fmt = (VOICE_VERSION << 4) | ((VOICE_AUDIO_FORMAT_WAV >> 4) & 0x0F);
    hdr->audio_fmt_sample_hi = ((VOICE_AUDIO_FORMAT_WAV & 0x0F) << 4) | ((VOICE_SAMPLE_RATE >> 8) & 0x0F);
    hdr->sample_mid = (VOICE_SAMPLE_RATE >> 0) & 0xFF;
    hdr->sample_lo_ch_f = ((VOICE_SAMPLE_RATE & 0x0F) << 4) | ((VOICE_CH & 0x03) << 2) | (frame_type & 0x03);
    hdr->frame_seq = htons(frame_seq);
    hdr->timestamp = htons((uint16_t)(time(NULL) & 0xFFFF));
    hdr->payload_len = htons(payload_len);
    // CRC16 需要单独计算
    hdr->crc16 = 0;  // TODO: 计算 CRC16
}

// MQTT 发布
int mqtt_publish_voice(MQTTClient client, const char *topic, 
                       voice_header_t *hdr, uint8_t *payload, int payload_len) {
    MQTTClient_message pubmsg = MQTTClient_message_initializer;
    pubmsg.payloadlen = sizeof(voice_header_t) + payload_len;
    pubmsg.payload = malloc(pubmsg.payloadlen);
    memcpy(pubmsg.payload, hdr, sizeof(voice_header_t));
    memcpy((uint8_t*)pubmsg.payload + sizeof(voice_header_t), payload, payload_len);
    pubmsg.qos = 1;
    pubmsg.retained = 0;
    
    return MQTTClient_publishMessage(client, topic, &pubmsg, NULL);
}
```

**QoS 选择**:
- **QoS 1** (至少一次): 推荐，平衡可靠性与延迟
- **QoS 0** (最多一次): 低延迟，但可能丢帧
- **QoS 2** (只有一次): 高可靠，但延迟较高

---

### 3.4 音频播放模块 (Audio Playback)

**职责**: 接收服务端 TTS 响应音频并播放

**播放流程**:
```
MQTT 接收 → 协议头解析 → 音频重组 → PortAudio 播放 → 扬声器
```

**关键设计**:

1. **缓冲策略**:
   - 预缓冲：累积 3-5 帧后开始播放 (60-100ms)
   - 动态缓冲：根据网络抖动调整

2. **播放参数**:
   ```c
   typedef struct {
       int sample_rate;      // 16000 Hz
       int channel_count;    // 1 (单声道)
       int bits_per_sample;  // 16 (16-bit PCM)
       int buffer_frames;    // 预缓冲帧数：5 帧 (100ms)
   } playback_config_t;
   ```

3. **打断机制**:
   - 用户再次说话时，停止当前播放
   - 通过 VAD 检测用户语音活动

**PortAudio 播放示例**:
```c
#include <portaudio.h>

static PaStream *g_playback_stream = NULL;
static ring_buffer_t *g_playback_buffer = NULL;

static int playback_callback(const void *inputBuffer, void *outputBuffer,
                             unsigned long framesPerBuffer,
                             const PaStreamCallbackTimeInfo* timeInfo,
                             PaStreamCallbackFlags statusFlags,
                             void *userData) {
    // 从环形缓冲区读取音频数据
    int16_t *out = (int16_t*)outputBuffer;
    int samples_read = ring_buffer_read(out, framesPerBuffer);
    
    // 如果缓冲区为空，填充静音
    if (samples_read < framesPerBuffer) {
        memset(out + samples_read, 0, (framesPerBuffer - samples_read) * sizeof(int16_t));
    }
    
    return paContinue;
}

int audio_playback_init() {
    PaError err = Pa_OpenDefaultStream(&g_playback_stream,
                                       0,              // input channels
                                       1,              // output channels
                                       paInt16,        // sample format
                                       16000,          // sample rate
                                       256,            // frames per buffer
                                       playback_callback,
                                       NULL);
    if (err != paNoError) return -1;
    
    err = Pa_StartStream(g_playback_stream);
    if (err != paNoError) return -1;
    
    return 0;
}

// 接收 TTS 音频数据
void playback_write(int16_t *samples, int num_samples) {
    ring_buffer_write(g_playback_buffer, samples, num_samples);
}
```

---

## 4. 目录结构

```
example/mock/virtual_voice_realtime/
├── README.md                    # 使用说明
├── Makefile                     # 编译脚本
├── config.yaml                  # 配置文件
│
├── src/                         # C 语言源代码
│   ├── main.c                   # 程序入口
│   ├── audio_capture.c          # PortAudio 采集封装
│   ├── audio_capture.h          # 采集头文件
│   ├── audio_playback.c         # PortAudio 播放封装
│   ├── audio_playback.h         # 播放头文件
│   ├── ring_buffer.c            # 环形缓冲区
│   ├── ring_buffer.h            # 缓冲区头文件
│   ├── vad.c                    # VAD 静音检测
│   ├── vad.h                    # VAD 头文件
│   ├── mqtt_client.c            # MQTT 客户端封装
│   ├── mqtt_client.h            # MQTT 头文件
│   ├── protocol.c               # 协议封装
│   ├── protocol.h               # 协议头文件
│   └── session.c                # 会话状态管理
│       └── session.h            # 会话头文件
│
├── include/                     # 公共头文件
│   ├── config.h                 # 配置宏定义
│   └── types.h                  # 类型定义
│
├── build/                       # 编译输出
│   ├── voice_capture            # Linux 可执行文件
│   ├── voice_capture.exe        # Windows 可执行文件
│   └── voice_capture.dylib      # macOS 动态库
│
└── scripts/                     # 辅助脚本
    ├── build.sh                 # Linux/macOS 编译脚本
    ├── build.bat                # Windows 编译脚本
    └── list_devices.c           # 列出可用音频设备
```

**编译说明**:

```bash
# Linux/macOS
cd example/mock/virtual_voice_realtime
make

# Windows (使用 MinGW)
build.bat
```

---

## 5. 状态机设计

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        语音会话状态机 (FSM)                                  │
│                                                                             │
│                    ┌─────────────┐                                          │
│         (开始) ───>│  IDLE       │                                          │
│                    │  空闲状态   │                                          │
│                    └──────┬──────┘                                          │
│                           │ 按下说话键 / 检测到语音                          │
│                           ▼                                                  │
│                    ┌─────────────┐                                          │
│                    │  LISTENING  │                                          │
│                    │  聆听状态   │──┐                                       │
│                    └──────┬──────┘  │                                       │
│                           │          │ 检测到静音超时                        │
│                           │ ASR 识别完成                                     │
│                           ▼          │                                       │
│                    ┌─────────────┐  │                                       │
│                    │  PROCESSING │◄─┘                                       │
│                    │  处理状态   │                                          │
│                    └──────┬──────┘                                          │
│                           │                                                  │
│                           │ LLM 响应生成                                     │
│                           ▼                                                  │
│                    ┌─────────────┐                                          │
│                    │  SPEAKING   │                                          │
│                    │  说话状态   │                                          │
│                    └──────┬──────┘                                          │
│                           │                                                  │
│                           │ TTS 播放完成 / 用户打断                          │
│                           ▼                                                  │
│                    ┌─────────────┐                                          │
│                    │  IDLE       │                                          │
│                    └─────────────┘                                          │
└─────────────────────────────────────────────────────────────────────────────┘
```

**状态说明**:

| 状态 | 描述 | 采集 | 播放 | MQTT 上行 | MQTT 下行 |
|------|------|------|------|-----------|-----------|
| IDLE | 空闲等待 | ❌ | ❌ | ❌ | ✅ |
| LISTENING | 正在录音 | ✅ | ❌ | ✅ | ✅ |
| PROCESSING | 服务端处理中 | ❌ | ❌ | ❌ | ✅ |
| SPEAKING | TTS 播放中 | ⚠️ (VAD 监听打断) | ✅ | ❌ | ✅ |

---

## 6. 配置设计

### 6.1 C 语言配置头 (config.h)

```c
#ifndef __VOICE_CONFIG_H__
#define __VOICE_CONFIG_H__

// 音频采集配置
#define CAPTURE_SAMPLE_RATE      (16000)      // 目标采样率 (重采样后)
#define CAPTURE_CHANNELS         (1)          // 通道数 (单声道)
#define CAPTURE_FRAME_DURATION_MS (20)        // 每帧时长 (毫秒)
#define CAPTURE_BUFFER_FRAMES    (4)          // 缓冲帧数

// VAD 配置
#define VAD_ENABLED              (1)          // 是否启用 VAD
#define VAD_THRESHOLD_DB         (-45)        // 能量阈值 (dB)
#define VAD_SILENCE_FRAMES       (3)          // 连续静音帧数判定静音

// MQTT 配置
#define MQTT_BROKER              "tcp://127.0.0.1:1883"
#define MQTT_USERNAME            "root"
#define MQTT_PASSWORD            "root123"
#define MQTT_CLIENT_ID           "voice_realtime_001"
#define MQTT_QOS                 (1)
#define MQTT_VENDOR              "test"
#define MQTT_DEVICE_TYPE         "T0001"
#define MQTT_DEVICE_SN           "A0001"

// 播放配置
#define PLAYBACK_SAMPLE_RATE     (16000)
#define PLAYBACK_CHANNELS        (1)
#define PLAYBACK_BUFFER_FRAMES   (5)         // 预缓冲帧数

// 日志配置
#define LOG_LEVEL                (LOG_DEBUG)

#endif
```

### 6.2 YAML 配置文件 (config.yaml, 用于 Go 控制程序)

```yaml
# 音频采集配置
capture:
  device_id: -1              # -1=默认设备，或指定设备 ID
  sample_rate: 16000         # 目标采样率 (重采样后)
  channels: 1                # 通道数 (单声道)
  frame_duration_ms: 20      # 每帧时长 (毫秒)
  buffer_frames: 4           # 缓冲帧数

# VAD 配置
vad:
  enabled: true              # 是否启用 VAD
  threshold_db: -45          # 能量阈值 (dB)
  silence_frames: 3          # 连续静音帧数判定静音

# MQTT 配置
mqtt:
  broker: "tcp://127.0.0.1:1883"
  username: "root"
  password: "root123"
  client_id: "voice_realtime_001"
  qos: 1
  topic:
    vendor: "test"
    device_type: "T0001"
    device_sn: "A0001"

# 播放配置
playback:
  device_id: -1
  sample_rate: 16000
  channels: 1
  buffer_frames: 5           # 预缓冲帧数
  auto_gain: true            # 自动增益

# 日志配置
logger:
  level: "debug"
  file: "storage/logs/voice_realtime.log"
```

---

## 7. 依赖管理

### 7.1 C 语言依赖库

| 库 | 功能 | 安装命令 |
|------|------|----------|
| **PortAudio** | 音频采集/播放 | `apt-get install portaudio19-dev` (Ubuntu) |
| **Paho MQTT C** | MQTT 客户端 | `apt-get install libpaho-mqtt-dev` (Ubuntu) |
| **cJSON** | JSON 解析 (配置解析) | `apt-get install libcjson-dev` (Ubuntu) |

**可选依赖**:
| 库 | 功能 | 安装命令 |
|------|------|----------|
| **libsamplerate** | 高质量重采样 | `apt-get install libsamplerate0-dev` |
| **WebRTC VAD** | VAD 静音检测 | 需手动编译 |

### 7.2 系统依赖

**Linux (Ubuntu/Debian)**:
```bash
# 基础依赖
sudo apt-get install portaudio19-dev libpaho-mqtt-dev libcjson-dev

# 可选：高质量重采样
sudo apt-get install libsamplerate0-dev
```

**macOS**:
```bash
brew install portaudio paho-mqtt-c cJSON
```

**Windows**:
- 使用 MSYS2:
```bash
pacman -S mingw-w64-x86_64-portaudio mingw-w64-x86_64-paho-mqtt mingw-w64-x86_64-cjson
```
- 或下载预编译库:
  - PortAudio: https://www.portaudio.com/download.html
  - Paho MQTT: https://github.com/eclipse/paho.mqtt.c/releases

### 7.3 Makefile 示例

```makefile
CC = gcc
CFLAGS = -Wall -Wextra -O2 -I./include -I./src
LDFLAGS = -lportaudio -lpaho-mqtt3c -lcjson -lpthread

TARGET = voice_capture
SRCS = src/main.c src/audio_capture.c src/audio_playback.c \
       src/ring_buffer.c src/vad.c src/mqtt_client.c \
       src/protocol.c src/session.c

OBJS = $(SRCS:.c=.o)

all: $(TARGET)

$(TARGET): $(OBJS)
	$(CC) -o $@ $^ $(LDFLAGS)

%.o: %.c
	$(CC) $(CFLAGS) -c -o $@ $<

clean:
	rm -f $(OBJS) $(TARGET)

.PHONY: all clean
```

---

## 8. API 设计

### 8.1 核心接口 (C 语言)

```c
// 会话状态
typedef enum {
    SESSION_IDLE,        // 空闲状态
    SESSION_LISTENING,   // 聆听状态 (录音中)
    SESSION_PROCESSING,  // 处理状态 (等待响应)
    SESSION_SPEAKING,    // 说话状态 (播放 TTS)
} session_state_t;

// 事件回调
typedef struct {
    void (*on_state_changed)(session_state_t old_state, session_state_t new_state);
    void (*on_asr_result)(const char *text);
    void (*on_tts_audio)(const int16_t *samples, int num_samples);
    void (*on_error)(int error_code, const char *message);
} event_callbacks_t;

// 会话配置
typedef struct {
    const char *mqtt_broker;
    const char *mqtt_username;
    const char *mqtt_password;
    const char *mqtt_client_id;
    const char *device_vendor;
    const char *device_type;
    const char *device_sn;
    event_callbacks_t callbacks;
} voice_session_config_t;

// 会话句柄
typedef struct voice_session voice_session_t;

// 创建会话
voice_session_t* voice_session_create(const voice_session_config_t *config);

// 启动会话
int voice_session_start(voice_session_t *session);

// 停止会话
int voice_session_stop(voice_session_t *session);

// 销毁会话
void voice_session_destroy(voice_session_t *session);

// 获取状态
session_state_t voice_session_get_state(voice_session_t *session);

// 发送音频帧 (由采集模块调用)
int voice_session_send_audio(voice_session_t *session, const int16_t *samples, int num_samples);
```

### 8.2 使用示例

```c
#include "session.h"
#include <stdio.h>
#include <signal.h>

static voice_session_t *g_session = NULL;

void on_state_changed(session_state_t old_state, session_state_t new_state) {
    printf("State: %d -> %d\n", old_state, new_state);
}

void on_asr_result(const char *text) {
    printf("ASR: %s\n", text);
}

void on_tts_audio(const int16_t *samples, int num_samples) {
    // 自动播放，无需处理
}

void on_error(int error_code, const char *message) {
    printf("Error %d: %s\n", error_code, message);
}

void signal_handler(int sig) {
    if (g_session) {
        voice_session_stop(g_session);
        voice_session_destroy(g_session);
    }
}

int main(int argc, char *argv[]) {
    signal(SIGINT, signal_handler);
    
    voice_session_config_t config = {
        .mqtt_broker = "tcp://127.0.0.1:1883",
        .mqtt_username = "root",
        .mqtt_password = "root123",
        .mqtt_client_id = "voice_realtime_001",
        .device_vendor = "test",
        .device_type = "T0001",
        .device_sn = "A0001",
        .callbacks = {
            .on_state_changed = on_state_changed,
            .on_asr_result = on_asr_result,
            .on_tts_audio = on_tts_audio,
            .on_error = on_error,
        },
    };
    
    g_session = voice_session_create(&config);
    if (!g_session) {
        fprintf(stderr, "Failed to create session\n");
        return 1;
    }
    
    if (voice_session_start(g_session) != 0) {
        fprintf(stderr, "Failed to start session\n");
        voice_session_destroy(g_session);
        return 1;
    }
    
    printf("Voice session started. Press Ctrl+C to exit.\n");
    
    // 主循环 (等待事件)
    while (1) {
        usleep(100000);  // 100ms
    }
    
    return 0;
}
```

---

## 9. 性能指标

| 指标 | 目标值 | 说明 |
|------|--------|------|
| 端到端延迟 | < 500ms | 从说话到听到 TTS 响应 |
| ASR 延迟 | < 200ms | 语音→文字的识别延迟 |
| 网络带宽 | ~64kbps | 16k/16bit/mono 原始 PCM |
| CPU 占用 | < 5% | 现代 CPU 单核占用 |
| 内存占用 | < 50MB | 包含缓冲区 |

---

## 10. 测试计划

### 10.1 单元测试

- [ ] `capture/`: 麦克风采集测试
- [ ] `preprocess/resample.go`: 重采样质量测试
- [ ] `preprocess/vad.go`: VAD 准确率测试
- [ ] `transport/protocol.go`: 协议封装测试

### 10.2 集成测试

- [ ] 端到端延迟测试
- [ ] 网络抖动模拟测试
- [ ] 长会话稳定性测试 (1 小时+)

### 10.3 用户体验测试

- [ ] 打断响应测试
- [ ] 方言/口音识别测试
- [ ] 背景噪音环境测试

---

## 11. 后续扩展

### 11.1 短期优化

1. **蓝牙耳麦支持**: 自动切换输入/输出设备
2. **回声消除 (AEC)**: 防止 TTS 播放被误识别
3. **自动增益 (AGC)**: 适应不同音量环境

### 11.2 长期规划

1. **WebSocket 传输**: 替代 MQTT，降低延迟
2. **离线 ASR**: 本地语音识别 (Vosk/Sherpa)
3. **多设备协同**: 手机 + 设备端同步测试

---

## 12. 相关文件

| 文件 | 描述 |
|------|------|
| `internal/pkg/mqtt/protocol/voice/voice.go` | 音频协议头定义 |
| `internal/pkg/mqtt/protocol/voice/message.go` | 音频消息构建 |
| `internal/pkg/mqtt/core/client.go` | MQTT 客户端封装 |
| `internal/service/voice/handler/chat.go` | 语音处理流水线 |
| `docs/protocol.md` | 通信协议规范 |

---

## 13. 附录：PortAudio C API 快速入门

### 13.1 采集示例

```c
#include <portaudio.h>
#include <stdio.h>

#define SAMPLE_RATE (16000)
#define FRAMES_PER_BUFFER (256)
#define NUM_CHANNELS (1)

static int audio_callback(const void *inputBuffer, void *outputBuffer,
                          unsigned long framesPerBuffer,
                          const PaStreamCallbackTimeInfo* timeInfo,
                          PaStreamCallbackFlags statusFlags,
                          void *userData) {
    const int16_t *in = (const int16_t*)inputBuffer;
    
    // 处理音频数据 (例如：写入环形缓冲区)
    for (unsigned long i = 0; i < framesPerBuffer; i++) {
        // process in[i]
    }
    
    return paContinue;
}

int main() {
    PaStream *stream;
    PaError err;
    
    err = Pa_Initialize();
    if (err != paNoError) {
        fprintf(stderr, "PortAudio init failed: %s\n", Pa_GetErrorText(err));
        return 1;
    }
    
    // 枚举设备
    int num_devices = Pa_GetDeviceCount();
    for (int i = 0; i < num_devices; i++) {
        const PaDeviceInfo *info = Pa_GetDeviceInfo(i);
        printf("Device %d: %s (in: %d, out: %d)\n",
               i, info->name,
               info->maxInputChannels,
               info->maxOutputChannels);
    }
    
    // 打开输入流
    err = Pa_OpenDefaultStream(&stream,
                               NUM_CHANNELS,   // input channels
                               0,              // output channels
                               paInt16,        // sample format
                               SAMPLE_RATE,
                               FRAMES_PER_BUFFER,
                               audio_callback,
                               NULL);
    if (err != paNoError) {
        fprintf(stderr, "Open stream failed: %s\n", Pa_GetErrorText(err));
        Pa_Terminate();
        return 1;
    }
    
    err = Pa_StartStream(stream);
    if (err != paNoError) {
        fprintf(stderr, "Start stream failed: %s\n", Pa_GetErrorText(err));
        Pa_CloseStream(stream);
        Pa_Terminate();
        return 1;
    }
    
    printf("Recording... Press Enter to stop.\n");
    getchar();
    
    err = Pa_StopStream(stream);
    Pa_CloseStream(stream);
    Pa_Terminate();
    
    return 0;
}
```

### 13.2 播放示例

```c
#include <portaudio.h>

static int playback_callback(const void *inputBuffer, void *outputBuffer,
                             unsigned long framesPerBuffer,
                             const PaStreamCallbackTimeInfo* timeInfo,
                             PaStreamCallbackFlags statusFlags,
                             void *userData) {
    int16_t *out = (int16_t*)outputBuffer;
    
    // 从缓冲区读取音频数据
    // 如果缓冲区为空，填充静音
    memset(out, 0, framesPerBuffer * sizeof(int16_t));
    
    return paContinue;
}

int audio_playback_init() {
    PaStream *stream;
    PaError err;
    
    err = Pa_OpenDefaultStream(&stream,
                               0,              // input channels
                               1,              // output channels
                               paInt16,
                               16000,
                               256,
                               playback_callback,
                               NULL);
    if (err != paNoError) return -1;
    
    err = Pa_StartStream(stream);
    if (err != paNoError) return -1;
    
    return 0;
}
```

### 13.3 注意事项

1. **编译链接**: 需要链接 PortAudio 库 `-lportaudio`
2. **权限**: 
   - macOS 需要麦克风权限 (Info.plist + 用户授权)
   - Linux 需要将用户加入 `audio` 组：`sudo usermod -aG audio $USER`
3. **采样率匹配**: 确保设备支持目标采样率，否则 PortAudio 可能自动重采样
4. **回调时间**: 回调函数应尽量短，避免阻塞，耗时操作放到工作线程
