package repository

import (
	"backend/internal/domain/model"
	"context"
)

type CommodityRepository interface {
	Migrate() error
	Save(ctx context.Context, stock model.Commodity) error
	GetLatestPrice(ctx context.Context, commodity string) (model.Commodity, error)
	GetPriceHistory(ctx context.Context, commodity string, limit int) ([]model.Commodity, error)
	HasRecentData(ctx context.Context) (bool, error)
}