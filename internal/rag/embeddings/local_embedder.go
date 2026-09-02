package rag

import (
	"context"
	"fmt"
	"path/filepath"

	models "docker-cli/internal/rag/models"

	"github.com/knights-analytics/hugot"
	"github.com/knights-analytics/hugot/options"
	"github.com/knights-analytics/hugot/pipelines"
)

type LocalEmbedder struct {
	session  *hugot.Session
	pipeline *pipelines.FeatureExtractionPipeline
}

func NewLocalEmbedder(ctx context.Context, modelDir string, cudaEnabled bool) (*LocalEmbedder, error) {
	modelPath, err := filepath.Abs(modelDir)
	if err != nil {
		return nil, fmt.Errorf("resolve model path: %w", err)
	}

	var session *hugot.Session
	if cudaEnabled {
		opts := []options.WithOption{
			options.WithCuda(map[string]string{
				"device_id": "0",
			}),
			options.WithOnnxLibraryPath("/usr/local/lib/"),
		}
		session, err = hugot.NewORTSession(ctx, opts...)
	} else {
		session, err = hugot.NewGoSession(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("create Hugot Go session: %w", err)
	}

	config := hugot.FeatureExtractionConfig{
		Name:      "bge-small-en-v1.5",
		ModelPath: modelPath,
	}

	pipeline, err := hugot.NewPipeline(
		session,
		config,
	)
	if err != nil {
		session.Destroy()
		return nil, fmt.Errorf("create embedding pipeline: %w", err)
	}

	return &LocalEmbedder{
		session:  session,
		pipeline: pipeline,
	}, nil
}

func (e *LocalEmbedder) Close() {
	if e.session != nil {
		e.session.Destroy()
	}
}

func (e *LocalEmbedder) Embed(ctx context.Context, chunk models.Chunk) (models.VectorDocument, error) {
	output, err := e.pipeline.RunPipeline(ctx, []string{chunk.String()})
	if err != nil {
		return models.VectorDocument{}, fmt.Errorf("generate embedding: %w", err)
	}

	if len(output.Embeddings) != 1 {
		return models.VectorDocument{}, fmt.Errorf(
			"expected one embedding, got %d",
			len(output.Embeddings),
		)
	}

	return models.NewVectorDocument(chunk, output.Embeddings[0]), nil
}

func (e *LocalEmbedder) EmbedBatch(
	ctx context.Context, chunks []models.Chunk,
) ([]models.VectorDocument, error) {
	if len(chunks) == 0 {
		return []models.VectorDocument{}, nil
	}

	texts := make([]string, len(chunks))
	for i, c := range chunks {
		texts[i] = c.String()
	}

	output, err := e.pipeline.RunPipeline(ctx, texts)
	if err != nil {
		return nil, fmt.Errorf("generate batch embeddings: %w", err)
	}

	result := make([]models.VectorDocument, len(output.Embeddings))

	for i, emd := range output.Embeddings {
		result[i] = models.NewVectorDocument(chunks[i], emd)
	}

	return result, nil
}

func (e *LocalEmbedder) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	output, err := e.pipeline.RunPipeline(ctx, []string{text})
	if err != nil {
		return nil, err
	}

	if len(output.Embeddings) != 1 {
		return nil, fmt.Errorf(
			"expected one embedding, got %d",
			len(output.Embeddings),
		)
	}
	return output.Embeddings[0], nil
}
