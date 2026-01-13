package metering

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCostRepository 使用 testify/mock
type MockCostRepository struct {
	mock.Mock
}

func (m *MockCostRepository) Save(ctx context.Context, record *CostRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

// TestRecord_Success 测试正常记录流程
func TestRecord_Success(t *testing.T) {
	start := time.Now()
	time.Sleep(2 * time.Second) // 模拟执行耗时
	end := time.Now()

	mockRepo := new(MockCostRepository)
	rules := map[string]PricingRule{
		"qwen-flash": {
			InputPrice:  0.0,
			OutputPrice: 0.0,
			Currency:    "CNY",
		},
	}
	calculator := NewStandardCostCalculator(rules)
	service := NewMeteringService(calculator, mockRepo)

	expectedCost := decimal.NewFromFloat(0.0).Mul(decimal.NewFromInt(22 + 66)).Div(decimal.NewFromInt(1000))

	mockRepo.On("Save", mock.Anything, mock.MatchedBy(func(r *CostRecord) bool {
		return r.SN == "device-123" &&
			r.ModelName == "qwen-flash" &&
			r.PromptTokens == 22 &&
			r.CompletionTokens == 66 &&
			r.Currency == "CNY" &&
			r.Cost.Equal(expectedCost) && // decimal 提供 Equal 方法
			r.DurationMS >= 0
	})).Return(nil)

	usage := Usage{
		Model:            "qwen-flash",
		PromptTokens:     22,
		CompletionTokens: 66,
		TotalTokens:      88,
		StartTime:        start,
		EndTime:          end,
	}

	err := service.Record(context.Background(), "device-123", usage)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
