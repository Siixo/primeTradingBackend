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
	chartURLFormat = "https://query1.finance.yahoo.com/v8/finance/chart/%s?interval=1d&range=1mo"
	userAgent      = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	
	// Conversion factors to KG
	metricTonToKg = 1000.0
	lbToKg        = 0.453592
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
	default:
		return ""
	}
}
