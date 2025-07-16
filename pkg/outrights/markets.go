package outrights

import (
	"fmt"
	"strconv"
	"strings"
)

// parsePayoff parses payoff expressions like "1|19x0" meaning 1 winner gets 1, 19 losers get 0
func parsePayoff(payoffExpr string) ([]int, error) {
	var payoff []int
	
	for _, expr := range strings.Split(payoffExpr, "|") {
		tokens := strings.Split(expr, "x")
		
		var n int
		var v int
		var err error
		
		if len(tokens) == 1 {
			// Single value, assume n=1
			n = 1
			v, err = strconv.Atoi(tokens[0])
		} else if len(tokens) == 2 {
			// n and value
			var err1 error
			n, err1 = strconv.Atoi(tokens[0])
			v, err = strconv.Atoi(tokens[1])
			if err1 != nil || err != nil {
				return nil, fmt.Errorf("invalid payoff format: %s", expr)
			}
		} else {
			return nil, fmt.Errorf("invalid payoff format: %s", expr)
		}
		
		if err != nil {
			return nil, fmt.Errorf("invalid payoff format: %s", expr)
		}
		
		for i := 0; i < n; i++ {
			payoff = append(payoff, v)
		}
	}
	
	return payoff, nil
}

// initIncludeMarket initializes a market with specific included teams
func initIncludeMarket(teamNames []string, market *Market) error {
	// Check for unknown teams
	for _, teamName := range market.Include {
		found := false
		for _, knownTeam := range teamNames {
			if teamName == knownTeam {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("%s market has unknown team %s", market.Name, teamName)
		}
	}
	
	market.Teams = make([]string, len(market.Include))
	copy(market.Teams, market.Include)
	
	// Parse and validate payoff
	if market.Payoff == "" {
		return fmt.Errorf("market %s has no payoff defined", market.Name)
	}
	
	parsedPayoff, err := parsePayoff(market.Payoff)
	if err != nil {
		return fmt.Errorf("error parsing payoff for market %s: %v", market.Name, err)
	}
	market.ParsedPayoff = parsedPayoff
	
	// Validate payoff length matches include teams count
	if len(market.ParsedPayoff) != len(market.Include) {
		return fmt.Errorf("%s include market payoff length (%d) does not match include teams count (%d)", 
			market.Name, len(market.ParsedPayoff), len(market.Include))
	}
	
	return nil
}

// initExcludeMarket initializes a market excluding specific teams
func initExcludeMarket(teamNames []string, market *Market) error {
	// Check for unknown teams
	for _, teamName := range market.Exclude {
		found := false
		for _, knownTeam := range teamNames {
			if teamName == knownTeam {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("%s market has unknown team %s", market.Name, teamName)
		}
	}
	
	// Include all teams except excluded ones
	market.Teams = []string{}
	for _, teamName := range teamNames {
		excluded := false
		for _, excludedTeam := range market.Exclude {
			if teamName == excludedTeam {
				excluded = true
				break
			}
		}
		if !excluded {
			market.Teams = append(market.Teams, teamName)
		}
	}
	
	// Parse and validate payoff
	if market.Payoff == "" {
		return fmt.Errorf("market %s has no payoff defined", market.Name)
	}
	
	parsedPayoff, err := parsePayoff(market.Payoff)
	if err != nil {
		return fmt.Errorf("error parsing payoff for market %s: %v", market.Name, err)
	}
	market.ParsedPayoff = parsedPayoff
	
	// Validate payoff length matches remaining teams count (total - excluded)
	expectedLength := len(teamNames) - len(market.Exclude)
	if len(market.ParsedPayoff) != expectedLength {
		return fmt.Errorf("%s exclude market payoff length (%d) does not match remaining teams count (%d)", 
			market.Name, len(market.ParsedPayoff), expectedLength)
	}
	
	return nil
}

// initStandardMarket initializes a market with all teams
func initStandardMarket(teamNames []string, market *Market) error {
	market.Teams = make([]string, len(teamNames))
	copy(market.Teams, teamNames)
	
	// Parse and validate payoff
	if market.Payoff == "" {
		return fmt.Errorf("market %s has no payoff defined", market.Name)
	}
	
	parsedPayoff, err := parsePayoff(market.Payoff)
	if err != nil {
		return fmt.Errorf("error parsing payoff for market %s: %v", market.Name, err)
	}
	market.ParsedPayoff = parsedPayoff
	
	// Validate payoff length matches all teams count
	if len(market.ParsedPayoff) != len(teamNames) {
		return fmt.Errorf("%s standard market payoff length (%d) does not match total teams count (%d)", 
			market.Name, len(market.ParsedPayoff), len(teamNames))
	}
	
	return nil
}

// InitMarkets initializes all markets with proper team lists and payoffs
func InitMarkets(teamNames []string, markets []Market) error {
	for i := range markets {
		market := &markets[i]
		
		// Validate that market doesn't have both include and exclude
		if len(market.Include) > 0 && len(market.Exclude) > 0 {
			return fmt.Errorf("market %s cannot have both include and exclude fields", market.Name)
		}
		
		// Initialize teams based on include/exclude
		var err error
		if len(market.Include) > 0 {
			err = initIncludeMarket(teamNames, market)
		} else if len(market.Exclude) > 0 {
			err = initExcludeMarket(teamNames, market)
		} else {
			err = initStandardMarket(teamNames, market)
		}
		
		if err != nil {
			return err
		}
	}
	
	return nil
}


