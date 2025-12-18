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


type LocalClient struct {
	Endpoint string `json:"endpoint"`
}


// Transfer 语音识别 音频转换为文本 -- 本地模型
func (c *LocalClient) Transfer(ctx context.Context, header *mqtt_voice.Header, payload []byte) error {
	return nil
}