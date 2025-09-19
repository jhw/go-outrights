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
	"github.com/jhw/go-outrights/pkg/outrights/endpoints"
)

func main() {
	// Default values
	resultsFile := "fixtures/ENG1-results.json"   // Default results file
	eventsFile := "fixtures/ENG1-training-events.json"   // Default training events file
	marketsFile := "fixtures/ENG1-markets.json" // Default markets file
	generations := 0 // 0 means use default
	npaths := 0      // 0 means use default
	rounds := 0      // 0 means use default
	timePowerWeighting := 0.0 // 0.0 means use default
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
		} else if strings.HasPrefix(arg, "--time-power-weighting=") {
			if t, err := strconv.ParseFloat(strings.TrimPrefix(arg, "--time-power-weighting="), 64); err == nil {
				timePowerWeighting = t
			} else {
				log.Fatalf("Invalid time-power-weighting: %s", arg)
			}
		} else if arg == "--debug" {
			debug = true
		} else if strings.HasPrefix(arg, "--results=") {
			resultsFile = strings.TrimPrefix(arg, "--results=")
		} else if strings.HasPrefix(arg, "--events=") {
			eventsFile = strings.TrimPrefix(arg, "--events=")
		} else if strings.HasPrefix(arg, "--markets=") {
			marketsFile = strings.TrimPrefix(arg, "--markets=")
		} else if arg == "--help" || arg == "-h" {
			fmt.Println("Usage: go run . [--results=filename] [--events=filename] [--markets=filename] [--generations=N] [--npaths=N] [--rounds=N] [--time-power-weighting=N] [--debug]")
			fmt.Println()
			fmt.Println("Options:")
			fmt.Println("  --results=filename      Results JSON file (default: fixtures/ENG1-results.json)")
			fmt.Println("  --events=filename       Training events JSON file (default: fixtures/ENG1-training-events.json)")
			fmt.Println("  --markets=filename      Markets JSON file (default: fixtures/ENG1-markets.json)")
			fmt.Println("  --generations=N         Number of genetic algorithm generations (default: 1000)")
			fmt.Println("  --npaths=N             Number of simulation paths (default: 5000)")
			fmt.Println("  --rounds=N             Number of rounds each team plays (default: 1)")
			fmt.Println("  --time-power-weighting=N Time power weighting (1.0=linear, >1=faster decay, <1=slower decay, default: 1.0)")
			fmt.Println("  --debug                Enable debug logging for genetic algorithm")
			fmt.Println("  --help, -h          Show this help message")
			fmt.Println()
			fmt.Println("Examples:")
			fmt.Println("  go run .                                    # Use default settings")
			fmt.Println("  go run . --generations=2000 --npaths=5000 --time-power-weighting=2.0   # High-quality run with exponential decay")
			fmt.Println("  go run . --results=fixtures/other-results.json --events=fixtures/other-events.json  # Use different files")
			fmt.Println("  go run . --markets=fixtures/other.json      # Use different markets file")
			fmt.Println("  go run . --debug                           # Enable debug logging")
			os.Exit(0)
		} else {
			log.Fatalf("Unknown argument: %s", arg)
		}
	}
	
	// Read and parse the results JSON file
	resultsData, err := os.ReadFile(resultsFile)
	if err != nil {
		log.Fatal(err)
	}
	
	var results []outrights.Result
	if err := json.Unmarshal(resultsData, &results); err != nil {
		log.Fatal(err)
	}
	
	// Read and parse the events JSON file
	eventsData, err := os.ReadFile(eventsFile)
	if err != nil {
		log.Fatal(err)
	}
	
	var events []outrights.Event
	if err := json.Unmarshal(eventsData, &events); err != nil {
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
	
	log.Printf("Processing %d results from %s and %d training events from %s", len(results), resultsFile, len(events), eventsFile)
	log.Printf("Loaded %d markets from %s", len(markets), marketsFile)
	log.Println("Starting simulation...")
	
	// Create options struct with overrides
	opts := outrights.SimOptions{
		Generations:        generations,
		NPaths:             npaths,
		Rounds:             rounds,
		TimePowerWeighting: timePowerWeighting,
		Debug:              debug,
	}
	
	result, err := endpoints.SimulateSeason(results, events, markets, make(map[string]int), opts)
	if err != nil {
		log.Fatalf("Simulation error: %v", err)
	}
	
	log.Printf("Home advantage: %.4f, Solver error: %.6f", result.HomeAdvantage, result.SolverError)
	log.Println()
	log.Println("Teams (sorted by expected season points):")
	log.Println("Team            \tPts\tPlayed\tGD\tPPG\tPoisson\tExp.Pts")
	log.Println("----            \t---\t------\t--\t---\t-------\t-------")
	for _, team := range result.Teams {
		teamName := team.Name
		if len(teamName) > 16 {
			teamName = teamName[:16]
		}
		log.Printf("%-16s\t%d\t%d\t%+d\t%.3f\t%.3f\t%.1f", 
			teamName, team.Points, team.Played, team.GoalDifference, team.PointsPerGameRating, team.PoissonRating, team.ExpectedSeasonPoints)
	}
	
	
	// Display marks table
	displayMarksTables(&result)
}

