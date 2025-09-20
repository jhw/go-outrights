package outrights

import (
	"fmt"
	"math"
)

// Mathematical utility functions

// rmsError calculates the root mean square error between two slices
// 
// Note: For match probabilities [home, draw, away], we include all three values
// despite mathematical redundancy (away = 1 - home - draw). This gives team 
// strength differences (home/away outcomes) double weight vs. draw probability
// in the error calculation, which is appropriate since draws are harder to
// predict and team ability should be the primary optimization target.
func rmsError(x, y []float64) float64 {
	if len(x) != len(y) {
		return math.Inf(1)
	}
	
	sum := 0.0
	for i := range x {
		diff := x[i] - y[i]
		sum += diff * diff
	}
	
	return math.Sqrt(sum / float64(len(x)))
}

// mean calculates the arithmetic mean of a slice
func mean(x []float64) float64 {
	if len(x) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range x {
		sum += v
	}
	return sum / float64(len(x))
}

// variance calculates the sample variance of a slice
func variance(x []float64) float64 {
	if len(x) <= 1 {
		return 0
	}
	m := mean(x)
	sum := 0.0
	for _, v := range x {
		diff := v - m
		sum += diff * diff
	}
	return sum / float64(len(x)-1) // Sample variance uses n-1
}

// stdDeviation calculates the standard deviation of a slice
func stdDeviation(x []float64) float64 {
	return math.Sqrt(variance(x))
}

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