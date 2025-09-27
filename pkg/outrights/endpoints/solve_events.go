package endpoints

import (
	"errors"
	"fmt"

	"github.com/jhw/go-outrights/pkg/outrights"
)

// EventMatch represents a single match for solve-events workflow
type EventMatch struct {
	Fixture       string    `json:"fixture"`        // "Home Team vs Away Team"  
	MatchOdds     [3]float64 `json:"match_odds"`     // [home_price, draw_price, away_price]
}

// SolveEventsRequest represents the input for solve-events workflow
type SolveEventsRequest struct {
	Matches       []EventMatch           `json:"matches"`
	HomeAdvantage float64                `json:"home_advantage"`
	CustomOptions map[string]interface{} `json:"custom_options,omitempty"` // Optional parameter overrides
}

// EventSolution represents the solution for a single event
type EventSolution struct {
	Fixture         string           `json:"fixture"`
	Lambdas         [2]float64       `json:"lambdas"`          // [home_lambda, away_lambda]
	Probabilities   [3]float64       `json:"probabilities"`    // [home_win, draw, away_win] 
	AsianHandicaps  [][2]interface{} `json:"asian_handicaps"`  // [(handicap, probabilities)]
	TotalGoals      [][2]interface{} `json:"total_goals"`      // [(line, [under, over])]
	SolverError     float64          `json:"solver_error"`     // Fit quality
}

// SolveEventsResult represents the output for solve-events workflow  
type SolveEventsResult struct {
	Solutions     []EventSolution `json:"solutions"`
	HomeAdvantage float64         `json:"home_advantage"`
}

// SolveEvents processes match odds and solves for lambdas and comprehensive betting markets
func SolveEvents(request SolveEventsRequest) (SolveEventsResult, error) {
	if len(request.Matches) == 0 {
		return SolveEventsResult{}, errors.New("no matches provided")
	}

	var solutions []EventSolution

	// Process each match independently using the fixed home advantage
	for _, match := range request.Matches {
		solution, err := solveIndividualMatch(match, request.HomeAdvantage, request.CustomOptions)
		if err != nil {
			return SolveEventsResult{}, fmt.Errorf("error solving match %s: %v", match.Fixture, err)
		}
		solutions = append(solutions, solution)
	}

	return SolveEventsResult{
		Solutions:     solutions,
		HomeAdvantage: request.HomeAdvantage,
	}, nil
}

// solveIndividualMatch solves for a single match using the existing solver infrastructure
func solveIndividualMatch(match EventMatch, homeAdvantage float64, customOptions map[string]interface{}) (EventSolution, error) {
	// Convert match odds prices to normalized probabilities
	matchOddsSlice := match.MatchOdds[:]
	targetProbs, err := outrights.NormalizeProbabilities(matchOddsSlice)
	if err != nil {
		return EventSolution{}, fmt.Errorf("error normalizing probabilities: %v", err)
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
	
	// Set up solver options for single match optimization with fixed home advantage
	// Optimized parameters based on stability analysis - "Larger_Pop" configuration
	// provides best balance of stability vs execution time
	options := map[string]interface{}{
		"generations":            100, // Optimal generations for accuracy
		"population_size":        20,  // Larger population for stability (was 8)
		"mutation_factor":        0.1, // Keep moderate mutation
		"elite_ratio":            0.2, // Higher elite ratio for stability (was 0.1)
		"init_std":               1.0, // Keep exploration capability
		"log_interval":           10,
		"decay_exponent":         0.5,
		"mutation_probability":   0.2, // Keep moderate mutation probability
		"debug":                  false,
		"use_league_table_init":  false, // Don't use league table init
		"home_advantage":         homeAdvantage, // Use fixed home advantage
	}

	// Override with custom options if provided
	if customOptions != nil {
		for key, value := range customOptions {
			options[key] = value
		}
		// Always ensure home advantage is set correctly
		options["home_advantage"] = homeAdvantage
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
	solverError := solverResp["error"].(float64)
	
	// Calculate final lambdas using the fixed home advantage
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

	return EventSolution{
		Fixture:        match.Fixture,
		Lambdas:        [2]float64{homeLambda, awayLambda},
		Probabilities:  [3]float64{probabilities[0], probabilities[1], probabilities[2]},
		AsianHandicaps: asianHandicaps,
		TotalGoals:     totalGoals,
		SolverError:    solverError,
	}, nil
}


