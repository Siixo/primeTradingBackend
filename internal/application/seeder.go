package application

import (
	"backend/internal/domain/model"
	"backend/internal/domain/repository"
	"log"
	"math/rand"
	"time"
)

func RunCommoditySeeder(repo repository.CommodityRepository) {
	// Simple check: do we have any data from the last 48 hours?
	// If not, we probably need a seed or help.
	recent, err := repo.HasRecentData()
	if err == nil && recent {
		return // We have recent data, don't re-seed
	}

	log.Println("No recent data found. Seeding 1 year of historical commodity data for POC...")

	startDate := time.Now().AddDate(-1, 0, 0)
	daysToSeed := 365

	// Starting prices in USD/kg
	prices := map[string]float64{
		"gold":     75000.0,
		"silver":   850.0,
		"copper":   9.5,
		"aluminum": 2.4,
		"brent":    0.65, // ~$88/barrel
	}

	for i := 0; i < daysToSeed; i++ {
		currentDate := startDate.AddDate(0, 0, i)

		// Create some market trends to make the correlations interesting
		// Gold and Silver follow trendA
		trendA := 1.0 + (rand.Float64()*0.02 - 0.01) // -1% to +1% daily shift
		// Copper, Aluminum, and Brent follow trendB
		trendB := 1.0 + (rand.Float64()*0.03 - 0.015) // -1.5% to +1.5% daily shift

		// Apply trends with some independent noise
		prices["gold"] = prices["gold"] * trendA * (1.0 + (rand.Float64()*0.01 - 0.005))
		prices["silver"] = prices["silver"] * trendA * (1.0 + (rand.Float64()*0.015 - 0.0075))
		
		prices["copper"] = prices["copper"] * trendB * (1.0 + (rand.Float64()*0.01 - 0.005))
		prices["aluminum"] = prices["aluminum"] * trendB * (1.0 + (rand.Float64()*0.01 - 0.005))
		prices["brent"] = prices["brent"] * trendB * (1.0 + (rand.Float64()*0.02 - 0.01))

		// Save each commodity for this day
		for name, price := range prices {
			err := repo.Save(model.Commodity{
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

	log.Println("✅ Successfully seeded 365 days of historical data.")
}
