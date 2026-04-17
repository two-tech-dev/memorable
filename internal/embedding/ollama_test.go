package embedding

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOllamaEmbed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/embed" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}

		var req ollamaEmbedRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if req.Model != "nomic-embed-text" {
			t.Errorf("expected model nomic-embed-text, got %q", req.Model)
		}

		resp := ollamaEmbedResponse{
			Embeddings: [][]float32{{0.1, 0.2, 0.3}},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewOllama(server.URL, "nomic-embed-text", 3)

	vec, err := provider.Embed(context.Background(), "hello")
	if err != nil {
		t.Fatalf("Embed: %v", err)
	}
	if len(vec) != 3 {
		t.Errorf("expected 3 dims, got %d", len(vec))
	}
	if vec[0] != 0.1 {
		t.Errorf("expected first value 0.1, got %f", vec[0])
	}
}

func TestOllamaEmbedBatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ollamaEmbedRequest
		json.NewDecoder(r.Body).Decode(&req)

		resp := ollamaEmbedResponse{
			Embeddings: make([][]float32, len(req.Input)),
		}
		for i := range req.Input {
			resp.Embeddings[i] = []float32{float32(i) * 0.1, 0.2, 0.3}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewOllama(server.URL, "nomic-embed-text", 3)

	vecs, err := provider.EmbedBatch(context.Background(), []string{"a", "b", "c"})
	if err != nil {
		t.Fatalf("EmbedBatch: %v", err)
	}
	if len(vecs) != 3 {
		t.Errorf("expected 3 vectors, got %d", len(vecs))
	}
}

func TestOllamaEmbedServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("model not found"))
	}))
	defer server.Close()

	provider := NewOllama(server.URL, "nonexistent", 3)

	_, err := provider.Embed(context.Background(), "hello")
	if err == nil {
		t.Error("expected error for server error response")
	}
}

func TestOllamaDimensions(t *testing.T) {
	provider := NewOllama("", "nomic-embed-text", 768)
	if provider.Dimensions() != 768 {
		t.Errorf("expected 768, got %d", provider.Dimensions())
	}
}

func TestOllamaDefaults(t *testing.T) {
	provider := NewOllama("", "", 0)
	if provider.baseURL != "http://localhost:11434" {
		t.Errorf("expected default base URL, got %q", provider.baseURL)
	}
	if provider.model != "nomic-embed-text" {
		t.Errorf("expected default model, got %q", provider.model)
	}
	if provider.dims != 768 {
		t.Errorf("expected default dims 768, got %d", provider.dims)
	}
}
