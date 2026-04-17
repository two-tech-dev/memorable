package graph

import (
	"context"
	"testing"
)

func TestInMemoryGraph_UpsertAndGet(t *testing.T) {
	g := NewInMemoryGraph()
	ctx := context.Background()

	e := &Entity{
		ID:        "ent_001",
		Name:      "React",
		Type:      "technology",
		MemoryIDs: []string{"mem_1"},
	}

	if err := g.UpsertEntity(ctx, e); err != nil {
		t.Fatal(err)
	}

	got, err := g.GetEntity(ctx, "ent_001")
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || got.Name != "React" {
		t.Fatalf("expected React, got %v", got)
	}

	// Upsert merge memory IDs
	e2 := &Entity{
		ID:        "ent_001",
		Name:      "React",
		Type:      "technology",
		MemoryIDs: []string{"mem_2"},
	}
	if err := g.UpsertEntity(ctx, e2); err != nil {
		t.Fatal(err)
	}

	got, _ = g.GetEntity(ctx, "ent_001")
	if len(got.MemoryIDs) != 2 {
		t.Fatalf("expected 2 memory IDs, got %d", len(got.MemoryIDs))
	}
}

func TestInMemoryGraph_FindEntities(t *testing.T) {
	g := NewInMemoryGraph()
	ctx := context.Background()

	entities := []*Entity{
		{ID: "e1", Name: "PostgreSQL", Type: "technology"},
		{ID: "e2", Name: "pgvector", Type: "technology"},
		{ID: "e3", Name: "Redis", Type: "technology"},
	}
	for _, e := range entities {
		_ = g.UpsertEntity(ctx, e)
	}

	found, err := g.FindEntities(ctx, "postg", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != 1 {
		t.Fatalf("expected 1 entity matching 'postg', got %d", len(found))
	}

	found2, _ := g.FindEntities(ctx, "pg", 10)
	if len(found2) != 1 {
		t.Fatalf("expected 1 entity matching 'pg', got %d", len(found2))
	}
}

func TestInMemoryGraph_Relations(t *testing.T) {
	g := NewInMemoryGraph()
	ctx := context.Background()

	_ = g.UpsertEntity(ctx, &Entity{ID: "e1", Name: "App"})
	_ = g.UpsertEntity(ctx, &Entity{ID: "e2", Name: "PostgreSQL"})

	rel := &Relation{
		ID:     "r1",
		FromID: "e1",
		ToID:   "e2",
		Type:   "uses",
		Weight: 1.0,
	}
	if err := g.UpsertRelation(ctx, rel); err != nil {
		t.Fatal(err)
	}

	rels, err := g.GetRelations(ctx, "e1")
	if err != nil {
		t.Fatal(err)
	}
	if len(rels) != 1 || rels[0].Type != "uses" {
		t.Fatalf("unexpected relations: %v", rels)
	}

	// Also accessible from the other side
	rels2, _ := g.GetRelations(ctx, "e2")
	if len(rels2) != 1 {
		t.Fatal("expected relation accessible from both sides")
	}
}

func TestInMemoryGraph_GetNeighbors(t *testing.T) {
	g := NewInMemoryGraph()
	ctx := context.Background()

	// A → B → C
	_ = g.UpsertEntity(ctx, &Entity{ID: "a", Name: "A"})
	_ = g.UpsertEntity(ctx, &Entity{ID: "b", Name: "B"})
	_ = g.UpsertEntity(ctx, &Entity{ID: "c", Name: "C"})
	_ = g.UpsertRelation(ctx, &Relation{ID: "r1", FromID: "a", ToID: "b", Type: "related_to"})
	_ = g.UpsertRelation(ctx, &Relation{ID: "r2", FromID: "b", ToID: "c", Type: "related_to"})

	// Depth 1: only B
	entities, relations, err := g.GetNeighbors(ctx, "a", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(entities) != 1 || entities[0].ID != "b" {
		t.Fatalf("depth 1: expected [b], got %v", entities)
	}
	if len(relations) != 1 {
		t.Fatalf("depth 1: expected 1 relation, got %d", len(relations))
	}

	// Depth 2: B and C
	entities2, relations2, _ := g.GetNeighbors(ctx, "a", 2)
	if len(entities2) != 2 {
		t.Fatalf("depth 2: expected 2 entities, got %d", len(entities2))
	}
	if len(relations2) != 2 {
		t.Fatalf("depth 2: expected 2 relations, got %d", len(relations2))
	}
}

func TestInMemoryGraph_Stats(t *testing.T) {
	g := NewInMemoryGraph()
	ctx := context.Background()

	_ = g.UpsertEntity(ctx, &Entity{ID: "e1", Name: "X"})
	_ = g.UpsertEntity(ctx, &Entity{ID: "e2", Name: "Y"})
	_ = g.UpsertRelation(ctx, &Relation{ID: "r1", FromID: "e1", ToID: "e2"})

	ec, rc := g.Stats()
	if ec != 2 || rc != 1 {
		t.Fatalf("expected 2 entities, 1 relation; got %d, %d", ec, rc)
	}
}

func TestExtractTriples(t *testing.T) {
	content := "The project uses PostgreSQL. React depends on Node. User migrated to Go."
	triples := ExtractTriples(content, "mem_123")

	if len(triples) != 3 {
		t.Fatalf("expected 3 triples, got %d: %+v", len(triples), triples)
	}

	// Check first triple
	if triples[0].Subject != "The project" || triples[0].Predicate != "uses" || triples[0].Object != "PostgreSQL" {
		t.Errorf("unexpected triple[0]: %+v", triples[0])
	}

	if triples[1].Predicate != "depends_on" {
		t.Errorf("expected depends_on, got %s", triples[1].Predicate)
	}

	if triples[2].Predicate != "supersedes" {
		t.Errorf("expected supersedes, got %s", triples[2].Predicate)
	}
}

func TestTriplesToGraph(t *testing.T) {
	triples := []Triple{
		{Subject: "App", Predicate: "uses", Object: "PostgreSQL", MemoryID: "m1"},
		{Subject: "App", Predicate: "uses", Object: "Redis", MemoryID: "m2"},
	}

	entities, relations := TriplesToGraph(triples)

	if len(entities) != 3 { // App, PostgreSQL, Redis
		t.Fatalf("expected 3 entities, got %d", len(entities))
	}
	if len(relations) != 2 {
		t.Fatalf("expected 2 relations, got %d", len(relations))
	}

	// App should have 2 memory IDs
	for _, e := range entities {
		if e.Name == "App" {
			if len(e.MemoryIDs) != 2 {
				t.Errorf("App should have 2 memory IDs, got %d", len(e.MemoryIDs))
			}
		}
	}
}

func TestInferEntityType(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"PostgreSQL", "technology"},
		{"react", "technology"},
		{"src/main.go", "file"},
		{"UserService", "concept"},
		{"docker-compose", "technology"},
	}

	for _, tt := range tests {
		got := inferEntityType(tt.name)
		if got != tt.expected {
			t.Errorf("inferEntityType(%q) = %q, want %q", tt.name, got, tt.expected)
		}
	}
}

func TestGetEntity_NotFound(t *testing.T) {
	g := NewInMemoryGraph()
	got, err := g.GetEntity(context.Background(), "nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatal("expected nil for nonexistent entity")
	}
}
