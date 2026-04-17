package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/two-tech-dev/memorable/internal/graph"
	"github.com/two-tech-dev/memorable/internal/heartbeat"
	"github.com/two-tech-dev/memorable/internal/memory"
)

// --- Input types (auto-generate JSON Schema via jsonschema struct tags) ---

type AddInput struct {
	Content  string         `json:"content" jsonschema:"The memory content to store"`
	Type     string         `json:"type" jsonschema:"Type of memory (fact conversation decision code_pattern correction)"`
	UserID   string         `json:"user_id,omitempty" jsonschema:"User scope identifier"`
	AgentID  string         `json:"agent_id,omitempty" jsonschema:"Agent scope identifier"`
	AppID    string         `json:"app_id,omitempty" jsonschema:"Application scope identifier"`
	RunID    string         `json:"run_id,omitempty" jsonschema:"Session/run scope identifier"`
	Metadata map[string]any `json:"metadata,omitempty" jsonschema:"Additional metadata as JSON object"`
}

type SearchInput struct {
	Query   string `json:"query" jsonschema:"Natural language search query"`
	Limit   int    `json:"limit,omitempty" jsonschema:"Maximum results to return (default 10)"`
	Type    string `json:"type,omitempty" jsonschema:"Filter by memory type"`
	UserID  string `json:"user_id,omitempty" jsonschema:"Filter by user scope"`
	AgentID string `json:"agent_id,omitempty" jsonschema:"Filter by agent scope"`
	AppID   string `json:"app_id,omitempty" jsonschema:"Filter by application scope"`
	RunID   string `json:"run_id,omitempty" jsonschema:"Filter by session/run scope"`
}

type GetInput struct {
	ID string `json:"id" jsonschema:"The memory UUID"`
}

type UpdateInput struct {
	ID       string         `json:"id" jsonschema:"The memory UUID to update"`
	Content  *string        `json:"content,omitempty" jsonschema:"New content (triggers re-embedding)"`
	Metadata map[string]any `json:"metadata,omitempty" jsonschema:"Metadata fields to merge"`
}

type DeleteInput struct {
	ID string `json:"id" jsonschema:"The memory UUID to delete"`
}

type ListInput struct {
	Type    string `json:"type,omitempty" jsonschema:"Filter by memory type"`
	UserID  string `json:"user_id,omitempty" jsonschema:"Filter by user scope"`
	AgentID string `json:"agent_id,omitempty" jsonschema:"Filter by agent scope"`
	AppID   string `json:"app_id,omitempty" jsonschema:"Filter by application scope"`
	RunID   string `json:"run_id,omitempty" jsonschema:"Filter by session/run scope"`
	Limit   int    `json:"limit,omitempty" jsonschema:"Max results (default 20)"`
	Offset  int    `json:"offset,omitempty" jsonschema:"Offset for pagination (default 0)"`
}

type StatsInput struct {
	UserID  string `json:"user_id,omitempty" jsonschema:"Filter stats by user scope"`
	AgentID string `json:"agent_id,omitempty" jsonschema:"Filter stats by agent scope"`
	AppID   string `json:"app_id,omitempty" jsonschema:"Filter stats by application scope"`
}

// --- Handlers ---

func addHandler(mgr *memory.Manager) mcp.ToolHandlerFor[AddInput, any] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input AddInput) (*mcp.CallToolResult, any, error) {
		filter := &memory.SearchFilter{
			UserID:  input.UserID,
			AgentID: input.AgentID,
			AppID:   input.AppID,
			RunID:   input.RunID,
		}
		mem, err := mgr.Add(ctx, input.Content, memory.MemoryType(input.Type), filter, input.Metadata)
		if err != nil {
			return nil, nil, err
		}
		return nil, mem, nil
	}
}

func searchHandler(mgr *memory.Manager) mcp.ToolHandlerFor[SearchInput, any] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input SearchInput) (*mcp.CallToolResult, any, error) {
		filter := &memory.SearchFilter{
			UserID:     input.UserID,
			AgentID:    input.AgentID,
			AppID:      input.AppID,
			RunID:      input.RunID,
			MemoryType: memory.MemoryType(input.Type),
		}
		results, err := mgr.Search(ctx, input.Query, input.Limit, filter)
		if err != nil {
			return nil, nil, err
		}
		return nil, results, nil
	}
}

func getHandler(mgr *memory.Manager) mcp.ToolHandlerFor[GetInput, any] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input GetInput) (*mcp.CallToolResult, any, error) {
		mem, err := mgr.Get(ctx, input.ID)
		if err != nil {
			return nil, nil, err
		}
		return nil, mem, nil
	}
}

func updateHandler(mgr *memory.Manager) mcp.ToolHandlerFor[UpdateInput, any] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input UpdateInput) (*mcp.CallToolResult, any, error) {
		mem, err := mgr.Update(ctx, input.ID, input.Content, input.Metadata)
		if err != nil {
			return nil, nil, err
		}
		return nil, mem, nil
	}
}

