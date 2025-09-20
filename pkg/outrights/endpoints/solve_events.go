package endpoints

import (
	"errors"
	"fmt"

	"github.com/jhw/go-outrights/pkg/outrights"
)

// SolveEvents processes match odds and solves for lambdas and comprehensive betting markets
func SolveEvents(request outrights.SolveEventsRequest) (outrights.SolveEventsResult, error) {
	if len(request.Matches) == 0 {
		return outrights.SolveEventsResult{}, errors.New("no matches provided")
	}

	var solutions []outrights.EventSolution

	// Process each match independently
	for _, match := range request.Matches {
		solution, err := solveIndividualMatch(match)
		if err != nil {
			return outrights.SolveEventsResult{}, fmt.Errorf("error solving match %s: %v", match.Fixture, err)
		}
		solutions = append(solutions, solution)
	}

	// Calculate average home advantage from all solutions
	avgHomeAdvantage := 0.0
	for _, solution := range solutions {
		// Extract home advantage from lambdas (difference between home and away after accounting for team strength)
		avgHomeAdvantage += solution.Lambdas[0] - solution.Lambdas[1]
	}
	avgHomeAdvantage /= float64(len(solutions))

	return outrights.SolveEventsResult{
		Solutions:     solutions,
		HomeAdvantage: avgHomeAdvantage,
	}, nil
}

// solveIndividualMatch solves for a single match using the existing solver infrastructure
func solveIndividualMatch(match outrights.EventMatch) (outrights.EventSolution, error) {
	// Convert match odds prices to normalized probabilities
	matchOddsSlice := match.MatchOdds[:]
	targetProbs, err := outrights.NormalizeProbabilities(matchOddsSlice)
	if err != nil {
		return outrights.EventSolution{}, fmt.Errorf("error normalizing probabilities: %v", err)
	}

	// Get team names from fixture
	homeTeam, awayTeam := outrights.ParseEventName(match.Fixture)
	
	// Create unique team names for this match to avoid conflicts
	uniqueHomeTeam := fmt.Sprintf("%s_match", homeTeam)
	uniqueAwayTeam := fmt.Sprintf("%s_match", awayTeam)
	uniqueFixture := fmt.Sprintf("%s vs %s", uniqueHomeTeam, uniqueAwayTeam)

	// Create a synthetic event for the solver with the target probabilities
	event := outrights.Event{
		Name: uniqueFixture,
		MatchOdds: outrights.MatchOdds{
			Prices: []float64{
				1.0 / targetProbs[0], // Convert back to prices for consistency
				1.0 / targetProbs[1],
				1.0 / targetProbs[2],
			},
		},
	}
	
	// Set up solver options for single match optimization
	options := map[string]interface{}{
		"generations":            50,  // Fewer generations for single match
		"population_size":        8,
		"mutation_factor":        0.1,
		"elite_ratio":            0.1,
		"init_std":               1.0,  // Higher variance for exploration
		"log_interval":           10,
		"decay_exponent":         0.5,
		"mutation_probability":   0.2,  // Higher mutation for exploration
		"debug":                  false,
		"use_league_table_init":  false, // Don't use league table init
	}

	// Initialize ratings with reasonable starting values for both teams
	ratings := map[string]float64{
		uniqueHomeTeam: 1.0,  // Start with equal ratings
		uniqueAwayTeam: 1.0,
	}

	// Create minimal inputs for solver
	events := []outrights.Event{event}
	results := []outrights.Result{} // Empty - we're solving from market prices only

	// Solve for optimal lambdas and home advantage
	solverResp := outrights.Solve(events, results, ratings, 1.0, options)

	// Extract results
	solvedRatings := solverResp["ratings"].(map[string]float64)
	homeAdvantage := solverResp["home_advantage"].(float64)
	solverError := solverResp["error"].(float64)
	
	// Calculate final lambdas
	homeLambda := solvedRatings[uniqueHomeTeam] + homeAdvantage
	awayLambda := solvedRatings[uniqueAwayTeam]

	// Create score matrix with solved lambdas using the existing ScoreMatrix from matrix.go
	ratings = map[string]float64{
		homeTeam: homeLambda - homeAdvantage, // Extract base rating
		awayTeam: awayLambda,
	}
	matrix := outrights.NewScoreMatrix(match.Fixture, ratings, homeAdvantage)

	// Generate comprehensive outputs using existing matrix methods
	probabilities := matrix.MatchOdds()
	asianHandicaps := matrix.AsianHandicaps()
	totalGoals := matrix.TotalGoals()

	return outrights.EventSolution{
		Fixture:        match.Fixture,
		Lambdas:        [2]float64{homeLambda, awayLambda},
		Probabilities:  [3]float64{probabilities[0], probabilities[1], probabilities[2]},
		AsianHandicaps: asianHandicaps,
		TotalGoals:     totalGoals,
		SolverError:    solverError,
	}, nil
}


