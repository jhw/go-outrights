package outrights

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
)


// Simulate processes events and markets and returns simulation results
func Simulate(events []Event, markets []Market, handicaps map[string]int, opts ...SimOptions) (SimulationResult, error) {
	// Set defaults
	generations := 1000
	npaths := 5000
	rounds := 1
	trainingSetSize := 60
	populationSize := 8
	mutationFactor := 0.1
	eliteRatio := 0.1
	initStd := 0.2
	logInterval := 10
	decayExponent := 0.5
	mutationProbability := 0.1
	debug := false
	
	// Override with provided options
	if len(opts) > 0 {
		if opts[0].Generations > 0 {
			generations = opts[0].Generations
		}
		if opts[0].NPaths > 0 {
			npaths = opts[0].NPaths
		}
		if opts[0].Rounds > 0 {
			rounds = opts[0].Rounds
		}
		if opts[0].TrainingSetSize > 0 {
			trainingSetSize = opts[0].TrainingSetSize
		}
		if opts[0].PopulationSize > 0 {
			populationSize = opts[0].PopulationSize
		}
		if opts[0].MutationFactor > 0 {
			mutationFactor = opts[0].MutationFactor
		}
		if opts[0].EliteRatio > 0 {
			eliteRatio = opts[0].EliteRatio
		}
		if opts[0].InitStd > 0 {
			initStd = opts[0].InitStd
		}
		if opts[0].LogInterval > 0 {
			logInterval = opts[0].LogInterval
		}
		if opts[0].DecayExponent > 0 {
			decayExponent = opts[0].DecayExponent
		}
		if opts[0].MutationProbability > 0 {
			mutationProbability = opts[0].MutationProbability
		}
		debug = opts[0].Debug
	}
	
	// Validate that events are not empty
	if len(events) == 0 {
		return SimulationResult{}, errors.New("events cannot be empty")
	}
	
	// Extract team names from events
	teamNamesMap := make(map[string]bool)
	for _, event := range events {
		homeTeam, awayTeam := parseEventName(event.Name)
		if homeTeam != "" && awayTeam != "" {
			teamNamesMap[homeTeam] = true
			teamNamesMap[awayTeam] = true
		}
	}
	
	teamNames := make([]string, 0, len(teamNamesMap))
	for name := range teamNamesMap {
		teamNames = append(teamNames, name)
	}
	
	// Validate that team names are not empty
	if len(teamNames) == 0 {
		return SimulationResult{}, errors.New("no valid team names found in events")
	}
	
	// Validate handicaps keys against extracted team names
	for teamName := range handicaps {
		found := false
		for _, validTeam := range teamNames {
			if teamName == validTeam {
				found = true
				break
			}
		}
		if !found {
			return SimulationResult{}, fmt.Errorf("handicaps contains unknown team: %s", teamName)
		}
	}
	
	// Sort events by date and name for consistent training set selection
	sort.Slice(events, func(i, j int) bool {
		if events[i].Date == events[j].Date {
			return events[i].Name < events[j].Name
		}
		return events[i].Date < events[j].Date
	})
	
	// Split events into training and prediction sets
	// Take the last trainingSetSize events for training
	trainingCount := trainingSetSize
	if trainingCount > len(events) {
		trainingCount = len(events)
	}
	
	startIndex := len(events) - trainingCount
	trainingEvents := events[startIndex:]
	predictionEvents := events[:startIndex]
	
	// Create simulation request
	req := SimulationRequest{
		Ratings:         make(map[string]float64),
		TrainingSet:     trainingEvents,
		Events:          predictionEvents,
		Handicaps:       handicaps,
		Markets:         markets,
		PopulationSize:  populationSize,
		MutationFactor:  mutationFactor,
		EliteRatio:      eliteRatio,
		InitStd:         initStd,
		LogInterval:     logInterval,
		DecayExponent:   decayExponent,
		MutationProbability: mutationProbability,
		NPaths:          npaths,
	}
	
	// Initialize ratings to 1.0 for all teams
	for _, name := range teamNames {
		req.Ratings[name] = 1.0
	}
	
	result, err := ProcessSimulation(req, generations, rounds, debug)
	if err != nil {
		return SimulationResult{}, err
	}
	return result, nil
}

