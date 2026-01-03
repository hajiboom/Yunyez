// Package llm 自然语言模型
package llm

import (
	"context"
	constant "yunyez/internal/common/constant"
	qwen "yunyez/internal/pkg/agent/llm/qwen"
)

type Agent interface {
	Chat(ctx context.Context, message string) (<-chan string, error)
}

type Strategy struct {
	Model Agent
}

func (s *Strategy) SetAgent(model string) *Strategy {
	switch model {
	case constant.ModelQwenLLM:
		s.Model = &QwenAgent{}
	default:
		s.Model = &LocalAgent{}
	}
	return s
}



type QwenAgent struct {
}

func (a *QwenAgent) Chat(ctx context.Context, message string) (<-chan string, error) {
	return qwen.QwenChat(ctx, message)
}


type LocalAgent struct {
}

func (a *LocalAgent) Chat(ctx context.Context, message string) (<-chan string, error) {
	return nil, nil
}