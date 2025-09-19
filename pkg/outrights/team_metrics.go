package outrights


// calcPPGRatings calculates points per game ratings for teams based on their Poisson ratings
func CalcPPGRatings(teamNames []string, ratings map[string]float64, homeAdvantage float64) map[string]float64 {
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
				matrix := newScoreMatrix(eventName, ratings, homeAdvantage)
				
				ppgRatings[homeTeam] += matrix.expectedHomePoints()
				ppgRatings[awayTeam] += matrix.expectedAwayPoints()
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