func deleteHandler(mgr *memory.Manager) mcp.ToolHandlerFor[DeleteInput, any] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input DeleteInput) (*mcp.CallToolResult, any, error) {
		if err := mgr.Delete(ctx, input.ID); err != nil {
			return nil, nil, err
		}
		return nil, map[string]string{"status": "deleted", "id": input.ID}, nil
	}
}

func listHandler(mgr *memory.Manager) mcp.ToolHandlerFor[ListInput, any] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input ListInput) (*mcp.CallToolResult, any, error) {
		filter := &memory.SearchFilter{
			UserID:     input.UserID,
			AgentID:    input.AgentID,
			AppID:      input.AppID,
			RunID:      input.RunID,
			MemoryType: memory.MemoryType(input.Type),
		}
		memories, total, err := mgr.List(ctx, filter, input.Limit, input.Offset)
		if err != nil {
			return nil, nil, err
		}
		return nil, map[string]any{"memories": memories, "total": total}, nil
	}
}

func statsHandler(mgr *memory.Manager) mcp.ToolHandlerFor[StatsInput, any] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input StatsInput) (*mcp.CallToolResult, any, error) {
		filter := &memory.SearchFilter{
			UserID:  input.UserID,
			AgentID: input.AgentID,
			AppID:   input.AppID,
		}
		stats, err := mgr.Stats(ctx, filter)
		if err != nil {
			return nil, nil, err
		}
		return nil, stats, nil
	}
}

// --- Heartbeat handlers ---

type HeartbeatInput struct {
	UserID  string `json:"user_id,omitempty" jsonschema:"Scope consolidation to user"`
	AgentID string `json:"agent_id,omitempty" jsonschema:"Scope consolidation to agent"`
	AppID   string `json:"app_id,omitempty" jsonschema:"Scope consolidation to application"`
	RunID   string `json:"run_id,omitempty" jsonschema:"Scope consolidation to session/run"`
}

func heartbeatHandler(c *heartbeat.Consolidator) mcp.ToolHandlerFor[HeartbeatInput, any] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input HeartbeatInput) (*mcp.CallToolResult, any, error) {
		filter := &memory.SearchFilter{
			UserID:  input.UserID,
			AgentID: input.AgentID,
			AppID:   input.AppID,
			RunID:   input.RunID,
		}
		result, err := c.Run(ctx, filter)
		if err != nil {
			return nil, nil, err
		}
		return nil, result, nil
	}
}

// --- Knowledge Graph handlers ---

type GraphAddInput struct {
	Content  string `json:"content" jsonschema:"Text to extract entities and relations from"`
	MemoryID string `json:"memory_id,omitempty" jsonschema:"Associated memory ID for provenance tracking"`
}

type GraphSearchInput struct {
	Query string `json:"query" jsonschema:"Entity name or partial name to search for"`
	Limit int    `json:"limit,omitempty" jsonschema:"Max results (default 10)"`
}

type GraphNeighborsInput struct {
	EntityID string `json:"entity_id" jsonschema:"Entity ID to get neighbors for"`
	Depth    int    `json:"depth,omitempty" jsonschema:"Traversal depth (default 1, max 3)"`
}

func graphAddHandler(g *graph.InMemoryGraph) mcp.ToolHandlerFor[GraphAddInput, any] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input GraphAddInput) (*mcp.CallToolResult, any, error) {
		triples := graph.ExtractTriples(input.Content, input.MemoryID)
		entities, relations := graph.TriplesToGraph(triples)

		for _, e := range entities {
			if err := g.UpsertEntity(ctx, e); err != nil {
				return nil, nil, err
			}
		}
		for _, r := range relations {
			if err := g.UpsertRelation(ctx, r); err != nil {
				return nil, nil, err
			}
		}

		return nil, map[string]any{
			"triples_extracted": len(triples),
			"entities_added":    len(entities),
			"relations_added":   len(relations),
			"triples":           triples,
		}, nil
	}
}

func graphSearchHandler(g *graph.InMemoryGraph) mcp.ToolHandlerFor[GraphSearchInput, any] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input GraphSearchInput) (*mcp.CallToolResult, any, error) {
		limit := input.Limit
		if limit <= 0 {
			limit = 10
		}
		entities, err := g.FindEntities(ctx, input.Query, limit)
		if err != nil {
			return nil, nil, err
		}
		return nil, entities, nil
	}
}

func graphNeighborsHandler(g *graph.InMemoryGraph) mcp.ToolHandlerFor[GraphNeighborsInput, any] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input GraphNeighborsInput) (*mcp.CallToolResult, any, error) {
		depth := input.Depth
		if depth <= 0 {
			depth = 1
		}
		if depth > 3 {
			depth = 3
		}
		entities, relations, err := g.GetNeighbors(ctx, input.EntityID, depth)
		if err != nil {
			return nil, nil, err
		}
		return nil, map[string]any{
			"entities":  entities,
			"relations": relations,
		}, nil
	}
}

type GraphStatsInput struct{}

func graphStatsHandler(g *graph.InMemoryGraph) mcp.ToolHandlerFor[GraphStatsInput, any] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input GraphStatsInput) (*mcp.CallToolResult, any, error) {
		ec, rc := g.Stats()
		return nil, map[string]any{
			"entity_count":   ec,
			"relation_count": rc,
		}, nil
	}
}
