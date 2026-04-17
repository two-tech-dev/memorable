# Changelog

All notable changes to Memorable will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- MCP server with stdio transport
- Seven tools: `memorable_add`, `memorable_search`, `memorable_get`, `memorable_update`, `memorable_delete`, `memorable_list`, `memorable_stats`
- Five memory types: fact, conversation, decision, code_pattern, correction
- Four-dimensional scoping: user_id, agent_id, app_id, run_id
- pgvector storage backend with auto-migration
- OpenAI embedding provider (text-embedding-3-small, text-embedding-3-large, ada-002)
- SHA-256 content deduplication within scope
- YAML configuration with environment variable overrides
- Dockerfile with multi-stage build
- Docker Compose stack (PostgreSQL + Memorable)
- GitHub Actions CI pipeline
