package rag

import (
	"context"
	"testing"

	rag "docker-cli/internal/rag/storage"
	models "docker-cli/internal/models"
)

type mockEmbedder struct {
	embedQueryFunc func(ctx context.Context, text string) ([]float32, error)
}

func (m *mockEmbedder) Embed(ctx context.Context, chunk models.Chunk) (models.VectorDocument, error) {
	return models.VectorDocument{}, nil
}

func (m *mockEmbedder) EmbedBatch(ctx context.Context, chunks []models.Chunk) ([]models.VectorDocument, error) {
	return nil, nil
}

func (m *mockEmbedder) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	if m.embedQueryFunc != nil {
		return m.embedQueryFunc(ctx, text)
	}
	return []float32{0.1, 0.2, 0.3}, nil
}

func (m *mockEmbedder) Close() {}

func TestRetriever_SearchUserReqeust(t *testing.T) {
	ctx := context.Background()

	// Create a vector index with test data
	index, err := rag.NewVectorIndex(ctx)
	if err != nil {
		t.Fatalf("rag.NewVectorIndex error: %v", err)
	}

	ch := make(chan models.VectorDocument, 2)
	go func() {
		ch <- models.VectorDocument{
			Id:         "doc1",
			Embeddings: []float32{1.0, 0.0},
			Content:    "Docker container management",
			MetaData: models.MetaData{Title: "Container Guide"},
		}
		ch <- models.VectorDocument{
			Id:         "doc2",
			Embeddings: []float32{0.0, 1.0},
			Content:    "Docker image building",
			MetaData: models.MetaData{Title: "Image Guide"},
		}
		close(ch)
	}()

	err = index.BuildVectorIndex(ctx, ch)
	if err != nil {
		t.Fatalf("BuildVectorIndex error: %v", err)
	}

	// Create retriever with mock embedder
	embedder := &mockEmbedder{
		embedQueryFunc: func(ctx context.Context, text string) ([]float32, error) {
			// Return embedding similar to first document
			return []float32{0.9, 0.1}, nil
		},
	}

	retriever := &Retriever{
		index:             index,
		embedder:          embedder,
		topKWhenSearching: 2,
	}

	results, err := retriever.SearchUserReqeust("How to manage containers?")
	if err != nil {
		t.Fatalf("SearchUserReqeust error: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Search returned %d results, want 2", len(results))
	}

	// First result should be about containers
	if results[0].Document.Id != "doc1" {
		t.Errorf("First result ID = %q, want doc1", results[0].Document.Id)
	}
	if results[0].Document.MetaData.Title != "Container Guide" {
		t.Errorf("First result title = %q, want Container Guide", results[0].Document.MetaData.Title)
	}
}

func TestRetriever_SearchUserReqeust_EmbedderError(t *testing.T) {
	ctx := context.Background()
	index, err := rag.NewVectorIndex(ctx)
	if err != nil {
		t.Fatalf("rag.NewVectorIndex error: %v", err)
	}

	embedder := &mockEmbedder{
		embedQueryFunc: func(ctx context.Context, text string) ([]float32, error) {
			return nil, context.Canceled
		},
	}

	retriever := &Retriever{
		index:             index,
		embedder:          embedder,
		topKWhenSearching: 2,
	}

	_, err = retriever.SearchUserReqeust("test query")
	if err == nil {
		t.Error("Expected error from embedder")
	}
}

func TestRetriever_SearchUserReqeust_EmptyIndex(t *testing.T) {
	ctx := context.Background()
	index, err := rag.NewVectorIndex(ctx)
	if err != nil {
		t.Fatalf("rag.NewVectorIndex error: %v", err)
	}

	embedder := &mockEmbedder{
		embedQueryFunc: func(ctx context.Context, text string) ([]float32, error) {
			return []float32{0.1, 0.2}, nil
		},
	}

	retriever := &Retriever{
		index:             index,
		embedder:          embedder,
		topKWhenSearching: 2,
	}

	results, err := retriever.SearchUserReqeust("test query")
	// Empty index search may error or return empty
	if err != nil {
		// Expected behavior - empty index can't search
		return
	}

	if len(results) != 0 {
		t.Errorf("Search on empty index returned %d results, want 0", len(results))
	}
}

func TestRetriever_SearchUserReqeust_TopKLimit(t *testing.T) {
	ctx := context.Background()
	index, err := rag.NewVectorIndex(ctx)
	if err != nil {
		t.Fatalf("rag.NewVectorIndex error: %v", err)
	}

	ch := make(chan models.VectorDocument, 5)
	go func() {
		for i := 0; i < 5; i++ {
			ch <- models.VectorDocument{
				Id:         "doc" + string(rune('0'+i)),
				Embeddings: []float32{float32(i), 0.0},
				Content:    "content",
				MetaData: models.MetaData{Title: "Doc"},
			}
		}
		close(ch)
	}()

	err = index.BuildVectorIndex(ctx, ch)
	if err != nil {
		t.Fatalf("BuildVectorIndex error: %v", err)
	}

	embedder := &mockEmbedder{
		embedQueryFunc: func(ctx context.Context, text string) ([]float32, error) {
			return []float32{2.5, 0.0}, nil
		},
	}

	retriever := &Retriever{
		index:             index,
		embedder:          embedder,
		topKWhenSearching: 3,
	}

	results, err := retriever.SearchUserReqeust("test query")
	if err != nil {
		t.Fatalf("SearchUserReqeust error: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Search with topK=3 returned %d results, want 3", len(results))
	}
}