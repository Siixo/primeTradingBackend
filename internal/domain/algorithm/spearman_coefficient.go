package algorithm

import (
	"errors"
	"sort"
)

// Spearman calculates the Spearman rank correlation coefficient for two slices of float64.
// It works by ranking the data and then calculating the Pearson correlation on those ranks.
func Spearman(x, y []float64) (float64, error) {
	if len(x) != len(y) {
		return 0, errors.New("input slices must have the same length")
	}
	if len(x) < 2 {
		return 0, errors.New("insufficient data for Spearman correlation (need at least 2 points)")
	}
	
	rankX := calculateRanks(x)
	rankY := calculateRanks(y)
	
	return Pearson(rankX, rankY)
}

// calculateRanks converts a slice of values into a slice of their ranks.
// Handles ties by averaging the ranks.
func calculateRanks(data []float64) []float64 {
	n := len(data)
	type observation struct {
		value float64
		index int
	}
	
	obs := make([]observation, n)
	for i, v := range data {
		obs[i] = observation{value: v, index: i}
	}
	
	// Sort by value
	sort.Slice(obs, func(i, j int) bool {
		return obs[i].value < obs[j].value
	})
	
	ranks := make([]float64, n)
	for i := 0; i < n; {
		j := i + 1
		for j < n && obs[j].value == obs[i].value {
			j++
		}
		
		// Ties between i and j-1
		averageRank := float64(i+j+1) / 2.0
		for k := i; k < j; k++ {
			ranks[obs[k].index] = averageRank
		}
		i = j
	}
	
	return ranks
}
