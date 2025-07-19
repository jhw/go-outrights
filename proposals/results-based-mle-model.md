# Results-Based Maximum Likelihood Estimation Model

## Executive Summary

This proposal outlines a complementary model to the existing market-fitting approach that would perform maximum likelihood estimation (MLE) against actual match results rather than betting odds. This "results-based model" would leverage decades of historical data across all English football leagues to produce true strength ratings that evolve over time using Bayesian updating principles.

## Current vs Proposed Approach

### Current Market-Fitting Model
- **Objective**: Minimize RMS error between Dixon-Coles predictions and market-implied probabilities
- **Training Data**: ~60 recent events with betting odds
- **Scope**: Single league, single season
- **Output**: Ratings calibrated to market expectations
- **Use Case**: Betting mark generation, market inefficiency detection

### Proposed Results-Based MLE Model  
- **Objective**: Maximize likelihood of observed match results (scorelines)
- **Training Data**: Decades of historical results across all leagues
- **Scope**: Multi-league, multi-season with temporal evolution
- **Output**: True strength ratings independent of market sentiment
- **Use Case**: Long-term strength assessment, cross-league comparisons, market divergence analysis

## Technical Foundation

### Dixon-Coles MLE Formulation

The likelihood function for a set of matches is:

```
L(Θ) = Π P(home_goals, away_goals | λ_home, λ_away, ρ)
```

Where:
- `P(i,j | λ_home, λ_away, ρ) = poisson(λ_home, i) × poisson(λ_away, j) × τ(i,j,ρ)`
- `τ(i,j,ρ)` is the Dixon-Coles adjustment for low-scoring games
- `λ_home = exp(α_home + γ + θ_home - θ_away)`
- `λ_away = exp(α_away + θ_away - θ_home)`

### Parameters to Estimate
- `θ_i`: Attack strength rating for team i
- `α_home, α_away`: Global home/away effects  
- `γ`: Home advantage parameter
- `ρ`: Dixon-Coles correlation parameter

## Data Requirements

### Historical Results Database
```json
{
  "match_id": "unique_identifier",
  "date": "2023-10-15",
  "season": "2023-24",
  "league": "ENG1",
  "home_team": "Arsenal", 
  "away_team": "Chelsea",
  "home_goals": 2,
  "away_goals": 1,
  "home_team_id": "arsenal_2023",
  "away_team_id": "chelsea_2023"
}
```

### Team Identity Tracking
Teams need unique identifiers that persist across:
- League promotions/relegations
- Name changes
- Ownership changes
- Season transitions

### Suggested Data Scope
- **Leagues**: ENG1, ENG2, ENG3, ENG4 (plus Conference if available)
- **Time Range**: 20+ seasons (sufficient for meaningful temporal patterns)
- **Minimum**: ~50,000 matches across all leagues and seasons

## Architecture Design

### Core Components

#### 1. Results Database Layer
```go
type MatchResult struct {
    ID           string    `db:"id"`
    Date         time.Time `db:"match_date"`
    Season       string    `db:"season"`
    League       string    `db:"league"` 
    HomeTeam     string    `db:"home_team"`
    AwayTeam     string    `db:"away_team"`
    HomeGoals    int       `db:"home_goals"`
    AwayGoals    int       `db:"away_goals"`
    HomeTeamID   string    `db:"home_team_id"`
    AwayTeamID   string    `db:"away_team_id"`
}
```

#### 2. Rating System
```go
type TeamRating struct {
    TeamID       string    `json:"team_id"`
    Season       string    `json:"season"`
    League       string    `json:"league"`
    AttackRating float64   `json:"attack_rating"`
    DefenseRating float64  `json:"defense_rating"`
    UpdatedAt    time.Time `json:"updated_at"`
    Confidence   float64   `json:"confidence"`
}
```

#### 3. MLE Solver
```go
type ResultsMLESolver struct {
    data           []MatchResult
    ratings        map[string]*TeamRating
    homeAdvantage  float64
    rho           float64
    priorVariance float64
}

func (s *ResultsMLESolver) MaximizeLikelihood() error {
    // Implement L-BFGS or similar gradient-based optimization
    // Objective: maximize sum of log-likelihoods
}
```

## Bayesian Updating Mechanism

### Rating Evolution Model
Teams' true strengths evolve over time following a random walk:

```
θ_i(t+1) = θ_i(t) + ε_i(t+1)
```

Where `ε_i(t+1) ~ N(0, σ²_evolution)`

### League Transition Handling
When teams change leagues, increase rating volatility temporarily:

```go
func (s *ResultsMLESolver) HandleLeagueTransition(teamID string, newLeague string) {
    rating := s.ratings[teamID]
    
    // Increase uncertainty for transition period
    transitionPeriod := 10 // matches
    rating.Confidence *= 0.5 // Reduce confidence
    
    // Allow faster adaptation
    s.updateVariance(teamID, s.priorVariance * 2.0, transitionPeriod)
}
```

### Temporal Weighting
Recent matches should have higher influence:

```go
func temporalWeight(matchDate time.Time, currentDate time.Time, halfLife time.Duration) float64 {
    age := currentDate.Sub(matchDate)
    return math.Exp(-age.Hours() / halfLife.Hours() * math.Ln2)
}
```

## Cross-League Rating System

### Unified Rating Space
All teams exist in the same rating space regardless of current league:

```
Rating Range: [-3.0, 3.0] (roughly corresponds to goal difference per match)
ENG1 average: ~0.5
ENG2 average: ~0.0  
ENG3 average: ~-0.5
ENG4 average: ~-1.0
```