// ProcessSimulation processes a simulation request and returns results
func ProcessSimulation(req SimulationRequest, generations int, rounds int, debug bool) (SimulationResult, error) {
	teamNames := make([]string, 0, len(req.Ratings))
	for name := range req.Ratings {
		teamNames = append(teamNames, name)
	}
	sort.Strings(teamNames)
	
	// Initialize markets
	if err := InitMarkets(teamNames, req.Markets); err != nil {
		return SimulationResult{}, err
	}
	
	// Calculate league table and remaining fixtures
	leagueTable := calcLeagueTable(teamNames, req.Events, req.Handicaps)
	remainingFixtures := calcRemainingFixtures(teamNames, req.Events, rounds)
	
	// Solve for ratings
	solver := newRatingsSolver()
	
	// Create options map
	options := map[string]interface{}{
		"population_size":        req.PopulationSize,
		"mutation_factor":        req.MutationFactor,
		"elite_ratio":            req.EliteRatio,
		"init_std":               req.InitStd,
		"log_interval":           req.LogInterval,
		"decay_exponent":         req.DecayExponent,
		"mutation_probability":   req.MutationProbability,
		"generations":            generations,
		"debug":                  debug,
	}
	
	// Solve for ratings using training data
	solverResp := solver.solve(req.TrainingSet, req.Ratings, req.Events, options)
	
	// Extract results
	poissonRatings := solverResp["ratings"].(map[string]float64)
	homeAdvantage := solverResp["home_advantage"].(float64)
	solverError := solverResp["error"].(float64)
	
	// Run simulation
	simPoints := newSimPoints(leagueTable, req.NPaths)
	
	for _, eventName := range remainingFixtures {
		simPoints.simulate(eventName, poissonRatings, homeAdvantage)
	}
	
	// Calculate position probabilities
	// positionProbs := calcPositionProbabilities(simPoints, req.Markets)
	
	// Calculate PPG ratings and expected points
	ppgRatings := calcPPGRatings(teamNames, poissonRatings, homeAdvantage)
	expectedPoints := calcExpectedSeasonPoints(teamNames, req.Events, req.Handicaps, remainingFixtures, poissonRatings, homeAdvantage)
	
	// Update league table with ratings and expected points
	for i := range leagueTable {
		if ppgRating, exists := ppgRatings[leagueTable[i].Name]; exists {
			leagueTable[i].PointsPerGameRating = ppgRating
		}
		if expPoints, exists := expectedPoints[leagueTable[i].Name]; exists {
			leagueTable[i].ExpectedSeasonPoints = expPoints
		}
		if poissonRating, exists := poissonRatings[leagueTable[i].Name]; exists {
			leagueTable[i].PoissonRating = poissonRating
		}
		
		// Calculate training errors
		trainingErrors := calcTrainingErrors(teamNames, req.TrainingSet, poissonRatings, homeAdvantage)
		if errors, exists := trainingErrors[leagueTable[i].Name]; exists {
			leagueTable[i].TrainingEvents = len(errors)
			leagueTable[i].MeanTrainingError = mean(errors)
			leagueTable[i].StdTrainingError = stdDeviation(errors)
		}
	}
	
	// Sort teams by expected season points (descending)
	sort.Slice(leagueTable, func(i, j int) bool {
		return leagueTable[i].ExpectedSeasonPoints > leagueTable[j].ExpectedSeasonPoints
	})
	
	// Calculate position probabilities for markets
	positionProbabilities := calcPositionProbabilities(simPoints, req.Markets)
	
	// Assign position probabilities to teams
	if defaultProbs, exists := positionProbabilities["default"]; exists {
		for i := range leagueTable {
			if teamProbs, exists := defaultProbs[leagueTable[i].Name]; exists {
				leagueTable[i].PositionProbabilities = teamProbs
			}
		}
	}
	
	// Calculate outright marks
	outrightMarks := calcOutrightMarks(positionProbabilities, req.Markets)
	
	return SimulationResult{
		Teams:         leagueTable,
		OutrightMarks: outrightMarks,
		HomeAdvantage: homeAdvantage,
		SolverError:   solverError,
	}, nil
}



func calcTrainingErrors(teamNames []string, events []Event, ratings map[string]float64, homeAdvantage float64) map[string][]float64 {
	errors := make(map[string][]float64)
	
	// Initialize error slices
	for _, name := range teamNames {
		errors[name] = make([]float64, 0)
	}
	
	for _, event := range events {
		homeTeam, awayTeam := parseEventName(event.Name)
		if homeTeam == "" || awayTeam == "" {
			continue
		}
		
		matrix := newScoreMatrix(event.Name, ratings, homeAdvantage)
		marketProbs := extractMarketProbabilities(event)
		
		// Calculate expected points from market probabilities
		expectedHomePoints := 3*marketProbs[0] + marketProbs[1]
		expectedAwayPoints := 3*marketProbs[2] + marketProbs[1]
		
		// Calculate actual points from model
		actualHomePoints := matrix.expectedHomePoints()
		actualAwayPoints := matrix.expectedAwayPoints()
		
		// Calculate errors
		homeError := math.Abs(actualHomePoints - expectedHomePoints)
		awayError := math.Abs(actualAwayPoints - expectedAwayPoints)
		
		errors[homeTeam] = append(errors[homeTeam], homeError)
		errors[awayTeam] = append(errors[awayTeam], awayError)
	}
	
	return errors
}

