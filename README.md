<p align="center">
  <img src="docs/assets/logo.svg" alt="Memorable" width="120" />
</p>

<h1 align="center">Memorable</h1>

<p align="center">
  <strong>Long-term memory engine for AI agents.</strong><br>
  Give your coding assistant a brain that persists across sessions.
</p>

<p align="center">
  <a href="#quick-start">Quick Start</a> •
  <a href="#auto-install">Auto Install</a> •
  <a href="#architecture">Architecture</a> •
  <a href="#mcp-tools">MCP Tools</a> •
  <a href="#configuration">Configuration</a> •
  <a href="CONTRIBUTING.md">Contributing</a>
</p>

<p align="center">
  <a href="https://github.com/two-tech-dev/memorable/actions/workflows/ci.yml"><img src="https://github.com/two-tech-dev/memorable/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://goreportcard.com/report/github.com/two-tech-dev/memorable"><img src="https://goreportcard.com/badge/github.com/two-tech-dev/memorable" alt="Go Report Card"></a>
  <a href="https://pkg.go.dev/github.com/two-tech-dev/memorable"><img src="https://pkg.go.dev/badge/github.com/two-tech-dev/memorable.svg" alt="Go Reference"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License: MIT"></a>
  <a href="https://github.com/two-tech-dev/memorable/releases"><img src="https://img.shields.io/github/v/release/two-tech-dev/memorable?include_prereleases" alt="Release"></a>
  <img src="https://img.shields.io/badge/go-1.21+-00ADD8?logo=go&logoColor=white" alt="Go 1.21+">
  <img src="https://img.shields.io/badge/MCP-compatible-8A2BE2?logo=data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjQiIGhlaWdodD0iMjQiIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48Y2lyY2xlIGN4PSIxMiIgY3k9IjEyIiByPSIxMCIgc3Ryb2tlPSJ3aGl0ZSIgc3Ryb2tlLXdpZHRoPSIyIi8+PC9zdmc+" alt="MCP Compatible">
  <a href="https://github.com/two-tech-dev/memorable/stargazers"><img src="https://img.shields.io/github/stars/two-tech-dev/memorable?style=social" alt="Stars"></a>
</p>

---

## What is Memorable?

