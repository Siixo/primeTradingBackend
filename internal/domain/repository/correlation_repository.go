package repository

import "backend/internal/domain/model"

type CorrelationRepository interface {
	Migrate() error
	// Save inserts a new correlation record
	Save(correlation *model.Correlation) error

	// SaveBatch inserts multiple records efficiently (useful since correlations are usually calculated in bulk)
	SaveBatch(correlations []*model.Correlation) error

	// GetLatest returns the most recent correlation between two specific commodities
	GetLatest(commodityA, commodityB string) (*model.Correlation, error)

	// GetHistory returns the historical correlations between two commodities, ordered by date descending
	GetHistory(commodityA, commodityB string, limit int) ([]*model.Correlation, error)

	// GetTopCorrelated returns the most strongly correlated pairs for a given commodity on a specific date
	// Can be ordered by closest to 1 or -1 (Pearson/Spearman)
	GetTopCorrelated(commodity string, limit int) ([]*model.Correlation, error)
}