// truncateString truncates a string to maxLen characters, adding "..." if truncated
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// compactMarketName converts market names to compact versions for table display
func compactMarketName(name string, maxLen int) string {
	// Common abbreviations for betting markets
	abbreviations := map[string]string{
		"Winner":                    "Win",
		"Relegation":                "Rlg",
		"To Stay Up":               "Stay",
		"Top Two":                  "T2",
		"Top Three":                "T3", 
		"Top Four":                 "T4",
		"Top Six":                  "T6",
		"Top Seven":                "T7",
		"Top Half":                 "TH",
		"Bottom Half":              "BH",
		"Outside Top Four":         "OT4",
		"Outside Top Six":          "OT6",
		"Bottom":                   "Bot",
		"Without Big Seven":        "WB7",
		"Without Man City":         "WMC",
		"Top London Club":          "TLC",
	}
	
	// Check for exact match first
	if abbrev, exists := abbreviations[name]; exists {
		return abbrev
	}
	
	// If no abbreviation found, truncate
	return truncateString(name, maxLen)
}

// displayMarksTables displays marks in a formatted table
func displayMarksTables(result *outrights.SimulationResult) {
	// Group marks by team
	teamMarks := make(map[string]map[string]float64)
	marketNames := make(map[string]bool)
	
	for _, mark := range result.OutrightMarks {
		if mark.Mark > 0 { // Only include non-zero marks
			if teamMarks[mark.Team] == nil {
				teamMarks[mark.Team] = make(map[string]float64)
			}
			teamMarks[mark.Team][mark.Market] = mark.Mark
			marketNames[mark.Market] = true
		}
	}
	
	if len(teamMarks) == 0 {
		log.Println("No marks to display")
		return
	}
	
	// Create ordered list of markets
	var markets []string
	for market := range marketNames {
		markets = append(markets, market)
	}
	sort.Strings(markets)
	
	// Create team lookup for expected points
	teamExpPoints := make(map[string]float64)
	for _, team := range result.Teams {
		teamExpPoints[team.Name] = team.ExpectedSeasonPoints
	}
	
	log.Println()
	log.Println("📊 MARK VALUES TABLE")
	
	// Calculate table width
	headerWidth := 13 + 1 + 6 // Team name + space + Expected Points columns
	for range markets {
		headerWidth += 7 // 6 chars + 1 space per market
	}
	
	// Print top border
	for i := 0; i < headerWidth; i++ {
		fmt.Print("═")
	}
	fmt.Println()
	
	// Print header
	fmt.Print("Team          ExpPts")
	for _, market := range markets {
		compactName := compactMarketName(market, 6)
		fmt.Printf(" %6s", compactName)
	}
	fmt.Println()
	
	// Print header separator
	fmt.Print("─────────────  ─────")
	for range markets {
		fmt.Print(" ──────")
	}
	fmt.Println()
	
	// Create list of teams sorted by expected points (descending)
	type teamData struct {
		name string
		expPoints float64
	}
	
	var teams []teamData
	for teamName := range teamMarks {
		teams = append(teams, teamData{
			name: teamName,
			expPoints: teamExpPoints[teamName],
		})
	}
	
	sort.Slice(teams, func(i, j int) bool {
		return teams[i].expPoints > teams[j].expPoints
	})
	
	// Print data rows
	for _, team := range teams {
		teamName := truncateString(team.name, 13)
		fmt.Printf("%-13s %6.1f", teamName, team.expPoints)
		
		for _, market := range markets {
			if mark, exists := teamMarks[team.name][market]; exists {
				fmt.Printf(" %6.3f", mark)
			} else {
				fmt.Print("       ") // Empty cell for no mark (7 spaces to match " %6.3f")
			}
		}
		fmt.Println()
	}
	
	// Print bottom border
	for i := 0; i < headerWidth; i++ {
		fmt.Print("═")
	}
	fmt.Println()
}