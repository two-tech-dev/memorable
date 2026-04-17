package profile

import (
	"strings"
	"testing"
)

func TestSoul_Observe(t *testing.T) {
	s := NewSoul("user1")

	s.Observe("language", "Go")
	s.Observe("language", "Go")
	s.Observe("editor", "VS Code")

	if s.Count() != 2 {
		t.Fatalf("expected 2 traits, got %d", s.Count())
	}

	trait, ok := s.Get("language")
	if !ok {
		t.Fatal("expected language trait")
	}
	if trait.Occurrences != 2 {
		t.Errorf("expected 2 occurrences, got %d", trait.Occurrences)
	}
	if trait.Value != "Go" {
		t.Errorf("expected Go, got %s", trait.Value)
	}
}

func TestSoul_TopTraits(t *testing.T) {
	s := NewSoul("user1")

	s.Observe("language", "Go")
	s.Observe("language", "Go")
	s.Observe("language", "Go")
	s.Observe("editor", "VS Code")
	s.Observe("editor", "VS Code")
	s.Observe("os", "Linux")

	top := s.TopTraits(2)
	if len(top) != 2 {
		t.Fatalf("expected 2 top traits, got %d", len(top))
	}
	if top[0].Key != "language" {
		t.Errorf("expected language as #1, got %s", top[0].Key)
	}
	if top[1].Key != "editor" {
		t.Errorf("expected editor as #2, got %s", top[1].Key)
	}
}

func TestSoul_TopTraits_MoreThanAvailable(t *testing.T) {
	s := NewSoul("user1")
	s.Observe("x", "1")

	top := s.TopTraits(10)
	if len(top) != 1 {
		t.Fatalf("expected 1, got %d", len(top))
	}
}

func TestSoul_Summary(t *testing.T) {
	s := NewSoul("user1")
	if !strings.Contains(s.Summary(), "No profile data") {
		t.Error("empty profile should say no data")
	}

	s.Observe("language", "Go")
	summary := s.Summary()
	if !strings.Contains(summary, "language: Go") {
		t.Errorf("summary should contain trait, got: %s", summary)
	}
}

func TestProfileStore_GetOrCreate(t *testing.T) {
	ps := NewProfileStore()

	s1 := ps.GetOrCreate("user1")
	s1.Observe("key", "val")

	s2 := ps.GetOrCreate("user1")
	if s2.Count() != 1 {
		t.Error("expected same profile instance")
	}

	if ps.Get("user2") != nil {
		t.Error("expected nil for unknown user")
	}
}
