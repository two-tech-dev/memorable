package memory

import (
	"context"
	"fmt"
	"testing"
)

// mockStore is a simple in-memory VectorStore for testing.
type mockStore struct {
	memories map[string]*Memory
	inserted int
	updated  int
	deleted  int
}

func newMockStore() *mockStore {
	return &mockStore{memories: make(map[string]*Memory)}
}

func (s *mockStore) Insert(_ context.Context, mem *Memory) error {
	s.memories[mem.ID] = mem
	s.inserted++
	return nil
}

func (s *mockStore) Update(_ context.Context, mem *Memory) error {
	if _, ok := s.memories[mem.ID]; !ok {
		return fmt.Errorf("not found")
	}
	s.memories[mem.ID] = mem
	s.updated++
	return nil
}

func (s *mockStore) Delete(_ context.Context, id string) error {
	if _, ok := s.memories[id]; !ok {
		return fmt.Errorf("not found")
	}
	delete(s.memories, id)
	s.deleted++
	return nil
}

func (s *mockStore) Get(_ context.Context, id string) (*Memory, error) {
	m, ok := s.memories[id]
	if !ok {
		return nil, nil
	}
	return m, nil
}

func (s *mockStore) GetByHash(_ context.Context, hash string, _ *SearchFilter) (*Memory, error) {
	for _, m := range s.memories {
		if m.ContentHash == hash {
			return m, nil
		}
	}
	return nil, nil
}

func (s *mockStore) Search(_ context.Context, _ []float32, k int, _ *SearchFilter) ([]SearchResult, error) {
	var results []SearchResult
	for _, m := range s.memories {
		results = append(results, SearchResult{Memory: *m, Score: 0.95, Distance: 0.05})
		if len(results) >= k {
			break
		}
	}
	return results, nil
}

func (s *mockStore) List(_ context.Context, _ *SearchFilter, limit, offset int) ([]*Memory, int, error) {
	var all []*Memory
	for _, m := range s.memories {
		all = append(all, m)
	}
	total := len(all)
	if offset >= total {
		return nil, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return all[offset:end], total, nil
}

func (s *mockStore) Stats(_ context.Context, _ *SearchFilter) (*Stats, error) {
	st := &Stats{
		Total:  len(s.memories),
		ByType: make(map[string]int),
	}
	for _, m := range s.memories {
		st.ByType[string(m.MemoryType)]++
	}
	return st, nil
}

func (s *mockStore) Close() error { return nil }

// mockEmbedder returns a fixed vector for testing.
type mockEmbedder struct {
	dims int
}

func (e *mockEmbedder) Embed(_ context.Context, _ string) ([]float32, error) {
	v := make([]float32, e.dims)
	for i := range v {
		v[i] = 0.1
	}
	return v, nil
}

func (e *mockEmbedder) EmbedBatch(_ context.Context, texts []string) ([][]float32, error) {
	results := make([][]float32, len(texts))
	for i := range texts {
		v := make([]float32, e.dims)
		for j := range v {
			v[j] = 0.1
		}
		results[i] = v
	}
	return results, nil
}

func (e *mockEmbedder) Dimensions() int { return e.dims }

func TestManagerAdd(t *testing.T) {
	store := newMockStore()
	embedder := &mockEmbedder{dims: 10}
	mgr := NewManager(store, embedder)
	ctx := context.Background()

	mem, err := mgr.Add(ctx, "Go is great", TypeFact, nil, nil)
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if mem.Content != "Go is great" {
		t.Errorf("expected content 'Go is great', got %q", mem.Content)
	}
	if mem.MemoryType != TypeFact {
		t.Errorf("expected type fact, got %q", mem.MemoryType)
	}
	if mem.ID == "" {
		t.Error("expected non-empty ID")
	}
	if store.inserted != 1 {
		t.Errorf("expected 1 insert, got %d", store.inserted)
	}
}

func TestManagerAddDedup(t *testing.T) {
	store := newMockStore()
	embedder := &mockEmbedder{dims: 10}
	mgr := NewManager(store, embedder)
	ctx := context.Background()

	m1, err := mgr.Add(ctx, "same content", TypeFact, nil, nil)
	if err != nil {
		t.Fatalf("Add first: %v", err)
	}

	m2, err := mgr.Add(ctx, "same content", TypeFact, nil, nil)
	if err != nil {
		t.Fatalf("Add second: %v", err)
	}

	if m1.ID != m2.ID {
		t.Error("duplicate content should return same memory")
	}
	if store.inserted != 1 {
		t.Errorf("expected 1 insert (dedup), got %d", store.inserted)
	}
}

func TestManagerAddValidation(t *testing.T) {
	store := newMockStore()
	embedder := &mockEmbedder{dims: 10}
	mgr := NewManager(store, embedder)
	ctx := context.Background()

	_, err := mgr.Add(ctx, "", TypeFact, nil, nil)
	if err == nil {
		t.Error("expected error for empty content")
	}

	_, err = mgr.Add(ctx, "content", MemoryType("invalid"), nil, nil)
	if err == nil {
		t.Error("expected error for invalid type")
	}
}

func TestManagerAddWithScope(t *testing.T) {
	store := newMockStore()
	embedder := &mockEmbedder{dims: 10}
	mgr := NewManager(store, embedder)
	ctx := context.Background()

	filter := &SearchFilter{UserID: "alice", AgentID: "cursor"}
	mem, err := mgr.Add(ctx, "user-scoped memory", TypeConversation, filter, nil)
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if mem.UserID != "alice" {
		t.Errorf("expected UserID alice, got %q", mem.UserID)
	}
	if mem.AgentID != "cursor" {
		t.Errorf("expected AgentID cursor, got %q", mem.AgentID)
	}
}

func TestManagerSearch(t *testing.T) {
	store := newMockStore()
	embedder := &mockEmbedder{dims: 10}
	mgr := NewManager(store, embedder)
	ctx := context.Background()

	_, _ = mgr.Add(ctx, "Go uses goroutines for concurrency", TypeFact, nil, nil)

	results, err := mgr.Search(ctx, "concurrency in Go", 5, nil)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected at least one result")
	}
}

