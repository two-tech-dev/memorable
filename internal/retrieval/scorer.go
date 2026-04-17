package retrieval

import (
	"math"
	"sort"
	"time"

	"github.com/two-tech-dev/memorable/internal/memory"
)

// Weights controls the relative importance of each scoring signal.
type Weights struct {
	Similarity float64 `json:"similarity" yaml:"similarity"` // vector cosine score
	Recency    float64 `json:"recency" yaml:"recency"`       // time decay
	Frequency  float64 `json:"frequency" yaml:"frequency"`   // access count boost
}

// DefaultWeights provides balanced defaults.
func DefaultWeights() Weights {
	return Weights{
		Similarity: 0.6,
		Recency:    0.25,
		Frequency:  0.15,
	}
}

// ScoredResult wraps a search result with hybrid scoring details.
type ScoredResult struct {
	Memory          memory.Memory `json:"memory"`
	SimilarityScore float64       `json:"similarity_score"`
	RecencyScore    float64       `json:"recency_score"`
	FrequencyScore  float64       `json:"frequency_score"`
	FinalScore      float64       `json:"final_score"`
}

// Scorer applies multi-signal ranking to search results.
type Scorer struct {
	weights    Weights
	halfLifeH float64 // recency half-life in hours
}

func NewScorer(weights Weights) *Scorer {
	return &Scorer{
		weights:    weights,
		halfLifeH: 168, // 7 days default half-life
	}
}

// SetHalfLife configures the recency decay half-life in hours.
// After this many hours, recency score drops to 0.5.
func (s *Scorer) SetHalfLife(hours float64) {
	if hours > 0 {
		s.halfLifeH = hours
	}
}

// Rank takes vector search results and re-ranks them using multi-signal scoring.
func (s *Scorer) Rank(results []memory.SearchResult, now time.Time) []ScoredResult {
	if len(results) == 0 {
		return nil
	}

	scored := make([]ScoredResult, len(results))

	// Collect frequency values for normalization
	maxFreq := 1.0
	for _, r := range results {
		if f := getFrequency(r.Memory); f > maxFreq {
			maxFreq = f
		}
	}

	for i, r := range results {
		simScore := r.Score
		recScore := s.recencyScore(r.Memory.UpdatedAt, now)
		freqScore := getFrequency(r.Memory) / maxFreq

		finalScore := s.weights.Similarity*simScore +
			s.weights.Recency*recScore +
			s.weights.Frequency*freqScore

		scored[i] = ScoredResult{
			Memory:          r.Memory,
			SimilarityScore: simScore,
			RecencyScore:    recScore,
			FrequencyScore:  freqScore,
			FinalScore:      finalScore,
		}
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].FinalScore > scored[j].FinalScore
	})

	return scored
}

// recencyScore computes exponential decay based on time since last update.
// Returns a value in [0, 1] where 1 = just now, 0.5 = one half-life ago.
func (s *Scorer) recencyScore(updatedAt, now time.Time) float64 {
	hoursSince := now.Sub(updatedAt).Hours()
	if hoursSince < 0 {
		hoursSince = 0
	}
	// Exponential decay: score = 2^(-t/halfLife)
	return math.Pow(2, -hoursSince/s.halfLifeH)
}

// getFrequency extracts the access_count from memory metadata.
func getFrequency(m memory.Memory) float64 {
	if m.Metadata == nil {
		return 1.0
	}
	if v, ok := m.Metadata["access_count"]; ok {
		switch n := v.(type) {
		case float64:
			if n > 0 {
				return n
			}
		case int:
			if n > 0 {
				return float64(n)
			}
		}
	}
	return 1.0
}
