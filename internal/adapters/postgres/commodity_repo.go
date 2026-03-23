package postgres

import (
	"backend/internal/domain/repository"
	"database/sql"

	"backend/internal/domain/model"
)

type CommodityRepository struct {
	db *sql.DB
}

func NewCommodityRepository(db *sql.DB) repository.CommodityRepository {
	return &CommodityRepository{db: db}
}

func (p *CommodityRepository) Migrate() error {
	query := `CREATE TABLE IF NOT EXISTS commodities (
		id					SERIAL PRIMARY KEY,
		name			VARCHAR(50) NOT NULL,
		date 			TIMESTAMP NOT NULL,
		price_kg 			FLOAT,
		unit 			VARCHAR(50) NOT NULL,
		fetched_at 			TIMESTAMP NOT NULL DEFAULT NOW()
	);`

	_, err := p.db.Exec(query)
	return err
}

func (p *CommodityRepository) Save(stock model.Commodity) error {
	query := `INSERT INTO commodities (name, date, price_kg, unit, fetched_at) 
			  VALUES ($1, $2, $3, $4, $5) 
			  ON CONFLICT DO NOTHING` // Or update if needed, but for history 'DO NOTHING' or just insert is common.
	_, err := p.db.Exec(query, stock.Name, stock.Date, stock.PriceKg, stock.Unit, stock.FetchedAt)
	return err
}

func (p *CommodityRepository) GetLatestPrice(commodity string) (model.Commodity, error) {
	query := `SELECT id, name, date, price_kg, unit, fetched_at
			  FROM commodities WHERE name=$1 ORDER BY date DESC LIMIT 1`
	row := p.db.QueryRow(query, commodity)
	var c model.Commodity
	err := row.Scan(&c.ID, &c.Name, &c.Date, &c.PriceKg, &c.Unit, &c.FetchedAt)
	if err != nil {
		return model.Commodity{}, err
	}
	return c, nil
}

func (p *CommodityRepository) GetPriceHistory(commodity string, limit int) ([]model.Commodity, error) {
	query := `SELECT id, name, date, price_kg, unit, fetched_at
			  FROM commodities WHERE name=$1 ORDER BY date DESC LIMIT $2`
	rows, err := p.db.Query(query, commodity, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []model.Commodity
	for rows.Next() {
		var c model.Commodity
		if err := rows.Scan(&c.ID, &c.Name, &c.Date, &c.PriceKg, &c.Unit, &c.FetchedAt); err != nil {
			return nil, err
		}
		history = append(history, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return history, nil
}