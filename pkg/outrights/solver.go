package outrights

import (
	"log"
	"math"
	"math/rand"
	"sort"
	"sync"
)

const (
	RatingMin = 0.0
	RatingMax = 6.0
	HomeAdvantageMin = 0.0
	HomeAdvantageMax = 1.5
)

type GeneticAlgorithm struct {
	maxIterations       int
	populationSize      int
	mutationFactor      float64
	eliteRatio          float64
	initStd             float64
	logInterval         int
	decayExponent       float64
	mutationProbability float64
	debug               bool
}

type Individual struct {
	Genes   []float64
	Fitness float64
}

type Population []Individual

func (p Population) Len() int           { return len(p) }
func (p Population) Less(i, j int) bool { return p[i].Fitness < p[j].Fitness }
func (p Population) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func newGeneticAlgorithm(options map[string]interface{}) *GeneticAlgorithm {
	ga := &GeneticAlgorithm{
		maxIterations:       options["generations"].(int),
		populationSize:      options["population_size"].(int),
		mutationFactor:      options["mutation_factor"].(float64),
		eliteRatio:          options["elite_ratio"].(float64),
		initStd:             options["init_std"].(float64),
		logInterval:         options["log_interval"].(int),
		decayExponent:       options["decay_exponent"].(float64),
		mutationProbability: options["mutation_probability"].(float64),
		debug:               options["debug"].(bool),
	}
	return ga
}

func (ga *GeneticAlgorithm) optimize(objectiveFn func([]float64) float64, x0 []float64, bounds [][]float64) ([]float64, float64) {
	nParams := len(x0)
	nElite := int(math.Max(1, float64(ga.populationSize)*ga.eliteRatio))
	
	log.Printf("Starting parallel genetic algorithm: %d generations, %d candidates per generation", ga.maxIterations, ga.populationSize)
	
	// Initialize population
	population := make(Population, ga.populationSize)
	
	// First individual: use provided initial guess
	population[0] = Individual{
		Genes: make([]float64, nParams),
	}
	copy(population[0].Genes, x0)
	
	// Remaining individuals: random within bounds
	for i := 1; i < ga.populationSize; i++ {
		genes := make([]float64, nParams)
		for j := 0; j < nParams; j++ {
			if bounds != nil && len(bounds[j]) == 2 {
				genes[j] = bounds[j][0] + rand.Float64()*(bounds[j][1]-bounds[j][0])
			} else {
				genes[j] = x0[j] + rand.NormFloat64()*ga.initStd
			}
		}
		population[i] = Individual{Genes: genes}
	}
	
	bestFitness := math.Inf(1)
	var bestSolution []float64
	
	for generation := 0; generation < ga.maxIterations; generation++ {
		// Evaluate fitness in parallel
		var wg sync.WaitGroup
		for i := range population {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				population[idx].Fitness = objectiveFn(population[idx].Genes)
			}(i)
		}
		wg.Wait()
		
		// Sort by fitness
		sort.Sort(population)
		
		// Update best solution
		if population[0].Fitness < bestFitness {
			bestFitness = population[0].Fitness
			bestSolution = make([]float64, nParams)
			copy(bestSolution, population[0].Genes)
		}
		
		// Log progress
		if ga.debug && (generation%ga.logInterval == 0 || generation == ga.maxIterations-1) {
			avgFitness := 0.0
			for _, ind := range population {
				avgFitness += ind.Fitness
			}
			avgFitness /= float64(len(population))
			
			timeRemaining := float64(ga.maxIterations-generation) / float64(ga.maxIterations)
			currentMutation := ga.mutationFactor * math.Pow(timeRemaining, 0.5)
			
			log.Printf("Generation %d/%d: best=%.6f, avg=%.6f, mutation=%.4f", 
				generation+1, ga.maxIterations, bestFitness, avgFitness, currentMutation)
		}
		
		
		// Create new population
		newPopulation := make(Population, ga.populationSize)
		
		// Keep elite unchanged
		for i := 0; i < nElite; i++ {
			newPopulation[i] = Individual{
				Genes:   make([]float64, nParams),
				Fitness: population[i].Fitness,
			}
			copy(newPopulation[i].Genes, population[i].Genes)
		}
		
		// Generate offspring
		timeRemaining := float64(ga.maxIterations-generation) / float64(ga.maxIterations)
		decayFactor := math.Pow(timeRemaining, ga.decayExponent)
		currentMutationFactor := ga.mutationFactor * decayFactor
		
		for i := nElite; i < ga.populationSize; i++ {
			// Select random elite parent
			parentIdx := rand.Intn(nElite)
			parent := population[parentIdx]
			
			// Create offspring
			offspring := Individual{
				Genes: make([]float64, nParams),
			}
			copy(offspring.Genes, parent.Genes)
			
			// Apply mutations
			for j := 0; j < nParams; j++ {
				if rand.Float64() < ga.mutationProbability {
					mutation := rand.NormFloat64() * currentMutationFactor
					offspring.Genes[j] += mutation
					
					// Clamp to bounds
					if bounds != nil && len(bounds[j]) == 2 {
						offspring.Genes[j] = math.Max(bounds[j][0], math.Min(bounds[j][1], offspring.Genes[j]))
					}
				}
			}
			
			newPopulation[i] = offspring
		}
		
		population = newPopulation
	}
	
	log.Printf("Parallel optimization completed. Final objective value: %.6f", bestFitness)
	return bestSolution, bestFitness
}

