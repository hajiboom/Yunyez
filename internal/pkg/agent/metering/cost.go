// Package metering model usage cost
package metering

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// Usage model usage
type Usage struct {
	Model            string    `json:"model"`             // model name
	PromptTokens     int       `json:"prompt_tokens"`     // prompt tokens count
	CompletionTokens int       `json:"completion_tokens"` // completion tokens count
	TotalTokens      int       `json:"total_tokens"`      // total tokens count
	StartTime        time.Time `json:"start_time"`        // start time
	EndTime          time.Time `json:"end_time"`          // end time
}

// Cost model usage cost
type Cost struct {
	Usage
	Data     decimal.Decimal `json:"data"`
	Currency string          `json:"currency"` //currency unit e.g. "CNY", "USD"
}

// CostCalculator cost calculator interface
type CostCalculator interface {
	// Calculate calculate cost
	Calculate(usage Usage) (Cost, error)
}

// PricingRule model pricing rule
type PricingRule struct {
	InputPrice  float64 // per 1K tokens in
	OutputPrice float64 // per 1K tokens out
	Currency    string
}

// StandardCostCalculator standard cost calculator
type StandardCostCalculator struct {
	rules map[string]PricingRule
}

// NewStandardCostCalculator create standard cost calculator
func NewStandardCostCalculator(rules map[string]PricingRule) *StandardCostCalculator {
	return &StandardCostCalculator{rules: rules}
}

// Calculate calculate cost
// Parameters:
//   - u: model usage
//
// Returns:
//   - Cost: model usage cost
//   - error: error information
func (c *StandardCostCalculator) Calculate(u Usage) (Cost, error) {
	rule, exists := c.rules[u.Model]
	if !exists {
		// unknown model, cost=0 but keep record
		return Cost{
			Usage:    u,
			Data:     decimal.NewFromFloat(0.0),
			Currency: "UNKNOWN",
		}, fmt.Errorf("unknown model: %s", u.Model)
	}
	promptCost := decimal.NewFromInt(int64(u.PromptTokens)).Div(decimal.NewFromInt(1000)).Mul(decimal.NewFromFloat(rule.InputPrice))
	completionCost := decimal.NewFromInt(int64(u.CompletionTokens)).Div(decimal.NewFromInt(1000)).Mul(decimal.NewFromFloat(rule.OutputPrice))
	totalCost := promptCost.Add(completionCost)

	return Cost{
		Usage:    u,
		Data:     totalCost,
		Currency: rule.Currency,
	}, nil
}
