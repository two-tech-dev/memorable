package profile

import (
	"sort"
	"sync"
	"time"
)

// Trait captures a single behavioral or preference pattern.
type Trait struct {
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	Occurrences int       `json:"occurrences"`
	FirstSeen   time.Time `json:"first_seen"`
	LastSeen    time.Time `json:"last_seen"`
}

// Soul aggregates patterns from memories into a user/agent profile.
type Soul struct {
	mu     sync.RWMutex
	UserID string           `json:"user_id"`
	Traits map[string]*Trait `json:"traits"` // key → Trait
}

func NewSoul(userID string) *Soul {
	return &Soul{
		UserID: userID,
		Traits: make(map[string]*Trait),
	}
}

// Observe records a trait observation. If the trait exists, increments count;
// otherwise creates it.
func (s *Soul) Observe(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()

	if t, ok := s.Traits[key]; ok {
		t.Value = value
		t.Occurrences++
		t.LastSeen = now
	} else {
		s.Traits[key] = &Trait{
			Key:         key,
			Value:       value,
			Occurrences: 1,
			FirstSeen:   now,
			LastSeen:    now,
		}
	}
}

// Get returns a trait by key.
func (s *Soul) Get(key string) (*Trait, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	t, ok := s.Traits[key]
	return t, ok
}

// TopTraits returns the N most frequently observed traits.
func (s *Soul) TopTraits(n int) []*Trait {
	s.mu.RLock()
	defer s.mu.RUnlock()

	traits := make([]*Trait, 0, len(s.Traits))
	for _, t := range s.Traits {
		traits = append(traits, t)
	}

	sort.Slice(traits, func(i, j int) bool {
		return traits[i].Occurrences > traits[j].Occurrences
	})

	if n > len(traits) {
		n = len(traits)
	}
	return traits[:n]
}

// Summary produces a text description of the profile.
func (s *Soul) Summary() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.Traits) == 0 {
		return "No profile data yet."
	}

	var b []byte
	b = append(b, "Profile for "+s.UserID+":\n"...)

	top := s.TopTraits(10)
	for _, t := range top {
		b = append(b, "- "+t.Key+": "+t.Value+"\n"...)
	}

	return string(b)
}

// Count returns the number of traits.
func (s *Soul) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.Traits)
}

// ProfileStore manages multiple Soul profiles.
type ProfileStore struct {
	mu       sync.RWMutex
	profiles map[string]*Soul
}

func NewProfileStore() *ProfileStore {
	return &ProfileStore{
		profiles: make(map[string]*Soul),
	}
}

// GetOrCreate returns the profile for a user, creating if needed.
func (ps *ProfileStore) GetOrCreate(userID string) *Soul {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if s, ok := ps.profiles[userID]; ok {
		return s
	}
	s := NewSoul(userID)
	ps.profiles[userID] = s
	return s
}

// Get returns an existing profile or nil.
func (ps *ProfileStore) Get(userID string) *Soul {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.profiles[userID]
}
