// Package common mqtt 常量定义
package common

const (
	VoiceVersion = 0x01 // 版本号 1.0

	// 音频帧类型
	VoiceFrameFull     = 0x01 // 完整帧
	VoiceFrameFragment = 0x02 // 分片帧
	VoiceFrameLast     = 0x03 // 最后一帧

	// 音频格式
	VoiceAudioFormatPcm   = 0x01 // pcm
	VoiceAudioFormatAac   = 0x02 // aac
	VoiceAudioFormatOpus  = 0x03 // opus
	VoiceAudioFormatMp3   = 0x04 // mp3
	VoiceAudioFormatG711A = 0x05 // g711a
	VoiceAudioFormatG711U = 0x06 // g711u
	VoiceAudioFormatWav   = 0x07 // wav
)

// AudioFormatString 获取音频格式
func AudioFormatString(format uint8) string {
	switch format {
	case VoiceAudioFormatPcm:
		return "pcm"
	case VoiceAudioFormatAac:
		return "aac"
	case VoiceAudioFormatOpus:
		return "opus"
	case VoiceAudioFormatMp3:
		return "mp3"
	case VoiceAudioFormatG711A:
		return "g711a"
	case VoiceAudioFormatG711U:
		return "g711u"
	case VoiceAudioFormatWav:
		return "wav"
	default:
		return "unknown"
	}
}