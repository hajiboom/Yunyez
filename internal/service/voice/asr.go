package voice

import (
	"context"
	mqtt_voice "yunyez/internal/pkg/mqtt/protocol/voice"
)

// ASR 语音识别接口
type ASR interface {
	// Transfer 语音识别 音频转换为文本
	Transfer(ctx context.Context, header *mqtt_voice.Header, payload []byte) error
}


// DouBaoASR 豆包语音识别
type DouBaoASR struct {}

func (d *DouBaoASR) Transfer(ctx context.Context, header *mqtt_voice.Header, payload []byte) error {
	// 调用豆包语音识别API
	return nil
}
