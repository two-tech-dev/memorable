package memory

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/two-tech-dev/memorable/internal/embedding"
)

type Manager struct {
	store    VectorStore
	embedder embedding.Provider
}

func NewManager(s VectorStore, e embedding.Provider) *Manager {
	return &Manager{store: s, embedder: e}
}

func (m *Manager) Add(ctx context.Context, content string, memType MemoryType, filter *SearchFilter, metadata map[string]any) (*Memory, error) {
	if !memType.Valid() {
		return nil, fmt.Errorf("invalid memory type: %s", memType)
	}
	if content == "" {
		return nil, fmt.Errorf("content cannot be empty")
	}

	hash := ContentHash(content)

	// Dedup: check if same content exists in same scope
	existing, err := m.store.GetByHash(ctx, hash, filter)
	if err != nil {
		return nil, fmt.Errorf("dedup check: %w", err)
	}
	if existing != nil {
		return existing, nil
	}

	// Embed content
	emb, err := m.embedder.Embed(ctx, content)
	if err != nil {
		return nil, fmt.Errorf("embed content: %w", err)
	}

	now := time.Now().UTC()
	mem := &Memory{
		ID:          uuid.New().String(),
		Content:     content,
		MemoryType:  memType,
		Metadata:    metadata,
		ContentHash: hash,
		Embedding:   emb,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if filter != nil {
		mem.UserID = filter.UserID
		mem.AgentID = filter.AgentID
		mem.AppID = filter.AppID
		mem.RunID = filter.RunID
	}

	if mem.Metadata == nil {
		mem.Metadata = make(map[string]any)
	}

	if err := m.store.Insert(ctx, mem); err != nil {
		return nil, fmt.Errorf("store memory: %w", err)
	}
	return mem, nil
}

func (m *Manager) Search(ctx context.Context, query string, k int, filter *SearchFilter) ([]SearchResult, error) {
	if query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}
	if k <= 0 {
		k = 10
	}

	emb, err := m.embedder.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}

	results, err := m.store.Search(ctx, emb, k, filter)
	if err != nil {
		return nil, fmt.Errorf("search memories: %w", err)
	}
	return results, nil
}

func (m *Manager) Get(ctx context.Context, id string) (*Memory, error) {
	mem, err := m.store.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get memory: %w", err)
	}
	if mem == nil {
		return nil, fmt.Errorf("memory not found: %s", id)
	}
	return mem, nil
}

func (m *Manager) Update(ctx context.Context, id string, content *string, metadata map[string]any) (*Memory, error) {
	mem, err := m.store.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get memory for update: %w", err)
	}
	if mem == nil {
		return nil, fmt.Errorf("memory not found: %s", id)
	}

	if content != nil && *content != mem.Content {
		mem.Content = *content
		mem.ContentHash = ContentHash(*content)

		emb, err := m.embedder.Embed(ctx, *content)
		if err != nil {
			return nil, fmt.Errorf("re-embed content: %w", err)
		}
		mem.Embedding = emb
	}

	if metadata != nil {
		if mem.Metadata == nil {
			mem.Metadata = make(map[string]any)
		}
		for k, v := range metadata {
			mem.Metadata[k] = v
		}
	}

	mem.UpdatedAt = time.Now().UTC()

	if err := m.store.Update(ctx, mem); err != nil {
		return nil, fmt.Errorf("update memory: %w", err)
	}
	return mem, nil
}

func (m *Manager) Delete(ctx context.Context, id string) error {
	return m.store.Delete(ctx, id)
}

func (m *Manager) List(ctx context.Context, filter *SearchFilter, limit, offset int) ([]*Memory, int, error) {
	if limit <= 0 {
		limit = 20
	}
	return m.store.List(ctx, filter, limit, offset)
}

func (m *Manager) Stats(ctx context.Context, filter *SearchFilter) (*Stats, error) {
	return m.store.Stats(ctx, filter)
}
