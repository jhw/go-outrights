package outrights

import (
	"sort"
	"strings"
)

func calcLeagueTable(teamNames []string, events []Event, handicaps map[string]int) []Team {
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
	
	// Process events
	for _, event := range events {
		homeTeam, awayTeam := parseEventName(event.Name)
		
		// Skip if we don't have match result data
		if len(event.Score) != 2 {
			continue
		}
		
		// Ensure teams exist
		if _, exists := teams[homeTeam]; !exists {
			teams[homeTeam] = &Team{Name: homeTeam}
		}
		if _, exists := teams[awayTeam]; !exists {
			teams[awayTeam] = &Team{Name: awayTeam}
		}
		
		homeGoals := event.Score[0]
		awayGoals := event.Score[1]
		
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

func calcRemainingFixtures(teamNames []string, events []Event, rounds int) []string {
	// Count how many times each fixture has been played
	playedCounts := make(map[string]int)
	
	// Count already played fixtures (only those with scores)
	for _, event := range events {
		if len(event.Score) == 2 {
			playedCounts[event.Name]++
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

func parseEventName(eventName string) (string, string) {
	parts := strings.Split(eventName, " vs ")
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

func getTeamNamesFromEvents(events []Event) []string {
	teamSet := make(map[string]bool)
	
	for _, event := range events {
		homeTeam, awayTeam := parseEventName(event.Name)
		if homeTeam != "" && awayTeam != "" {
			teamSet[homeTeam] = true
			teamSet[awayTeam] = true
		}
	}
	
	teamNames := make([]string, 0, len(teamSet))
	for name := range teamSet {
		teamNames = append(teamNames, name)
	}
	
	sort.Strings(teamNames)
	return teamNames
}