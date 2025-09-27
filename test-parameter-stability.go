package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jhw/go-outrights/pkg/outrights/endpoints"
)

// ParameterSet represents a configuration to test
type ParameterSet struct {
	Name               string
	Generations        int
	PopulationSize     int
	MutationFactor     float64
	EliteRatio         float64
	InitStd            float64
	MutationProb       float64
	DecayExponent      float64
}

// TestResult holds the results of a parameter test
type TestResult struct {
	ParameterSet
	HomeLambdaMean  float64
	HomeLambdaStd   float64
	AwayLambdaMean  float64
	AwayLambdaStd   float64
	ErrorMean       float64
	ErrorStd        float64
	ExecutionTime   time.Duration
}

func main() {
	// Check for help flag
	if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		fmt.Println("Usage: go run test-parameter-stability.go")
		fmt.Println()
		fmt.Println("This script tests different parameter configurations for the events solver")
		fmt.Println("to find the most stable set of parameters.")
		os.Exit(0)
	}

	// Define parameter sets to test
	parameterSets := []ParameterSet{
		{
			Name:           "Default",
			Generations:    100,
			PopulationSize: 8,
			MutationFactor: 0.1,
			EliteRatio:     0.1,
			InitStd:        1.0,
			MutationProb:   0.2,
			DecayExponent:  0.5,
		},
		{
			Name:           "Larger_Pop",
			Generations:    100,
			PopulationSize: 20,
			MutationFactor: 0.1,
			EliteRatio:     0.2,  // Keep 4 elites
			InitStd:        1.0,
			MutationProb:   0.2,
			DecayExponent:  0.5,
		},
		{
			Name:           "Lower_Mutation",
			Generations:    100,
			PopulationSize: 20,
			MutationFactor: 0.05,
			EliteRatio:     0.2,
			InitStd:        0.5,   // Lower init variance
			MutationProb:   0.1,   // Lower mutation probability
			DecayExponent:  0.5,
		},
		{
			Name:           "More_Stable",
			Generations:    150,   // More generations
			PopulationSize: 30,    // Larger population
			MutationFactor: 0.03,  // Very low mutation
			EliteRatio:     0.3,   // Keep more elites
			InitStd:        0.2,   // Very low initial variance
			MutationProb:   0.05,  // Very low mutation probability
			DecayExponent:  0.8,   // Slower decay
		},
		{
			Name:           "Conservative",
			Generations:    200,   // Even more generations
			PopulationSize: 50,    // Large population
			MutationFactor: 0.01,  // Minimal mutation
			EliteRatio:     0.4,   // Keep many elites
			InitStd:        0.1,   // Minimal initial variance
			MutationProb:   0.02,  // Minimal mutation probability
			DecayExponent:  1.0,   // No decay
		},
	}

	var results []TestResult

	// Test each parameter set
	for _, paramSet := range parameterSets {
		fmt.Printf("\n=== Testing Parameter Set: %s ===\n", paramSet.Name)
		result := testParameterSet(paramSet)
		results = append(results, result)
	}

	// Display comparison
	fmt.Println("\n" + strings.Repeat("=", 120))
	fmt.Println("PARAMETER STABILITY COMPARISON")
	fmt.Println(strings.Repeat("=", 120))
	
	fmt.Printf("%-15s %10s %10s %10s %10s %10s %10s %10s\n", 
		"Config", "HomeLStd", "AwayLStd", "ErrorStd", "Time(ms)", "Pop", "Gen", "MutProb")
	fmt.Println(strings.Repeat("-", 120))

	for _, result := range results {
		fmt.Printf("%-15s %10.6f %10.6f %10.8f %10.0f %10d %10d %10.3f\n",
			result.Name,
			result.HomeLambdaStd,
			result.AwayLambdaStd, 
			result.ErrorStd,
			float64(result.ExecutionTime.Nanoseconds())/1e6,
			result.PopulationSize,
			result.Generations,
			result.MutationProb)
	}

	// Find best configuration
	bestConfig := results[0]
	bestScore := results[0].HomeLambdaStd + results[0].AwayLambdaStd + results[0].ErrorStd*1000  // Weight error more

	for _, result := range results[1:] {
		score := result.HomeLambdaStd + result.AwayLambdaStd + result.ErrorStd*1000
		if score < bestScore {
			bestScore = score
			bestConfig = result
		}
	}

	fmt.Printf("\n🏆 BEST CONFIGURATION: %s\n", bestConfig.Name)
	fmt.Printf("   Home Lambda Std: %.6f\n", bestConfig.HomeLambdaStd)
	fmt.Printf("   Away Lambda Std: %.6f\n", bestConfig.AwayLambdaStd) 
	fmt.Printf("   Error Std: %.8f\n", bestConfig.ErrorStd)
	fmt.Printf("   Execution Time: %.0fms\n", float64(bestConfig.ExecutionTime.Nanoseconds())/1e6)
}

