package alphavantage

import (
	"backend/internal/domain/model"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	baseURL = "https://www.alphavantage.co/query"
	goldPricezURL = "https://goldpricez.com/api/rates/currency/usd/measure/ounce/metal/all"
	alphaVantageMinInterval = 1200 * time.Millisecond
	
	// Conversion factors to KG
	troyOunceToKg = 0.0311035
	metricTonToKg = 1000.0
	barrelToKg    = 136.0 // Approximate for Brent Oil
)

type Client struct {
	httpClient         *http.Client
	alphaVantageAPIKey string
	goldPricezAPIKey   string
	alphaVantageMu     sync.Mutex
	lastAlphaCallAt    time.Time
}

func NewClient(httpClient *http.Client, alphaVantageKey, goldPricezKey string) *Client {
	return &Client{
		httpClient:         httpClient,
		alphaVantageAPIKey: alphaVantageKey,
		goldPricezAPIKey:   goldPricezKey,
	}
}

// AlphaVantageExchangeRate handles CURRENCY_EXCHANGE_RATE responses
type AlphaVantageExchangeRate struct {
	RealtimeExchangeRate struct {
		FromCurrencyCode string `json:"1. From_Currency Code"`
		ExchangeRate     string `json:"5. Exchange Rate"`
		LastRefreshed    string `json:"6. Last Refreshed"`
	} `json:"Realtime Currency Exchange Rate"`
}

// AlphaVantageCommodity handles COPPER, BRENT, ALUMINUM responses
type AlphaVantageCommodity struct {
	Name string `json:"name"`
	Unit string `json:"unit"`
	Data []struct {
		Date  string `json:"date"`
		Value string `json:"value"`
	} `json:"data"`
}

type alphaVantageErrorResponse struct {
	Information  string `json:"Information"`
	Note         string `json:"Note"`
	ErrorMessage string `json:"Error Message"`
}

type GoldPricezRates struct {
	OuncePriceUSD            string `json:"ounce_price_usd"`
	SilverOuncePriceAskUSD   string `json:"silver_ounce_price_ask_usd"`
	GMTUpdated               string `json:"gmt_ounce_price_usd_updated"`
}

func (c *Client) FetchPrice(ctx context.Context, symbol string) (*model.Commodity, error) {
	s := strings.ToLower(symbol)
	
	switch s {
	case "gold", "silver":
		return c.fetchPreciousMetal(ctx, s)
	case "copper", "brent", "aluminum", "aluminium":
		return c.fetchIndustrialCommodity(ctx, s)
	default:
		return nil, fmt.Errorf("unsupported symbol: %s", symbol)
	}
}

func (c *Client) fetchPreciousMetal(ctx context.Context, metal string) (*model.Commodity, error) {
	if c.goldPricezAPIKey == "" {
		return nil, errors.New("missing GOLD_PRICEZ_API_KEY")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, goldPricezURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-KEY", c.goldPricezAPIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("goldpricez returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data GoldPricezRates
	if err := json.Unmarshal(body, &data); err != nil {
		// Some GoldPricez plans return a JSON string that contains JSON.
		var wrapped string
		if errWrapped := json.Unmarshal(body, &wrapped); errWrapped != nil {
			return nil, fmt.Errorf("invalid goldpricez payload: %w", err)
		}
		if errUnwrapped := json.Unmarshal([]byte(wrapped), &data); errUnwrapped != nil {
			return nil, fmt.Errorf("invalid wrapped goldpricez payload: %w", errUnwrapped)
		}
	}

	priceOunceRaw := data.OuncePriceUSD
	if metal == "silver" {
		priceOunceRaw = data.SilverOuncePriceAskUSD
	}

	if priceOunceRaw == "" {
		return nil, fmt.Errorf("no %s ounce price returned by goldpricez", metal)
	}

	priceOunce, err := strconv.ParseFloat(priceOunceRaw, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid %s price '%s': %w", metal, priceOunceRaw, err)
	}

	priceKg := priceOunce / troyOunceToKg

	date := time.Now()
	if data.GMTUpdated != "" {
		if parsed, err := time.Parse("02-01-2006 03:04:05 pm", data.GMTUpdated); err == nil {
			date = parsed
		}
	}

	return &model.Commodity{
		Name:      metal,
		Date:      date,
		PriceKg:   priceKg,
		Unit:      "USD/kg",
		FetchedAt: time.Now(),
	}, nil
}

func (c *Client) fetchIndustrialCommodity(ctx context.Context, commodity string) (*model.Commodity, error) {
	function := strings.ToUpper(commodity)
	if function == "ALUMINIUM" {
		function = "ALUMINUM"
	}

	if c.alphaVantageAPIKey == "" {
		return nil, errors.New("missing ALPHA_VANTAGE_API_KEY")
	}

	c.waitForAlphaVantageSlot()

	reqURL := fmt.Sprintf("%s?function=%s&apikey=%s", baseURL, function, c.alphaVantageAPIKey)
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Some industrial commodities might return errors in the body
	body, _ := io.ReadAll(resp.Body)
	if err := parseAlphaVantageError(body); err != nil {
		return nil, fmt.Errorf("alphavantage %s request failed: %w", strings.ToLower(function), err)
	}

	var data AlphaVantageCommodity
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	if len(data.Data) == 0 {
		return nil, fmt.Errorf("alphavantage returned no history data for %s", commodity)
	}

	// Take the latest value
	latest := data.Data[0]
	price, _ := strconv.ParseFloat(latest.Value, 64)
	
	var priceKg float64
	switch function {
	case "COPPER", "ALUMINUM":
		// These are in dollars per metric ton
		priceKg = price / metricTonToKg
	case "BRENT":
		// Brent is in dollars per barrel
		priceKg = price / barrelToKg
	}

	date, _ := time.Parse("2006-01-02", latest.Date)

	return &model.Commodity{
		Name:      commodity,
		Date:      date,
		PriceKg:   priceKg,
		Unit:      "USD/kg",
		FetchedAt: time.Now(),
	}, nil
}

func (c *Client) waitForAlphaVantageSlot() {
	c.alphaVantageMu.Lock()
	defer c.alphaVantageMu.Unlock()

	if !c.lastAlphaCallAt.IsZero() {
		wait := alphaVantageMinInterval - time.Since(c.lastAlphaCallAt)
		if wait > 0 {
			time.Sleep(wait)
		}
	}

	c.lastAlphaCallAt = time.Now()
}

func parseAlphaVantageError(body []byte) error {
	var apiErr alphaVantageErrorResponse
	if err := json.Unmarshal(body, &apiErr); err != nil {
		return nil
	}

	if apiErr.ErrorMessage != "" {
		return errors.New(strings.TrimSpace(apiErr.ErrorMessage))
	}
	if apiErr.Note != "" {
		return errors.New(strings.TrimSpace(apiErr.Note))
	}
	if apiErr.Information != "" && !strings.Contains(string(body), `"data"`) {
		return errors.New(strings.TrimSpace(apiErr.Information))
	}

	return nil
}
