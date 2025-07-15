package main

import (
	"encoding/json"
	"testing"
	"time"
)

func TestSimulateExample(t *testing.T) {
	// Example request similar to demo.py
	req := SimulationRequest{
		Ratings: map[string]float64{
			"Team A": 2.5,
			"Team B": 1.8,
			"Team C": 3.2,
			"Team D": 1.5,
		},
		TrainingSet: []Event{
			{
				Name: "Team A vs Team B",
				Date: time.Now().AddDate(0, 0, -7),
				MatchOdds: MatchOdds{
					Prices: []float64{2.1, 3.2, 3.8},
				},
			},
			{
				Name: "Team C vs Team D",
				Date: time.Now().AddDate(0, 0, -5),
				MatchOdds: MatchOdds{
					Prices: []float64{1.4, 4.5, 8.0},
				},
			},
		},
		Events: []Event{
			{
				Name: "Team A vs Team B",
				Date: time.Now().AddDate(0, 0, -7),
				MatchOdds: MatchOdds{
					Prices: []float64{2.1, 3.2, 3.8},
				},
			},
		},
		Handicaps: map[string]float64{},
		Markets: []Market{
			{
				Name:   "Winner",
				Payoff: []float64{1, 0, 0, 0},
			},
		},
		Rounds:              1,
		MaxIterations:       100, // Reduced for testing
		PopulationSize:      4,   // Reduced for testing
		MutationFactor:      0.1,
		EliteRatio:          0.2,
		InitStd:             1.0,
		LogInterval:         10,
		DecayExponent:       0.5,
		MutationProbability: 0.3,
		ExplorationInterval: 50,
		NExplorationPoints:  10,
		ExcellentError:      0.03,
		MaxError:            0.05,
		NPaths:              100, // Reduced for testing
	}
	
	result := simulate(req)
	
	// Basic validation
	if len(result.Teams) != 4 {
		t.Errorf("Expected 4 teams, got %d", len(result.Teams))
	}
	
	if result.HomeAdvantage <= 0 {
		t.Errorf("Expected positive home advantage, got %f", result.HomeAdvantage)
	}
	
	if result.SolverError < 0 {
		t.Errorf("Expected non-negative solver error, got %f", result.SolverError)
	}
	
	// Check that teams have ratings
	for _, team := range result.Teams {
		if team.PoissonRating <= 0 {
			t.Errorf("Team %s has invalid rating: %f", team.Name, team.PoissonRating)
		}
		
		if len(team.PositionProbabilities) != 4 {
			t.Errorf("Team %s has wrong number of position probabilities: %d", team.Name, len(team.PositionProbabilities))
		}
	}
	
	// Check outright marks
	if len(result.OutrightMarks) == 0 {
		t.Error("Expected outright marks, got none")
	}
	
	// Verify JSON serialization works
	_, err := json.Marshal(result)
	if err != nil {
		t.Errorf("Failed to marshal result to JSON: %v", err)
	}
}

func TestScoreMatrix(t *testing.T) {
	ratings := map[string]float64{
		"Team A": 2.5,
		"Team B": 1.8,
	}
	
	matrix := NewScoreMatrix("Team A vs Team B", ratings, 1.2)
	
	// Test match odds
	odds := matrix.MatchOdds()
	if len(odds) != 3 {
		t.Errorf("Expected 3 odds, got %d", len(odds))
	}
	
	// Odds should sum to 1.0
	sum := odds[0] + odds[1] + odds[2]
	if abs(sum-1.0) > 0.001 {
		t.Errorf("Match odds don't sum to 1.0: %f", sum)
	}
	
	// Test expected points
	homePoints := matrix.ExpectedHomePoints()
	awayPoints := matrix.ExpectedAwayPoints()
	
	if homePoints <= 0 || homePoints > 3 {
		t.Errorf("Invalid home expected points: %f", homePoints)
	}
	
	if awayPoints <= 0 || awayPoints > 3 {
		t.Errorf("Invalid away expected points: %f", awayPoints)
	}
	
	// Test simulation
	scores := matrix.SimulateScores(100)
	if len(scores) != 100 {
		t.Errorf("Expected 100 simulated scores, got %d", len(scores))
	}
	
	for i, score := range scores {
		if len(score) != 2 {
			t.Errorf("Score %d has wrong length: %d", i, len(score))
		}
		if score[0] < 0 || score[1] < 0 {
			t.Errorf("Score %d has negative goals: %v", i, score)
		}
	}
}

func TestGeneticAlgorithm(t *testing.T) {
	// Simple optimization problem: minimize (x-2)^2 + (y-3)^2
	objectiveFn := func(params []float64) float64 {
		x, y := params[0], params[1]
		return (x-2)*(x-2) + (y-3)*(y-3)
	}
	
	options := map[string]interface{}{
		"max_iterations":       50,
		"population_size":      10,
		"mutation_factor":      0.1,
		"elite_ratio":          0.2,
		"init_std":             1.0,
		"log_interval":         10,
		"decay_exponent":       0.5,
		"mutation_probability": 0.3,
		"excellent_error":      0.01,
		"max_error":            0.1,
	}
	
	ga := NewGeneticAlgorithm(options)
	x0 := []float64{0, 0}
	bounds := [][]float64{{-5, 5}, {-5, 5}}
	
	solution, fitness := ga.Optimize(objectiveFn, x0, bounds)
	
	// Should find optimum near (2, 3)
	if abs(solution[0]-2) > 0.5 || abs(solution[1]-3) > 0.5 {
		t.Errorf("GA didn't find good solution: %v (fitness: %f)", solution, fitness)
	}
	
	if fitness > 0.1 {
		t.Errorf("GA didn't achieve good fitness: %f", fitness)
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}