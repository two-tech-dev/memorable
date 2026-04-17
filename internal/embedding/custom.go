package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// CustomProvider is a generic OpenAI-compatible embedding provider.
// Works with OpenRouter, Together AI, Voyage AI, Mistral, Cohere, Azure OpenAI,
// Fireworks, Anyscale, Groq, DeepInfra, and any other API that implements the
// OpenAI /v1/embeddings endpoint.
type CustomProvider struct {
	baseURL string
	apiKey  string
	model   string
	dims    int
	client  *http.Client
	headers map[string]string
}

func NewCustom(baseURL, apiKey, model string, dims int, headers map[string]string) *CustomProvider {
	if dims <= 0 {
		dims = 1536
	}
	return &CustomProvider{
		baseURL: baseURL,
		apiKey:  apiKey,
		model:   model,
		dims:    dims,
		client:  &http.Client{},
		headers: headers,
	}
}

type customEmbedRequest struct {
	Input          any    `json:"input"`
	Model          string `json:"model"`
	EncodingFormat string `json:"encoding_format,omitempty"`
}

type customEmbedResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (p *CustomProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	results, err := p.embedRaw(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("custom embed: empty response")
	}
	return results[0], nil
}

func (p *CustomProvider) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	return p.embedRaw(ctx, texts)
}

func (p *CustomProvider) Dimensions() int {
	return p.dims
}

func (p *CustomProvider) embedRaw(ctx context.Context, inputs []string) ([][]float32, error) {
	var input any = inputs
	if len(inputs) == 1 {
		input = inputs[0]
	}

	body, err := json.Marshal(customEmbedRequest{
		Input: input,
		Model: p.model,
	})
	if err != nil {
		return nil, fmt.Errorf("custom embed marshal: %w", err)
	}

	endpoint := p.baseURL + "/embeddings"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("custom embed request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

	// Apply custom headers (e.g. X-Title, HTTP-Referer for OpenRouter)
	for k, v := range p.headers {
		req.Header.Set(k, v)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("custom embed request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("custom embed read: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("custom embed: status %d: %s", resp.StatusCode, string(respBody))
	}

	var result customEmbedResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("custom embed decode: %w", err)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("custom embed error: %s", result.Error.Message)
	}

	embeddings := make([][]float32, len(result.Data))
	for _, d := range result.Data {
		embeddings[d.Index] = d.Embedding
	}

	return embeddings, nil
}
