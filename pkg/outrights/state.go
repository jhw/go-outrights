package outrights

import (
	"sort"
	"strings"
)

func CalcLeagueTable(teamNames []string, results []Result, handicaps map[string]int) []Team {
	teams := make(map[string]*Team)
	
	// Initialize teams
	for _, name := range teamNames {
		teams[name] = &Team{
			Name:           name,
			Points:         0,
			GoalDifference: 0,
			Played:         0,
		}
	}
	
	// Apply handicaps
	for name, handicap := range handicaps {
		if team, exists := teams[name]; exists {
			team.Points += handicap
		}
	}
	
	// Process results
	for _, result := range results {
		homeTeam, awayTeam := ParseEventName(result.Name)
		
		// Skip if we don't have match result data
		if len(result.Score) != 2 {
			continue
		}
		
		// Ensure teams exist
		if _, exists := teams[homeTeam]; !exists {
			teams[homeTeam] = &Team{Name: homeTeam}
		}
		if _, exists := teams[awayTeam]; !exists {
			teams[awayTeam] = &Team{Name: awayTeam}
		}
		
		homeGoals := result.Score[0]
		awayGoals := result.Score[1]
		
		// Calculate points
		if homeGoals > awayGoals {
			// Home team wins
			teams[homeTeam].Points += 3
		} else if homeGoals < awayGoals {
			// Away team wins
			teams[awayTeam].Points += 3
		} else {
			// Draw
			teams[homeTeam].Points += 1
			teams[awayTeam].Points += 1
		}
		
		// Update goal difference and games played
		teams[homeTeam].GoalDifference += homeGoals - awayGoals
		teams[awayTeam].GoalDifference += awayGoals - homeGoals
		teams[homeTeam].Played += 1
		teams[awayTeam].Played += 1
	}
	
	// Convert to slice and sort
	result := make([]Team, 0, len(teams))
	for _, team := range teams {
		result = append(result, *team)
	}
	
	// Sort by points (descending), then by goal difference (descending)
	sort.Slice(result, func(i, j int) bool {
		if result[i].Points == result[j].Points {
			return result[i].GoalDifference > result[j].GoalDifference
		}
		return result[i].Points > result[j].Points
	})
	
	return result
}

func CalcRemainingFixtures(teamNames []string, results []Result, rounds int) []string {
	// Count how many times each fixture has been played
	playedCounts := make(map[string]int)
	
	// Count already played fixtures
	for _, result := range results {
		if len(result.Score) == 2 {
			playedCounts[result.Name]++
		}
	}
	
	var remainingFixtures []string
	
	// Generate all possible fixtures (each team plays every other team home and away)
	for i, homeTeam := range teamNames {
		for j, awayTeam := range teamNames {
			if i != j {
				fixtureName := homeTeam + " vs " + awayTeam
				playedCount := playedCounts[fixtureName]
				
				// Add remaining fixtures for this matchup
				for k := playedCount; k < rounds; k++ {
					remainingFixtures = append(remainingFixtures, fixtureName)
				}
			}
		}
	}
	
	return remainingFixtures
}

func ParseEventName(eventName string) (string, string) {
	parts := strings.Split(eventName, " vs ")
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

