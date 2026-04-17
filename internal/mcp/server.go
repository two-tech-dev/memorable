package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/two-tech-dev/memorable/internal/graph"
	"github.com/two-tech-dev/memorable/internal/heartbeat"
	"github.com/two-tech-dev/memorable/internal/memory"
)

// Deps bundles all dependencies for MCP handlers.
type Deps struct {
	Manager      *memory.Manager
	Consolidator *heartbeat.Consolidator
	Graph        *graph.InMemoryGraph
}

func NewServer(deps *Deps) *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{
		Name:    "memorable",
		Version: "0.2.0",
	}, nil)

	registerTools(s, deps)
	return s
}

func registerTools(s *mcp.Server, deps *Deps) {
	mgr := deps.Manager

	mcp.AddTool(s, &mcp.Tool{
		Name:        "memorable_add",
		Description: "Add a new memory. Returns the created memory or existing one if duplicate content is found in the same scope.",
	}, addHandler(mgr))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "memorable_search",
		Description: "Search memories by semantic similarity. Returns the most relevant memories matching the query.",
	}, searchHandler(mgr))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "memorable_get",
		Description: "Get a specific memory by its ID.",
	}, getHandler(mgr))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "memorable_update",
		Description: "Update an existing memory's content and/or metadata. Re-embeds if content changes.",
	}, updateHandler(mgr))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "memorable_delete",
		Description: "Delete a memory by its ID.",
	}, deleteHandler(mgr))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "memorable_list",
		Description: "List memories with optional filters and pagination.",
	}, listHandler(mgr))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "memorable_stats",
		Description: "Get memory statistics: total count, breakdown by type, time range.",
	}, statsHandler(mgr))

	// Phase 7: Heartbeat/Self-Reflection
	mcp.AddTool(s, &mcp.Tool{
		Name:        "memorable_heartbeat",
		Description: "Run a heartbeat consolidation cycle. Analyzes stored memories, finds clusters of related content, generates insights (summaries, contradictions, patterns).",
	}, heartbeatHandler(deps.Consolidator))

	// Phase 8: Knowledge Graph
	mcp.AddTool(s, &mcp.Tool{
		Name:        "memorable_graph_add",
		Description: "Extract entities and relations from text and add them to the knowledge graph.",
	}, graphAddHandler(deps.Graph))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "memorable_graph_search",
		Description: "Search for entities in the knowledge graph by name.",
	}, graphSearchHandler(deps.Graph))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "memorable_graph_neighbors",
		Description: "Get entities and relations connected to a given entity, traversing up to the specified depth.",
	}, graphNeighborsHandler(deps.Graph))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "memorable_graph_stats",
		Description: "Get knowledge graph statistics: entity and relation counts.",
	}, graphStatsHandler(deps.Graph))
}

func RunStdio(ctx context.Context, s *mcp.Server) error {
	return s.Run(ctx, &mcp.StdioTransport{})
}
