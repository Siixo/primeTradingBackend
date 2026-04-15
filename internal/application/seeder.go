package application

import (
	"backend/internal/domain/model"
	"backend/internal/domain/repository"
	"context"
	"log"
	"math/rand"
	"time"
)

func RunCommoditySeeder(ctx context.Context, repo repository.CommodityRepository) {
	recent, err := repo.HasRecentData(ctx)
	if err == nil && recent {
		return
	}

	log.Println("No recent data found. Seeding 1 year of historical commodity data for POC...")

	startDate := time.Now().AddDate(-1, 0, 0)
	daysToSeed := 365

	prices := map[string]float64{
		"gold":     75000.0,
		"silver":   850.0,
		"copper":   9.5,
		"aluminum": 2.4,
		"brent":    0.65,
	}

	for i := 0; i < daysToSeed; i++ {
		currentDate := startDate.AddDate(0, 0, i)

		trendA := 1.0 + (rand.Float64()*0.02 - 0.01)
		trendB := 1.0 + (rand.Float64()*0.03 - 0.015)

		prices["gold"] = prices["gold"] * trendA * (1.0 + (rand.Float64()*0.01 - 0.005))
		prices["silver"] = prices["silver"] * trendA * (1.0 + (rand.Float64()*0.015 - 0.0075))
		
		prices["copper"] = prices["copper"] * trendB * (1.0 + (rand.Float64()*0.01 - 0.005))
		prices["aluminum"] = prices["aluminum"] * trendB * (1.0 + (rand.Float64()*0.01 - 0.005))
		prices["brent"] = prices["brent"] * trendB * (1.0 + (rand.Float64()*0.02 - 0.01))

		for name, price := range prices {
			err := repo.Save(ctx, model.Commodity{
				Name:      name,
				Date:      currentDate,
				PriceKg:   price,
				Unit:      "USD/kg",
				FetchedAt: time.Now(),
			})
			if err != nil {
				log.Printf("Seeder failed to save %s: %v", name, err)
			}
		}
	}

	log.Println("Successfully seeded 365 days of historical data.")
}
