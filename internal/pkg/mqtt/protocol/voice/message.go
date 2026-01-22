package voice

import (
	"time"

	mqtt_common "yunyez/internal/pkg/mqtt/common"

	"github.com/sigurn/crc16"
)

var (
	// CRC16 校验表
	crc16Table = crc16.MakeTable(crc16.CRC16_MAXIM)
)

// AudioConfig 音频消息属性
type AudioConfig struct {
	AudioSampleRate uint16 // 音频采样率
	AudioChannel    uint8  // 音频通道数
	AudioFormat     uint8  // 音频格式
}

// BuildPayload 构建音频消息 payload（包含协议头）
// 参数：
//   - seq: 消息序号 从0开始递增
//   - data: 消息数据 字节切片
//   - frameType: 音频帧类型
//   - config: 语音配置[采样率、格式、声道数]
//
// 返回值:
//   - []byte: 包含协议头的音频消息 payload
func BuildPayload(seq uint16, data []byte, frameType uint8, config AudioConfig) []byte {
	header := &Header{
		Version:     mqtt_common.VoiceVersion,
		AudioFormat: config.AudioFormat,
		SampleRate:  config.AudioSampleRate,
		Ch:          config.AudioChannel,
		F:           frameType,
		FrameSeq:    seq,
		Timestamp:   uint16(time.Now().Unix() & 0xFFFF),
		PayloadLen:  uint16(len(data)),
		CRC16:       0,
	}

	headerBytes := header.Marshal()
	payload := make([]byte, len(headerBytes)+len(data))
	copy(payload, headerBytes)
	copy(payload[len(headerBytes):], data)

	// 计算CRC16校验值
	header.CRC16 = crc16.Checksum(payload[:len(headerBytes)-2], crc16Table)
	// 重新序列化头信息
	headerBytes = header.Marshal()
	copy(payload, headerBytes)

	return payload
}

// BuildFullPayload 构建完整音频消息 payload（包含协议头）
// 参数：
//   - seq: 消息序号 从0开始递增
//   - data: 消息数据 字节切片
//   - config: 语音配置[采样率、格式、声道数]
//
// 返回值:
//   - []byte: 包含协议头的完整音频消息 payload
func BuildFullPayload(seq uint16, data []byte, config AudioConfig) []byte {
	return BuildPayload(seq, data, mqtt_common.VoiceFrameFull, config)
}

// BuildStreamPayload 构建流式音频消息 payload（包含协议头）
// 参数：
//   - seq: 消息序号 从0开始递增
//   - data: 消息数据 字节切片
//   - config: 语音配置[采样率、格式、声道数]
//   - isLast: 是否为最后一帧
//
// 返回值:
//   - []byte: 包含协议头的流式音频消息 payload
func BuildStreamPayload(seq uint16, data []byte, config AudioConfig, isLast bool) []byte {
	frameType := mqtt_common.VoiceFrameFragment
	if isLast {
		frameType = mqtt_common.VoiceFrameLast
	}
	return BuildPayload(seq, data, uint8(frameType), config)
}
