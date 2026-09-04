package rag

import (
	"context"

	models "docker-cli/internal/models"
)

type Embedder interface {
	Embed(ctx context.Context, chunk models.Chunk) (models.VectorDocument, error)
	EmbedBatch(ctx context.Context, chunk []models.Chunk) ([]models.VectorDocument, error)
	EmbedQuery(ctx context.Context, text string) ([]float32, error)
	Close()
}
