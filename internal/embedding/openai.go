package embedding

import (
	"context"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

type OpenAIProvider struct {
	client *openai.Client
	model  openai.EmbeddingModel
	dims   int
}

func NewOpenAI(apiKey, model, baseURL string) *OpenAIProvider {
	m := openai.SmallEmbedding3
	dims := 1536
	switch model {
	case "text-embedding-3-large":
		m = openai.LargeEmbedding3
		dims = 3072
	case "text-embedding-ada-002":
		m = openai.AdaEmbeddingV2
		dims = 1536
	}

	cfg := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		cfg.BaseURL = baseURL
	}

	return &OpenAIProvider{
		client: openai.NewClientWithConfig(cfg),
		model:  m,
		dims:   dims,
	}
}

func (p *OpenAIProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	resp, err := p.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Input: []string{text},
		Model: p.model,
	})
	if err != nil {
		return nil, fmt.Errorf("openai embed: %w", err)
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("openai embed: empty response")
	}
	return resp.Data[0].Embedding, nil
}

func (p *OpenAIProvider) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	resp, err := p.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Input: texts,
		Model: p.model,
	})
	if err != nil {
		return nil, fmt.Errorf("openai embed batch: %w", err)
	}
	results := make([][]float32, len(resp.Data))
	for i, d := range resp.Data {
		results[i] = d.Embedding
	}
	return results, nil
}

func (p *OpenAIProvider) Dimensions() int {
	return p.dims
}
