package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/jhw/go-outrights/pkg/outrights"
)

func main() {
	// Default values
	eventsFile := "fixtures/ENG1-events.json"   // Default events file
	marketsFile := "fixtures/ENG1-markets.json" // Default markets file
	generations := 0 // 0 means use default
	npaths := 0      // 0 means use default
	rounds := 0      // 0 means use default
	trainingSetSize := 0 // 0 means use default
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
		} else if strings.HasPrefix(arg, "--training-set-size=") {
			if t, err := strconv.Atoi(strings.TrimPrefix(arg, "--training-set-size=")); err == nil {
				trainingSetSize = t
			} else {
				log.Fatalf("Invalid training-set-size: %s", arg)
			}
		} else if arg == "--debug" {
			debug = true
		} else if strings.HasPrefix(arg, "--events=") {
			eventsFile = strings.TrimPrefix(arg, "--events=")
		} else if strings.HasPrefix(arg, "--markets=") {
			marketsFile = strings.TrimPrefix(arg, "--markets=")
		} else if arg == "--help" || arg == "-h" {
			fmt.Println("Usage: go run . [--events=filename] [--markets=filename] [--generations=N] [--npaths=N] [--rounds=N] [--training-set-size=N] [--debug]")
			fmt.Println()
			fmt.Println("Options:")
			fmt.Println("  --events=filename       Events JSON file (default: fixtures/ENG1-events.json)")
			fmt.Println("  --markets=filename      Markets JSON file (default: fixtures/ENG1-markets.json)")
			fmt.Println("  --generations=N         Number of genetic algorithm generations (default: 1000)")
			fmt.Println("  --npaths=N             Number of simulation paths (default: 5000)")
			fmt.Println("  --rounds=N             Number of rounds each team plays (default: 1)")
			fmt.Println("  --training-set-size=N  Number of recent events for training (default: 60)")
			fmt.Println("  --debug                Enable debug logging for genetic algorithm")
			fmt.Println("  --help, -h          Show this help message")
			fmt.Println()
			fmt.Println("Examples:")
			fmt.Println("  go run .                                    # Use default settings")
			fmt.Println("  go run . --generations=2000 --npaths=5000   # Quick high-quality run")
			fmt.Println("  go run . --events=fixtures/other.json      # Use different events file")
			fmt.Println("  go run . --markets=fixtures/other.json      # Use different markets file")
			fmt.Println("  go run . --debug                           # Enable debug logging")
			os.Exit(0)
		} else {
			log.Fatalf("Unknown argument: %s", arg)
		}
	}
	
	// Read and parse the events JSON file
	data, err := os.ReadFile(eventsFile)
	if err != nil {
		log.Fatal(err)
	}
	
	var events []outrights.Event
	if err := json.Unmarshal(data, &events); err != nil {
		log.Fatal(err)
	}
	
	// Read and parse the markets JSON file
	marketsData, err := os.ReadFile(marketsFile)
	if err != nil {
		log.Fatal(err)
	}
	
	var markets []outrights.Market
	if err := json.Unmarshal(marketsData, &markets); err != nil {
		log.Fatal(err)
	}
	
	log.Printf("Processing %s with %d events", eventsFile, len(events))
	log.Printf("Loaded %d markets from %s", len(markets), marketsFile)
	log.Println("Starting simulation...")
	
	// Create options struct with overrides
	opts := outrights.SimOptions{
		Generations:     generations,
		NPaths:          npaths,
		Rounds:          rounds,
		TrainingSetSize: trainingSetSize,
		Debug:           debug,
	}
	
	result, err := outrights.Simulate(events, markets, make(map[string]int), opts)
	if err != nil {
		log.Fatalf("Simulation error: %v", err)
	}
	
	log.Printf("Home advantage: %.4f, Solver error: %.6f", result.HomeAdvantage, result.SolverError)
	log.Println()
	log.Println("Teams (sorted by expected season points):")
	log.Println("Team            \tPts\tPlayed\tGD\tPPG\tPoisson\tExp.Pts\tEv\tError\tStdErr")
	log.Println("----            \t---\t------\t--\t---\t-------\t-------\t--\t-----\t------")
	for _, team := range result.Teams {
		teamName := team.Name
		if len(teamName) > 16 {
			teamName = teamName[:16]
		}
		log.Printf("%-16s\t%d\t%d\t%+d\t%.3f\t%.3f\t%.1f\t%d\t%.3f\t%.3f", 
			teamName, team.Points, team.Played, team.GoalDifference, team.PointsPerGameRating, team.PoissonRating, team.ExpectedSeasonPoints, team.TrainingEvents, team.MeanTrainingError, team.StdTrainingError)
	}
	
	log.Println()
	log.Println("Position Probabilities:")
	log.Println("Team            \tPosition Probabilities")
	log.Println("----            \t----------------------")
	for _, team := range result.Teams {
		teamName := team.Name
		if len(teamName) > 16 {
			teamName = teamName[:16]
		}
		log.Printf("%-16s\t", teamName)
		
		for i, prob := range team.PositionProbabilities {
			if prob > 0.001 { // Only show probabilities > 0.1%
				log.Printf("P%d:%.3f ", i+1, prob)
			}
		}
		log.Println()
	}
	
	log.Println()
	log.Println("Outright marks:")
	// Group marks by market and filter out zeros
	marketGroups := make(map[string][]outrights.OutrightMark)
	for _, mark := range result.OutrightMarks {
		if mark.Mark > 0 { // Only include non-zero marks
			marketGroups[mark.Market] = append(marketGroups[mark.Market], mark)
		}
	}
	
	// Print marks grouped by market
	for marketName, marks := range marketGroups {
		if len(marks) > 0 { // Only print markets with non-zero marks
			// Sort marks by value (descending)
			sort.Slice(marks, func(i, j int) bool {
				return marks[i].Mark > marks[j].Mark
			})
			
			log.Printf("%s:", marketName)
			for _, mark := range marks {
				log.Printf("  - %s: %.3f", mark.Team, mark.Mark)
			}
		}
	}
}