package embedding

import (
	"context"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiProvider struct {
	client *genai.Client
	model  string
	dims   int
}

func NewGemini(ctx context.Context, apiKey, model string) (*GeminiProvider, error) {
	if model == "" {
		model = "text-embedding-004"
	}

	dims := 768
	switch model {
	case "text-embedding-004":
		dims = 768
	case "embedding-001":
		dims = 768
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("gemini client: %w", err)
	}

	return &GeminiProvider{
		client: client,
		model:  model,
		dims:   dims,
	}, nil
}

func (p *GeminiProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	em := p.client.EmbeddingModel(p.model)
	res, err := em.EmbedContent(ctx, genai.Text(text))
	if err != nil {
		return nil, fmt.Errorf("gemini embed: %w", err)
	}
	if res == nil || res.Embedding == nil {
		return nil, fmt.Errorf("gemini embed: empty response")
	}
	return res.Embedding.Values, nil
}

func (p *GeminiProvider) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	em := p.client.EmbeddingModel(p.model)
	batch := em.NewBatch()
	for _, t := range texts {
		batch.AddContent(genai.Text(t))
	}

	res, err := em.BatchEmbedContents(ctx, batch)
	if err != nil {
		return nil, fmt.Errorf("gemini embed batch: %w", err)
	}

	results := make([][]float32, len(res.Embeddings))
	for i, e := range res.Embeddings {
		results[i] = e.Values
	}
	return results, nil
}

func (p *GeminiProvider) Dimensions() int {
	return p.dims
}

func (p *GeminiProvider) Close() error {
	return p.client.Close()
}
