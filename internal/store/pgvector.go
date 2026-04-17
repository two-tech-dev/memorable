package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pgvector "github.com/pgvector/pgvector-go"

	"github.com/two-tech-dev/memorable/internal/memory"
)

type PgvectorStore struct {
	pool  *pgxpool.Pool
	table string
	dims  int
}

func NewPgvector(ctx context.Context, dsn, table string, dims int) (*PgvectorStore, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("pgvector connect: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pgvector ping: %w", err)
	}

	s := &PgvectorStore{pool: pool, table: table, dims: dims}
	if err := s.migrate(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pgvector migrate: %w", err)
	}
	return s, nil
}

func (s *PgvectorStore) migrate(ctx context.Context) error {
	queries := []string{
		`CREATE EXTENSION IF NOT EXISTS vector`,
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			content TEXT NOT NULL,
			memory_type VARCHAR(50) NOT NULL,
			user_id VARCHAR(255),
			agent_id VARCHAR(255),
			app_id VARCHAR(255),
			run_id VARCHAR(255),
			metadata JSONB DEFAULT '{}',
			content_hash VARCHAR(64) NOT NULL,
			embedding vector(%d),
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)`, s.table, s.dims),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_embedding ON %s USING hnsw (embedding vector_cosine_ops)`, s.table, s.table),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_user ON %s (user_id)`, s.table, s.table),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_type ON %s (memory_type)`, s.table, s.table),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_hash ON %s (content_hash)`, s.table, s.table),
	}
	for _, q := range queries {
		if _, err := s.pool.Exec(ctx, q); err != nil {
			return fmt.Errorf("migrate query: %w", err)
		}
	}
	return nil
}

func (s *PgvectorStore) Insert(ctx context.Context, mem *memory.Memory) error {
	meta, err := json.Marshal(mem.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	query := fmt.Sprintf(`INSERT INTO %s (id, content, memory_type, user_id, agent_id, app_id, run_id, metadata, content_hash, embedding, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`, s.table)

	_, err = s.pool.Exec(ctx, query,
		mem.ID, mem.Content, string(mem.MemoryType),
		mem.UserID, mem.AgentID, mem.AppID, mem.RunID,
		meta, mem.ContentHash,
		pgvector.NewVector(mem.Embedding),
		mem.CreatedAt, mem.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert memory: %w", err)
	}
	return nil
}

func (s *PgvectorStore) Update(ctx context.Context, mem *memory.Memory) error {
	meta, err := json.Marshal(mem.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	query := fmt.Sprintf(`UPDATE %s SET content=$1, metadata=$2, content_hash=$3, embedding=$4, updated_at=$5 WHERE id=$6`, s.table)
	_, err = s.pool.Exec(ctx, query,
		mem.Content, meta, mem.ContentHash,
		pgvector.NewVector(mem.Embedding),
		time.Now().UTC(), mem.ID,
	)
	if err != nil {
		return fmt.Errorf("update memory: %w", err)
	}
	return nil
}

func (s *PgvectorStore) Delete(ctx context.Context, id string) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE id=$1`, s.table)
	_, err := s.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete memory: %w", err)
	}
	return nil
}

func (s *PgvectorStore) Get(ctx context.Context, id string) (*memory.Memory, error) {
	query := fmt.Sprintf(`SELECT id, content, memory_type, user_id, agent_id, app_id, run_id, metadata, content_hash, embedding, created_at, updated_at FROM %s WHERE id=$1`, s.table)
	return s.scanOne(ctx, query, id)
}

func (s *PgvectorStore) GetByHash(ctx context.Context, contentHash string, filter *memory.SearchFilter) (*memory.Memory, error) {
	where, args := buildWhere(filter, 2)
	query := fmt.Sprintf(`SELECT id, content, memory_type, user_id, agent_id, app_id, run_id, metadata, content_hash, embedding, created_at, updated_at FROM %s WHERE content_hash=$1%s LIMIT 1`, s.table, where)
	args = append([]any{contentHash}, args...)
	return s.scanOne(ctx, query, args...)
}

func (s *PgvectorStore) Search(ctx context.Context, vector []float32, k int, filter *memory.SearchFilter) ([]memory.SearchResult, error) {
	where, args := buildWhere(filter, 3)
	query := fmt.Sprintf(`SELECT id, content, memory_type, user_id, agent_id, app_id, run_id, metadata, content_hash, embedding, created_at, updated_at, 1 - (embedding <=> $1) AS score, embedding <=> $1 AS distance FROM %s WHERE embedding IS NOT NULL%s ORDER BY distance LIMIT $2`, s.table, where)
	args = append([]any{pgvector.NewVector(vector), k}, args...)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("search memories: %w", err)
	}
	defer rows.Close()

	var results []memory.SearchResult
	for rows.Next() {
		var r memory.SearchResult
		var meta []byte
		var emb pgvector.Vector
		err := rows.Scan(
			&r.Memory.ID, &r.Memory.Content, &r.Memory.MemoryType,
			&r.Memory.UserID, &r.Memory.AgentID, &r.Memory.AppID, &r.Memory.RunID,
			&meta, &r.Memory.ContentHash, &emb,
			&r.Memory.CreatedAt, &r.Memory.UpdatedAt,
			&r.Score, &r.Distance,
		)
		if err != nil {
			return nil, fmt.Errorf("scan search result: %w", err)
		}
		_ = json.Unmarshal(meta, &r.Memory.Metadata) // metadata is optional; ignore parse errors
		r.Memory.Embedding = emb.Slice()
		results = append(results, r)
	}
	return results, rows.Err()
}

