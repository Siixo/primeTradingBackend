package postgres

import (
	"backend/internal/domain/repository"
	"database/sql"
	"errors"
	"time"

	"backend/internal/domain/model"
)

type StockRepository struct {
	db *sql.DB
}

func NewStockRepository(db *sql.DB) repository.StockRepository {
	return &StockRepository{db: db}
}

func (p *StockRepository) Save(stock model.Stock) error {
	_ = stock
	return errors.New("Save not implemented")
}

func (p *StockRepository) GetLatestPrice(commodity string) (model.Stock, error) {
	_ = commodity
	return model.Stock{}, errors.New("GetLatestPrice not implemented")
}

func (p *StockRepository) GetPriceHistory(commodity string, limit int) ([]model.Stock, error) {
	_ = commodity
	_ = limit
	_ = time.Now()
	return nil, errors.New("GetPriceHistory not implemented")
}