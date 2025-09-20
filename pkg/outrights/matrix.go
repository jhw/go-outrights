package outrights

import (
	"math"
	"math/rand"
)

const (
	DefaultN   = 11
	DefaultRho = 0.1
	GDMultiplier = 1e-4
	NoiseMultiplier = 1e-8
)


type ScoreMatrix struct {
	HomeLambda  float64
	AwayLambda  float64
	Rho         float64
	Matrix      [][]float64
	N           int
}

func NewScoreMatrix(eventName string, ratings map[string]float64, homeAdvantage float64) *ScoreMatrix {
	homeTeam, awayTeam := ParseEventName(eventName)
	homeLambda := ratings[homeTeam] + homeAdvantage
	awayLambda := ratings[awayTeam]
	
	sm := &ScoreMatrix{
		HomeLambda: homeLambda,
		AwayLambda: awayLambda,
		Rho:        DefaultRho,
		N:          DefaultN,
	}
	
	sm.initMatrix()
	return sm
}

func (sm *ScoreMatrix) initMatrix() {
	sm.Matrix = make([][]float64, sm.N)
	for i := range sm.Matrix {
		sm.Matrix[i] = make([]float64, sm.N)
	}
	
	for i := 0; i < sm.N; i++ {
		for j := 0; j < sm.N; j++ {
			homeProb := poissonProb(sm.HomeLambda, i)
			awayProb := poissonProb(sm.AwayLambda, j)
			adjustment := dixonColesAdjustment(i, j, sm.Rho)
			sm.Matrix[i][j] = homeProb * awayProb * adjustment
		}
	}
}

func (sm *ScoreMatrix) probability(maskFn func(i, j int) bool) float64 {
	total := 0.0
	for i := 0; i < sm.N; i++ {
		for j := 0; j < sm.N; j++ {
			if maskFn(i, j) {
				total += sm.Matrix[i][j]
			}
		}
	}
	return total
}

func (sm *ScoreMatrix) MatchOdds() []float64 {
	homeWin := sm.probability(func(i, j int) bool { return i > j })
	draw := sm.probability(func(i, j int) bool { return i == j })
	awayWin := sm.probability(func(i, j int) bool { return i < j })
	
	// Normalize
	total := homeWin + draw + awayWin
	return []float64{homeWin / total, draw / total, awayWin / total}
}

func (sm *ScoreMatrix) expectedHomePoints() float64 {
	odds := sm.MatchOdds()
	return 3*odds[0] + odds[1]
}

func (sm *ScoreMatrix) expectedAwayPoints() float64 {
	odds := sm.MatchOdds()
	return 3*odds[2] + odds[1]
}

func (sm *ScoreMatrix) simulateScores(nPaths int) [][]int {
	// Flatten matrix and create cumulative distribution
	var flatMatrix []float64
	var indices [][]int
	
	for i := 0; i < sm.N; i++ {
		for j := 0; j < sm.N; j++ {
			flatMatrix = append(flatMatrix, sm.Matrix[i][j])
			indices = append(indices, []int{i, j})
		}
	}
	
	// Normalize
	total := 0.0
	for _, prob := range flatMatrix {
		total += prob
	}
	for i := range flatMatrix {
		flatMatrix[i] /= total
	}
	
	// Create cumulative distribution
	cumulative := make([]float64, len(flatMatrix))
	cumulative[0] = flatMatrix[0]
	for i := 1; i < len(flatMatrix); i++ {
		cumulative[i] = cumulative[i-1] + flatMatrix[i]
	}
	
	// Sample
	results := make([][]int, nPaths)
	for path := 0; path < nPaths; path++ {
		r := rand.Float64()
		for i, cum := range cumulative {
			if r <= cum {
				results[path] = []int{indices[i][0], indices[i][1]}
				break
			}
		}
	}
	
	return results
}

// asianHandicaps calculates Asian handicap probabilities at half-point intervals
func (sm *ScoreMatrix) AsianHandicaps() [][2]interface{} {
	var handicaps [][2]interface{}
	
	// Calculate handicaps from -4.5 to +4.5 (based on N-1 to handle matrix bounds)
	maxHandicap := float64(sm.N - 1)
	for handicap := -maxHandicap + 0.5; handicap <= maxHandicap - 0.5; handicap += 0.5 {
		var probs interface{}
		
		if handicap == float64(int(handicap)) {
			// Integer handicap: [home_win, draw, away_win]
			homeWin := sm.probability(func(i, j int) bool { return float64(i) + handicap > float64(j) })
			draw := sm.probability(func(i, j int) bool { return float64(i) + handicap == float64(j) })
			awayWin := sm.probability(func(i, j int) bool { return float64(i) + handicap < float64(j) })
			
			total := homeWin + draw + awayWin
			probs = [3]float64{homeWin / total, draw / total, awayWin / total}
		} else {
			// Half handicap: [home_win, away_win] 
			homeWin := sm.probability(func(i, j int) bool { return float64(i) + handicap > float64(j) })
			awayWin := sm.probability(func(i, j int) bool { return float64(i) + handicap < float64(j) })
			
			total := homeWin + awayWin
			probs = [2]float64{homeWin / total, awayWin / total}
		}
		
		handicaps = append(handicaps, [2]interface{}{handicap, probs})
	}
	
	return handicaps
}

// totalGoals calculates over/under total goals probabilities at half-point intervals
func (sm *ScoreMatrix) TotalGoals() [][2]interface{} {
	var totals [][2]interface{}
	
	// Calculate totals from 0.5 to (N-1)*2 - 0.5 goals
	maxGoals := float64(sm.N*2 - 2)
	for line := 0.5; line <= maxGoals - 0.5; line += 1.0 {
		under := sm.probability(func(i, j int) bool { return float64(i + j) < line })
		over := sm.probability(func(i, j int) bool { return float64(i + j) > line })
		
		total := under + over
		probs := [2]float64{under / total, over / total}
		
		totals = append(totals, [2]interface{}{line, probs})
	}
	
	return totals
}

// factorial calculates the factorial of n
func factorial(n int) float64 {
	if n <= 1 {
		return 1
	}
	result := 1.0
	for i := 2; i <= n; i++ {
		result *= float64(i)
	}
	return result
}

// poissonProb calculates the Poisson probability for lambda and k
func poissonProb(lambda float64, k int) float64 {
	return math.Pow(lambda, float64(k)) * math.Exp(-lambda) / factorial(k)
}

// dixonColesAdjustment applies Dixon-Coles adjustment for low-scoring games
func dixonColesAdjustment(i, j int, rho float64) float64 {
	switch {
	case i == 0 && j == 0:
		return 1 - (float64(i*j) * rho)
	case i == 0 && j == 1:
		return 1 + (rho / 2)
	case i == 1 && j == 0:
		return 1 + (rho / 2)
	case i == 1 && j == 1:
		return 1 - rho
	default:
		return 1
	}
}

