package retrieval

import (
	"testing"
	"time"

	"github.com/two-tech-dev/memorable/internal/memory"
)

func TestDefaultWeights(t *testing.T) {
	w := DefaultWeights()
	sum := w.Similarity + w.Recency + w.Frequency
	if sum < 0.99 || sum > 1.01 {
		t.Fatalf("weights should sum to ~1.0, got %f", sum)
	}
}

func TestRecencyScore(t *testing.T) {
	s := NewScorer(DefaultWeights())
	now := time.Now()

	// Just now → ~1.0
	score := s.recencyScore(now, now)
	if score < 0.99 {
		t.Errorf("expected ~1.0 for now, got %f", score)
	}

	// One half-life ago → ~0.5
	halfLifeAgo := now.Add(-time.Duration(s.halfLifeH) * time.Hour)
	score = s.recencyScore(halfLifeAgo, now)
	if score < 0.49 || score > 0.51 {
		t.Errorf("expected ~0.5 for one half-life ago, got %f", score)
	}

	// Very old → near 0
	ancient := now.Add(-time.Duration(s.halfLifeH*10) * time.Hour)
	score = s.recencyScore(ancient, now)
	if score > 0.01 {
		t.Errorf("expected near 0 for ancient, got %f", score)
	}
}

func TestRank(t *testing.T) {
	s := NewScorer(DefaultWeights())
	now := time.Now()

	results := []memory.SearchResult{
		{
			Memory: memory.Memory{
				ID: "old_high_sim", Content: "old but relevant",
				UpdatedAt: now.Add(-720 * time.Hour), // 30 days old
				Metadata:  map[string]any{"access_count": 1.0},
			},
			Score: 0.95,
		},
		{
			Memory: memory.Memory{
				ID: "new_med_sim", Content: "new and somewhat relevant",
				UpdatedAt: now.Add(-1 * time.Hour), // 1 hour old
				Metadata:  map[string]any{"access_count": 5.0},
			},
			Score: 0.70,
		},
	}

	ranked := s.Rank(results, now)
	if len(ranked) != 2 {
		t.Fatalf("expected 2 results, got %d", len(ranked))
	}

	// With recency + frequency boost, the newer one might rank higher
	// depending on weights. Just verify scoring is applied.
	for _, r := range ranked {
		if r.FinalScore <= 0 {
			t.Errorf("expected positive final score, got %f", r.FinalScore)
		}
		if r.SimilarityScore <= 0 {
			t.Errorf("expected positive similarity score, got %f", r.SimilarityScore)
		}
	}

	// Results should be sorted by FinalScore descending
	if ranked[0].FinalScore < ranked[1].FinalScore {
		t.Error("results not sorted by FinalScore descending")
	}
}

func TestRankEmpty(t *testing.T) {
	s := NewScorer(DefaultWeights())
	ranked := s.Rank(nil, time.Now())
	if ranked != nil {
		t.Fatalf("expected nil for empty input, got %v", ranked)
	}
}

func TestGetFrequency(t *testing.T) {
	tests := []struct {
		name     string
		metadata map[string]any
		expected float64
	}{
		{"nil metadata", nil, 1.0},
		{"no access_count", map[string]any{"foo": "bar"}, 1.0},
		{"float count", map[string]any{"access_count": 5.0}, 5.0},
		{"int count", map[string]any{"access_count": 3}, 3.0},
		{"zero count", map[string]any{"access_count": 0.0}, 1.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := memory.Memory{Metadata: tt.metadata}
			got := getFrequency(m)
			if got != tt.expected {
				t.Errorf("getFrequency() = %f, want %f", got, tt.expected)
			}
		})
	}
}

func TestSetHalfLife(t *testing.T) {
	s := NewScorer(DefaultWeights())
	s.SetHalfLife(24)
	if s.halfLifeH != 24 {
		t.Errorf("expected 24h half-life, got %f", s.halfLifeH)
	}

	// Negative should be ignored
	s.SetHalfLife(-10)
	if s.halfLifeH != 24 {
		t.Errorf("negative should be ignored, got %f", s.halfLifeH)
	}
}
