package memory

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"
)

type MemoryType string

const (
	TypeFact         MemoryType = "fact"
	TypeConversation MemoryType = "conversation"
	TypeDecision     MemoryType = "decision"
	TypeCodePattern  MemoryType = "code_pattern"
	TypeCorrection   MemoryType = "correction"
)

var ValidMemoryTypes = map[MemoryType]bool{
	TypeFact:         true,
	TypeConversation: true,
	TypeDecision:     true,
	TypeCodePattern:  true,
	TypeCorrection:   true,
}

type Memory struct {
	ID          string         `json:"id"`
	Content     string         `json:"content"`
	MemoryType  MemoryType     `json:"memory_type"`
	UserID      string         `json:"user_id,omitempty"`
	AgentID     string         `json:"agent_id,omitempty"`
	AppID       string         `json:"app_id,omitempty"`
	RunID       string         `json:"run_id,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	ContentHash string         `json:"content_hash"`
	Embedding   []float32      `json:"-"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type SearchResult struct {
	Memory   Memory  `json:"memory"`
	Score    float64 `json:"score"`
	Distance float64 `json:"distance"`
}

type SearchFilter struct {
	UserID     string     `json:"user_id,omitempty"`
	AgentID    string     `json:"agent_id,omitempty"`
	AppID      string     `json:"app_id,omitempty"`
	RunID      string     `json:"run_id,omitempty"`
	MemoryType MemoryType `json:"memory_type,omitempty"`
	Since      *time.Time `json:"since,omitempty"`
	Until      *time.Time `json:"until,omitempty"`
}

type Stats struct {
	Total       int            `json:"total"`
	ByType      map[string]int `json:"by_type"`
	ByUser      map[string]int `json:"by_user,omitempty"`
	OldestAt    *time.Time     `json:"oldest_at,omitempty"`
	NewestAt    *time.Time     `json:"newest_at,omitempty"`
}

func ContentHash(content string) string {
	h := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", h)
}

func (mt MemoryType) Valid() bool {
	return ValidMemoryTypes[mt]
}

// VectorStore is the storage interface for memories.
// Implementations live in internal/store.
type VectorStore interface {
	Insert(ctx context.Context, mem *Memory) error
	Update(ctx context.Context, mem *Memory) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*Memory, error)
	GetByHash(ctx context.Context, contentHash string, filter *SearchFilter) (*Memory, error)
	Search(ctx context.Context, vector []float32, k int, filter *SearchFilter) ([]SearchResult, error)
	List(ctx context.Context, filter *SearchFilter, limit, offset int) ([]*Memory, int, error)
	Stats(ctx context.Context, filter *SearchFilter) (*Stats, error)
	Close() error
}
