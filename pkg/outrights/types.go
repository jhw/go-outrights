package outrights


type MatchOdds struct {
	Prices []float64 `json:"prices"`
}

type Result struct {
	Name  string `json:"name"`
	Date  string `json:"date"`
	Score []int  `json:"score"`
}

type Event struct {
	Name      string    `json:"name"`
	Date      string    `json:"date"`
	MatchOdds MatchOdds `json:"match_odds"`
}

type Market struct {
	Name         string    `json:"name"`
	Payoff       string    `json:"payoff"`
	ParsedPayoff []float64 `json:"-"` // Parsed version, not serialized
	Teams        []string  `json:"teams,omitempty"`
	Include      []string  `json:"include,omitempty"`
	Exclude      []string  `json:"exclude,omitempty"`
}

type Team struct {
	Name                   string    `json:"name"`
	Points                 int       `json:"points"`
	GoalDifference         int       `json:"goal_difference"`
	Played                 int       `json:"played"`
	PointsPerGameRating    float64   `json:"points_per_game_rating"`
	PoissonRating          float64   `json:"poisson_rating"`
	ExpectedSeasonPoints   float64   `json:"expected_season_points"`
	PositionProbabilities  []float64 `json:"position_probabilities"`
}

type OutrightMark struct {
	Market string  `json:"market"`
	Team   string  `json:"team"`
	Mark   float64 `json:"mark"`
}

type FixtureOdds struct {
	Fixture         string          `json:"fixture"`          // "Home Team vs Away Team"
	Probabilities   [3]float64      `json:"probabilities"`    // [home_win, draw, away_win]
	AsianHandicaps  [][2]interface{} `json:"asian_handicaps"`  // [(handicap, [home_win, away_win] or [home_win, draw, away_win])]
	TotalGoals      [][2]interface{} `json:"total_goals"`      // [(line, [under, over])]
	Lambdas         [2]float64      `json:"lambdas"`          // [home_lambda, away_lambda]
}

