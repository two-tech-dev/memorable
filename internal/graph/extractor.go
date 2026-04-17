package graph

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"
)

// ExtractTriples extracts subject-predicate-object triples from text content.
// Uses simple pattern matching for MVP; can be upgraded to LLM-based extraction later.
func ExtractTriples(content, memoryID string) []Triple {
	var triples []Triple
	sentences := splitSentences(content)

	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if len(sentence) < 5 {
			continue
		}

		// Pattern: "X uses Y", "X depends on Y", "X is a Y"
		for _, pattern := range relationPatterns {
			if t, ok := matchPattern(sentence, pattern, memoryID); ok {
				triples = append(triples, t)
			}
		}
	}

	return triples
}

type patternDef struct {
	marker    string
	predicate string
}

var relationPatterns = []patternDef{
	{" uses ", "uses"},
	{" depends on ", "depends_on"},
	{" is a ", "is_a"},
	{" is an ", "is_a"},
	{" created ", "created"},
	{" built ", "created"},
	{" works with ", "works_with"},
	{" replaced ", "supersedes"},
	{" migrated to ", "supersedes"},
	{" switched to ", "supersedes"},
	{" is part of ", "part_of"},
	{" belongs to ", "part_of"},
	{" implemented ", "created"},
	{" added ", "created"},
	{" configured ", "configured"},
	{" prefers ", "prefers"},
	{" chose ", "prefers"},
}

func matchPattern(sentence string, pattern patternDef, memoryID string) (Triple, bool) {
	lower := strings.ToLower(sentence)
	idx := strings.Index(lower, pattern.marker)
	if idx < 0 {
		return Triple{}, false
	}

	subject := strings.TrimSpace(sentence[:idx])
	object := strings.TrimSpace(sentence[idx+len(pattern.marker):])

	// Clean up subject/object
	subject = trimQuotes(subject)
	object = trimQuotes(object)

	if len(subject) < 2 || len(object) < 2 || len(subject) > 100 || len(object) > 100 {
		return Triple{}, false
	}

	return Triple{
		Subject:   subject,
		Predicate: pattern.predicate,
		Object:    object,
		MemoryID:  memoryID,
	}, true
}

// TriplesToGraph converts extracted triples into entities and relations.
func TriplesToGraph(triples []Triple) ([]*Entity, []*Relation) {
	entityMap := make(map[string]*Entity) // lowercase name → Entity
	var relations []*Relation

	now := time.Now().UTC()

	for _, t := range triples {
		subjectKey := strings.ToLower(t.Subject)
		objectKey := strings.ToLower(t.Object)

		// Upsert subject entity
		if _, ok := entityMap[subjectKey]; !ok {
			entityMap[subjectKey] = &Entity{
				ID:        entityID(t.Subject),
				Name:      t.Subject,
				Type:      inferEntityType(t.Subject),
				MemoryIDs: []string{t.MemoryID},
				CreatedAt: now,
				UpdatedAt: now,
			}
		} else {
			addMemoryID(entityMap[subjectKey], t.MemoryID)
		}

		// Upsert object entity
		if _, ok := entityMap[objectKey]; !ok {
			entityMap[objectKey] = &Entity{
				ID:        entityID(t.Object),
				Name:      t.Object,
				Type:      inferEntityType(t.Object),
				MemoryIDs: []string{t.MemoryID},
				CreatedAt: now,
				UpdatedAt: now,
			}
		} else {
			addMemoryID(entityMap[objectKey], t.MemoryID)
		}

		// Create relation
		fromID := entityMap[subjectKey].ID
		toID := entityMap[objectKey].ID
		relID := relationID(fromID, t.Predicate, toID)

		relations = append(relations, &Relation{
			ID:        relID,
			FromID:    fromID,
			ToID:      toID,
			Type:      t.Predicate,
			Weight:    1.0,
			MemoryIDs: []string{t.MemoryID},
			CreatedAt: now,
		})
	}

	var entities []*Entity
	for _, e := range entityMap {
		entities = append(entities, e)
	}

	return entities, relations
}

func entityID(name string) string {
	h := sha256.Sum256([]byte(strings.ToLower(name)))
	return fmt.Sprintf("ent_%x", h[:8])
}

func relationID(fromID, predicate, toID string) string {
	h := sha256.Sum256([]byte(fromID + ":" + predicate + ":" + toID))
	return fmt.Sprintf("rel_%x", h[:8])
}

func addMemoryID(e *Entity, memoryID string) {
	for _, id := range e.MemoryIDs {
		if id == memoryID {
			return
		}
	}
	e.MemoryIDs = append(e.MemoryIDs, memoryID)
}

// inferEntityType guesses the entity type from naming patterns.
func inferEntityType(name string) string {
	lower := strings.ToLower(name)

	// File paths
	if strings.Contains(lower, ".") && (strings.Contains(lower, "/") || strings.Contains(lower, "\\")) {
		return "file"
	}
	if strings.HasSuffix(lower, ".go") || strings.HasSuffix(lower, ".ts") ||
		strings.HasSuffix(lower, ".py") || strings.HasSuffix(lower, ".js") {
		return "file"
	}

	// Known tech words
	techWords := []string{"react", "go", "python", "rust", "docker", "kubernetes", "postgres",
		"redis", "nginx", "node", "typescript", "javascript", "vue", "angular", "django",
		"flask", "fastapi", "grpc", "graphql", "rest", "api", "sdk", "cli", "mcp",
		"openai", "gemini", "ollama", "pgvector", "git"}
	for _, tw := range techWords {
		if lower == tw || strings.Contains(lower, tw) {
			return "technology"
		}
	}

	// Capitalized single word likely a name
	if len(name) > 1 && name[0] >= 'A' && name[0] <= 'Z' && !strings.Contains(name, " ") {
		return "concept"
	}

	return "concept"
}

func splitSentences(text string) []string {
	var sentences []string
	var current strings.Builder
	for _, r := range text {
		current.WriteRune(r)
		if r == '.' || r == '!' || r == '?' || r == '\n' {
			s := strings.TrimSpace(current.String())
			if len(s) > 0 {
				sentences = append(sentences, s)
			}
			current.Reset()
		}
	}
	if s := strings.TrimSpace(current.String()); len(s) > 0 {
		sentences = append(sentences, s)
	}
	return sentences
}

func trimQuotes(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			s = s[1 : len(s)-1]
		}
	}
	// Remove trailing punctuation
	s = strings.TrimRight(s, ".,;:!?")
	return strings.TrimSpace(s)
}
