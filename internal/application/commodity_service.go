package application

import (
	"backend/internal/domain/model"
	"backend/internal/domain/repository"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
)

// MetalPriceProvider is the outbound port for fetching live metal prices.
// Implementations live in the adapters layer (e.g. adapters/goldpricez).
type MetalPriceProvider interface {
	FetchPrice(metal string) (*model.Commodity, error)
	FetchHistory(symbol string) ([]model.Commodity, error)
	FetchCommodity() (*model.Commodity, error)
}

type CommodityRepositoryPort interface {
	GetPriceHistory(commodity string, limit int) ([]model.Commodity, error)
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

func (s *CommodityService) GetCommodityByType(commodityType string) (*model.Commodity, error) {
	if commodityType == "" {
		return nil, errors.New("'type' query parameter is required")
	}

	commodity := strings.ToLower(commodityType)
	if commodity != "gold" && commodity != "silver" && commodity != "copper" && commodity != "aluminum" && commodity != "brent" {
		return nil, errors.New("unknown commodity type")
	}

	return s.priceProvider.FetchPrice(commodity)
}

func (s *CommodityService) GetHistory(name string, limit int) ([]model.Commodity, error) {
	if limit <= 0 {
		limit = 100
	}
	return s.commodityRepo.GetPriceHistory(name, limit)
}

func (s *CommodityService) UpdatePreciousPrices() error {
	return s.updateSymbols([]string{"gold", "silver"})
}

func (s *CommodityService) UpdateIndustrialPrices() error {
	return s.updateSymbols([]string{"copper", "aluminum", "brent"})
}

// UpdateMetalPrices is kept for compatibility and executes both update groups.
func (s *CommodityService) UpdateMetalPrices() error {
	if err := s.UpdatePreciousPrices(); err != nil {
		if err2 := s.UpdateIndustrialPrices(); err2 != nil {
			return fmt.Errorf("precious update failed: %v; industrial update failed: %v", err, err2)
		}
		return err
	}
	return s.UpdateIndustrialPrices()
}

func (s *CommodityService) updateSymbols(symbols []string) error {

	successes := 0
	var failed []string
	for _, symbol := range symbols {
		// First, try to fetch the historical points (e.g. last 24h)
		history, err := s.priceProvider.FetchHistory(symbol)
		if err == nil && len(history) > 0 {
			for _, c := range history {
				if err := s.commodityRepo.Save(c); err != nil {
					// Ignore individual save errors (e.g. duplicate key)
					continue
				}
			}
			s.clearLastError(symbol)
			successes++
			continue
		}

		// Fallback to single price fetch if history fails or isn't supported
		commodity, err := s.priceProvider.FetchPrice(symbol)
		if err != nil {
			s.setLastError(symbol, err)
			failed = append(failed, fmt.Sprintf("fetch %s: %v", symbol, err))
			continue
		}

		if err := s.commodityRepo.Save(*commodity); err != nil {
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

	if len(failed) > 0 {
		// Keep the updater moving when at least one symbol is available.
		return nil
	}

	return nil
}

func (s *CommodityService) GetStatuses() ([]model.CommodityStatus, error) {
	symbols := []string{"gold", "silver", "copper", "aluminum", "brent"}
	statuses := make([]model.CommodityStatus, 0, len(symbols))

	for _, symbol := range symbols {
		status := model.CommodityStatus{
			Name:   symbol,
			Source: sourceFor(symbol),
		}

		latest, err := s.commodityRepo.GetLatestPrice(symbol)
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
