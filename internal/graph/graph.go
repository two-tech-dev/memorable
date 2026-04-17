package graph

import (
	"context"
	"sync"
	"time"
)

// Entity represents a named concept extracted from memories.
type Entity struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Type      string         `json:"type"` // person | project | technology | concept | file | decision
	MemoryIDs []string       `json:"memory_ids"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// Relation represents a directed edge between two entities.
type Relation struct {
	ID        string    `json:"id"`
	FromID    string    `json:"from_id"`
	ToID      string    `json:"to_id"`
	Type      string    `json:"type"` // uses | depends_on | created_by | related_to | supersedes | part_of
	Weight    float64   `json:"weight"`
	MemoryIDs []string  `json:"memory_ids"`
	CreatedAt time.Time `json:"created_at"`
}

// Triple is a subject-predicate-object extracted from text.
type Triple struct {
	Subject   string `json:"subject"`
	Predicate string `json:"predicate"`
	Object    string `json:"object"`
	MemoryID  string `json:"memory_id"`
}

// EntityStore defines the persistence interface for entities and relations.
type EntityStore interface {
	UpsertEntity(ctx context.Context, entity *Entity) error
	GetEntity(ctx context.Context, id string) (*Entity, error)
	FindEntities(ctx context.Context, query string, limit int) ([]*Entity, error)
	UpsertRelation(ctx context.Context, rel *Relation) error
	GetRelations(ctx context.Context, entityID string) ([]*Relation, error)
	GetNeighbors(ctx context.Context, entityID string, depth int) ([]*Entity, []*Relation, error)
}

// InMemoryGraph is a basic in-memory knowledge graph for MVP.
type InMemoryGraph struct {
	mu        sync.RWMutex
	entities  map[string]*Entity
	relations map[string]*Relation
	// index: entity name → entity ID (lowercase)
	nameIndex map[string]string
	// adjacency: entity ID → []relation IDs
	adjacency map[string][]string
}

func NewInMemoryGraph() *InMemoryGraph {
	return &InMemoryGraph{
		entities:  make(map[string]*Entity),
		relations: make(map[string]*Relation),
		nameIndex: make(map[string]string),
		adjacency: make(map[string][]string),
	}
}

func (g *InMemoryGraph) UpsertEntity(_ context.Context, entity *Entity) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if existing, ok := g.entities[entity.ID]; ok {
		existing.Name = entity.Name
		existing.Type = entity.Type
		existing.UpdatedAt = time.Now().UTC()
		// Merge memory IDs
		seen := make(map[string]bool)
		for _, id := range existing.MemoryIDs {
			seen[id] = true
		}
		for _, id := range entity.MemoryIDs {
			if !seen[id] {
				existing.MemoryIDs = append(existing.MemoryIDs, id)
			}
		}
	} else {
		g.entities[entity.ID] = entity
	}

	g.nameIndex[lowercase(entity.Name)] = entity.ID
	return nil
}

func (g *InMemoryGraph) GetEntity(_ context.Context, id string) (*Entity, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	e, ok := g.entities[id]
	if !ok {
		return nil, nil
	}
	return e, nil
}

func (g *InMemoryGraph) FindEntities(_ context.Context, query string, limit int) ([]*Entity, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	q := lowercase(query)
	var results []*Entity
	for name, id := range g.nameIndex {
		if contains(name, q) {
			if e, ok := g.entities[id]; ok {
				results = append(results, e)
			}
		}
		if len(results) >= limit {
			break
		}
	}
	return results, nil
}

func (g *InMemoryGraph) UpsertRelation(_ context.Context, rel *Relation) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.relations[rel.ID] = rel
	g.adjacency[rel.FromID] = appendUnique(g.adjacency[rel.FromID], rel.ID)
	g.adjacency[rel.ToID] = appendUnique(g.adjacency[rel.ToID], rel.ID)
	return nil
}

func (g *InMemoryGraph) GetRelations(_ context.Context, entityID string) ([]*Relation, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var results []*Relation
	for _, relID := range g.adjacency[entityID] {
		if r, ok := g.relations[relID]; ok {
			results = append(results, r)
		}
	}
	return results, nil
}

func (g *InMemoryGraph) GetNeighbors(_ context.Context, entityID string, depth int) ([]*Entity, []*Relation, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	visitedEntities := make(map[string]bool)
	visitedRelations := make(map[string]bool)
	var entities []*Entity
	var relations []*Relation

	queue := []string{entityID}
	visitedEntities[entityID] = true

	for d := 0; d < depth && len(queue) > 0; d++ {
		var nextQueue []string
		for _, eid := range queue {
			for _, relID := range g.adjacency[eid] {
				if visitedRelations[relID] {
					continue
				}
				visitedRelations[relID] = true

				rel := g.relations[relID]
				relations = append(relations, rel)

				neighborID := rel.ToID
				if neighborID == eid {
					neighborID = rel.FromID
				}

				if !visitedEntities[neighborID] {
					visitedEntities[neighborID] = true
					if e, ok := g.entities[neighborID]; ok {
						entities = append(entities, e)
					}
					nextQueue = append(nextQueue, neighborID)
				}
			}
		}
		queue = nextQueue
	}

	return entities, relations, nil
}

// Stats returns graph statistics.
func (g *InMemoryGraph) Stats() (entityCount, relationCount int) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.entities), len(g.relations)
}

func lowercase(s string) string {
	b := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}

func contains(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func appendUnique(slice []string, val string) []string {
	for _, v := range slice {
		if v == val {
			return slice
		}
	}
	return append(slice, val)
}
