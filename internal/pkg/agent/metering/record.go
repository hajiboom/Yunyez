// Package metering metering database record model
package metering

import (
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// CostRecord 是写入数据库的实体
type CostRecord struct {
	gorm.Model
	SN               string          `gorm:"not null;index:idx_sn"`         // 序列号
	ModelName        string          `gorm:"not null;index:idx_model_name"` // 模型名
	PromptTokens     int             `gorm:"not null"`                      // 输入 token 数
	CompletionTokens int             `gorm:"not null"`                      // 输出 token 数
	TotalTokens      int             `gorm:"->;virtual"`                    // 虚拟字段（或可存实际值）
	Cost             decimal.Decimal `gorm:"not null;type:numeric(10,6)"`   // 成本数值
	Currency         string          `gorm:"size:3;default:'CNY'"`          // 货币单位
	DurationMS       int64           `gorm:"not null"`                      // 持续时间
}
