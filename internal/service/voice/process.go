// Package voice provides voice processing services.
package voice

import (
	"context"
	"fmt"
	"path/filepath"
	config "yunyez/internal/common/config"
	tools "yunyez/internal/common/tools"
	logger "yunyez/internal/pkg/logger"
	mqtt_constant "yunyez/internal/pkg/mqtt/common"
	mqtt_voice "yunyez/internal/pkg/mqtt/protocol/voice"
	nlu "yunyez/internal/pkg/agent/nlu"
)

var (
	audioStorage = config.GetString("audio.storage") // 音频临时存储目录
	asrClient    = NewLocalASRClient()               // ASR 语音识别客户端 - 本地模型
	nluClient    = nlu.NewClient()                   // NLU 自然语言理解客户端 - 本地模型
	fragmentMgr  = NewFragmentManager()              // 分片帧管理
)

// ProcessFull 处理完整帧
// 参数：
//   - ctx: 上下文对象
//   - clientID: 客户端ID(设备序列号)
//   - header: 音频消息头信息
//   - payload: 音频消息 payload 字节切片
//
// 返回值:
//   - error: 处理过程中遇到的错误，若成功则为 nil
func ProcessFull(ctx context.Context, clientID string, header *mqtt_voice.Header, payload []byte) error {
	
	logger.Info(ctx, "ProcessFull", map[string]any{
		"clientID": clientID,
		"header":   header,
	})
	
	// 暂存完整帧
	ext := mqtt_constant.AudioFormatString(header.AudioFormat)
	// example: storage/tmp/audio/[device_sn]/1694567890_0.wav
	audioPath := filepath.Join(audioStorage, clientID,
		fmt.Sprintf("%d_%d.%s", header.Timestamp, header.FrameSeq, ext))
	ok, err := tools.WriteFile(audioPath, payload)
	if err != nil {
		logger.Error(ctx, "write audio file failed", map[string]interface{}{
			"error": err.Error(),
			"path":  audioPath,
		})
		return fmt.Errorf("write audio file failed: %w", err)
	} else if !ok {
		return fmt.Errorf("write audio file failed: file exists")
	}

	// asr 识别
	text, err := asrClient.Transfer(ctx, payload)
	if err != nil {
		return fmt.Errorf("asr transfer failed: %w", err)
	}

	// nlu 理解
	input := &nlu.Input{
		Text: text,
	}
	intent, err := nluClient.Predict(input)
	if err != nil {
		return fmt.Errorf("nlu predict failed: %w", err)
	}
	// TODO: response to mqtt
	fmt.Printf("[ASR] client=%s, text=%s, intent=%+v\n", clientID, text, intent)

	return nil
}

func ProcessFragment(ctx context.Context, clientID string, header *mqtt_voice.Header, payload []byte) error {
	// 暂存分片帧
	// 检查是否为最后一帧
	// 合并分片帧
	// 合并音频数据
	// 处理完整帧

	return nil
}