func testParameterSet(paramSet ParameterSet) TestResult {
	// Convert probabilities 0.5/0.3/0.2 to prices: 1/prob
	homePrice := 1.0 / 0.5  // = 2.0
	drawPrice := 1.0 / 0.3  // = 3.33...
	awayPrice := 1.0 / 0.2  // = 5.0

	// Create the same match 16 times
	var matches []endpoints.EventMatch
	for i := 0; i < 16; i++ {
		matches = append(matches, endpoints.EventMatch{
			Fixture:   fmt.Sprintf("TestTeamA vs TestTeamB (Run %d)", i+1),
			MatchOdds: [3]float64{homePrice, drawPrice, awayPrice},
		})
	}

	// Create modified solve-events request
	request := endpoints.SolveEventsRequest{
		Matches:       matches,
		HomeAdvantage: 0.3,
	}

	// Temporarily modify the default parameters by creating a custom version
	// We'll have to modify the solve_events.go file or create a custom version

	fmt.Printf("Testing with: Pop=%d, Gen=%d, MutFac=%.3f, Elite=%.2f, InitStd=%.2f, MutProb=%.3f\n", 
		paramSet.PopulationSize, paramSet.Generations, paramSet.MutationFactor, 
		paramSet.EliteRatio, paramSet.InitStd, paramSet.MutationProb)

	// Temporarily create a custom version for testing
	startTime := time.Now()

	// We need to create a custom version that accepts these parameters
	// For now, let's create a test version that allows parameter passing
	result := runCustomSolveEvents(request, paramSet)

	executionTime := time.Since(startTime)

	// Collect statistics
	var homeLambdas, awayLambdas, errors []float64
	for _, solution := range result.Solutions {
		homeLambdas = append(homeLambdas, solution.Lambdas[0])
		awayLambdas = append(awayLambdas, solution.Lambdas[1])
		errors = append(errors, solution.SolverError)
	}

	// Calculate statistics
	homeMean, homeStd := meanStd(homeLambdas)
	awayMean, awayStd := meanStd(awayLambdas)
	errorMean, errorStd := meanStd(errors)

	fmt.Printf("Results: HomeLStd=%.6f, AwayLStd=%.6f, ErrorStd=%.8f, Time=%.0fms\n",
		homeStd, awayStd, errorStd, float64(executionTime.Nanoseconds())/1e6)

	return TestResult{
		ParameterSet:    paramSet,
		HomeLambdaMean:  homeMean,
		HomeLambdaStd:   homeStd,
		AwayLambdaMean:  awayMean,
		AwayLambdaStd:   awayStd,
		ErrorMean:       errorMean,
		ErrorStd:        errorStd,
		ExecutionTime:   executionTime,
	}
}

// Custom solve function that accepts parameter overrides
func runCustomSolveEvents(request endpoints.SolveEventsRequest, params ParameterSet) endpoints.SolveEventsResult {
	// Set custom options based on the parameter set
	customOptions := map[string]interface{}{
		"generations":          params.Generations,
		"population_size":      params.PopulationSize,
		"mutation_factor":      params.MutationFactor,
		"elite_ratio":          params.EliteRatio,
		"init_std":            params.InitStd,
		"mutation_probability": params.MutationProb,
		"decay_exponent":      params.DecayExponent,
		"log_interval":        10,
		"debug":               false,
		"use_league_table_init": false,
	}
	
	// Add custom options to the request
	request.CustomOptions = customOptions
	
	result, err := endpoints.SolveEvents(request)
	if err != nil {
		log.Fatalf("Solve-events error: %v", err)
	}
	
	return result
}

// Helper functions for statistics (same as before)
func minMax(values []float64) (float64, float64) {
	if len(values) == 0 {
		return 0, 0
	}
	min, max := values[0], values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return min, max
}

func meanStd(values []float64) (float64, float64) {
	if len(values) == 0 {
		return 0, 0
	}
	
	// Calculate mean
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))
	
	// Calculate standard deviation
	sumSquaredDiff := 0.0
	for _, v := range values {
		diff := v - mean
		sumSquaredDiff += diff * diff
	}
	std := 0.0
	if len(values) > 1 {
		std = sumSquaredDiff / float64(len(values)-1)
		// Take square root for standard deviation
		std = sqrt(std)
	}
	
	return mean, std
}

// Simple square root implementation
func sqrt(x float64) float64 {
	if x == 0 {
		return 0
	}
	// Newton's method
	guess := x / 2
	for i := 0; i < 20; i++ { // 20 iterations should be enough
		nextGuess := (guess + x/guess) / 2
		if abs(guess-nextGuess) < 1e-10 {
			break
		}
		guess = nextGuess
	}
	return guess
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}