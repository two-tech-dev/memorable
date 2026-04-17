package heartbeat

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/two-tech-dev/memorable/internal/memory"
)

// Insight represents a derived insight from memory consolidation.
type Insight struct {
	Content    string         `json:"content"`
	SourceIDs  []string       `json:"source_ids"`
	InsightType string        `json:"insight_type"` // summary | contradiction | pattern
	Confidence float64        `json:"confidence"`
	CreatedAt  time.Time      `json:"created_at"`
}

// ConsolidationResult holds the output of a heartbeat cycle.
type ConsolidationResult struct {
	Insights       []Insight `json:"insights"`
	MergedCount    int       `json:"merged_count"`
	ProcessedCount int       `json:"processed_count"`
	Duration       time.Duration `json:"duration"`
}

// Consolidator analyzes stored memories and produces insights.
type Consolidator struct {
	mgr            *memory.Manager
	similarityThreshold float64
}

func NewConsolidator(mgr *memory.Manager) *Consolidator {
	return &Consolidator{
		mgr:                 mgr,
		similarityThreshold: 0.85,
	}
}

// SetSimilarityThreshold adjusts how aggressively memories are grouped.
// Higher values (closer to 1.0) mean only very similar memories are grouped.
func (c *Consolidator) SetSimilarityThreshold(t float64) {
	if t > 0 && t <= 1.0 {
		c.similarityThreshold = t
	}
}

// Run executes a heartbeat consolidation cycle for the given scope.
// It finds clusters of similar memories and generates insights.
func (c *Consolidator) Run(ctx context.Context, filter *memory.SearchFilter) (*ConsolidationResult, error) {
	start := time.Now()

	memories, total, err := c.mgr.List(ctx, filter, 200, 0)
	if err != nil {
		return nil, fmt.Errorf("heartbeat list: %w", err)
	}

	result := &ConsolidationResult{
		ProcessedCount: total,
	}

	if len(memories) < 2 {
		result.Duration = time.Since(start)
		return result, nil
	}

	// Find clusters of related memories using pairwise search
	clusters := c.findClusters(ctx, memories)

	for _, cluster := range clusters {
		if len(cluster) < 2 {
			continue
		}

		insight := c.synthesizeCluster(cluster)
		if insight != nil {
			result.Insights = append(result.Insights, *insight)
			result.MergedCount += len(cluster)
		}
	}

	// Detect contradictions between memories of the same type
	contradictions := c.detectContradictions(ctx, memories)
	result.Insights = append(result.Insights, contradictions...)

	result.Duration = time.Since(start)
	return result, nil
}

// findClusters groups memories by searching for similar content.
func (c *Consolidator) findClusters(ctx context.Context, memories []*memory.Memory) [][]*memory.Memory {
	visited := make(map[string]bool)
	var clusters [][]*memory.Memory

	for _, mem := range memories {
		if visited[mem.ID] {
			continue
		}

		results, err := c.mgr.Search(ctx, mem.Content, 10, nil)
		if err != nil {
			continue
		}

		var cluster []*memory.Memory
		cluster = append(cluster, mem)
		visited[mem.ID] = true

		for _, r := range results {
			if r.Memory.ID == mem.ID {
				continue
			}
			if visited[r.Memory.ID] {
				continue
			}
			if r.Score >= c.similarityThreshold {
				m := r.Memory
				cluster = append(cluster, &m)
				visited[r.Memory.ID] = true
			}
		}

		if len(cluster) >= 2 {
			clusters = append(clusters, cluster)
		}
	}

	return clusters
}

// synthesizeCluster creates a summary insight from a group of related memories.
func (c *Consolidator) synthesizeCluster(cluster []*memory.Memory) *Insight {
	if len(cluster) == 0 {
		return nil
	}

	// Sort by creation time
	sort.Slice(cluster, func(i, j int) bool {
		return cluster[i].CreatedAt.Before(cluster[j].CreatedAt)
	})

	var ids []string
	var contents []string
	for _, m := range cluster {
		ids = append(ids, m.ID)
		contents = append(contents, m.Content)
	}

	// Build a summary: take the most recent as the primary, note the count
	summary := fmt.Sprintf(
		"Consolidated from %d related memories. Latest: %s",
		len(cluster),
		cluster[len(cluster)-1].Content,
	)

	return &Insight{
		Content:     summary,
		SourceIDs:   ids,
		InsightType: "summary",
		Confidence:  0.8,
		CreatedAt:   time.Now().UTC(),
	}
}

// detectContradictions finds memories with high similarity but potentially
// conflicting content (e.g., corrections that supersede older facts).
func (c *Consolidator) detectContradictions(ctx context.Context, memories []*memory.Memory) []Insight {
	var insights []Insight

	corrections := filterByType(memories, memory.TypeCorrection)
	facts := filterByType(memories, memory.TypeFact)

	for _, correction := range corrections {
		for _, fact := range facts {
			if correction.CreatedAt.Before(fact.CreatedAt) {
				continue
			}

			results, err := c.mgr.Search(ctx, correction.Content, 3, nil)
			if err != nil {
				continue
			}

			for _, r := range results {
				if r.Memory.ID == fact.ID && r.Score > 0.7 {
					insights = append(insights, Insight{
						Content: fmt.Sprintf(
							"Potential contradiction: fact %q may be superseded by correction %q",
							truncate(fact.Content, 80),
							truncate(correction.Content, 80),
						),
						SourceIDs:   []string{fact.ID, correction.ID},
						InsightType: "contradiction",
						Confidence:  r.Score,
						CreatedAt:   time.Now().UTC(),
					})
				}
			}
		}
	}

	return insights
}

func filterByType(memories []*memory.Memory, t memory.MemoryType) []*memory.Memory {
	var result []*memory.Memory
	for _, m := range memories {
		if m.MemoryType == t {
			result = append(result, m)
		}
	}
	return result
}

func truncate(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
