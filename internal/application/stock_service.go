package application

import (
	"backend/internal/domain/model"
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	stdhttp "net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func LoadStocks(name string) (map[time.Time]float32, error) {
	path := fmt.Sprintf("./assets/chart_%s.csv", name)
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %q: %w", path, err)
	}
	defer file.Close()

	stocks := make(map[time.Time]float32)
	buf := bufio.NewScanner(file)
	for i := 0; buf.Scan(); i++ {
		line := buf.Text()
		if i == 0 {
			// skip header
			continue
		}

		// Split manually on ;
		parts := strings.Split(line, ";")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid line %d: %q", i+1, line)
		}

		// Clean quotes
		dateStr := strings.Trim(parts[0], `"`)
		valueStr := strings.Trim(parts[1], `"`)

		// Parse date
		date, err := time.Parse("01/02/2006", dateStr)
		if err != nil {
			return nil, fmt.Errorf("invalid date %q on line %d: %w", dateStr, i+1, err)
		}

		// Replace comma with dot for float
		valueStr = strings.ReplaceAll(valueStr, ",", ".")
		val, err := strconv.ParseFloat(valueStr, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid value %q on line %d: %w", valueStr, i+1, err)
		}

		stocks[date] = float32(val)
	}

	if err := buf.Err(); err != nil {
		return nil, fmt.Errorf("error reading file %q: %w", path, err)
	}

	return stocks, nil
}

type StockService struct {
	stock map[string]map[time.Time]float32
}

func NewStockService() *StockService {
	goldStocks, err := LoadStocks("gold")
	if err != nil {
		log.Fatalf("Error loading stocks: %v", err)
	}
	silverStocks, err := LoadStocks("silver")
	if err != nil {
		log.Fatalf("Error loading stocks: %v", err)
	}

	return &StockService{
		stock: map[string]map[time.Time]float32{
			"gold":   goldStocks,
			"silver": silverStocks,
		},
	}
}

func (s *StockService) GetStocks(writer stdhttp.ResponseWriter, request *stdhttp.Request) {
	// Read query parameter "type"
	stockType := request.URL.Query().Get("type")
	if stockType == "" {
		stdhttp.Error(writer, "'type' query parameter is required", stdhttp.StatusBadRequest)
		return
	}

	stocks, ok := s.stock[stockType]
	if !ok {
		stdhttp.Error(writer, "unknown stock type", stdhttp.StatusNotFound)
		return
	}
	resp := make(map[string]float32)
	for k, v := range stocks {
		resp[k.Format("2006-01-02")] = v
	}

	// Send JSON response
	writer.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(writer).Encode(resp); err != nil {
		stdhttp.Error(writer, fmt.Sprintf("Failed to encode JSON: %v", err), stdhttp.StatusInternalServerError)
		return
	}
}

// MetalPriceResponse represents the JSON response from the goldpricez.com API.
// Prices are in USD per kg.
// API docs: http://goldpricez.com/about/api
// Example URL: http://goldpricez.com/api/rates/currency/usd/measure/kg

// fetchMetalPrice fetches the live price of a metal in USD per kg
// from the goldpricez.com API.
// metal must be "gold" or "silver".
func fetchMetalPrice(metal string) (*model.Stock, error) {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("API_KEY environment variable is not set")
	}

	apiURL := fmt.Sprintf(
		"https://goldpricez.com/api/%s/rates/currency/usd/measure/kg",
		strings.ToLower(metal),
	)

	req, err := stdhttp.NewRequest(stdhttp.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("X-API-KEY", apiKey)

	client := &stdhttp.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != stdhttp.StatusOK {
		return nil, fmt.Errorf("API returned non-200 status: %d", resp.StatusCode)
	}

	var price model.Stock
	if err := json.NewDecoder(resp.Body).Decode(&price); err != nil {
		return nil, fmt.Errorf("failed to decode API response: %w", err)
	}

	return &price, nil
}

// fetchGoldPrice returns the live gold price in USD per kg.
func (s *StockService) fetchGoldPrice() (*model.Stock, error) {
	return fetchMetalPrice("gold")
}

// fetchSilverPrice returns the live silver price in USD per kg.
func (s *StockService) fetchSilverPrice() (*model.Stock, error) {
	return fetchMetalPrice("silver")
}

// UpdateMetalPrices fetches the latest gold and silver prices from the API.
// Skeleton in place; persistence and in-memory update flow will be implemented next.
func (s *StockService) UpdateMetalPrices() error {
	return nil 
}	