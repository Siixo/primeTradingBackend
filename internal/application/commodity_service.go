package application

import (
	"backend/internal/domain/model"
	"backend/internal/domain/repository"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
)

// MetalPriceProvider is the outbound port for fetching live metal prices.
type MetalPriceProvider interface {
	FetchPrice(ctx context.Context, metal string) (*model.Commodity, error)
}

type CommodityService struct {
	priceProvider MetalPriceProvider
	commodityRepo repository.CommodityRepository
	statusMu      sync.RWMutex
	lastErrors    map[string]string
}

func NewCommodityService(priceProvider MetalPriceProvider, commodityRepo repository.CommodityRepository) *CommodityService {
	return &CommodityService{
		priceProvider: priceProvider,
		commodityRepo: commodityRepo,
		lastErrors:    make(map[string]string),
	}
}

func (s *CommodityService) GetCommodityByType(ctx context.Context, commodityType string) (*model.Commodity, error) {
	if commodityType == "" {
		return nil, errors.New("'type' query parameter is required")
	}

	commodity := strings.ToLower(commodityType)
	if commodity != "gold" && commodity != "silver" && commodity != "copper" && commodity != "aluminum" && commodity != "brent" {
		return nil, errors.New("unknown commodity type")
	}

	return s.priceProvider.FetchPrice(ctx, commodity)
}

func (s *CommodityService) GetHistory(ctx context.Context, name string, limit int) ([]model.Commodity, error) {
	if limit <= 0 {
		limit = 100
	}
	return s.commodityRepo.GetPriceHistory(ctx, name, limit)
}

func (s *CommodityService) UpdatePreciousPrices(ctx context.Context) error {
	return s.updateSymbols(ctx, []string{"gold", "silver"})
}

func (s *CommodityService) UpdateIndustrialPrices(ctx context.Context) error {
	return s.updateSymbols(ctx, []string{"copper", "aluminum", "brent"})
}

func (s *CommodityService) updateSymbols(ctx context.Context, symbols []string) error {

	successes := 0
	var failed []string
	for _, symbol := range symbols {
		commodity, err := s.priceProvider.FetchPrice(ctx, symbol)
		if err != nil {
			s.setLastError(symbol, err)
			failed = append(failed, fmt.Sprintf("fetch %s: %v", symbol, err))
			continue
		}

		if err := s.commodityRepo.Save(ctx, *commodity); err != nil {
			s.setLastError(symbol, err)
			failed = append(failed, fmt.Sprintf("save %s: %v", symbol, err))
			continue
		}

		s.clearLastError(symbol)
		successes++
	}

	if successes == 0 {
		if len(failed) == 0 {
			return errors.New("no symbols processed")
		}
		return fmt.Errorf("all symbol updates failed: %s", strings.Join(failed, "; "))
	}

	return nil
}

func (s *CommodityService) GetStatuses(ctx context.Context) ([]model.CommodityStatus, error) {
	symbols := []string{"gold", "silver", "copper", "aluminum", "brent"}
	statuses := make([]model.CommodityStatus, 0, len(symbols))

	for _, symbol := range symbols {
		status := model.CommodityStatus{
			Name:   symbol,
			Source: sourceFor(symbol),
		}

		latest, err := s.commodityRepo.GetLatestPrice(ctx, symbol)
		if err == nil {
			status.Available = true
			status.LastDate = &latest.Date
		} else if !errors.Is(err, sql.ErrNoRows) {
			status.LastError = err.Error()
		}

		if lastErr := s.getLastError(symbol); lastErr != "" {
			status.LastError = lastErr
		}

		statuses = append(statuses, status)
	}

	return statuses, nil
}

func sourceFor(symbol string) string {
	switch symbol {
	case "gold", "silver":
		return "GoldPriceZ"
	case "copper", "aluminum", "brent":
		return "AlphaVantage"
	default:
		return "Official API"
	}
}

func (s *CommodityService) setLastError(symbol string, err error) {
	s.statusMu.Lock()
	defer s.statusMu.Unlock()
	s.lastErrors[symbol] = err.Error()
}

func (s *CommodityService) clearLastError(symbol string) {
	s.statusMu.Lock()
	defer s.statusMu.Unlock()
	delete(s.lastErrors, symbol)
}

func (s *CommodityService) getLastError(symbol string) string {
	s.statusMu.RLock()
	defer s.statusMu.RUnlock()
	return s.lastErrors[symbol]
}
