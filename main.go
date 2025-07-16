package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/jhw/go-outrights/pkg/outrights"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run . <filename> [--generations=N] [--npaths=N] [--rounds=N] [--debug]")
	}
	
	filename := os.Args[1]
	generations := 0 // 0 means use default
	npaths := 0      // 0 means use default
	rounds := 0      // 0 means use default
	debug := false   // default false
	
	// Parse named arguments
	for i := 2; i < len(os.Args); i++ {
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
	
	// Create options struct with overrides
	opts := outrights.ProcessEventsFileOptions{
		Generations: generations,
		NPaths:      npaths,
		Rounds:      rounds,
		Debug:       debug,
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