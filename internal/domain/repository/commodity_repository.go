package repository

import "backend/internal/domain/model"

type CommodityRepository interface {
	Migrate() error
	Save(stock model.Commodity) error
	GetLatestPrice(commodity string) (model.Commodity, error)
	GetPriceHistory(commodity string, limit int) ([]model.Commodity, error)
}