func TestManagerSearchValidation(t *testing.T) {
	store := newMockStore()
	embedder := &mockEmbedder{dims: 10}
	mgr := NewManager(store, embedder)
	ctx := context.Background()

	_, err := mgr.Search(ctx, "", 5, nil)
	if err == nil {
		t.Error("expected error for empty query")
	}
}

func TestManagerGet(t *testing.T) {
	store := newMockStore()
	embedder := &mockEmbedder{dims: 10}
	mgr := NewManager(store, embedder)
	ctx := context.Background()

	mem, _ := mgr.Add(ctx, "test memory", TypeFact, nil, nil)

	got, err := mgr.Get(ctx, mem.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Content != "test memory" {
		t.Errorf("expected content 'test memory', got %q", got.Content)
	}
}

func TestManagerGetNotFound(t *testing.T) {
	store := newMockStore()
	embedder := &mockEmbedder{dims: 10}
	mgr := NewManager(store, embedder)
	ctx := context.Background()

	_, err := mgr.Get(ctx, "nonexistent-id")
	if err == nil {
		t.Error("expected error for nonexistent ID")
	}
}

func TestManagerUpdate(t *testing.T) {
	store := newMockStore()
	embedder := &mockEmbedder{dims: 10}
	mgr := NewManager(store, embedder)
	ctx := context.Background()

	mem, _ := mgr.Add(ctx, "original", TypeFact, nil, nil)

	newContent := "updated content"
	updated, err := mgr.Update(ctx, mem.ID, &newContent, map[string]any{"reviewed": true})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Content != "updated content" {
		t.Errorf("expected updated content, got %q", updated.Content)
	}
	if updated.Metadata["reviewed"] != true {
		t.Error("expected metadata 'reviewed' to be true")
	}
}

func TestManagerUpdateNotFound(t *testing.T) {
	store := newMockStore()
	embedder := &mockEmbedder{dims: 10}
	mgr := NewManager(store, embedder)
	ctx := context.Background()

	newContent := "updated"
	_, err := mgr.Update(ctx, "nonexistent", &newContent, nil)
	if err == nil {
		t.Error("expected error for nonexistent ID")
	}
}

func TestManagerDelete(t *testing.T) {
	store := newMockStore()
	embedder := &mockEmbedder{dims: 10}
	mgr := NewManager(store, embedder)
	ctx := context.Background()

	mem, _ := mgr.Add(ctx, "to delete", TypeFact, nil, nil)

	if err := mgr.Delete(ctx, mem.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := mgr.Get(ctx, mem.ID)
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestManagerList(t *testing.T) {
	store := newMockStore()
	embedder := &mockEmbedder{dims: 10}
	mgr := NewManager(store, embedder)
	ctx := context.Background()

	_, _ = mgr.Add(ctx, "memory one", TypeFact, nil, nil)
	_, _ = mgr.Add(ctx, "memory two", TypeDecision, nil, nil)

	memories, total, err := mgr.List(ctx, nil, 10, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
	if len(memories) != 2 {
		t.Errorf("expected 2 memories, got %d", len(memories))
	}
}

func TestManagerStats(t *testing.T) {
	store := newMockStore()
	embedder := &mockEmbedder{dims: 10}
	mgr := NewManager(store, embedder)
	ctx := context.Background()

	_, _ = mgr.Add(ctx, "fact one", TypeFact, nil, nil)
	_, _ = mgr.Add(ctx, "decision one", TypeDecision, nil, nil)
	_, _ = mgr.Add(ctx, "fact two", TypeFact, nil, nil)

	stats, err := mgr.Stats(ctx, nil)
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if stats.Total != 3 {
		t.Errorf("expected total 3, got %d", stats.Total)
	}
	if stats.ByType["fact"] != 2 {
		t.Errorf("expected 2 facts, got %d", stats.ByType["fact"])
	}
	if stats.ByType["decision"] != 1 {
		t.Errorf("expected 1 decision, got %d", stats.ByType["decision"])
	}
}
