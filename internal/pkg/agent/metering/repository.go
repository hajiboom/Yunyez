// Package metering provides the metering service.
// defined the cost record repository interface and methods
package metering

import (
	"context"

	postgre "yunyez/internal/pkg/postgre"
)

type CostRepository interface {
	Save(ctx context.Context, metering *CostRecord) error
}

type PostgreCostRepository struct {
	db *postgre.Client
}


func NewPostgreCostRepository(db *postgre.Client) *PostgreCostRepository {
	return &PostgreCostRepository{db: db}
}

func (r *PostgreCostRepository) Save(ctx context.Context, metering *CostRecord) error {
	// write to postgresql
	if err := r.db.DB.WithContext(ctx).Create(metering).Error; err != nil {
		return err
	}

	return nil
}