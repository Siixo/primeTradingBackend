package yahoofinance

import (
	"backend/internal/domain/model"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	chartURLFormat = "https://query1.finance.yahoo.com/v8/finance/chart/%s?interval=5m&range=1d"
	userAgent      = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	
	// Conversion factors to KG
	metricTonToKg = 1000.0
	lbToKg        = 0.453592
	barrelToKg    = 136.0 // Approximate for Brent Oil
	troyOunceToKg = 0.0311035
)

type Client struct {
	client *http.Client
}

func NewClient() *Client {
	return &Client{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// YahooChartResponse represents the parts of the Yahoo JSON we care about
type YahooChartResponse struct {
	Chart struct {
		Result []struct {
			Meta struct {
				Symbol string `json:"symbol"`
			} `json:"meta"`
			Timestamp []int64 `json:"timestamp"`
			Indicators struct {
				Quote []struct {
					Close []float64 `json:"close"`
				} `json:"quote"`
			} `json:"indicators"`
		} `json:"result"`
		Error interface{} `json:"error"`
	} `json:"chart"`
}

func (c *Client) FetchPrice(commodity string) (*model.Commodity, error) {
	ticker := c.getTickerFor(commodity)
	if ticker == "" {
		return nil, fmt.Errorf("unsupported commodity: %s", commodity)
	}

	url := fmt.Sprintf(chartURLFormat, ticker)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("yahoo finance returned status %d: %s", resp.StatusCode, string(body))
	}

	var chartResp YahooChartResponse
	if err := json.NewDecoder(resp.Body).Decode(&chartResp); err != nil {
		return nil, fmt.Errorf("failed to decode yahoo response: %w", err)
	}

	if len(chartResp.Chart.Result) == 0 || len(chartResp.Chart.Result[0].Timestamp) == 0 {
		return nil, fmt.Errorf("no data returned for %s", commodity)
	}

	result := chartResp.Chart.Result[0]
	lastIdx := len(result.Timestamp) - 1
	
	// Sometimes the last value is null in Yahoo's JSON
	for lastIdx >= 0 && (len(result.Indicators.Quote[0].Close) <= lastIdx || result.Indicators.Quote[0].Close[lastIdx] == 0) {
		lastIdx--
	}

	if lastIdx < 0 {
		return nil, fmt.Errorf("no valid closing price found for %s", commodity)
	}

	price := result.Indicators.Quote[0].Close[lastIdx]
	timestamp := result.Timestamp[lastIdx]

	var priceKg float64
	switch ticker {
	case "HG=F":
		// Copper futures are in US Cents per pound (lb)
		// Convert cents to dollars, then to KG
		priceKg = (price / 100.0) / lbToKg
	case "ALI=F":
		// Aluminum futures are in Dollars per metric ton
		priceKg = price / metricTonToKg
	case "BZ=F":
		// Brent Oil is in Dollars per barrel
		priceKg = price / barrelToKg
	case "GC=F", "SI=F":
		// Gold and Silver are in Dollars per troy ounce
		priceKg = price / troyOunceToKg
	}

	return &model.Commodity{
		Name:      commodity,
		Date:      time.Unix(timestamp, 0),
		PriceKg:   priceKg,
		Unit:      "USD/kg",
		FetchedAt: time.Now(),
	}, nil
}

func (c *Client) FetchCommodity() (*model.Commodity, error) {
	return nil, fmt.Errorf("use FetchPrice with specific symbol")
}

func (c *Client) getTickerFor(commodity string) string {
	switch strings.ToLower(commodity) {
	case "copper":
		return "HG=F"
	case "aluminum", "aluminium":
		return "ALI=F"
	case "brent":
		return "BZ=F"
	case "gold":
		return "GC=F"
	case "silver":
		return "SI=F"
	default:
		return ""
	}
}

func (c *Client) FetchHistory(commodity string) ([]model.Commodity, error) {
	ticker := c.getTickerFor(commodity)
	if ticker == "" {
		return nil, fmt.Errorf("unsupported commodity: %s", commodity)
	}

	url := fmt.Sprintf(chartURLFormat, ticker)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("yahoo finance returned status %d: %s", resp.StatusCode, string(body))
	}

	var chartResp YahooChartResponse
	if err := json.NewDecoder(resp.Body).Decode(&chartResp); err != nil {
		return nil, fmt.Errorf("failed to decode yahoo response: %w", err)
	}

	if len(chartResp.Chart.Result) == 0 || len(chartResp.Chart.Result[0].Timestamp) == 0 {
		return nil, fmt.Errorf("no data returned for %s", commodity)
	}

	result := chartResp.Chart.Result[0]
	var history []model.Commodity

	for i, timestamp := range result.Timestamp {
		if len(result.Indicators.Quote[0].Close) <= i {
			continue
		}
		
		price := result.Indicators.Quote[0].Close[i]
		if price == 0 {
			continue
		}

		var priceKg float64
		switch ticker {
		case "HG=F":
			priceKg = (price / 100.0) / lbToKg
		case "ALI=F":
			priceKg = price / metricTonToKg
		case "BZ=F":
			priceKg = price / barrelToKg
		case "GC=F", "SI=F":
			priceKg = price / troyOunceToKg
		}

		history = append(history, model.Commodity{
			Name:      commodity,
			Date:      time.Unix(timestamp, 0),
			PriceKg:   priceKg,
			Unit:      "USD/kg",
			FetchedAt: time.Now(),
		})
	}

	return history, nil
}