type RatingsSolver struct{}

func newRatingsSolver() *RatingsSolver {
	return &RatingsSolver{}
}

func (rs *RatingsSolver) calcError(events []Event, ratings map[string]float64, homeAdvantage, timePowerWeighting float64) float64 {
	var totalWeightedError float64
	var totalWeight float64
	
	for i, event := range events {
		matrix := newScoreMatrix(event.Name, ratings, homeAdvantage)
		modelOdds := matrix.matchOdds()
		marketProbs := extractMarketProbabilities(event)
		
		error := rmsError(modelOdds, marketProbs)
		weight := calculateTimePowerWeight(i, len(events), timePowerWeighting)
		
		totalWeightedError += error * weight
		totalWeight += weight
	}
	
	if totalWeight == 0 {
		return 0
	}
	return totalWeightedError / totalWeight
}

func (rs *RatingsSolver) optimizeRatings(events []Event, ratings map[string]float64, homeAdvantage, timePowerWeighting float64, options map[string]interface{}) {
	log.Printf("Starting ratings optimization for %d teams with fixed home advantage %.6f", len(ratings), homeAdvantage)
	
	teamNames := make([]string, 0, len(ratings))
	for name := range ratings {
		teamNames = append(teamNames, name)
	}
	sort.Strings(teamNames)
	
	// Create initial solution and bounds
	x0 := make([]float64, len(teamNames))
	bounds := make([][]float64, len(teamNames))
	for i, name := range teamNames {
		x0[i] = ratings[name]
		bounds[i] = []float64{RatingMin, RatingMax}
	}
	
	// Objective function
	objectiveFn := func(params []float64) float64 {
		tempRatings := make(map[string]float64)
		for i, name := range teamNames {
			tempRatings[name] = params[i]
		}
		return rs.calcError(events, tempRatings, homeAdvantage, timePowerWeighting)
	}
	
	// Optimize
	ga := newGeneticAlgorithm(options)
	solution, fitness := ga.optimize(objectiveFn, x0, bounds)
	
	// Update ratings
	for i, name := range teamNames {
		ratings[name] = solution[i]
	}
	
	log.Printf("Ratings optimization completed with final error: %.6f", fitness)
}

