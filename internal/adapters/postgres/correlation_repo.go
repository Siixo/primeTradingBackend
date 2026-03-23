package postgres

import (
	"backend/internal/domain/model"
	"backend/internal/domain/repository"
	"database/sql"
)

type CorrelationRepository struct {
	db *sql.DB
}

func NewCorrelationRepository(db *sql.DB) repository.CorrelationRepository {
	return &CorrelationRepository{db}
}

func (p *CorrelationRepository) Migrate() error {
	query := `CREATE TABLE IF NOT EXISTS correlations (
		id					SERIAL PRIMARY KEY,
		commodity_a			VARCHAR(50) NOT NULL,
		commodity_b 		VARCHAR(50) NOT NULL,
		correlationDate 	TIMESTAMP NOT NULL,
		pearsonR 			FLOAT,
		spearmanRho 		FLOAT,
		dataPoints 			INT NOT NULL,
		createdAt 			TIMESTAMP NOT NULL DEFAULT NOW()
	);`

	_, err := p.db.Exec(query)
	return err
}


func (p *CorrelationRepository) Save(correlation *model.Correlation) error {
	query := `INSERT INTO correlations(commodity_a, commodity_b, correlationDate, pearsonR, spearmanRho, dataPoints) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := p.db.Exec(query, correlation.CommodityA, correlation.CommodityB, correlation.CorrelationDate, correlation.PearsonR, correlation.SpearmanRho, correlation.DataPoints)
	return err
}

func (p *CorrelationRepository) SaveBatch(correlations []*model.Correlation) error {
	query := `INSERT INTO correlations(commodity_a, commodity_b, correlationDate, pearsonR, spearmanRho, dataPoints) VALUES ($1, $2, $3, $4, $5, $6)`
	for _, correlation := range correlations {
		_, err := p.db.Exec(query, correlation.CommodityA, correlation.CommodityB, correlation.CorrelationDate, correlation.PearsonR, correlation.SpearmanRho, correlation.DataPoints)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *CorrelationRepository) GetLatest(commodityA, commodityB string) (*model.Correlation, error) {
	query := `SELECT id, commodity_a, commodity_b, correlationDate, pearsonR, spearmanRho, dataPoints, createdAt FROM correlations WHERE commodity_a=$1 AND commodity_b=$2 ORDER BY correlationDate DESC LIMIT 1`
	row := p.db.QueryRow(query, commodityA, commodityB)
	var correlation model.Correlation
	err := row.Scan(&correlation.ID, &correlation.CommodityA, &correlation.CommodityB, &correlation.CorrelationDate, &correlation.PearsonR, &correlation.SpearmanRho, &correlation.DataPoints, &correlation.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &correlation, nil
}

func (p *CorrelationRepository) GetHistory(commodityA, commodityB string, limit int) ([]*model.Correlation, error) {
	query := `SELECT id, commodity_a, commodity_b, correlationDate, pearsonR, spearmanRho, dataPoints, createdAt FROM correlations WHERE commodity_a=$1 AND commodity_b=$2 ORDER BY correlationDate DESC LIMIT $3`
	
	rows, err := p.db.Query(query, commodityA, commodityB, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var history []*model.Correlation
	for rows.Next() {
		var c model.Correlation

		err := rows.Scan(&c.ID, &c.CommodityA, &c.CommodityB, &c.CorrelationDate, &c.PearsonR, &c.SpearmanRho, &c.DataPoints, &c.CreatedAt)
		if err != nil {
			return nil, err
		}
		history = append(history, &c)
	}
	// Always check if there were errors encountered during the iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return history, nil
}
	
func (p *CorrelationRepository) GetTopCorrelated(commodity string, limit int) ([]*model.Correlation, error) {
	// Look for the commodity in either column A or B
	// ABS(pearsonR) makes sure we find the strongest correlations (e.g., -0.9 is stronger than 0.5)
	query := `SELECT id, commodity_a, commodity_b, correlationDate, pearsonR, spearmanRho, dataPoints, createdAt FROM correlations WHERE commodity_a=$1 OR commodity_b=$1 ORDER BY ABS(pearsonR) DESC LIMIT $2`
	
	rows, err := p.db.Query(query, commodity, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var correlations []*model.Correlation
	for rows.Next() {
		var c model.Correlation
		err := rows.Scan(&c.ID, &c.CommodityA, &c.CommodityB, &c.CorrelationDate, &c.PearsonR, &c.SpearmanRho, &c.DataPoints, &c.CreatedAt)
		if err != nil {
			return nil, err
		}
		correlations = append(correlations, &c)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return correlations, nil
}