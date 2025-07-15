package outrights

import (
	"math/rand"
	"sort"
)

type SimPoints struct {
	NPaths    int
	TeamNames []string
	Points    [][]float64
}

func newSimPoints(leagueTable []Team, nPaths int) *SimPoints {
	sp := &SimPoints{
		NPaths:    nPaths,
		TeamNames: make([]string, len(leagueTable)),
		Points:    make([][]float64, len(leagueTable)),
	}
	
	for i, team := range leagueTable {
		sp.TeamNames[i] = team.Name
		sp.Points[i] = make([]float64, nPaths)
		
		// Initialize with current points + small adjustments for goal difference and noise
		pointsWithAdjustments := team.Points + 
			GDMultiplier*float64(team.GoalDifference) + 
			NoiseMultiplier*(rand.Float64()-0.5)
		
		for j := 0; j < nPaths; j++ {
			sp.Points[i][j] = pointsWithAdjustments
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
		
		// Update points
		sp.Points[teamIndex][i] += points + GDMultiplier*goalDifference
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
		
		// Update points
		sp.Points[teamIndex][i] += points + GDMultiplier*goalDifference
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
	
	// Extract points for selected teams
	selectedPoints := make([][]float64, len(selectedIndices))
	for i, idx := range selectedIndices {
		selectedPoints[i] = sp.Points[idx]
	}
	
	// Calculate positions for each path
	positions := make([][]int, len(selectedIndices))
	for i := range positions {
		positions[i] = make([]int, sp.NPaths)
	}
	
	for path := 0; path < sp.NPaths; path++ {
		// Create array of team points for this path
		teamPoints := make([]struct {
			TeamIndex int
			Points    float64
		}, len(selectedIndices))
		
		for i := range selectedIndices {
			teamPoints[i] = struct {
				TeamIndex int
				Points    float64
			}{
				TeamIndex: i,
				Points:    selectedPoints[i][path],
			}
		}
		
		// Sort by points (descending) to get positions
		sort.Slice(teamPoints, func(i, j int) bool {
			return teamPoints[i].Points > teamPoints[j].Points
		})
		
		// Assign positions (0 = first place, 1 = second place, etc.)
		for pos, team := range teamPoints {
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