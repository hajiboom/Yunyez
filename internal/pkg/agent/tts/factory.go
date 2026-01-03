// Package tts provides text-to-speech services
package tts

import (
	"context"
	"fmt"
	"yunyez/internal/common/config"
	"yunyez/internal/common/constant"
	"yunyez/internal/pkg/logger"
)

// NewTTSClientFromConfig creates a new TTS client based on the configuration.
func NewTTSClientFromConfig(configReader ConfigReader) (Service, error) {
	model := configReader.GetString("tts.model")
	switch model {
		case constant.ModelChatTTS: // ChatTTS
			return NewChatTTS(ChatTTSConfig{
				Endpoint:    configReader.GetString("tts.chat.endpoint"),
				Voice:       configReader.GetString("tts.chat.params.voice"),
				Rate:        configReader.GetString("tts.chat.params.rate"),
				Pitch:       configReader.GetString("tts.chat.params.pitch"),
				Volume:      configReader.GetString("tts.chat.params.volume"),
				Temperature: configReader.GetString("tts.chat.params.temperature"),
			}), nil
		case constant.ModelEdgeTTS: // EdgeTTS
			return NewEdgeTTS(EdgeTTSConfig{
				Endpoint: configReader.GetString("tts.edge.endpoint"),
				Voice: configReader.GetString("tts.edge.params.voice"),
				Rate:    configReader.GetString("tts.edge.params.rate"),
				Pitch:   configReader.GetString("tts.edge.params.pitch"),
				Volume:  configReader.GetString("tts.edge.params.volume"),
			}), nil
		default:
			return nil, fmt.Errorf("unsupported TTS model: %s", model)
	}
}

// NewTTSClient creates a new TTS client based on the global configuration.
func NewTTSClient() Service {
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
