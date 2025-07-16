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

func poissonProb(lambda float64, k int) float64 {
	return math.Pow(lambda, float64(k)) * math.Exp(-lambda) / factorial(k)
}

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

type ScoreMatrix struct {
	HomeLambda  float64
	AwayLambda  float64
	Rho         float64
	Matrix      [][]float64
	N           int
}

func newScoreMatrix(eventName string, ratings map[string]float64, homeAdvantage float64) *ScoreMatrix {
	homeTeam, awayTeam := parseEventName(eventName)
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

func (sm *ScoreMatrix) matchOdds() []float64 {
	homeWin := sm.probability(func(i, j int) bool { return i > j })
	draw := sm.probability(func(i, j int) bool { return i == j })
	awayWin := sm.probability(func(i, j int) bool { return i < j })
	
	// Normalize
	total := homeWin + draw + awayWin
	return []float64{homeWin / total, draw / total, awayWin / total}
}

func (sm *ScoreMatrix) expectedHomePoints() float64 {
	odds := sm.matchOdds()
	return 3*odds[0] + odds[1]
}

func (sm *ScoreMatrix) expectedAwayPoints() float64 {
	odds := sm.matchOdds()
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

func extractMarketProbabilities(event Event) []float64 {
	prices := event.MatchOdds.Prices
	probs := make([]float64, len(prices))
	overround := 0.0
	
	for i, price := range prices {
		probs[i] = 1.0 / price
		overround += probs[i]
	}
	
	// Normalize
	for i := range probs {
		probs[i] /= overround
	}
	
	return probs
}

func rmsError(x, y []float64) float64 {
	if len(x) != len(y) {
		return math.Inf(1)
	}
	
	sum := 0.0
	for i := range x {
		diff := x[i] - y[i]
		sum += diff * diff
	}
	
	return math.Sqrt(sum / float64(len(x)))
}