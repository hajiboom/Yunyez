// Package handler provides voice processing services.
// including:
// - voice: the voice processing service
// - asr: the voice recognition service
// - nlu: the natural language understanding service
// - chat: the natural language conversation service
// - tts: the text-to-speech service
package handler

import (
	"context"
	"fmt"
	config "yunyez/internal/common/config"
	llm "yunyez/internal/pkg/agent/llm"
	_ "yunyez/internal/pkg/agent/llm/qwen"
	logger "yunyez/internal/pkg/logger"
)

var (
	agentStrategy *llm.Strategy                   // llm model strategy
	model         = config.GetString("agent.model") // default chat model
)

func init() {
	// 初始化智能体模型
	agentStrategy = &llm.Strategy{}
	agentStrategy.SetAgent(model)
}

// Chat response the natural language conversation response
// step:
// 1. call the llm model to chat
// 2. generate the audio data from the chat response by tts
// 3. send the audio data to the device by mqtt
// Parameters:
//   - ctx: the context.Context object
//   - message: the natural language message to chat
//
// Returns:
//   - string: the chat response message
//   - error: the error object if the chat failed
func Chat(ctx context.Context, message string) (string, error) {
	_, err := agentStrategy.Model.Chat(ctx, message)
	if err != nil {
		logger.Error(ctx, "agent chat failed", map[string]any{
			"error":   err.Error(),
			"message": message,
		})
		return "", fmt.Errorf("agent chat failed: %w", err)
	}

	return "", nil
}
