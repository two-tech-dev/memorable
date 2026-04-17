package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type OllamaProvider struct {
	baseURL string
	model   string
	dims    int
	client  *http.Client
}

func NewOllama(baseURL, model string, dims int) *OllamaProvider {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if model == "" {
		model = "nomic-embed-text"
	}
	if dims <= 0 {
		dims = 768
	}

	return &OllamaProvider{
		baseURL: baseURL,
		model:   model,
		dims:    dims,
		client:  &http.Client{},
	}
}

type ollamaEmbedRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type ollamaEmbedResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
}

func (p *OllamaProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	results, err := p.embedRaw(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("ollama embed: empty response")
	}
	return results[0], nil
}

func (p *OllamaProvider) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	return p.embedRaw(ctx, texts)
}

func (p *OllamaProvider) Dimensions() int {
	return p.dims
}

func (p *OllamaProvider) embedRaw(ctx context.Context, inputs []string) ([][]float32, error) {
	body, err := json.Marshal(ollamaEmbedRequest{
		Model: p.model,
		Input: inputs,
	})
	if err != nil {
		return nil, fmt.Errorf("ollama marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/api/embed", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("ollama request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama embed: status %d: %s", resp.StatusCode, string(respBody))
	}

	var result ollamaEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("ollama decode: %w", err)
	}

	return result.Embeddings, nil
}
