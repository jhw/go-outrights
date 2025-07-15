package outrights


type MatchOdds struct {
	Prices []float64 `json:"prices"`
}

type Event struct {
	Name      string    `json:"name"`
	Date      string    `json:"date"`
	Score     []int     `json:"score,omitempty"`
	MatchOdds MatchOdds `json:"match_odds"`
}

type Market struct {
	Name   string    `json:"name"`
	Payoff []float64 `json:"payoff"`
	Teams  []string  `json:"teams,omitempty"`
}

type Team struct {
	Name                   string    `json:"name"`
	Points                 float64   `json:"points"`
	GoalDifference         int       `json:"goal_difference"`
	PointsPerGameRating    float64   `json:"points_per_game_rating"`
	PoissonRating          float64   `json:"poisson_rating"`
	ExpectedSeasonPoints   float64   `json:"expected_season_points"`
	PositionProbabilities  []float64 `json:"position_probabilities"`
	TrainingEvents         int       `json:"training_events"`
	MeanTrainingError      float64   `json:"mean_training_error"`
	StdTrainingError       float64   `json:"std_training_error"`
}

type OutrightMark struct {
	Market string  `json:"market"`
	Team   string  `json:"team"`
	Mark   float64 `json:"mark"`
}

type SimulationResult struct {
	Teams           []Team         `json:"teams"`
	OutrightMarks   []OutrightMark `json:"outright_marks"`
	HomeAdvantage   float64        `json:"home_advantage"`
	SolverError     float64        `json:"solver_error"`
}

type SimulationRequest struct {
	Ratings     map[string]float64 `json:"ratings"`
	TrainingSet []Event            `json:"training_set"`
	Events      []Event            `json:"events"`
	Handicaps   map[string]float64 `json:"handicaps"`
	Markets     []Market           `json:"markets"`
	Rounds      int                `json:"rounds"`
	
	// Solver parameters
	PopulationSize        int     `json:"population_size"`
	MutationFactor        float64 `json:"mutation_factor"`
	EliteRatio            float64 `json:"elite_ratio"`
	InitStd               float64 `json:"init_std"`
	LogInterval           int     `json:"log_interval"`
	DecayExponent         float64 `json:"decay_exponent"`
	MutationProbability   float64 `json:"mutation_probability"`
	NPaths                int     `json:"n_paths"`
}