# Contributing to Memorable

Thanks for your interest in contributing to Memorable! This document covers everything you need to get started.

## Development Setup

### Prerequisites

- Go 1.21 or later
- PostgreSQL 15+ with [pgvector](https://github.com/pgvector/pgvector) extension
- Docker and Docker Compose (optional, for running PostgreSQL)
- [golangci-lint](https://golangci-lint.run/) (for linting)

### Clone and Build

```bash
git clone https://github.com/two-tech-dev/memorable.git
cd memorable
make build
```

### Start PostgreSQL

The easiest way to get a development database:

```bash
docker compose up -d postgres
```

This starts PostgreSQL 16 with pgvector on port `5432`, database `memorable`, no password.

### Configure

```bash
cp config.example.yaml memorable.yaml
# Edit memorable.yaml with your OpenAI key, or:
export OPENAI_API_KEY=sk-...
```

### Run

```bash
make run
```

## Project Structure

```
cmd/memorable/       Entry point
internal/
  config/            YAML config + env overrides
  embedding/         Embedding provider interface + OpenAI
  mcp/               MCP server + tool handlers
  memory/            Core types, Manager, VectorStore interface
  store/             Storage implementations (pgvector)
```

Code in `internal/` is private to this module. The public API is the MCP tool interface.

## Making Changes

### Branch Naming

- `feat/short-description` — New features
- `fix/short-description` — Bug fixes
- `docs/short-description` — Documentation changes
- `refactor/short-description` — Code changes that don't fix bugs or add features

### Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Use `golangci-lint run` before committing
- Wrap errors with context: `fmt.Errorf("operation: %w", err)`
- Keep functions focused and short
- Interface definitions live close to their consumers, not their implementations

### Commit Messages

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add SQLite storage backend
fix: handle nil metadata in search results
docs: update MCP tool examples
refactor: extract scope filter builder
```

### Testing

```bash
# Unit tests
make test

# Integration tests (requires running PostgreSQL)
make test-integration
```

Write tests for new features and bug fixes. Place test files next to the code they test (`foo_test.go` alongside `foo.go`).

## Pull Request Process

1. Fork the repo and create your branch from `main`
2. Make your changes with tests
3. Run `make lint` and `make test`
4. Push and open a Pull Request
5. Fill out the PR template describing your changes
6. Wait for CI to pass and a maintainer review

## Reporting Issues

- Use GitHub Issues for bugs and feature requests
- Include Go version, OS, and PostgreSQL version for bug reports
- Include steps to reproduce the problem

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).
