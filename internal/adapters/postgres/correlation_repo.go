package postgres

import (
	"backend/internal/domain/model"
	"backend/internal/domain/repository"
	"context"
	"database/sql"
	"fmt"
	"strings"
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


func (p *CorrelationRepository) Save(ctx context.Context, correlation *model.Correlation) error {
	query := `INSERT INTO correlations(commodity_a, commodity_b, correlationDate, pearsonR, spearmanRho, dataPoints) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := p.db.ExecContext(ctx, query, correlation.CommodityA, correlation.CommodityB, correlation.CorrelationDate, correlation.PearsonR, correlation.SpearmanRho, correlation.DataPoints)
	return err
}

func (p *CorrelationRepository) SaveBatch(ctx context.Context, correlations []*model.Correlation) error {
	if len(correlations) == 0 {
		return nil
	}

	var b strings.Builder
	b.WriteString("INSERT INTO correlations(commodity_a, commodity_b, correlationDate, pearsonR, spearmanRho, dataPoints) VALUES ")

	args := make([]interface{}, 0, len(correlations)*6)
	for i, c := range correlations {
		if i > 0 {
			b.WriteString(", ")
		}
		base := i * 6
		fmt.Fprintf(&b, "($%d, $%d, $%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4, base+5, base+6)
		args = append(args, c.CommodityA, c.CommodityB, c.CorrelationDate, c.PearsonR, c.SpearmanRho, c.DataPoints)
	}

	_, err := p.db.ExecContext(ctx, b.String(), args...)
	return err
}

func (p *CorrelationRepository) GetLatest(ctx context.Context, commodityA, commodityB string) (*model.Correlation, error) {
	query := `SELECT id, commodity_a, commodity_b, correlationDate, pearsonR, spearmanRho, dataPoints, createdAt FROM correlations WHERE commodity_a=$1 AND commodity_b=$2 ORDER BY correlationDate DESC LIMIT 1`
	row := p.db.QueryRowContext(ctx, query, commodityA, commodityB)
	var correlation model.Correlation
	err := row.Scan(&correlation.ID, &correlation.CommodityA, &correlation.CommodityB, &correlation.CorrelationDate, &correlation.PearsonR, &correlation.SpearmanRho, &correlation.DataPoints, &correlation.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &correlation, nil
}

func (p *CorrelationRepository) GetHistory(ctx context.Context, commodityA, commodityB string, limit int) ([]*model.Correlation, error) {
	query := `SELECT id, commodity_a, commodity_b, correlationDate, pearsonR, spearmanRho, dataPoints, createdAt FROM correlations WHERE commodity_a=$1 AND commodity_b=$2 ORDER BY correlationDate DESC LIMIT $3`
	
	rows, err := p.db.QueryContext(ctx, query, commodityA, commodityB, limit)
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
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return history, nil
}
	
func (p *CorrelationRepository) GetTopCorrelated(ctx context.Context, commodity string, limit int) ([]*model.Correlation, error) {
	query := `SELECT id, commodity_a, commodity_b, correlationDate, pearsonR, spearmanRho, dataPoints, createdAt FROM correlations WHERE commodity_a=$1 OR commodity_b=$1 ORDER BY ABS(pearsonR) DESC LIMIT $2`
	
	rows, err := p.db.QueryContext(ctx, query, commodity, limit)
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