package common

const (
	VOICE_VERSION        = 0x01 // 版本号 1.0

	// 音频帧类型
	VOICE_FRAME_FULL     = 0x01 // 完整帧
	VOICE_FRAME_FRAGMENT = 0x02 // 分片帧
	VOICE_FRAME_LAST     = 0x03 // 最后一帧

	// 音频格式
	VOICE_AUDIO_FORMAT_PCM   = 0x01 // pcm
	VOICE_AUDIO_FORMAT_AAC   = 0x02 // aac
	VOICE_AUDIO_FORMAT_OPUS  = 0x03 // opus
	VOICE_AUDIO_FORMAT_MP3   = 0x04 // mp3
	VOICE_AUDIO_FORMAT_G711A = 0x05 // g711a
	VOICE_AUDIO_FORMAT_G711U = 0x06 // g711u
	VOICE_AUDIO_FORMAT_WAV   = 0x07 // wav
)
