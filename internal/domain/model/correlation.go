package model

import "time"

type Correlation struct {
    ID              int64     `json:"id" db:"id"`
    CommodityA      string    `json:"commodity_a" db:"commodity_a"`
    CommodityB      string    `json:"commodity_b" db:"commodity_b"`
    CorrelationDate time.Time `json:"correlation_date" db:"correlation_date"`
    PearsonR        float64   `json:"pearson_r" db:"pearson_r"`
    SpearmanRho     float64   `json:"spearman_rho" db:"spearman_rho"`
    DataPoints      int       `json:"data_points" db:"data_points"`
    CreatedAt       time.Time `json:"created_at,omitempty" db:"created_at"`
}