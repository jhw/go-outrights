package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/jhw/go-outrights/pkg/outrights"
)

func main() {
	// Default values
	filename := "fixtures/events.json" // Default events file
	generations := 0 // 0 means use default
	npaths := 0      // 0 means use default
	rounds := 0      // 0 means use default
	debug := false   // default false
	
	// Parse named arguments
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if strings.HasPrefix(arg, "--generations=") {
			if g, err := strconv.Atoi(strings.TrimPrefix(arg, "--generations=")); err == nil {
				generations = g
			} else {
				log.Fatalf("Invalid generations: %s", arg)
			}
		} else if strings.HasPrefix(arg, "--npaths=") {
			if n, err := strconv.Atoi(strings.TrimPrefix(arg, "--npaths=")); err == nil {
				npaths = n
			} else {
				log.Fatalf("Invalid npaths: %s", arg)
			}
		} else if strings.HasPrefix(arg, "--rounds=") {
			if r, err := strconv.Atoi(strings.TrimPrefix(arg, "--rounds=")); err == nil {
				rounds = r
			} else {
				log.Fatalf("Invalid rounds: %s", arg)
			}
		} else if arg == "--debug" {
			debug = true
		} else if strings.HasPrefix(arg, "--events=") {
			filename = strings.TrimPrefix(arg, "--events=")
		} else if arg == "--help" || arg == "-h" {
			fmt.Println("Usage: go run . [--events=filename] [--generations=N] [--npaths=N] [--rounds=N] [--debug]")
			fmt.Println()
			fmt.Println("Options:")
			fmt.Println("  --events=filename    Events JSON file (default: fixtures/events.json)")
			fmt.Println("  --generations=N      Number of genetic algorithm generations (default: 1000)")
			fmt.Println("  --npaths=N          Number of simulation paths (default: 5000)")
			fmt.Println("  --rounds=N          Number of rounds each team plays (default: 1)")
			fmt.Println("  --debug             Enable debug logging for genetic algorithm")
			fmt.Println("  --help, -h          Show this help message")
			fmt.Println()
			fmt.Println("Examples:")
			fmt.Println("  go run .                                    # Use default settings")
			fmt.Println("  go run . --generations=2000 --npaths=5000   # Quick high-quality run")
			fmt.Println("  go run . --events=fixtures/other.json       # Use different events file")
			fmt.Println("  go run . --debug                           # Enable debug logging")
			os.Exit(0)
		} else {
			log.Fatalf("Unknown argument: %s", arg)
		}
	}
	
	// Read and parse the JSON file
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	
	var events []outrights.Event
	if err := json.Unmarshal(data, &events); err != nil {
		log.Fatal(err)
	}
	
	log.Printf("Processing %s with %d events", filename, len(events))
	log.Println("Starting simulation...")
	
	// Extract team names to calculate winner payoff
	var teamNames []string
	teamNamesMap := make(map[string]bool)
	for _, event := range events {
		parts := strings.Split(event.Name, " vs ")
		if len(parts) == 2 {
			teamNamesMap[parts[0]] = true
			teamNamesMap[parts[1]] = true
		}
	}
	for name := range teamNamesMap {
		teamNames = append(teamNames, name)
	}
	
	// Create Winner market (like Python: winner_payoff = f"1|{len(team_names)-1}x0")
	winnerPayoff := fmt.Sprintf("1|%dx0", len(teamNames)-1)
	markets := []outrights.Market{
		{
			Name:   "Winner",
			Payoff: winnerPayoff,
		},
	}
	
	// Create options struct with overrides
	opts := outrights.ProcessEventsFileOptions{
		Generations: generations,
		NPaths:      npaths,
		Rounds:      rounds,
		Debug:       debug,
		Markets:     markets,
	}
	
	result := outrights.ProcessEventsFile(events, opts)
	
	log.Printf("Home advantage: %.4f, Solver error: %.6f", result.HomeAdvantage, result.SolverError)
	log.Println()
	log.Println("Teams (sorted by points per game rating):")
	for _, team := range result.Teams {
		log.Printf("- %s: %.1f pts (%d played, %+d GD), PPG rating: %.3f, Poisson rating: %.3f, Expected season: %.1f pts", 
			team.Name, team.Points, team.Played, team.GoalDifference, team.PointsPerGameRating, team.PoissonRating, team.ExpectedSeasonPoints)
	}
	
	log.Println()
	log.Println("Outright marks:")
	for _, mark := range result.OutrightMarks {
		log.Printf("- %s: %.3f", mark.Team, mark.Mark)
	}
}