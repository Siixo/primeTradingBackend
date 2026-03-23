package model

import "time"

type Commodity struct {
    ID         int64     `json:"id" db:"id"`
    Name       string    `json:"commodity" db:"commodity_name"`
    Date       time.Time `json:"date" db:"date"`
    PriceKg    float64   `json:"price_kg" db:"price_kg"`
    Unit       string    `json:"unit" db:"unit"`
    FetchedAt  time.Time `json:"fetched_at,omitempty" db:"fetched_at"`
}

type CommodityStatus struct {
	Name       string    `json:"name"`
	Source     string    `json:"source"`
	Available  bool      `json:"available"`
    LastDate   *time.Time `json:"last_date,omitempty"`
	LastError  string    `json:"last_error,omitempty"`
}
