// Package tts provides text-to-speech services
package tts

import (
	"context"
	"fmt"
	"yunyez/internal/common/config"
	"yunyez/internal/common/constant"
	"yunyez/internal/pkg/logger"
)

// ConfigReader 配置读取接口
type ConfigReader interface {
	// GetString reads the TTS configuration as a string.
	GetString(key string) string
	// GetFloat64 reads the TTS configuration as a float64.
	GetFloat64(key string) float64
}

// NewTTSClientFromConfig creates a new TTS client based on the configuration.
func NewTTSClientFromConfig(configReader ConfigReader) (Service, error) {
	model := configReader.GetString("tts.model")
	protocol := configReader.GetString("tts.protocol")
	
	// 构建配置
	cfg := Config{
		Model:    model,
		Protocol: Protocol(protocol),
	}
	
	// 根据协议类型设置端点
	if protocol == "grpc" {
		cfg.GRPCEndpoint = configReader.GetString("tts.grpc_endpoint")
	} else {
		cfg.HTTPEndpoint = configReader.GetString("tts.edge.endpoint")
	}
	
	// 读取 Edge TTS 参数
	cfg.Voice = configReader.GetString("tts.edge.params.voice")
	cfg.Rate = configReader.GetString("tts.edge.params.rate")
	cfg.Pitch = configReader.GetString("tts.edge.params.pitch")
	cfg.Volume = configReader.GetString("tts.edge.params.volume")
	
	// 读取 ChatTTS 参数
	cfg.Temperature = configReader.GetString("tts.chat.params.temperature")
	
	switch model {
	case constant.ModelChatTTS: // ChatTTS
		return NewTTSClient(cfg)
	case constant.ModelEdgeTTS: // EdgeTTS
		return NewTTSClient(cfg)
	default:
		return nil, fmt.Errorf("unsupported TTS model: %s", model)
	}
}

// CreateTTSClient creates a new TTS client based on the global configuration.
// 注意：这个函数是为了兼容旧代码，推荐使用 NewTTSClientFromConfig
func CreateTTSClient() Service {
	service, err := NewTTSClientFromConfig(&globalConfig{})
	if err != nil {
		logger.Error(context.Background(), "tts.NewTTSClientFromConfig failed", map[string]any{
			"error": err.Error(),
		})
		return nil
	}
	return service
}

// globalConfig implements ConfigReader interface for global configuration.
type globalConfig struct {
}

// GetString implements ConfigReader interface.
func (g *globalConfig) GetString(key string) string {
	return config.GetString(key)
}

// GetFloat64 implements ConfigReader interface.
func (g *globalConfig) GetFloat64(key string) float64 {
	return config.GetFloat64(key)
}
