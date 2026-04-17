package memory

import (
	"testing"
	"time"
)

func TestContentHash(t *testing.T) {
	h1 := ContentHash("hello world")
	h2 := ContentHash("hello world")
	h3 := ContentHash("different content")

	if h1 != h2 {
		t.Errorf("same content should produce same hash: %q != %q", h1, h2)
	}
	if h1 == h3 {
		t.Error("different content should produce different hash")
	}
	if len(h1) != 64 {
		t.Errorf("SHA-256 hex should be 64 chars, got %d", len(h1))
	}
}

func TestMemoryTypeValid(t *testing.T) {
	tests := []struct {
		mt    MemoryType
		valid bool
	}{
		{TypeFact, true},
		{TypeConversation, true},
		{TypeDecision, true},
		{TypeCodePattern, true},
		{TypeCorrection, true},
		{MemoryType("unknown"), false},
		{MemoryType(""), false},
	}

	for _, tt := range tests {
		if got := tt.mt.Valid(); got != tt.valid {
			t.Errorf("MemoryType(%q).Valid() = %v, want %v", tt.mt, got, tt.valid)
		}
	}
}

func TestMemoryStruct(t *testing.T) {
	now := time.Now().UTC()
	m := Memory{
		ID:          "test-id",
		Content:     "some content",
		MemoryType:  TypeFact,
		UserID:      "alice",
		AgentID:     "cursor",
		AppID:       "my-app",
		RunID:       "run-1",
		Metadata:    map[string]any{"key": "value"},
		ContentHash: ContentHash("some content"),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if m.ID != "test-id" {
		t.Errorf("expected ID 'test-id', got %q", m.ID)
	}
	if m.MemoryType != TypeFact {
		t.Errorf("expected type fact, got %q", m.MemoryType)
	}
	if m.Metadata["key"] != "value" {
		t.Errorf("expected metadata key=value, got %v", m.Metadata["key"])
	}
}

func TestSearchFilter(t *testing.T) {
	since := time.Now().Add(-24 * time.Hour)
	f := SearchFilter{
		UserID:     "alice",
		MemoryType: TypeDecision,
		Since:      &since,
	}

	if f.UserID != "alice" {
		t.Errorf("expected UserID alice, got %q", f.UserID)
	}
	if f.MemoryType != TypeDecision {
		t.Errorf("expected type decision, got %q", f.MemoryType)
	}
	if f.Since == nil {
		t.Error("expected Since to be set")
	}
}
