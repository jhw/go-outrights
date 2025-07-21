package outrights

import (
	"math/rand"
	"sort"
)

type SimPoints struct {
	NPaths         int
	TeamNames      []string
	Points         [][]float64
	GoalDifference [][]float64
}

func newSimPoints(leagueTable []Team, nPaths int) *SimPoints {
	sp := &SimPoints{
		NPaths:         nPaths,
		TeamNames:      make([]string, len(leagueTable)),
		Points:         make([][]float64, len(leagueTable)),
		GoalDifference: make([][]float64, len(leagueTable)),
	}
	
	for i, team := range leagueTable {
		sp.TeamNames[i] = team.Name
		sp.Points[i] = make([]float64, nPaths)
		sp.GoalDifference[i] = make([]float64, nPaths)
		
		// Initialize with current points + noise
		pointsWithNoise := float64(team.Points) + NoiseMultiplier*(rand.Float64()-0.5)
		// Initialize with current goal difference + noise
		gdWithNoise := float64(team.GoalDifference) + NoiseMultiplier*(rand.Float64()-0.5)
		
		for j := 0; j < nPaths; j++ {
			sp.Points[i][j] = pointsWithNoise
			sp.GoalDifference[i][j] = gdWithNoise
		}
	}
	
	return sp
}

func (sp *SimPoints) getTeamIndex(teamName string) int {
	for i, name := range sp.TeamNames {
		if name == teamName {
			return i
		}
	}
	return -1
}

func (sp *SimPoints) simulate(eventName string, ratings map[string]float64, homeAdvantage float64) {
	matrix := newScoreMatrix(eventName, ratings, homeAdvantage)
	scores := matrix.simulateScores(sp.NPaths)
	sp.updateEvent(eventName, scores)
}

func (sp *SimPoints) updateHomeTeam(teamName string, scores [][]int) {
	teamIndex := sp.getTeamIndex(teamName)
	if teamIndex == -1 {
		return
	}
	
	for i, score := range scores {
		homeGoals := score[0]
		awayGoals := score[1]
		
		// Calculate points
		points := 0.0
		if homeGoals > awayGoals {
			points = 3.0
		} else if homeGoals == awayGoals {
			points = 1.0
		}
		
		// Calculate goal difference
		goalDifference := float64(homeGoals - awayGoals)
		
		// Update points and goal difference separately
		sp.Points[teamIndex][i] += points
		sp.GoalDifference[teamIndex][i] += goalDifference
	}
}

func (sp *SimPoints) updateAwayTeam(teamName string, scores [][]int) {
	teamIndex := sp.getTeamIndex(teamName)
	if teamIndex == -1 {
		return
	}
	
	for i, score := range scores {
		homeGoals := score[0]
		awayGoals := score[1]
		
		// Calculate points
		points := 0.0
		if awayGoals > homeGoals {
			points = 3.0
		} else if homeGoals == awayGoals {
			points = 1.0
		}
		
		// Calculate goal difference
		goalDifference := float64(awayGoals - homeGoals)
		
		// Update points and goal difference separately
		sp.Points[teamIndex][i] += points
		sp.GoalDifference[teamIndex][i] += goalDifference
	}
}

func (sp *SimPoints) updateEvent(eventName string, scores [][]int) {
	homeTeam, awayTeam := parseEventName(eventName)
	sp.updateHomeTeam(homeTeam, scores)
	sp.updateAwayTeam(awayTeam, scores)
}

func (sp *SimPoints) positionProbabilities(teamNames []string) map[string][]float64 {
	if teamNames == nil {
		teamNames = sp.TeamNames
	}
	
	// Create mask for selected teams
	selectedIndices := make([]int, 0, len(teamNames))
	for _, name := range teamNames {
		if idx := sp.getTeamIndex(name); idx >= 0 {
			selectedIndices = append(selectedIndices, idx)
		}
	}
	
	if len(selectedIndices) == 0 {
		return make(map[string][]float64)
	}
	
	// Extract points and goal difference for selected teams
	selectedPoints := make([][]float64, len(selectedIndices))
	selectedGoalDifference := make([][]float64, len(selectedIndices))
	for i, idx := range selectedIndices {
		selectedPoints[i] = sp.Points[idx]
		selectedGoalDifference[i] = sp.GoalDifference[idx]
	}
	
	// Calculate positions for each path
	positions := make([][]int, len(selectedIndices))
	for i := range positions {
		positions[i] = make([]int, sp.NPaths)
	}
	
	for path := 0; path < sp.NPaths; path++ {
		// Create array of team data for this path
		teamData := make([]struct {
			TeamIndex int
			Points    float64
			GD        float64
		}, len(selectedIndices))
		
		for i := range selectedIndices {
			teamData[i] = struct {
				TeamIndex int
				Points    float64
				GD        float64
			}{
				TeamIndex: i,
				Points:    selectedPoints[i][path],
				GD:        selectedGoalDifference[i][path],
			}
		}
		
		// Sort by points first, then by goal difference (both descending) to get positions
		sort.Slice(teamData, func(i, j int) bool {
			if teamData[i].Points == teamData[j].Points {
				return teamData[i].GD > teamData[j].GD
			}
			return teamData[i].Points > teamData[j].Points
		})
		
		// Assign positions (0 = first place, 1 = second place, etc.)
		for pos, team := range teamData {
			positions[team.TeamIndex][path] = pos
		}
	}
	
	// Calculate probabilities
	probabilities := make(map[string][]float64)
	for _, name := range teamNames {
		if idx := sp.getTeamIndex(name); idx >= 0 {
			probs := make([]float64, len(selectedIndices))
			
			// Find which index in selectedIndices this team corresponds to
			selectedIdx := -1
			for j, selIdx := range selectedIndices {
				if selIdx == idx {
					selectedIdx = j
					break
				}
			}
			
			if selectedIdx >= 0 {
				// Count occurrences of each position
				for path := 0; path < sp.NPaths; path++ {
					pos := positions[selectedIdx][path]
					probs[pos] += 1.0 / float64(sp.NPaths)
				}
			}
			
			probabilities[name] = probs
		}
	}
	
	return probabilities
}