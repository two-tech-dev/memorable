package embedding

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCustomProvider_Embed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/embeddings" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("X-Title") != "Memorable" {
			t.Errorf("expected custom header X-Title=Memorable, got %s", r.Header.Get("X-Title"))
		}

		var req customEmbedRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if req.Model != "test-model" {
			t.Errorf("expected model test-model, got %s", req.Model)
		}

		if err := json.NewEncoder(w).Encode(customEmbedResponse{
			Data: []struct {
				Embedding []float32 `json:"embedding"`
				Index     int       `json:"index"`
			}{
				{Embedding: []float32{0.1, 0.2, 0.3}, Index: 0},
			},
		}); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	p := NewCustom(server.URL, "test-key", "test-model", 3, map[string]string{
		"X-Title": "Memorable",
	})

	vec, err := p.Embed(context.Background(), "hello")
	if err != nil {
		t.Fatal(err)
	}
	if len(vec) != 3 {
		t.Fatalf("expected 3 dims, got %d", len(vec))
	}
	if vec[0] != 0.1 {
		t.Errorf("expected 0.1, got %f", vec[0])
	}
}

func TestCustomProvider_EmbedBatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode(customEmbedResponse{
			Data: []struct {
				Embedding []float32 `json:"embedding"`
				Index     int       `json:"index"`
			}{
				{Embedding: []float32{0.1, 0.2}, Index: 0},
				{Embedding: []float32{0.3, 0.4}, Index: 1},
			},
		}); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	p := NewCustom(server.URL, "key", "model", 2, nil)

	vecs, err := p.EmbedBatch(context.Background(), []string{"a", "b"})
	if err != nil {
		t.Fatal(err)
	}
	if len(vecs) != 2 {
		t.Fatalf("expected 2 results, got %d", len(vecs))
	}
}

func TestCustomProvider_ErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"message":"Invalid API key"}}`))
	}))
	defer server.Close()

	p := NewCustom(server.URL, "bad-key", "model", 3, nil)
	_, err := p.Embed(context.Background(), "hello")
	if err == nil {
		t.Fatal("expected error for 401")
	}
}

func TestCustomProvider_Dimensions(t *testing.T) {
	p := NewCustom("http://localhost", "", "model", 768, nil)
	if p.Dimensions() != 768 {
		t.Errorf("expected 768, got %d", p.Dimensions())
	}
}

func TestCustomProvider_DefaultDimensions(t *testing.T) {
	p := NewCustom("http://localhost", "", "model", 0, nil)
	if p.Dimensions() != 1536 {
		t.Errorf("expected default 1536, got %d", p.Dimensions())
	}
}
