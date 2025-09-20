package endpoints

import (
	"errors"
	"fmt"
	"sort"
	
	"github.com/jhw/go-outrights/pkg/outrights"
)


// SimulateSeason processes events and markets and returns simulation results
func SimulateSeason(results []outrights.Result, events []outrights.Event, markets []outrights.Market, handicaps map[string]int, opts ...outrights.SimOptions) (outrights.SimulationResult, error) {
	// Set defaults
	generations := 1000
	npaths := 5000
	rounds := 1
	timePowerWeighting := 1.0
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
		if opts[0].TimePowerWeighting > 0 {
			timePowerWeighting = opts[0].TimePowerWeighting
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
		return outrights.SimulationResult{}, errors.New("events cannot be empty")
	}
	
	if len(results) == 0 {
		return outrights.SimulationResult{}, errors.New("results cannot be empty")
	}
	
	// Extract team names from results
	teamNamesMap := make(map[string]bool)
	for _, result := range results {
		homeTeam, awayTeam := outrights.ParseEventName(result.Name)
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
		return outrights.SimulationResult{}, errors.New("no valid team names found in results")
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
			return outrights.SimulationResult{}, fmt.Errorf("handicaps contains unknown team: %s", teamName)
		}
	}
	
	// Sort events by date and name for consistent time-based weighting
	sort.Slice(events, func(i, j int) bool {
		if events[i].Date == events[j].Date {
			return events[i].Name < events[j].Name
		}
		return events[i].Date < events[j].Date
	})
	
	// Create simulation request
	req := outrights.SimulationRequest{
		Ratings:         make(map[string]float64),
		Results:         results,
		Events:          events,
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
		TimePowerWeighting: timePowerWeighting,
	}
	
	// Initialize ratings to 1.0 for all teams
	for _, name := range teamNames {
		req.Ratings[name] = 1.0
	}
	
	result, err := ProcessSimulation(req, generations, rounds, debug)
	if err != nil {
		return outrights.SimulationResult{}, err
	}
	return result, nil
}

// ProcessSimulation processes a simulation request and returns results
func ProcessSimulation(req outrights.SimulationRequest, generations int, rounds int, debug bool) (outrights.SimulationResult, error) {
	teamNames := make([]string, 0, len(req.Ratings))
	for name := range req.Ratings {
		teamNames = append(teamNames, name)
	}
	sort.Strings(teamNames)
	
	// Initialize markets
	if err := outrights.InitMarkets(teamNames, req.Markets); err != nil {
		return outrights.SimulationResult{}, err
	}
	
	// Calculate league table and remaining fixtures
	leagueTable := outrights.CalcLeagueTable(teamNames, req.Results, req.Handicaps)
	remainingFixtures := outrights.CalcRemainingFixtures(teamNames, req.Results, rounds)
	
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
	
	// Solve for ratings using events for training and results for initialization
	solverResp := outrights.Solve(req.Events, req.Results, req.Ratings, req.TimePowerWeighting, options)
	
	// Extract results
	poissonRatings := solverResp["ratings"].(map[string]float64)
	homeAdvantage := solverResp["home_advantage"].(float64)
	solverError := solverResp["error"].(float64)
	
	// Run simulation
	simPoints := outrights.NewSimPoints(leagueTable, req.NPaths)
	
	for _, eventName := range remainingFixtures {
		simPoints.Simulate(eventName, poissonRatings, homeAdvantage)
	}
	
	// Calculate position probabilities
	// positionProbs := calcPositionProbabilities(simPoints, req.Markets)
	
	// Calculate PPG ratings 
	ppgRatings := calcPPGRatings(teamNames, poissonRatings, homeAdvantage)
	
	// Calculate expected points from the actual simulation results (not deterministic calculation)
	expectedPoints := calculateExpectedSeasonPoints(simPoints)
	
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
		
	}
	
	// Sort teams by expected season points (descending)
	sort.Slice(leagueTable, func(i, j int) bool {
		return leagueTable[i].ExpectedSeasonPoints > leagueTable[j].ExpectedSeasonPoints
	})
	
	// Calculate position probabilities for markets
	positionProbabilities := outrights.CalcPositionProbabilities(simPoints, req.Markets)
	
	// Assign position probabilities to teams
	if defaultProbs, exists := positionProbabilities["default"]; exists {
		for i := range leagueTable {
			if teamProbs, exists := defaultProbs[leagueTable[i].Name]; exists {
				leagueTable[i].PositionProbabilities = teamProbs
			}
		}
	}
	
	// Calculate outright marks
	outrightMarks := outrights.CalcOutrightMarks(positionProbabilities, req.Markets)
	
	// Calculate fixture odds for all possible team matchups
	fixtureOdds := outrights.CalcAllFixtureOdds(teamNames, poissonRatings, homeAdvantage)
	
	return outrights.SimulationResult{
		Teams:         leagueTable,
		OutrightMarks: outrightMarks,
		FixtureOdds:   fixtureOdds,
		HomeAdvantage: homeAdvantage,
		SolverError:   solverError,
	}, nil
}

// calcPPGRatings calculates points per game ratings for teams based on their Poisson ratings
func calcPPGRatings(teamNames []string, ratings map[string]float64, homeAdvantage float64) map[string]float64 {
	ppgRatings := make(map[string]float64)
	
	// Initialize ratings
	for _, name := range teamNames {
		ppgRatings[name] = 0.0
	}
	
	// Calculate expected points for each team against every other team
	for _, homeTeam := range teamNames {
		for _, awayTeam := range teamNames {
			if homeTeam != awayTeam {
				eventName := homeTeam + " vs " + awayTeam
				matrix := outrights.NewScoreMatrix(eventName, ratings, homeAdvantage)
				odds := matrix.MatchOdds()
				
				// Expected points: home wins = 3 pts, draw = 1 pt each, away win = 0/3 pts
				ppgRatings[homeTeam] += 3*odds[0] + odds[1]  // 3*home_win + 1*draw
				ppgRatings[awayTeam] += 3*odds[2] + odds[1]  // 3*away_win + 1*draw
			}
		}
	}
	
	// Normalize by total number of games each team plays
	// Each team plays against every other team both home and away
	totalGames := float64(2 * (len(teamNames) - 1))
	for name := range ppgRatings {
		ppgRatings[name] /= totalGames
	}
	
	return ppgRatings
}

// calculateExpectedSeasonPoints calculates expected season points from the actual simulation results
func calculateExpectedSeasonPoints(simPoints *outrights.SimPoints) map[string]float64 {
	teamNames, points, nPaths := simPoints.GetSimulationData()
	expectedPoints := make(map[string]float64)
	
	for i, teamName := range teamNames {
		totalPoints := 0.0
		for path := 0; path < nPaths; path++ {
			totalPoints += float64(points[i][path])
		}
		expectedPoints[teamName] = totalPoints / float64(nPaths)
	}
	
	return expectedPoints
}