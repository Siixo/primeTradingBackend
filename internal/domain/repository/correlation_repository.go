package repository

import (
	"backend/internal/domain/model"
	"context"
)

type CorrelationRepository interface {
	Migrate() error
	Save(ctx context.Context, correlation *model.Correlation) error
	SaveBatch(ctx context.Context, correlations []*model.Correlation) error
	GetLatest(ctx context.Context, commodityA, commodityB string) (*model.Correlation, error)
	GetHistory(ctx context.Context, commodityA, commodityB string, limit int) ([]*model.Correlation, error)
	GetTopCorrelated(ctx context.Context, commodity string, limit int) ([]*model.Correlation, error)
}