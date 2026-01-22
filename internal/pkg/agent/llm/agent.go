// Package llm natural language model agent
package llm

import (
	"context"
	constant "yunyez/internal/common/constant"
	qwen "yunyez/internal/pkg/agent/llm/qwen"
	metering "yunyez/internal/pkg/agent/metering"
)

// Agent natural language model agent interface
type Agent interface {
	Chat(ctx context.Context, clientID, message string) (<-chan string, <-chan *metering.Usage, error)
}


// Strategy natural language model agent strategy
type Strategy struct {
	Model          Agent
}

// SetAgent set the agent model to choose different service
// - qwen
// - local
func (s *Strategy) SetAgent(model string) *Strategy {
	switch model {
	case constant.ModelQwenLLM:
		s.Model = &QwenAgent{}
	default:
		s.Model = &LocalAgent{}
	}
	return s
}

// QwenAgent qwen model agent
type QwenAgent struct {
}

// Chat qwen model agent chat
func (a *QwenAgent) Chat(ctx context.Context, clientID, message string) (<-chan string, <-chan *metering.Usage, error) {
	return qwen.QwenChat(ctx, clientID, message)
}

// LocalAgent local model agent
type LocalAgent struct {
}

// Chat local model agent chat
func (a *LocalAgent) Chat(ctx context.Context, clientID, message string) (<-chan string, <-chan *metering.Usage, error) {
	return nil, nil, nil
}
