package outrights

import (
	"fmt"
)

// Mathematical utility functions

// sumProduct calculates the dot product of two float64 slices
func sumProduct(x, y []float64) float64 {
	if len(x) != len(y) {
		return 0
	}
	
	sum := 0.0
	for i := range x {
		sum += x[i] * y[i]
	}
	return sum
}


// NormalizeProbabilities converts betting prices to normalized probabilities
// Takes prices (e.g., [2.0, 3.5, 2.8]) and returns probabilities that sum to 1.0
func NormalizeProbabilities(prices []float64) ([]float64, error) {
	if len(prices) == 0 {
		return nil, fmt.Errorf("no prices provided")
	}
	
	// Check all prices are positive
	for i, price := range prices {
		if price <= 0 {
			return nil, fmt.Errorf("price at index %d must be positive, got %f", i, price)
		}
	}

	// Convert prices to implied probabilities
	probs := make([]float64, len(prices))
	total := 0.0
	
	for i, price := range prices {
		probs[i] = 1.0 / price
		total += probs[i]
	}

	// Normalize to sum to 1.0
	for i := range probs {
		probs[i] /= total
	}

	return probs, nil
}