package adapters

import (
	"backend/internal/application"
	"backend/internal/domain/model"
	"fmt"
)

type CompositeProvider struct {
	providers map[string]application.MetalPriceProvider
	defaultPr application.MetalPriceProvider
}

func NewCompositeProvider(defaultPr application.MetalPriceProvider) *CompositeProvider {
	return &CompositeProvider{
		providers: make(map[string]application.MetalPriceProvider),
		defaultPr: defaultPr,
	}
}

func (c *CompositeProvider) Register(commodity string, provider application.MetalPriceProvider) {
	c.providers[commodity] = provider
}

func (c *CompositeProvider) FetchPrice(commodity string) (*model.Commodity, error) {
	if pr, ok := c.providers[commodity]; ok {
		return pr.FetchPrice(commodity)
	}
	if c.defaultPr != nil {
		return c.defaultPr.FetchPrice(commodity)
	}
	return nil, fmt.Errorf("no provider registered for commodity: %s", commodity)
}

func (c *CompositeProvider) FetchCommodity() (*model.Commodity, error) {
	if c.defaultPr != nil {
		return c.defaultPr.FetchCommodity()
	}
	return nil, fmt.Errorf("no default provider for FetchCommodity")
}
