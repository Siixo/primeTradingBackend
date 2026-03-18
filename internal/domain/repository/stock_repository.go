package repository 

import "backend/internal/domain/model"

type StockRepository interface {
	Save(stock model.Stock) error
	GetLatestPrice(commodity string) (model.Stock, error)
	GetPriceHistory(commodity string, limit int) ([]model.Stock, error)
}