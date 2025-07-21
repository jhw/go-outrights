package outrights

import (
	"sort"
	"strings"
)

// calcPositionProbabilities calculates position probabilities for each market using simulation results
func calcPositionProbabilities(simPoints *SimPoints, markets []Market) map[string]map[string][]float64 {
	positionProbs := make(map[string]map[string][]float64)
	
	// Cache to avoid duplicate calculations for same team sets
	cache := make(map[string]map[string][]float64)
	
	// Helper function to get cache key from team names
	getCacheKey := func(teamNames []string) string {
		if teamNames == nil {
			return "default"
		}
		// Sort team names for consistent cache key
		sorted := make([]string, len(teamNames))
		copy(sorted, teamNames)
		sort.Strings(sorted)
		return strings.Join(sorted, ",")
	}
	
	// Default probabilities for all teams
	defaultKey := getCacheKey(nil)
	if _, exists := cache[defaultKey]; !exists {
		cache[defaultKey] = simPoints.positionProbabilities(nil)
	}
	positionProbs["default"] = cache[defaultKey]
	
	// Market-specific probabilities
	for _, market := range markets {
		if len(market.Teams) > 0 {
			cacheKey := getCacheKey(market.Teams)
			if _, exists := cache[cacheKey]; !exists {
				cache[cacheKey] = simPoints.positionProbabilities(market.Teams)
			}
			positionProbs[market.Name] = cache[cacheKey]
		}
	}
	
	return positionProbs
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

// calcOutrightMarks calculates outright marks for each market based on position probabilities
func calcOutrightMarks(positionProbabilities map[string]map[string][]float64, markets []Market) []OutrightMark {
	var marks []OutrightMark
	
	for _, market := range markets {
		groupKey := "default"
		if len(market.Teams) > 0 {
			groupKey = market.Name
		}
		
		if groupProbs, exists := positionProbabilities[groupKey]; exists {
			for _, teamName := range market.Teams {
				if teamProbs, exists := groupProbs[teamName]; exists {
					// Convert []float64 to []float64 for calculation
					payoffFloat := make([]float64, len(market.ParsedPayoff))
					for i, v := range market.ParsedPayoff {
						payoffFloat[i] = v
					}
					markValue := sumProduct(teamProbs, payoffFloat)
					marks = append(marks, OutrightMark{
						Market: market.Name,
						Team:   teamName,
						Mark:   markValue,
					})
				}
			}
		}
	}
	
	return marks
}