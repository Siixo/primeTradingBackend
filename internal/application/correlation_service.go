package application

import (
	"backend/internal/domain/algorithm"
	"backend/internal/domain/model"
	"backend/internal/domain/repository"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

type CorrelationService struct {
	correlationRepo repository.CorrelationRepository
	commodityRepo   repository.CommodityRepository
}

func NewCorrelationService(correlationRepo repository.CorrelationRepository, commodityRepo repository.CommodityRepository) *CorrelationService {
	return &CorrelationService{
		correlationRepo: correlationRepo,
		commodityRepo:   commodityRepo,
	}
}

func (s *CorrelationService) GetCorrelationByType(correlationType string) (*model.Correlation, error) {
	if correlationType == "" {
		return nil, errors.New("'type' query parameter is required")
	}

	// Assuming correlationType is in format "commodityA-commodityB" (e.g., "gold-silver")
	parts := strings.Split(correlationType, "-")
	if len(parts) != 2 {
		return nil, errors.New("invalid correlation type format. Expected 'commodityA-commodityB'")
	}

	commodityA := strings.ToLower(parts[0])
	commodityB := strings.ToLower(parts[1])

	correlation, err := s.correlationRepo.GetLatest(commodityA, commodityB)
	if err != nil {
		return nil, err
	}

	return correlation, nil
}

// UpdateCorrelations computes and saves the latest correlation for a given pair.
func (s *CorrelationService) UpdateCorrelations(commodityA, commodityB string) error {
	const historyLimit = 100 // Last 100 points
	
	historyA, err := s.commodityRepo.GetPriceHistory(commodityA, historyLimit)
	if err != nil {
		return fmt.Errorf("fetch %s history: %w", commodityA, err)
	}
	
	historyB, err := s.commodityRepo.GetPriceHistory(commodityB, historyLimit)
	if err != nil {
		return fmt.Errorf("fetch %s history: %w", commodityB, err)
	}

	// Aligned data
	var x []float64
	var y []float64
	
	// Fast lookup for historyB by date
	mapB := make(map[string]float64)
	for _, c := range historyB {
		// Use date only for alignment (especially for daily data like Brent)
		dateKey := c.Date.Format("2006-01-02")
		mapB[dateKey] = c.PriceKg
	}
	
	for _, c := range historyA {
		dateKey := c.Date.Format("2006-01-02")
		if valB, ok := mapB[dateKey]; ok {
			x = append(x, c.PriceKg)
			y = append(y, valB)
		}
	}
	
	if len(x) < 5 { // Arbitrary minimum for correlation
		return fmt.Errorf("insufficient overlapping data points (found %d)", len(x))
	}
	
	pearsonR, err := algorithm.Pearson(x, y)
	if err != nil {
		return fmt.Errorf("calculate pearson: %w", err)
	}
	if math.IsNaN(pearsonR) || math.IsInf(pearsonR, 0) {
		pearsonR = 0
	}
	
	spearmanRho, err := algorithm.Spearman(x, y)
	if err != nil {
		return fmt.Errorf("calculate spearman: %w", err)
	}
	if math.IsNaN(spearmanRho) || math.IsInf(spearmanRho, 0) {
		spearmanRho = 0
	}
	
	correlation := &model.Correlation{
		CommodityA:      commodityA,
		CommodityB:      commodityB,
		CorrelationDate: time.Now(),
		PearsonR:        pearsonR,
		SpearmanRho:     spearmanRho,
		DataPoints:      len(x),
	}
	
	return s.correlationRepo.Save(correlation)
}

func (s *CorrelationService) GetHistory(commodityA, commodityB string, limit int) ([]*model.Correlation, error) {
	if limit <= 0 {
		limit = 100
	}
	return s.correlationRepo.GetHistory(commodityA, commodityB, limit)
}
