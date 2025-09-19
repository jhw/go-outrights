package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jhw/go-outrights/pkg/outrights"
	"github.com/jhw/go-outrights/pkg/outrights/endpoints"
)

func main() {
	// Check for help flag
	if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		fmt.Println("Usage: go run demo-events.go")
		fmt.Println()
		fmt.Println("This demo shows the solve-events workflow using sample match odds data.")
		fmt.Println("It takes match odds prices, solves for lambdas, and generates comprehensive")
		fmt.Println("betting market outputs including Asian handicaps and total goals.")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  go run demo-events.go              # Run with sample data")
		fmt.Println("  go run demo-events.go --help       # Show this help message")
		os.Exit(0)
	}

	// Sample match data with realistic Premier League odds
	sampleMatches := []outrights.EventMatch{
		{
			Fixture:   "Liverpool vs Arsenal", 
			MatchOdds: [3]float64{2.10, 3.40, 3.50}, // [home, draw, away] prices
		},
		{
			Fixture:   "Man City vs Chelsea",
			MatchOdds: [3]float64{1.75, 3.80, 4.20},
		},
		{
			Fixture:   "Southampton vs Brighton",
			MatchOdds: [3]float64{2.80, 3.20, 2.60},
		},
		{
			Fixture:   "Arsenal vs Tottenham", 
			MatchOdds: [3]float64{1.90, 3.60, 4.00},
		},
	}

	// Create solve-events request
	request := outrights.SolveEventsRequest{
		Matches: sampleMatches,
	}

	log.Printf("Processing %d matches with solve-events workflow", len(request.Matches))
	log.Println("Starting lambda solving from match odds...")

	// Process the solve-events workflow
	result, err := endpoints.SolveEvents(request)
	if err != nil {
		log.Fatalf("Solve-events error: %v", err)
	}

	log.Printf("Solved lambdas for all matches. Average home advantage: %.4f", result.HomeAdvantage)
	log.Println()

	// Display results for each match
	for _, solution := range result.Solutions {
		fmt.Printf("=== %s ===\n", solution.Fixture)
		fmt.Printf("Lambdas: Home=%.3f, Away=%.3f\n", solution.Lambdas[0], solution.Lambdas[1])
		fmt.Printf("Solver Error: %.6f\n", solution.SolverError)
		fmt.Printf("Match Odds: Home=%.3f, Draw=%.3f, Away=%.3f\n", 
			solution.Probabilities[0], solution.Probabilities[1], solution.Probabilities[2])
		
		// Show a selection of Asian handicaps
		fmt.Printf("\nKey Asian Handicaps:\n")
		for _, handicap := range solution.AsianHandicaps {
			line := handicap[0]
			probs := handicap[1]
			
			// Only show handicaps around the even money range
			if lineFloat, ok := line.(float64); ok && lineFloat >= -2.5 && lineFloat <= 2.5 {
				if probArray, ok := probs.([2]float64); ok {
					fmt.Printf("  %+.1f: Home=%.3f, Away=%.3f\n", line, probArray[0], probArray[1])
				} else if probArray, ok := probs.([3]float64); ok {
					fmt.Printf("  %+.1f: Home=%.3f, Draw=%.3f, Away=%.3f\n", line, probArray[0], probArray[1], probArray[2])
				}
			}
		}
		
		// Show popular total goals markets
		fmt.Printf("\nPopular Total Goals Markets:\n")
		for _, total := range solution.TotalGoals {
			line := total[0]
			probs := total[1].([2]float64)
			
			// Show common lines
			if lineFloat, ok := line.(float64); ok && 
				(lineFloat == 0.5 || lineFloat == 1.5 || lineFloat == 2.5 || lineFloat == 3.5 || lineFloat == 4.5) {
				fmt.Printf("  O/U %.1f: Under=%.3f, Over=%.3f\n", line, probs[0], probs[1])
			}
		}
		
		fmt.Println()
		fmt.Println(strings.Repeat("─", 60))
		fmt.Println()
	}

	// Show JSON output for first match (for API/integration reference)
	if len(result.Solutions) > 0 {
		fmt.Println("📊 SAMPLE JSON OUTPUT (first match):")
		jsonData, err := json.MarshalIndent(result.Solutions[0], "", "  ")
		if err != nil {
			log.Printf("Error marshaling JSON: %v", err)
		} else {
			// Show first 800 characters to keep output manageable
			jsonStr := string(jsonData)
			if len(jsonStr) > 800 {
				jsonStr = jsonStr[:800] + "..."
			}
			fmt.Println(jsonStr)
		}
	}
}