func (s *PgvectorStore) List(ctx context.Context, filter *memory.SearchFilter, limit, offset int) ([]*memory.Memory, int, error) {
	where, args := buildWhere(filter, 3)

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE 1=1%s`, s.table, where)
	countArgs := make([]any, len(args))
	copy(countArgs, args)

	var total int
	if err := s.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count memories: %w", err)
	}

	query := fmt.Sprintf(`SELECT id, content, memory_type, user_id, agent_id, app_id, run_id, metadata, content_hash, embedding, created_at, updated_at FROM %s WHERE 1=1%s ORDER BY created_at DESC LIMIT $1 OFFSET $2`, s.table, where)
	args = append([]any{limit, offset}, args...)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list memories: %w", err)
	}
	defer rows.Close()

	var memories []*memory.Memory
	for rows.Next() {
		m, err := scanMemory(rows)
		if err != nil {
			return nil, 0, err
		}
		memories = append(memories, m)
	}
	return memories, total, rows.Err()
}

func (s *PgvectorStore) Stats(ctx context.Context, filter *memory.SearchFilter) (*memory.Stats, error) {
	where, args := buildWhere(filter, 1)

	stats := &memory.Stats{
		ByType: make(map[string]int),
		ByUser: make(map[string]int),
	}

	// Total count
	q := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE 1=1%s`, s.table, where)
	if err := s.pool.QueryRow(ctx, q, args...).Scan(&stats.Total); err != nil {
		return nil, fmt.Errorf("stats total: %w", err)
	}

	// By type
	q = fmt.Sprintf(`SELECT memory_type, COUNT(*) FROM %s WHERE 1=1%s GROUP BY memory_type`, s.table, where)
	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("stats by type: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var t string
		var c int
		if err := rows.Scan(&t, &c); err != nil {
			return nil, err
		}
		stats.ByType[t] = c
	}

	// Time range
	q = fmt.Sprintf(`SELECT MIN(created_at), MAX(created_at) FROM %s WHERE 1=1%s`, s.table, where)
	var oldest, newest *time.Time
	if err := s.pool.QueryRow(ctx, q, args...).Scan(&oldest, &newest); err != nil {
		return nil, fmt.Errorf("stats time range: %w", err)
	}
	stats.OldestAt = oldest
	stats.NewestAt = newest

	return stats, nil
}

func (s *PgvectorStore) Close() error {
	s.pool.Close()
	return nil
}

func (s *PgvectorStore) scanOne(ctx context.Context, query string, args ...any) (*memory.Memory, error) {
	row := s.pool.QueryRow(ctx, query, args...)
	var m memory.Memory
	var meta []byte
	var emb pgvector.Vector
	err := row.Scan(
		&m.ID, &m.Content, &m.MemoryType,
		&m.UserID, &m.AgentID, &m.AppID, &m.RunID,
		&meta, &m.ContentHash, &emb,
		&m.CreatedAt, &m.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan memory: %w", err)
	}
	_ = json.Unmarshal(meta, &m.Metadata) // metadata is optional; ignore parse errors
	m.Embedding = emb.Slice()
	return &m, nil
}

func scanMemory(rows pgx.Rows) (*memory.Memory, error) {
	var m memory.Memory
	var meta []byte
	var emb pgvector.Vector
	err := rows.Scan(
		&m.ID, &m.Content, &m.MemoryType,
		&m.UserID, &m.AgentID, &m.AppID, &m.RunID,
		&meta, &m.ContentHash, &emb,
		&m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan memory: %w", err)
	}
	_ = json.Unmarshal(meta, &m.Metadata) // metadata is optional; ignore parse errors
	m.Embedding = emb.Slice()
	return &m, nil
}

func buildWhere(filter *memory.SearchFilter, nextParam int) (string, []any) {
	if filter == nil {
		return "", nil
	}
	var clauses []string
	var args []any

	if filter.UserID != "" {
		clauses = append(clauses, fmt.Sprintf("user_id=$%d", nextParam))
		args = append(args, filter.UserID)
		nextParam++
	}
	if filter.AgentID != "" {
		clauses = append(clauses, fmt.Sprintf("agent_id=$%d", nextParam))
		args = append(args, filter.AgentID)
		nextParam++
	}
	if filter.AppID != "" {
		clauses = append(clauses, fmt.Sprintf("app_id=$%d", nextParam))
		args = append(args, filter.AppID)
		nextParam++
	}
	if filter.RunID != "" {
		clauses = append(clauses, fmt.Sprintf("run_id=$%d", nextParam))
		args = append(args, filter.RunID)
		nextParam++
	}
	if filter.MemoryType != "" {
		clauses = append(clauses, fmt.Sprintf("memory_type=$%d", nextParam))
		args = append(args, string(filter.MemoryType))
		nextParam++
	}
	if filter.Since != nil {
		clauses = append(clauses, fmt.Sprintf("created_at >= $%d", nextParam))
		args = append(args, *filter.Since)
		nextParam++
	}
	if filter.Until != nil {
		clauses = append(clauses, fmt.Sprintf("created_at <= $%d", nextParam))
		args = append(args, *filter.Until)
	}

	if len(clauses) == 0 {
		return "", nil
	}
	return " AND " + strings.Join(clauses, " AND "), args
}