func calcPPGRatings(teamNames []string, ratings map[string]float64, homeAdvantage float64) map[string]float64 {
	ppgRatings := make(map[string]float64)
	
	// Initialize all teams to 0
	for _, name := range teamNames {
		ppgRatings[name] = 0.0
	}
	
	// Calculate expected points per game for each team
	// Each team plays every other team once at home and once away
	for i, homeTeam := range teamNames {
		for j, awayTeam := range teamNames {
			if i != j {
				eventName := homeTeam + " vs " + awayTeam
				matrix := newScoreMatrix(eventName, ratings, homeAdvantage)
				
				// Add expected points for this specific game
				ppgRatings[homeTeam] += matrix.expectedHomePoints()
				ppgRatings[awayTeam] += matrix.expectedAwayPoints()
			}
		}
	}
	
	// Normalize by total number of games each team plays
	// Each team plays 2*(n-1) games total (n-1 home + n-1 away)
	totalGames := float64(2 * (len(teamNames) - 1))
	for name := range ppgRatings {
		ppgRatings[name] = ppgRatings[name] / totalGames
	}
	
	return ppgRatings
}

func calcExpectedSeasonPoints(teamNames []string, events []Event, handicaps map[string]int, 
	remainingFixtures []string, ratings map[string]float64, homeAdvantage float64) map[string]float64 {
	
	// Start with current league table points
	leagueTable := calcLeagueTable(teamNames, events, handicaps)
	expPoints := make(map[string]float64)
	
	for _, team := range leagueTable {
		expPoints[team.Name] = float64(team.Points)
	}
	
	// Add expected points from remaining fixtures
	for _, eventName := range remainingFixtures {
		matrix := newScoreMatrix(eventName, ratings, homeAdvantage)
		homeTeam, awayTeam := parseEventName(eventName)
		
		if homeTeam != "" && awayTeam != "" {
			expPoints[homeTeam] += matrix.expectedHomePoints()
			expPoints[awayTeam] += matrix.expectedAwayPoints()
		}
	}
	
	return expPoints
}

func calcPositionProbabilities(simPoints *SimPoints, markets []Market) map[string]map[string][]float64 {
	positionProbs := make(map[string]map[string][]float64)
	
	// Cache to avoid duplicate calculations for same team sets
	cache := make(map[string]map[string][]float64)
	
	// Helper function to get cache key from team names
	getCacheKey := func(teamNames []string) string {
		if teamNames == nil {
			return "default"
		}
		// Sort team names for consistent cache key
		sorted := make([]string, len(teamNames))
		copy(sorted, teamNames)
		sort.Strings(sorted)
		return strings.Join(sorted, ",")
	}
	
	// Default probabilities for all teams
	defaultKey := getCacheKey(nil)
	if _, exists := cache[defaultKey]; !exists {
		cache[defaultKey] = simPoints.positionProbabilities(nil)
	}
	positionProbs["default"] = cache[defaultKey]
	
	// Market-specific probabilities
	for _, market := range markets {
		if len(market.Teams) > 0 {
			cacheKey := getCacheKey(market.Teams)
			if _, exists := cache[cacheKey]; !exists {
				cache[cacheKey] = simPoints.positionProbabilities(market.Teams)
			}
			positionProbs[market.Name] = cache[cacheKey]
		}
	}
	
	return positionProbs
}

func sumProduct(x, y []float64) float64 {
	if len(x) != len(y) {
		return 0
	}
	
	sum := 0.0
	for i := range x {
		sum += x[i] * y[i]
	}
	return sum
}

func calcOutrightMarks(positionProbabilities map[string]map[string][]float64, markets []Market) []OutrightMark {
	var marks []OutrightMark
	
	for _, market := range markets {
		groupKey := "default"
		if len(market.Teams) > 0 {
			groupKey = market.Name
		}
		
		if groupProbs, exists := positionProbabilities[groupKey]; exists {
			for _, teamName := range market.Teams {
				if teamProbs, exists := groupProbs[teamName]; exists {
					// Convert []int to []float64 for calculation
					payoffFloat := make([]float64, len(market.ParsedPayoff))
					for i, v := range market.ParsedPayoff {
						payoffFloat[i] = float64(v)
					}
					markValue := sumProduct(teamProbs, payoffFloat)
					marks = append(marks, OutrightMark{
						Market: market.Name,
						Team:   teamName,
						Mark:   markValue,
					})
				}
			}
		}
	}
	
	return marks
}

