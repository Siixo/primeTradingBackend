package algorithm

import (
	"errors"
	"math"
)

// Pearson calculates the Pearson correlation coefficient for two slices of float64.
// It assumes the slices have the same length and at least 2 elements.
func Pearson(x, y []float64) (float64, error) {
	n := len(x)
	if n < 2 {
		return 0, errors.New("insufficient data for Pearson correlation (need at least 2 points)")
	}

	var sumX, sumY, sumXY, sumX2, sumY2 float64
	for i := 0; i < n; i++ {
		sumX += x[i]
		sumY += y[i]
		sumXY += x[i] * y[i]
		sumX2 += x[i] * x[i]
		sumY2 += y[i] * y[i]
	}

	num := float64(n)*sumXY - sumX*sumY
	den := math.Sqrt((float64(n)*sumX2 - sumX*sumX) * (float64(n)*sumY2 - sumY*sumY))

	if den == 0 {
		return 0, nil
	}

	return num / den, nil
}
