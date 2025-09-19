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
	targetProbs, err := normalizeProbabilities(match.MatchOdds)
	if err != nil {
		return outrights.EventSolution{}, fmt.Errorf("error normalizing probabilities: %v", err)
	}

	// Create a synthetic event for the solver with the target probabilities
	event := outrights.Event{
		Name: match.Fixture,
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
		"generations":            100,
		"population_size":        8,
		"mutation_factor":        0.1,
		"elite_ratio":            0.1,
		"init_std":               0.5,
		"log_interval":           25,
		"decay_exponent":         0.5,
		"mutation_probability":   0.1,
		"debug":                  false,
		"use_league_table_init":  false,
	}

	// Create minimal inputs for solver
	events := []outrights.Event{event}
	results := []outrights.Result{} // Empty - we're solving from market prices only
	ratings := map[string]float64{} // Will be initialized by solver

	// Solve for optimal lambdas and home advantage
	solverResp := outrights.Solve(events, results, ratings, 1.0, options)

	// Extract results
	solvedRatings := solverResp["ratings"].(map[string]float64)
	homeAdvantage := solverResp["home_advantage"].(float64)
	solverError := solverResp["error"].(float64)

	// Get team names from fixture
	homeTeam, awayTeam := outrights.ParseEventName(match.Fixture)
	
	// Calculate final lambdas
	homeLambda := solvedRatings[homeTeam] + homeAdvantage
	awayLambda := solvedRatings[awayTeam]

	// Create score matrix with solved lambdas
	matrix := createScoreMatrixFromLambdas(match.Fixture, homeLambda, awayLambda)

	// Generate comprehensive outputs using existing matrix methods
	probabilities := matrix.matchOdds()
	asianHandicaps := matrix.asianHandicaps()
	totalGoals := matrix.totalGoals()

	return outrights.EventSolution{
		Fixture:        match.Fixture,
		Lambdas:        [2]float64{homeLambda, awayLambda},
		Probabilities:  [3]float64{probabilities[0], probabilities[1], probabilities[2]},
		AsianHandicaps: asianHandicaps,
		TotalGoals:     totalGoals,
		SolverError:    solverError,
	}, nil
}

// normalizeProbabilities converts match odds prices to normalized probabilities
func normalizeProbabilities(prices [3]float64) ([3]float64, error) {
	if prices[0] <= 0 || prices[1] <= 0 || prices[2] <= 0 {
		return [3]float64{}, errors.New("all prices must be positive")
	}

	// Convert prices to implied probabilities
	probs := [3]float64{
		1.0 / prices[0],
		1.0 / prices[1],
		1.0 / prices[2],
	}

	// Calculate total (overround)
	total := probs[0] + probs[1] + probs[2]

	// Normalize to sum to 1.0
	return [3]float64{
		probs[0] / total,
		probs[1] / total,
		probs[2] / total,
	}, nil
}

// createScoreMatrixFromLambdas creates a score matrix directly from lambda values
func createScoreMatrixFromLambdas(fixture string, homeLambda, awayLambda float64) *scoreMatrix {
	sm := &scoreMatrix{
		HomeLambda: homeLambda,
		AwayLambda: awayLambda,
		Rho:        0.1, // Default rho value
		N:          11,  // Default matrix size
	}
	
	sm.initMatrix()
	return sm
}

// scoreMatrix is a local version that exposes the matrix methods we need
type scoreMatrix struct {
	HomeLambda  float64
	AwayLambda  float64
	Rho         float64
	Matrix      [][]float64
	N           int
}

func (sm *scoreMatrix) initMatrix() {
	sm.Matrix = make([][]float64, sm.N)
	for i := range sm.Matrix {
		sm.Matrix[i] = make([]float64, sm.N)
	}
	
	for i := 0; i < sm.N; i++ {
		for j := 0; j < sm.N; j++ {
			homeProb := outrights.PoissonProb(sm.HomeLambda, i)
			awayProb := outrights.PoissonProb(sm.AwayLambda, j)
			adjustment := outrights.DixonColesAdjustment(i, j, sm.Rho)
			sm.Matrix[i][j] = homeProb * awayProb * adjustment
		}
	}
}

func (sm *scoreMatrix) probability(maskFn func(i, j int) bool) float64 {
	total := 0.0
	for i := 0; i < sm.N; i++ {
		for j := 0; j < sm.N; j++ {
			if maskFn(i, j) {
				total += sm.Matrix[i][j]
			}
		}
	}
	return total
}

func (sm *scoreMatrix) matchOdds() []float64 {
	homeWin := sm.probability(func(i, j int) bool { return i > j })
	draw := sm.probability(func(i, j int) bool { return i == j })
	awayWin := sm.probability(func(i, j int) bool { return i < j })
	
	total := homeWin + draw + awayWin
	return []float64{homeWin / total, draw / total, awayWin / total}
}

func (sm *scoreMatrix) asianHandicaps() [][2]interface{} {
	var handicaps [][2]interface{}
	
	maxHandicap := float64(sm.N - 1)
	for handicap := -maxHandicap + 0.5; handicap <= maxHandicap - 0.5; handicap += 0.5 {
		var probs interface{}
		
		if handicap == float64(int(handicap)) {
			homeWin := sm.probability(func(i, j int) bool { return float64(i) - float64(j) > handicap })
			draw := sm.probability(func(i, j int) bool { return float64(i) - float64(j) == handicap })
			awayWin := sm.probability(func(i, j int) bool { return float64(i) - float64(j) < handicap })
			
			total := homeWin + draw + awayWin
			probs = [3]float64{homeWin / total, draw / total, awayWin / total}
		} else {
			homeWin := sm.probability(func(i, j int) bool { return float64(i) - float64(j) > handicap })
			awayWin := sm.probability(func(i, j int) bool { return float64(i) - float64(j) < handicap })
			
			total := homeWin + awayWin
			probs = [2]float64{homeWin / total, awayWin / total}
		}
		
		handicaps = append(handicaps, [2]interface{}{handicap, probs})
	}
	
	return handicaps
}

func (sm *scoreMatrix) totalGoals() [][2]interface{} {
	var totals [][2]interface{}
	
	maxGoals := float64(sm.N*2 - 2)
	for line := 0.5; line <= maxGoals - 0.5; line += 1.0 {
		under := sm.probability(func(i, j int) bool { return float64(i + j) < line })
		over := sm.probability(func(i, j int) bool { return float64(i + j) > line })
		
		total := under + over
		probs := [2]float64{under / total, over / total}
		
		totals = append(totals, [2]interface{}{line, probs})
	}
	
	return totals
}