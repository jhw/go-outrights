package outrights

import (
	"errors"
	"fmt"
	"sort"
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
	
	// Calculate PPG ratings 
	ppgRatings := calcPPGRatings(teamNames, poissonRatings, homeAdvantage)
	
	// Calculate expected points from the actual simulation results (not deterministic calculation)
	expectedPoints := simPoints.calculateExpectedSeasonPoints()
	
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
	
	// Calculate fixture odds for all possible team matchups
	fixtureOdds := calcAllFixtureOdds(teamNames, poissonRatings, homeAdvantage)
	
	return SimulationResult{
		Teams:         leagueTable,
		OutrightMarks: outrightMarks,
		FixtureOdds:   fixtureOdds,
		HomeAdvantage: homeAdvantage,
		SolverError:   solverError,
	}, nil
}









