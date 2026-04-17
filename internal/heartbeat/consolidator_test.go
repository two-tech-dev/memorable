package heartbeat

import (
	"testing"
	"time"

	"github.com/two-tech-dev/memorable/internal/memory"
)

func TestConsolidatorNewAndThreshold(t *testing.T) {
	c := NewConsolidator(nil)
	if c.similarityThreshold != 0.85 {
		t.Errorf("expected default threshold 0.85, got %f", c.similarityThreshold)
	}

	c.SetSimilarityThreshold(0.9)
	if c.similarityThreshold != 0.9 {
		t.Errorf("expected threshold 0.9, got %f", c.similarityThreshold)
	}

	// Invalid thresholds should be ignored
	c.SetSimilarityThreshold(0)
	if c.similarityThreshold != 0.9 {
		t.Error("threshold 0 should be ignored")
	}
	c.SetSimilarityThreshold(1.5)
	if c.similarityThreshold != 0.9 {
		t.Error("threshold >1 should be ignored")
	}
}

func TestFilterByType(t *testing.T) {
	memories := []*memory.Memory{
		{ID: "1", MemoryType: memory.TypeFact},
		{ID: "2", MemoryType: memory.TypeCorrection},
		{ID: "3", MemoryType: memory.TypeFact},
		{ID: "4", MemoryType: memory.TypeDecision},
	}

	facts := filterByType(memories, memory.TypeFact)
	if len(facts) != 2 {
		t.Errorf("expected 2 facts, got %d", len(facts))
	}

	corrections := filterByType(memories, memory.TypeCorrection)
	if len(corrections) != 1 {
		t.Errorf("expected 1 correction, got %d", len(corrections))
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"short", 10, "short"},
		{"this is a longer string", 10, "this is..."},
		{"exactly10!", 10, "exactly10!"},
		{"", 5, ""},
	}

	for _, tt := range tests {
		got := truncate(tt.input, tt.maxLen)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
		}
	}
}

func TestSynthesizeCluster(t *testing.T) {
	c := NewConsolidator(nil)
	now := time.Now().UTC()

	cluster := []*memory.Memory{
		{ID: "a", Content: "Go is fast", CreatedAt: now.Add(-time.Hour)},
		{ID: "b", Content: "Go compiles quickly", CreatedAt: now},
	}

	insight := c.synthesizeCluster(cluster)
	if insight == nil {
		t.Fatal("expected non-nil insight")
	}
	if insight.InsightType != "summary" {
		t.Errorf("expected type summary, got %q", insight.InsightType)
	}
	if len(insight.SourceIDs) != 2 {
		t.Errorf("expected 2 source IDs, got %d", len(insight.SourceIDs))
	}
}

func TestSynthesizeClusterEmpty(t *testing.T) {
	c := NewConsolidator(nil)
	insight := c.synthesizeCluster(nil)
	if insight != nil {
		t.Error("expected nil insight for empty cluster")
	}
}
