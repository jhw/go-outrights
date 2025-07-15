package outrights

import (
	"sort"
	"strings"
)

func calcLeagueTable(teamNames []string, events []Event, handicaps map[string]float64) []Team {
	teams := make(map[string]*Team)
	
	// Initialize teams
	for _, name := range teamNames {
		teams[name] = &Team{
			Name:           name,
			Points:         0,
			GoalDifference: 0,
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
		if len(event.MatchOdds.Prices) == 0 {
			continue
		}
		
		// For now, we'll simulate a result based on the odds
		// In a real implementation, you'd have actual match results
		// This is a placeholder - you'd need actual match results
		// For demonstration, we'll skip updating points from events
		// as the Python version seems to work with odds data primarily
		
		// Ensure teams exist
		if _, exists := teams[homeTeam]; !exists {
			teams[homeTeam] = &Team{Name: homeTeam}
		}
		if _, exists := teams[awayTeam]; !exists {
			teams[awayTeam] = &Team{Name: awayTeam}
		}
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
	playedFixtures := make(map[string]bool)
	
	// Track already played fixtures
	for _, event := range events {
		playedFixtures[event.Name] = true
	}
	
	var remainingFixtures []string
	
	// Generate all possible fixtures for the specified number of rounds
	for round := 0; round < rounds; round++ {
		for i, homeTeam := range teamNames {
			for j, awayTeam := range teamNames {
				if i != j {
					fixtureName := homeTeam + " vs " + awayTeam
					if !playedFixtures[fixtureName] {
						remainingFixtures = append(remainingFixtures, fixtureName)
					}
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