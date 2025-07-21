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
func poissonProb(lambda float64, k int) float64 {
	return math.Pow(lambda, float64(k)) * math.Exp(-lambda) / factorial(k)
}

// dixonColesAdjustment applies Dixon-Coles adjustment for low-scoring games
func dixonColesAdjustment(i, j int, rho float64) float64 {
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