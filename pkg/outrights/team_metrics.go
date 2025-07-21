package outrights

import (
	"math"
)

// calcTrainingErrors calculates training errors for each team based on actual vs predicted outcomes
func calcTrainingErrors(teamNames []string, events []Event, ratings map[string]float64, homeAdvantage float64) map[string][]float64 {
	errors := make(map[string][]float64)
	
	// Initialize error slices
	for _, name := range teamNames {
		errors[name] = make([]float64, 0)
	}
	
	for _, event := range events {
		if len(event.MatchOdds.Prices) != 3 {
			continue
		}
		
		homeTeam, awayTeam := parseEventName(event.Name)
		if homeTeam == "" || awayTeam == "" {
			continue
		}
		
		// Skip if teams not in our list
		if _, exists := ratings[homeTeam]; !exists {
			continue
		}
		if _, exists := ratings[awayTeam]; !exists {
			continue
		}
		
		// Extract market probabilities
		marketProbs := extractMarketProbabilities(event)
		
		// Calculate expected points from market
		expectedHomePoints := 3*marketProbs[0] + marketProbs[1]
		expectedAwayPoints := 3*marketProbs[2] + marketProbs[1]
		
		// Calculate actual points from model
		matrix := newScoreMatrix(event.Name, ratings, homeAdvantage)
		actualHomePoints := matrix.expectedHomePoints()
		actualAwayPoints := matrix.expectedAwayPoints()
		
		// Calculate errors
		homeError := math.Abs(actualHomePoints - expectedHomePoints)
		awayError := math.Abs(actualAwayPoints - expectedAwayPoints)
		
		errors[homeTeam] = append(errors[homeTeam], homeError)
		errors[awayTeam] = append(errors[awayTeam], awayError)
	}
	
	return errors
}

// calcPPGRatings calculates points per game ratings for teams based on their Poisson ratings
func calcPPGRatings(teamNames []string, ratings map[string]float64, homeAdvantage float64) map[string]float64 {
	ppgRatings := make(map[string]float64)
	
	// Initialize ratings
	for _, name := range teamNames {
		ppgRatings[name] = 0.0
	}
	
	// Calculate expected points for each team against every other team
	for _, homeTeam := range teamNames {
		for _, awayTeam := range teamNames {
			if homeTeam != awayTeam {
				eventName := homeTeam + " vs " + awayTeam
				matrix := newScoreMatrix(eventName, ratings, homeAdvantage)
				
				ppgRatings[homeTeam] += matrix.expectedHomePoints()
				ppgRatings[awayTeam] += matrix.expectedAwayPoints()
			}
		}
	}
	
	// Normalize by total number of games each team plays
	totalGames := float64(len(teamNames) - 1)
	for name := range ppgRatings {
		ppgRatings[name] /= totalGames
	}
	
	return ppgRatings
}