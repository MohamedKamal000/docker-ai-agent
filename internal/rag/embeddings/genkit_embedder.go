package rag

import (
	"context"

	models "docker-cli/internal/models"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

type GenkitEmbedder struct {
	g         *genkit.Genkit
	modelName string
}

func NewGenkitEmbedder(ctx context.Context, modelOptions genkit.GenkitOption, modelName string) *GenkitEmbedder {
	return &GenkitEmbedder{
		g:         genkit.Init(ctx, modelOptions, genkit.WithDefaultModel(modelName)),
		modelName: modelName,
	}
}

func (ge *GenkitEmbedder) Embed(ctx context.Context, chunk models.Chunk) (models.VectorDocument, error) {
	resp, err := genkit.Embed(ctx, ge.g, ai.WithEmbedderName(ge.modelName), ai.WithTextDocs(chunk.String()))
	if err != nil {
		return models.VectorDocument{}, err
	}
	embeddings := resp.Embeddings[0].Embedding
	return models.NewVectorDocument(chunk, embeddings), nil
}

func (ge *GenkitEmbedder) EmbedBatch(ctx context.Context, chunks []models.Chunk) ([]models.VectorDocument, error) {
	length := len(chunks)
	text := make([]string, length)
	for index, chunk := range chunks {
		text[index] = chunk.String()
	}

	resp, err := genkit.Embed(ctx, ge.g, ai.WithEmbedderName(ge.modelName), ai.WithTextDocs(text...))
	if err != nil {
		return []models.VectorDocument{}, err
	}

	result := make([]models.VectorDocument, length)
	for index, emd := range resp.Embeddings {
		result[index] = models.NewVectorDocument(chunks[index], emd.Embedding)
	}

	return result, nil
}

func (ge *GenkitEmbedder) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	resp, err := genkit.Embed(ctx, ge.g, ai.WithEmbedderName(ge.modelName), ai.WithTextDocs(text))
	if err != nil {
		return nil, err
	}
	embeddings := resp.Embeddings[0].Embedding
	return embeddings, nil
}

func (ge *GenkitEmbedder) Close() {
	// left this way so it applies to the interface
}
