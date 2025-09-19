package outrights

import (
	"fmt"
	"sort"
	"strings"
)

// calcPositionProbabilities calculates position probabilities for each market using simulation results
func CalcPositionProbabilities(simPoints *SimPoints, markets []Market) map[string]map[string][]float64 {
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

// calcOutrightMarks calculates outright marks for each market based on position probabilities
func CalcOutrightMarks(positionProbabilities map[string]map[string][]float64, markets []Market) []OutrightMark {
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

// calcAllFixtureOdds calculates match odds for all possible team matchups in the league
func CalcAllFixtureOdds(teamNames []string, ratings map[string]float64, homeAdvantage float64) []FixtureOdds {
	var fixtureOdds []FixtureOdds
	
	// Generate odds for all team combinations (n * (n-1) fixtures)
	for i, homeTeam := range teamNames {
		for j, awayTeam := range teamNames {
			if i != j { // Skip same team vs same team
				fixture := fmt.Sprintf("%s vs %s", homeTeam, awayTeam)
				
				// Create score matrix for this matchup
				matrix := newScoreMatrix(fixture, ratings, homeAdvantage)
				
				// Get match probabilities [home_win, draw, away_win]
				probabilities := matrix.matchOdds()
				
				// Get Asian handicaps
				asianHandicaps := matrix.asianHandicaps()
				
				// Get total goals over/under
				totalGoals := matrix.totalGoals()
				
				// Get lambda values
				lambdas := [2]float64{matrix.HomeLambda, matrix.AwayLambda}
				
				fixtureOdds = append(fixtureOdds, FixtureOdds{
					Fixture:        fixture,
					Probabilities:  [3]float64{probabilities[0], probabilities[1], probabilities[2]},
					AsianHandicaps: asianHandicaps,
					TotalGoals:     totalGoals,
					Lambdas:        lambdas,
				})
			}
		}
	}
	
	// Sort by fixture name for consistent output
	sort.Slice(fixtureOdds, func(i, j int) bool {
		return fixtureOdds[i].Fixture < fixtureOdds[j].Fixture
	})
	
	return fixtureOdds
}