Memorable is an [MCP](https://modelcontextprotocol.io/) server that gives AI coding agents—Cursor, Claude Code, GitHub Copilot, Windsurf—persistent long-term memory backed by semantic vector search.

Without Memorable, every conversation starts from zero. With it, your agent recalls past decisions, learned patterns, corrections, and project context across sessions.

### Why another memory tool?

| Feature                     | Memorable                                                  | Plain file notes  | Other memory servers  |
| --------------------------- | ---------------------------------------------------------- | ----------------- | --------------------- |
| Semantic search             | **Yes** (pgvector cosine similarity)                       | No (keyword grep) | Varies                |
| Dedup                       | **SHA-256 content hashing**                                | Manual            | Rare                  |
| Multi-dimensional scoping   | **user × agent × app × run**                               | None              | Usually single-tenant |
| Five memory types           | **fact, conversation, decision, code_pattern, correction** | Unstructured      | Usually one           |
| MCP-native                  | **12 typed tools, auto-schema**                            | N/A               | Some                  |
| Pluggable embeddings        | **OpenAI, Gemini, Ollama, OpenRouter, custom**             | N/A               | Rarely                |
| Knowledge graph             | **Entity-relation extraction + traversal**                 | N/A               | No                    |
| Heartbeat / self-reflection | **Periodic consolidation + contradiction detection**       | N/A               | No                    |
| Hybrid retrieval            | **Vector + recency + frequency scoring**                   | N/A               | Vector only           |

---

## Quick Start

### Prerequisites

- **Go 1.21+**
- **PostgreSQL 15+** with the [`pgvector`](https://github.com/pgvector/pgvector) extension
- **Embedding API key** — one of: [OpenAI](https://platform.openai.com/), [Google Gemini](https://ai.google.dev/), [OpenRouter](https://openrouter.ai/), or [Ollama](https://ollama.com/) (local, no key needed). Any OpenAI-compatible API also works via the `custom` provider.

### 1. Install

**From source:**

```bash
git clone https://github.com/two-tech-dev/memorable.git
cd memorable
make build
```

The binary is written to `bin/memorable`.

**With Go:**

```bash
go install github.com/two-tech-dev/memorable/cmd/memorable@latest
```

### 2. Start PostgreSQL with pgvector

Using Docker:

```bash
docker run -d --name memorable-pg \
  -e POSTGRES_DB=memorable \
  -e POSTGRES_HOST_AUTH_METHOD=trust \
  -p 5432:5432 \
  pgvector/pgvector:pg16
```

Or install pgvector on an existing PostgreSQL instance.

### 3. Configure

Copy the example config and set your API key:

```bash
cp config.example.yaml memorable.yaml
```

Edit `memorable.yaml` or use environment variables:

```bash
# Pick your embedding provider:
export OPENAI_API_KEY=sk-...       # OpenAI
export GEMINI_API_KEY=AI...         # Google Gemini
# Or use Ollama locally (no key needed)

export MEMORABLE_DSN=postgres://localhost:5432/memorable?sslmode=disable
```

### 4. Run

```bash
bin/memorable
# or
make run
```

The server starts on stdio and auto-migrates the database schema on first connect.

---

## Auto Install

Memorable includes install scripts that automatically configure MCP for your AI agents.

### One-liner (all agents)

**macOS / Linux:**

```bash
./scripts/install.sh
```

**Windows (PowerShell):**

```powershell
.\scripts\install.ps1
```

This detects and configures **Cursor**, **Claude Desktop / Claude Code**, **VS Code (GitHub Copilot)**, and **Windsurf** in one command.

### Target a specific agent

```bash
# Linux/macOS
./scripts/install.sh --agent cursor
./scripts/install.sh --agent claude --config ~/memorable.yaml

# Windows
.\scripts\install.ps1 -Agent cursor
.\scripts\install.ps1 -Agent claude -Config "C:\Users\me\memorable.yaml"
```

**Supported agents:** `cursor`, `claude`, `copilot`, `windsurf`, `all`

### What the scripts do

| Agent                 | Config file                                                                                                                         | Servers key  |
| --------------------- | ----------------------------------------------------------------------------------------------------------------------------------- | ------------ |
| **Cursor**            | `~/.cursor/mcp.json`                                                                                                                | `mcpServers` |
| **Claude**            | `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) / `%APPDATA%\Claude\claude_desktop_config.json` (Windows) | `mcpServers` |
| **VS Code (Copilot)** | `.vscode/mcp.json` (workspace)                                                                                                      | `servers`    |
| **Windsurf**          | `~/.codeium/windsurf/mcp_config.json`                                                                                               | `mcpServers` |

The scripts safely merge into existing config files — your other MCP servers are preserved.

---

## Manual Configuration

<details>
<summary>Click to expand manual agent configuration</summary>

### Cursor

Add to `.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "memorable": {
      "command": "memorable",
      "args": ["-config", "/path/to/memorable.yaml"]
    }
  }
}
```

### Claude Code

Add to your MCP config:

```json
{
  "mcpServers": {
    "memorable": {
      "command": "memorable",
      "args": []
    }
  }
}
```

### VS Code (GitHub Copilot)

Add to `.vscode/mcp.json`:

```json
{
  "servers": {
    "memorable": {
      "type": "stdio",
      "command": "memorable",
      "args": ["-config", "memorable.yaml"]
    }
  }
}
```

### Windsurf

Add to `~/.codeium/windsurf/mcp_config.json`:

```json
{
  "mcpServers": {
    "memorable": {
      "command": "memorable"
    }
  }
}
```

</details>

---

## Architecture

```
┌───────────────────────────────────────────────────────────────┐
│               AI Agent (Cursor, Claude, Copilot, Windsurf)    │
│                          ↕ MCP (stdio)                        │
├───────────────────────────────────────────────────────────────┤
│                        MCP Server (12 tools)                  │
│  ┌──────┐ ┌────────┐ ┌──────────┐ ┌───────┐ ┌────────────┐  │
│  │ CRUD │ │ search │ │heartbeat │ │ graph │ │   stats    │  │
│  └──┬───┘ └───┬────┘ └────┬─────┘ └───┬───┘ └─────┬──────┘  │
│     └─────────┴────────────┴───────────┴───────────┘         │
│                      Memory Manager                           │
│            (dedup · embed · CRUD · scope)                     │
├────────┬──────────────────┬──────────────────┬───────────────┤
│  L1    │  L2 Vector DB    │  Knowledge       │  L3 Soul/     │
│  Cache │  (pgvector)      │  Graph           │  Profile      │
│  (LRU) │                  │  (entity-rel)    │  (traits)     │
├────────┴──────────────────┴──────────────────┴───────────────┤
│  Embedding Provider                   Hybrid Retrieval        │
│  (OpenAI / Gemini / Ollama / Custom)  (vector+recency+freq)  │
└───────────────────────────────────────────────────────────────┘
                           ↕
                    ┌──────────────┐
                    │  PostgreSQL  │
                    │  + pgvector  │
                    └──────────────┘
```

### Project Layout

```
memorable/
├── cmd/memorable/       # Application entry point
│   └── main.go
├── internal/
│   ├── cache/           # L1 LRU cache (generic, thread-safe)
│   ├── config/          # YAML config loader + env overrides
│   ├── embedding/       # Embedding providers (OpenAI, Gemini, Ollama, Custom)
│   ├── graph/           # Knowledge graph (entities, relations, triple extraction)
│   ├── heartbeat/       # Self-reflection & memory consolidation
│   ├── mcp/             # MCP server, tool registration, typed handlers
│   ├── memory/          # Core types, Manager (CRUD + dedup), VectorStore interface
│   ├── profile/         # L3 Soul/Profile (user trait accumulation)
│   ├── retrieval/       # Hybrid scoring (vector + recency + frequency)
│   └── store/           # Storage implementations (pgvector)
├── scripts/             # Auto-install scripts (bash + PowerShell)
├── docs/                # Documentation assets
├── config.example.yaml  # Example configuration
├── Makefile             # Build, test, lint targets
└── .github/workflows/   # CI pipeline
```

### Memory Types

| Type           | Description                                 | Example                                                              |
| -------------- | ------------------------------------------- | -------------------------------------------------------------------- |
| `fact`         | Factual knowledge the agent should remember | "This project uses Go 1.21 with modules."                            |
| `conversation` | Key points from conversations               | "User prefers tabs over spaces."                                     |
| `decision`     | Architectural or design decisions           | "We chose pgvector over Pinecone for self-hosting."                  |
| `code_pattern` | Reusable patterns and idioms                | "Error wrapping: always use fmt.Errorf with %w."                     |
| `correction`   | Mistakes and their fixes                    | "Don't import from internal/store directly; use memory.VectorStore." |

### Scoping

Every memory is scoped along four dimensions, all optional:

| Dimension  | Purpose             | Example                    |
| ---------- | ------------------- | -------------------------- |
| `user_id`  | Per-user isolation  | `"alice"`                  |
| `agent_id` | Per-agent context   | `"cursor"`, `"claude"`     |
| `app_id`   | Per-project context | `"memorable"`, `"my-saas"` |
| `run_id`   | Per-session context | `"session-2024-01-15"`     |

Memories with no scope are global. Scopes are combined as filters during search and retrieval.

---

## MCP Tools

Memorable exposes **12 tools** through the MCP protocol:

### Memory CRUD

#### `memorable_add`

Store a new memory. Automatically deduplicates by content hash within the same scope.

```json
{
  "content": "The project uses pgvector for vector similarity search",
  "type": "fact",
  "app_id": "my-project"
}
```

#### `memorable_search`

Search memories by semantic similarity. The query is embedded and compared against stored vectors using cosine distance.

```json
{
  "query": "which database do we use?",
  "limit": 5,
  "app_id": "my-project"
}
```

#### `memorable_get`

Retrieve a specific memory by UUID.

```json
{ "id": "550e8400-e29b-41d4-a716-446655440000" }
```

#### `memorable_update`

Update content (triggers re-embedding) and/or merge metadata.

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "content": "Updated: we migrated from pgvector to Qdrant",
  "metadata": { "reviewed": true }
}
```

#### `memorable_delete`

Delete a memory by UUID.

```json
{ "id": "550e8400-e29b-41d4-a716-446655440000" }
```

#### `memorable_list`

List memories with filters and pagination.

```json
{
  "type": "decision",
  "app_id": "my-project",
  "limit": 20,
  "offset": 0
}
```

#### `memorable_stats`

Get aggregate statistics: total count, breakdown by type, time range.

```json
{ "user_id": "alice" }
```

### Heartbeat / Self-Reflection

#### `memorable_heartbeat`

Run a consolidation cycle. Analyzes stored memories, finds clusters of similar content, and generates insights (summaries, contradictions, patterns).

```json
{ "user_id": "alice", "app_id": "my-project" }
```

Returns:

- **Summaries** — consolidated clusters of related memories
- **Contradictions** — detects when corrections supersede older facts
- **Patterns** — recurring themes across memories

### Knowledge Graph

#### `memorable_graph_add`

Extract entities and relations from text and add them to the knowledge graph.

```json
{
  "content": "The project uses PostgreSQL. React depends on Node.",
  "memory_id": "550e8400-..."
}
```

#### `memorable_graph_search`

Search for entities in the knowledge graph by name.

```json
{ "query": "postgres", "limit": 5 }
```

#### `memorable_graph_neighbors`

Get entities and relations connected to a given entity, with configurable traversal depth.

```json
{ "entity_id": "ent_abc123", "depth": 2 }
```

#### `memorable_graph_stats`

Get knowledge graph statistics: entity and relation counts.

```json
{}
```

---

## Configuration

Memorable looks for configuration in this order:

1. `--config` flag (explicit path)
2. `./memorable.yaml` (current directory)
3. `~/.memorable/config.yaml` (home directory)
4. Built-in defaults

### Full Reference

```yaml
# Server transport
server:
  transport: stdio # stdio | http (future)

# Storage backend
storage:
  backend: pgvector # pgvector | sqlite (future)
  pgvector:
    dsn: "postgres://localhost:5432/memorable?sslmode=disable"
    table_name: memories
    vector_dimensions: 1536 # Must match embedding model

# Embedding provider (openai | gemini | ollama | custom)
embedding:
  provider: openai

  openai:
    api_key: ${OPENAI_API_KEY}
    model: text-embedding-3-small
    # base_url: https://custom-api.example.com/v1  # For OpenAI-compatible APIs

  gemini:
    api_key: ${GEMINI_API_KEY}
    model: text-embedding-004

  ollama:
    base_url: http://localhost:11434
    model: nomic-embed-text
    dims: 768

  # Generic OpenAI-compatible provider — works with any /v1/embeddings API
  custom:
    base_url: https://openrouter.ai/api/v1 # Required
    api_key: ${CUSTOM_API_KEY}
    model: openai/text-embedding-3-small # Required
    dims: 1536 # Required
    headers: # Optional
      X-Title: Memorable
```

### Environment Variables

| Variable         | Description                   | Overrides                  |
| ---------------- | ----------------------------- | -------------------------- |
| `OPENAI_API_KEY` | OpenAI API key for embeddings | `embedding.openai.api_key` |
| `GEMINI_API_KEY` | Google Gemini API key         | `embedding.gemini.api_key` |
| `CUSTOM_API_KEY` | API key for custom provider   | `embedding.custom.api_key` |
| `MEMORABLE_DSN`  | PostgreSQL connection string  | `storage.pgvector.dsn`     |

### Embedding Providers

#### OpenAI

| Model                    | Dimensions | Notes                                 |
| ------------------------ | ---------- | ------------------------------------- |
| `text-embedding-3-small` | 1536       | Default. Best cost/performance ratio. |
| `text-embedding-3-large` | 3072       | Higher accuracy, 2× storage.          |
| `text-embedding-ada-002` | 1536       | Legacy.                               |

Set `base_url` to use OpenAI-compatible APIs like **Voyage AI**, **Together AI**, or **Azure OpenAI**.

#### Google Gemini

| Model                | Dimensions | Notes        |
| -------------------- | ---------- | ------------ |
| `text-embedding-004` | 768        | Recommended. |
| `embedding-001`      | 768        | Legacy.      |

#### Ollama (Local)

| Model               | Dimensions | Notes                       |
| ------------------- | ---------- | --------------------------- |
| `nomic-embed-text`  | 768        | Good general-purpose model. |
| `mxbai-embed-large` | 1024       | Higher accuracy.            |
| `all-minilm`        | 384        | Smallest, fastest.          |

Run `ollama pull nomic-embed-text` to download the model, then set `provider: ollama` in config. No API key required.

#### Custom Provider (OpenRouter, Together, Voyage, Mistral, Cohere, etc.)

Set `provider: custom` and point `base_url` at any API that implements the OpenAI `/v1/embeddings` endpoint.

**OpenRouter:**

```yaml
embedding:
  provider: custom
  custom:
    base_url: https://openrouter.ai/api/v1
    api_key: ${CUSTOM_API_KEY}
    model: openai/text-embedding-3-small
    dims: 1536
    headers:
      X-Title: Memorable
      HTTP-Referer: https://github.com/two-tech-dev/memorable
```

**Together AI:**

```yaml
embedding:
  provider: custom
  custom:
    base_url: https://api.together.xyz/v1
    api_key: ${CUSTOM_API_KEY}
    model: togethercomputer/m2-bert-80M-8k-retrieval
    dims: 768
```

**Voyage AI:**

```yaml
embedding:
  provider: custom
  custom:
    base_url: https://api.voyageai.com/v1
    api_key: ${CUSTOM_API_KEY}
    model: voyage-3
    dims: 1024
```

**Mistral:**

```yaml
embedding:
  provider: custom
  custom:
    base_url: https://api.mistral.ai/v1
    api_key: ${CUSTOM_API_KEY}
    model: mistral-embed
    dims: 1024
```

**Azure OpenAI:**

```yaml
embedding:
  provider: custom
  custom:
    base_url: https://YOUR-RESOURCE.openai.azure.com/openai/deployments/YOUR-DEPLOYMENT
    api_key: ${CUSTOM_API_KEY}
    model: text-embedding-3-small
    dims: 1536
    headers:
      api-key: ${CUSTOM_API_KEY}
```

> **Note:** Set `dims` to match the actual output dimensions of your chosen model, and ensure `storage.pgvector.vector_dimensions` matches.

---

## Development

```bash
# Build
make build

# Run tests
make test

# Run linter
make lint

# Clean build artifacts
make clean

```

See [CONTRIBUTING.md](CONTRIBUTING.md) for the full development guide.

---

## License

[MIT](LICENSE) © 2026 [Two Tech Dev](https://github.com/two-tech-dev)