### League Strength Anchoring
Use promotion/relegation results to anchor relative league strengths:

```go
func (s *ResultsMLESolver) anchorLeagueStrengths() {
    // Teams that get promoted and perform well indicate league strength gap
    // Teams that get relegated and struggle indicate league strength gap
}
```

## Implementation Strategy

### Phase 1: Historical Data Pipeline
1. Build match results database
2. Implement team identity tracking
3. Create data validation and cleaning processes
4. Historical data ingestion (20+ seasons)

### Phase 2: Core MLE Engine
1. Implement likelihood calculation
2. Build gradient-based optimizer  
3. Add Bayesian updating mechanism
4. Cross-league rating calibration

### Phase 3: Temporal Dynamics
1. Time-weighted likelihood
2. Rating evolution modeling
3. League transition detection and handling
4. Confidence interval estimation

### Phase 4: Integration & Validation
1. API endpoint for rating queries
2. Comparison framework with market-fitting model
3. Historical backtesting
4. Performance metrics and validation

## Expected Outcomes

### Primary Benefits
1. **True Strength Assessment**: Ratings independent of market sentiment and short-term form
2. **Cross-League Comparisons**: Meaningful strength comparisons across divisions
3. **Temporal Insights**: Understanding of team development and decline patterns
4. **Market Divergence Detection**: Identify when market and results-based models disagree significantly

### Use Cases
1. **Long-term Value Betting**: Identify teams whose market odds don't reflect historical strength
2. **Promotion/Relegation Markets**: Better assessment of teams crossing league boundaries  
3. **Manager/Player Impact Analysis**: Isolate rating changes from structural changes
4. **League Strength Evolution**: Track how league competitiveness changes over time

## Integration with Current System

### API Design
```go
type RatingService interface {
    GetCurrentRating(teamID string) (*TeamRating, error)
    GetHistoricalRatings(teamID string, fromDate, toDate time.Time) ([]TeamRating, error)
    CompareWithMarketModel(teamA, teamB string) (*ModelComparison, error)
    GetLeagueStrengths(season string) (map[string]float64, error)
}

type ModelComparison struct {
    MarketModelPrediction   []float64 `json:"market_model"`
    ResultsModelPrediction  []float64 `json:"results_model"`
    Divergence             float64   `json:"divergence"`
    ConfidenceInterval     []float64 `json:"confidence_interval"`
}
```

### Decision Framework
```go
func (s *SimulationService) GetRecommendation(homeTeam, awayTeam string) *BettingRecommendation {
    marketPred := s.marketModel.Predict(homeTeam, awayTeam)
    resultsPred := s.resultsModel.Predict(homeTeam, awayTeam)
    
    divergence := calculateDivergence(marketPred, resultsPred)
    
    if divergence > threshold {
        return &BettingRecommendation{
            Recommended: true,
            Rationale: "Significant model divergence detected",
            MarketModel: marketPred,
            ResultsModel: resultsPred,
            Confidence: calculateConfidence(divergence),
        }
    }
    
    return &BettingRecommendation{Recommended: false}
}
```

## Implementation Roadmap

### Month 1: Data Foundation
- [ ] Design and implement match results database schema
- [ ] Build team identity tracking system
- [ ] Create data ingestion pipeline
- [ ] Historical data acquisition and cleaning

### Month 2: Core Algorithm  
- [ ] Implement Dixon-Coles likelihood calculation
- [ ] Build MLE optimization engine
- [ ] Add basic Bayesian updating
- [ ] Unit tests and validation framework

### Month 3: Advanced Features
- [ ] Temporal weighting system
- [ ] League transition handling
- [ ] Cross-league rating calibration  
- [ ] Confidence interval estimation

### Month 4: Integration & Validation
- [ ] API development
- [ ] Integration with existing system
- [ ] Historical backtesting
- [ ] Performance comparison with market model

## Risk Assessment

### Technical Risks
- **Computational Complexity**: Large-scale MLE optimization may be slow
  - *Mitigation*: Incremental updates, efficient gradient computation
- **Data Quality**: Historical data may have inconsistencies
  - *Mitigation*: Extensive validation, manual verification of key matches
- **Model Overfitting**: Too many parameters for available data
  - *Mitigation*: Regularization, cross-validation, parameter constraints

### Business Risks
- **Limited Immediate Value**: Results may not immediately improve betting performance
  - *Mitigation*: Focus on long-term value and model diversification benefits
- **Model Complexity**: Difficult to explain and maintain
  - *Mitigation*: Comprehensive documentation, modular design

## Success Metrics

### Quantitative Metrics
1. **Likelihood Improvement**: Higher likelihood on held-out test sets
2. **Predictive Accuracy**: Better prediction of actual match results
3. **Model Divergence Value**: Profitable betting opportunities when models disagree
4. **Rating Stability**: Smooth rating evolution without excessive volatility

### Qualitative Metrics  
1. **Intuitive Results**: Ratings should align with football knowledge
2. **Cross-League Coherence**: Promoted teams should show appropriate rating changes
3. **Historical Consistency**: Major events (new managers, transfers) should be reflected in ratings

## Conclusion

This results-based MLE model represents a natural evolution of the current system, providing a complementary perspective that focuses on fundamental team strength rather than market sentiment. The combination of both models should provide a more robust foundation for identifying betting value and understanding team performance dynamics across the English football pyramid.

The implementation is technically feasible using existing infrastructure and would provide significant strategic advantages in terms of market analysis and long-term value identification.