func (rs *RatingsSolver) optimizeRatingsAndBias(events []Event, ratings map[string]float64, timePowerWeighting float64, options map[string]interface{}) float64 {
	log.Printf("Starting joint optimization of %d team ratings and home advantage", len(ratings))
	
	teamNames := make([]string, 0, len(ratings))
	for name := range ratings {
		teamNames = append(teamNames, name)
	}
	sort.Strings(teamNames)
	
	// Create initial solution and bounds
	x0 := make([]float64, len(teamNames)+1)
	bounds := make([][]float64, len(teamNames)+1)
	
	for i, name := range teamNames {
		x0[i] = ratings[name]
		bounds[i] = []float64{RatingMin, RatingMax}
	}
	
	// Home advantage parameter
	x0[len(teamNames)] = (HomeAdvantageMin + HomeAdvantageMax) / 2
	bounds[len(teamNames)] = []float64{HomeAdvantageMin, HomeAdvantageMax}
	
	// Objective function
	objectiveFn := func(params []float64) float64 {
		tempRatings := make(map[string]float64)
		for i, name := range teamNames {
			tempRatings[name] = params[i]
		}
		homeAdvantage := params[len(teamNames)]
		return rs.calcError(events, tempRatings, homeAdvantage, timePowerWeighting)
	}
	
	// Optimize
	ga := newGeneticAlgorithm(options)
	solution, fitness := ga.optimize(objectiveFn, x0, bounds)
	
	// Update ratings and get home advantage
	for i, name := range teamNames {
		ratings[name] = solution[i]
	}
	homeAdvantage := solution[len(teamNames)]
	
	log.Printf("Joint optimization completed with final error: %.6f, home advantage: %.6f", fitness, homeAdvantage)
	return homeAdvantage
}

func (rs *RatingsSolver) initializeRatingsFromLeagueTable(teamNames []string, events []Event) map[string]float64 {
	leagueTable := calcLeagueTable(teamNames, events, make(map[string]int))
	
	// Check if we have any results
	hasResults := false
	for _, team := range leagueTable {
		if team.Points > 0 {
			hasResults = true
			break
		}
	}
	
	if !hasResults {
		log.Printf("No match events found, using random initialization")
		ratings := make(map[string]float64)
		for _, name := range teamNames {
			ratings[name] = RatingMin + rand.Float64()*(RatingMax-RatingMin)
		}
		return ratings
	}
	
	// Map league position to rating range
	ratingSpan := RatingMax - RatingMin
	ratings := make(map[string]float64)
	
	for i, team := range leagueTable {
		// Linear mapping: best team gets max rating, worst gets min rating
		positionRatio := 0.0
		if len(leagueTable) > 1 {
			positionRatio = float64(i) / float64(len(leagueTable)-1)
		}
		rating := RatingMax - (positionRatio * ratingSpan)
		ratings[team.Name] = rating
	}
	
	topTeam := leagueTable[0]
	log.Printf("Initialized ratings from league table: %s (%d pts) = %.2f", 
		topTeam.Name, topTeam.Points, ratings[topTeam.Name])
	
	return ratings
}

func (rs *RatingsSolver) solve(events []Event, ratings map[string]float64, timePowerWeighting float64, options map[string]interface{}) map[string]interface{} {
	log.Printf("Starting solver with %d events, max_iterations=%d", len(events), options["generations"].(int))
	
	// Initialize ratings from league table if events with scores are provided
	useLeagueTableInit := true
	if val, exists := options["use_league_table_init"]; exists {
		useLeagueTableInit = val.(bool)
	}
	if useLeagueTableInit {
		// Check if we have any events with scores for initialization
		hasScores := false
		for _, event := range events {
			if len(event.Score) > 0 {
				hasScores = true
				break
			}
		}
		
		if hasScores {
			teamNames := make([]string, 0, len(ratings))
			for name := range ratings {
				teamNames = append(teamNames, name)
			}
			sort.Strings(teamNames)
			
			leagueTableRatings := rs.initializeRatingsFromLeagueTable(teamNames, events)
			for name, rating := range leagueTableRatings {
				ratings[name] = rating
			}
		}
	}
	
	var homeAdvantage float64
	
	// Check if home advantage is provided
	if ha, exists := options["home_advantage"]; exists {
		homeAdvantage = ha.(float64)
		rs.optimizeRatings(events, ratings, homeAdvantage, timePowerWeighting, options)
	} else {
		homeAdvantage = rs.optimizeRatingsAndBias(events, ratings, timePowerWeighting, options)
	}
	
	error := rs.calcError(events, ratings, homeAdvantage, timePowerWeighting)
	log.Printf("Solver completed with final error: %.6f", error)
	
	return map[string]interface{}{
		"ratings":        ratings,
		"home_advantage": homeAdvantage,
		"error":          error,
	}
}

