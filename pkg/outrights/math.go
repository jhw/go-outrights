package outrights

import "math"

// Mathematical utility functions

// factorial calculates the factorial of n
func factorial(n int) float64 {
	if n <= 1 {
		return 1
	}
	result := 1.0
	for i := 2; i <= n; i++ {
		result *= float64(i)
	}
	return result
}

// poissonProb calculates the Poisson probability for lambda and k
func PoissonProb(lambda float64, k int) float64 {
	return math.Pow(lambda, float64(k)) * math.Exp(-lambda) / factorial(k)
}

// dixonColesAdjustment applies Dixon-Coles adjustment for low-scoring games
func DixonColesAdjustment(i, j int, rho float64) float64 {
	switch {
	case i == 0 && j == 0:
		return 1 - (float64(i*j) * rho)
	case i == 0 && j == 1:
		return 1 + (rho / 2)
	case i == 1 && j == 0:
		return 1 + (rho / 2)
	case i == 1 && j == 1:
		return 1 - rho
	default:
		return 1
	}
}

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

// calculateTimePowerWeight calculates time power weighting for events
// Most recent event gets weight 1.0, oldest gets weight 0.0
// Power controls the decay curve: 1.0 = linear, >1 = faster decay, <1 = slower decay
func calculateTimePowerWeight(eventIndex, totalEvents int, power float64) float64 {
	if totalEvents <= 1 {
		return 1.0
	}
	// Convert index to ratio where 0 = oldest, 1 = newest
	ratio := float64(eventIndex) / float64(totalEvents-1)
	return math.Pow(ratio, power)
}