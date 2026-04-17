package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Storage   StorageConfig   `yaml:"storage"`
	Embedding EmbeddingConfig `yaml:"embedding"`
}

type ServerConfig struct {
	Transport string `yaml:"transport"` // stdio | http
}

type StorageConfig struct {
	Backend  string         `yaml:"backend"` // pgvector | sqlite
	Pgvector PgvectorConfig `yaml:"pgvector"`
}

type PgvectorConfig struct {
	DSN              string `yaml:"dsn"`
	TableName        string `yaml:"table_name"`
	VectorDimensions int    `yaml:"vector_dimensions"`
}

type EmbeddingConfig struct {
	Provider string             `yaml:"provider"` // openai | gemini | ollama | custom
	OpenAI   OpenAIEmbedConfig  `yaml:"openai"`
	Gemini   GeminiEmbedConfig  `yaml:"gemini"`
	Ollama   OllamaEmbedConfig  `yaml:"ollama"`
	Custom   CustomEmbedConfig  `yaml:"custom"`
}

type OpenAIEmbedConfig struct {
	APIKey  string `yaml:"api_key"`
	Model   string `yaml:"model"`
	BaseURL string `yaml:"base_url"` // Custom base URL for OpenAI-compatible APIs (Voyage, Together, Azure, etc.)
}

type GeminiEmbedConfig struct {
	APIKey string `yaml:"api_key"`
	Model  string `yaml:"model"`
}

type OllamaEmbedConfig struct {
	BaseURL string `yaml:"base_url"`
	Model   string `yaml:"model"`
	Dims    int    `yaml:"dims"`
}

type CustomEmbedConfig struct {
	BaseURL string            `yaml:"base_url"` // Required. e.g. https://openrouter.ai/api/v1
	APIKey  string            `yaml:"api_key"`
	Model   string            `yaml:"model"`    // Required. e.g. openai/text-embedding-3-small
	Dims    int               `yaml:"dims"`     // Required. Must match model output dimensions
	Headers map[string]string `yaml:"headers"` // Optional extra HTTP headers
}

func Load() (*Config, error) {
	paths := configSearchPaths()
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return loadFromFile(p)
		}
	}
	return defaultConfig(), nil
}

func LoadFromPath(path string) (*Config, error) {
	return loadFromFile(path)
}

func loadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}

	cfg := defaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}

	applyEnvOverrides(cfg)
	return cfg, nil
}

func defaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Transport: "stdio",
		},
		Storage: StorageConfig{
			Backend: "pgvector",
			Pgvector: PgvectorConfig{
				DSN:              "postgres://localhost:5432/memorable?sslmode=disable",
				TableName:        "memories",
				VectorDimensions: 1536,
			},
		},
		Embedding: EmbeddingConfig{
			Provider: "openai",
			OpenAI: OpenAIEmbedConfig{
				Model: "text-embedding-3-small",
			},
		},
	}
}

func configSearchPaths() []string {
	paths := []string{"memorable.yaml"}
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".memorable", "config.yaml"))
	}
	return paths
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("OPENAI_API_KEY"); v != "" && cfg.Embedding.OpenAI.APIKey == "" {
		cfg.Embedding.OpenAI.APIKey = v
	}
	if v := os.Getenv("GEMINI_API_KEY"); v != "" && cfg.Embedding.Gemini.APIKey == "" {
		cfg.Embedding.Gemini.APIKey = v
	}
	if v := os.Getenv("MEMORABLE_DSN"); v != "" {
		cfg.Storage.Pgvector.DSN = v
	}

	if v := os.Getenv("CUSTOM_API_KEY"); v != "" && cfg.Embedding.Custom.APIKey == "" {
		cfg.Embedding.Custom.APIKey = v
	}

	// Expand ${ENV_VAR} references in api keys
	cfg.Embedding.OpenAI.APIKey = expandEnvRef(cfg.Embedding.OpenAI.APIKey)
	cfg.Embedding.Gemini.APIKey = expandEnvRef(cfg.Embedding.Gemini.APIKey)
	cfg.Embedding.Custom.APIKey = expandEnvRef(cfg.Embedding.Custom.APIKey)
}

func expandEnvRef(val string) string {
	if strings.HasPrefix(val, "${") && strings.HasSuffix(val, "}") {
		envName := val[2 : len(val)-1]
		return os.Getenv(envName)
	}
	return val
}
