// Package metering provides the metering service.
package metering

import (
	"context"
	"sync"
	"yunyez/internal/pkg/logger"
	postgre "yunyez/internal/pkg/postgre"
)

var (
	once          sync.Once

	dbClient = postgre.GetClient()
)

// MeteringService 成本计算服务
type MeteringService struct {
	calculator CostCalculator
	repo       CostRepository
}

// NewMeteringService 允许外部构造（便于测试）
func NewMeteringService(calc CostCalculator, repo CostRepository) *MeteringService {
	return &MeteringService{calculator: calc, repo: repo}
}

// Initialize 初始化成本计算服务
// params:
// - rules map[string]PricingRule 成本计算规则
func Initialize(rules map[string]PricingRule) *MeteringService {
	var ms *MeteringService
	once.Do(func() {
		calculator := NewStandardCostCalculator(rules)
		repo := NewPostgreCostRepository(dbClient)
		ms = &MeteringService{
			calculator: calculator,
			repo:       repo,
		}
	})
	return ms
}

// Record 记录成本
// params:
// - ctx context.Context 上下文
// - clientID string 客户端 ID
// - usage Usage 成本使用记录
// returns:
// - error 错误信息
func (ms *MeteringService) Record(ctx context.Context, clientID string, usage Usage) error {
	if ms == nil {
		return nil // 未初始化，静默跳过（开发模式安全）
	}
	
	cost, err := ms.calculator.Calculate(usage)
	if err != nil {
		logger.Warn(ctx, "cost calculation failed", map[string]any{
			"error": err.Error(),
			"usage": usage,
		})
		return err
	}
	durationMs := usage.EndTime.Sub(usage.StartTime).Milliseconds()
	record := &CostRecord{
		SN:               clientID,
		ModelName:        cost.Model,
		PromptTokens:     cost.PromptTokens,
		CompletionTokens: cost.CompletionTokens,
		TotalTokens:      cost.TotalTokens,
		Cost:             cost.Data,
		Currency:         cost.Currency,
		DurationMS:       durationMs,
	}

	return ms.repo.Save(ctx, record)
}
