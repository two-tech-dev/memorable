package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()

	if cfg.Server.Transport != "stdio" {
		t.Errorf("expected transport stdio, got %q", cfg.Server.Transport)
	}
	if cfg.Storage.Backend != "pgvector" {
		t.Errorf("expected backend pgvector, got %q", cfg.Storage.Backend)
	}
	if cfg.Storage.Pgvector.TableName != "memories" {
		t.Errorf("expected table memories, got %q", cfg.Storage.Pgvector.TableName)
	}
	if cfg.Storage.Pgvector.VectorDimensions != 1536 {
		t.Errorf("expected dims 1536, got %d", cfg.Storage.Pgvector.VectorDimensions)
	}
	if cfg.Embedding.Provider != "openai" {
		t.Errorf("expected provider openai, got %q", cfg.Embedding.Provider)
	}
	if cfg.Embedding.OpenAI.Model != "text-embedding-3-small" {
		t.Errorf("expected model text-embedding-3-small, got %q", cfg.Embedding.OpenAI.Model)
	}
}

func TestLoadFromFile(t *testing.T) {
	t.Setenv("MEMORABLE_DSN", "")
	content := `
server:
  transport: stdio
storage:
  backend: pgvector
  pgvector:
    dsn: "postgres://test:5432/testdb"
    table_name: custom_memories
    vector_dimensions: 768
embedding:
  provider: gemini
  gemini:
    api_key: test-key
    model: text-embedding-004
`
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFromPath(path)
	if err != nil {
		t.Fatalf("LoadFromPath: %v", err)
	}

	if cfg.Storage.Pgvector.DSN != "postgres://test:5432/testdb" {
		t.Errorf("expected custom DSN, got %q", cfg.Storage.Pgvector.DSN)
	}
	if cfg.Storage.Pgvector.TableName != "custom_memories" {
		t.Errorf("expected custom_memories, got %q", cfg.Storage.Pgvector.TableName)
	}
	if cfg.Storage.Pgvector.VectorDimensions != 768 {
		t.Errorf("expected 768 dims, got %d", cfg.Storage.Pgvector.VectorDimensions)
	}
	if cfg.Embedding.Provider != "gemini" {
		t.Errorf("expected provider gemini, got %q", cfg.Embedding.Provider)
	}
	if cfg.Embedding.Gemini.APIKey != "test-key" {
		t.Errorf("expected api_key test-key, got %q", cfg.Embedding.Gemini.APIKey)
	}
}

func TestEnvOverrides(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "sk-test-from-env")
	t.Setenv("GEMINI_API_KEY", "gemini-test-from-env")
	t.Setenv("MEMORABLE_DSN", "postgres://env:5432/envdb")

	cfg := defaultConfig()
	applyEnvOverrides(cfg)

	if cfg.Embedding.OpenAI.APIKey != "sk-test-from-env" {
		t.Errorf("expected OpenAI key from env, got %q", cfg.Embedding.OpenAI.APIKey)
	}
	if cfg.Embedding.Gemini.APIKey != "gemini-test-from-env" {
		t.Errorf("expected Gemini key from env, got %q", cfg.Embedding.Gemini.APIKey)
	}
	if cfg.Storage.Pgvector.DSN != "postgres://env:5432/envdb" {
		t.Errorf("expected DSN from env, got %q", cfg.Storage.Pgvector.DSN)
	}
}

func TestExpandEnvRef(t *testing.T) {
	t.Setenv("MY_SECRET", "secret-value")

	got := expandEnvRef("${MY_SECRET}")
	if got != "secret-value" {
		t.Errorf("expected 'secret-value', got %q", got)
	}

	got = expandEnvRef("plain-value")
	if got != "plain-value" {
		t.Errorf("expected 'plain-value', got %q", got)
	}

	got = expandEnvRef("")
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestLoadNoFile(t *testing.T) {
	// When no config file exists, Load() returns defaults
	orig, _ := os.Getwd()
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(orig) }()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Server.Transport != "stdio" {
		t.Errorf("expected default transport, got %q", cfg.Server.Transport)
	}
}

func TestConfigSearchPaths(t *testing.T) {
	paths := configSearchPaths()
	if len(paths) == 0 {
		t.Error("expected at least one search path")
	}
	if paths[0] != "memorable.yaml" {
		t.Errorf("expected first path to be 'memorable.yaml', got %q", paths[0])
	}
}
