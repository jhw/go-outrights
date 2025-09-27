package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jhw/go-outrights/pkg/outrights/endpoints"
)

func main() {
	// Check for help flag
	if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		fmt.Println("Usage: go run analyze-events-stability.go")
		fmt.Println()
		fmt.Println("This demo analyzes stability of the events solver by running the same")
		fmt.Println("event 16 times and examining variation in lambdas and solver error.")
		os.Exit(0)
	}

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

	// Create solve-events request with home advantage of 0.3
	request := endpoints.SolveEventsRequest{
		Matches:       matches,
		HomeAdvantage: 0.3,
	}

	log.Printf("Processing %d identical matches with solve-events workflow", len(request.Matches))
	log.Println("Target probabilities: Home=0.500, Draw=0.300, Away=0.200")
	log.Printf("Home advantage: %.3f", request.HomeAdvantage)
	log.Println("Starting lambda solving...")

	// Process the solve-events workflow
	result, err := endpoints.SolveEvents(request)
	if err != nil {
		log.Fatalf("Solve-events error: %v", err)
	}

	log.Println()
	log.Println("Results (should be identical but may vary due to genetic algorithm):")
	log.Println()

	// Print header
	fmt.Printf("%-10s %12s %12s %15s\n", "Run", "Home Lambda", "Away Lambda", "Solver Error")
	fmt.Println("-------------------------------------------------------")

	// Collect statistics
	var homeLambdas, awayLambdas, errors []float64

	// Display results for each match
	for i, solution := range result.Solutions {
		fmt.Printf("%-10d %12.6f %12.6f %15.8f\n", i+1, 
			solution.Lambdas[0], solution.Lambdas[1], solution.SolverError)
		
		homeLambdas = append(homeLambdas, solution.Lambdas[0])
		awayLambdas = append(awayLambdas, solution.Lambdas[1])
		errors = append(errors, solution.SolverError)
	}

	// Calculate and display statistics
	fmt.Println()
	fmt.Println("Statistical Analysis:")
	fmt.Println("====================")
	
	// Home lambda stats
	homeMin, homeMax := minMax(homeLambdas)
	homeMean, homeStd := meanStd(homeLambdas)
	fmt.Printf("Home Lambda - Mean: %.6f, Std: %.6f, Min: %.6f, Max: %.6f, Range: %.6f\n",
		homeMean, homeStd, homeMin, homeMax, homeMax-homeMin)

	// Away lambda stats  
	awayMin, awayMax := minMax(awayLambdas)
	awayMean, awayStd := meanStd(awayLambdas)
	fmt.Printf("Away Lambda - Mean: %.6f, Std: %.6f, Min: %.6f, Max: %.6f, Range: %.6f\n",
		awayMean, awayStd, awayMin, awayMax, awayMax-awayMin)

	// Error stats
	errorMin, errorMax := minMax(errors)
	errorMean, errorStd := meanStd(errors)
	fmt.Printf("Solver Error - Mean: %.8f, Std: %.8f, Min: %.8f, Max: %.8f, Range: %.8f\n",
		errorMean, errorStd, errorMin, errorMax, errorMax-errorMin)

	fmt.Println()
	fmt.Printf("Expected Home Lambda (base + home advantage): %.3f + %.3f = %.3f\n", 
		homeMean-0.3, 0.3, homeMean)
	fmt.Printf("Expected Away Lambda (should match away mean): %.3f\n", awayMean)
}

// Helper functions for statistics
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