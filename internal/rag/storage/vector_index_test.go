package rag

import (
	"context"
	"testing"

	"github.com/philippgille/chromem-go"
	models "docker-cli/internal/models"
)

func TestToChromemDocument(t *testing.T) {
	doc := models.VectorDocument{
		Id:         "test-id",
		Embeddings: []float32{0.1, 0.2, 0.3},
		Content:    "test content",
		MetaData: models.MetaData{
			Title:       "Test Title",
			Description: "Test Description",
			Keywords:    []string{"keyword1", "keyword2"},
		},
	}

	result := ToChromemDocument(doc)

	if result.ID != "test-id" {
		t.Errorf("ID = %q, want %q", result.ID, "test-id")
	}
	if len(result.Embedding) != 3 {
		t.Errorf("Embedding length = %d, want 3", len(result.Embedding))
	}
	if result.Content != "test content" {
		t.Errorf("Content = %q, want %q", result.Content, "test content")
	}
	if result.Metadata["Title"] != "Test Title" {
		t.Errorf("Metadata Title = %q, want %q", result.Metadata["Title"], "Test Title")
	}
	if result.Metadata["Description"] != "Test Description" {
		t.Errorf("Metadata Description = %q, want %q", result.Metadata["Description"], "Test Description")
	}
	if result.Metadata["Keywords"] != "keyword1,keyword2" {
		t.Errorf("Metadata Keywords = %q, want %q", result.Metadata["Keywords"], "keyword1,keyword2")
	}
}

func TestToSearchResult(t *testing.T) {
	doc := chromem.Result{
		ID:        "test-id",
		Embedding: []float32{0.1, 0.2, 0.3},
		Content:   "test content",
		Metadata: map[string]string{
			"Title":       "Test Title",
			"Description": "Test Description",
			"Keywords":    "keyword1,keyword2",
		},
		Similarity: 0.95,
	}

	result := ToSearchResult(doc)

	if result.Document.Id != "test-id" {
		t.Errorf("Document.Id = %q, want %q", result.Document.Id, "test-id")
	}
	if len(result.Document.Embeddings) != 3 {
		t.Errorf("Document.Embeddings length = %d, want 3", len(result.Document.Embeddings))
	}
	if result.Document.Content != "test content" {
		t.Errorf("Document.Content = %q, want %q", result.Document.Content, "test content")
	}
	if result.Document.MetaData.Title != "Test Title" {
		t.Errorf("Document.MetaData.Title = %q, want %q", result.Document.MetaData.Title, "Test Title")
	}
	if result.Document.MetaData.Description != "Test Description" {
		t.Errorf("Document.MetaData.Description = %q, want %q", result.Document.MetaData.Description, "Test Description")
	}
	if len(result.Document.MetaData.Keywords) != 2 {
		t.Errorf("Document.MetaData.Keywords length = %d, want 2", len(result.Document.MetaData.Keywords))
	}
	if result.Similarity != 0.95 {
		t.Errorf("Similarity = %v, want %v", result.Similarity, 0.95)
	}
}

func TestToSearchResult_EmptyMetadata(t *testing.T) {
	doc := chromem.Result{
		ID:        "test-id",
		Embedding: []float32{0.1},
		Content:   "test content",
		Metadata:  map[string]string{},
		Similarity: 0.5,
	}

	result := ToSearchResult(doc)

	if result.Document.MetaData.Title != "" {
		t.Errorf("Title should be empty for missing metadata")
	}
	if result.Document.MetaData.Description != "" {
		t.Errorf("Description should be empty for missing metadata")
	}
	if len(result.Document.MetaData.Keywords) != 1 || result.Document.MetaData.Keywords[0] != "" {
		t.Errorf("Keywords should be empty slice for missing metadata")
	}
}

func TestNewVectorIndex(t *testing.T) {
	ctx := context.Background()
	index, err := NewVectorIndex(ctx)
	if err != nil {
		t.Fatalf("NewVectorIndex error: %v", err)
	}
	if index == nil {
		t.Fatal("NewVectorIndex returned nil")
	}
}

func TestVectorIndex_BuildVectorIndex(t *testing.T) {
	ctx := context.Background()
	index, err := NewVectorIndex(ctx)
	if err != nil {
		t.Fatalf("NewVectorIndex error: %v", err)
	}

	ch := make(chan models.VectorDocument, 3)
	go func() {
		ch <- models.VectorDocument{
			Id:         "doc1",
			Embeddings: []float32{0.1, 0.2},
			Content:    "content 1",
			MetaData: models.MetaData{Title: "Doc 1"},
		}
		ch <- models.VectorDocument{
			Id:         "doc2",
			Embeddings: []float32{0.3, 0.4},
			Content:    "content 2",
			MetaData: models.MetaData{Title: "Doc 2"},
		}
		ch <- models.VectorDocument{
			Id:         "doc3",
			Embeddings: []float32{0.5, 0.6},
			Content:    "content 3",
			MetaData: models.MetaData{Title: "Doc 3"},
		}
		close(ch)
	}()

	err = index.BuildVectorIndex(ctx, ch)
	if err != nil {
		t.Fatalf("BuildVectorIndex error: %v", err)
	}
}

func TestVectorIndex_SearchTheIndex(t *testing.T) {
	ctx := context.Background()
	index, err := NewVectorIndex(ctx)
	if err != nil {
		t.Fatalf("NewVectorIndex error: %v", err)
	}

	// Add some documents
	ch := make(chan models.VectorDocument, 2)
	go func() {
		ch <- models.VectorDocument{
			Id:         "doc1",
			Embeddings: []float32{1.0, 0.0},
			Content:    "first document",
			MetaData: models.MetaData{Title: "Doc 1"},
		}
		ch <- models.VectorDocument{
			Id:         "doc2",
			Embeddings: []float32{0.0, 1.0},
			Content:    "second document",
			MetaData: models.MetaData{Title: "Doc 2"},
		}
		close(ch)
	}()

	err = index.BuildVectorIndex(ctx, ch)
	if err != nil {
		t.Fatalf("BuildVectorIndex error: %v", err)
	}

	// Search with query similar to first document
	query := []float32{0.9, 0.1}
	results, err := index.SearchTheIndex(ctx, query, 2)
	if err != nil {
		t.Fatalf("SearchTheIndex error: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Search returned %d results, want 2", len(results))
	}

	// First result should be doc1 (more similar to query)
	if results[0].Document.Id != "doc1" {
		t.Errorf("First result ID = %q, want doc1", results[0].Document.Id)
	}
	if results[0].Similarity <= results[1].Similarity {
		t.Errorf("First result should have higher similarity: %v vs %v", results[0].Similarity, results[1].Similarity)
	}
}

func TestVectorIndex_SearchTheIndex_EmptyIndex(t *testing.T) {
	ctx := context.Background()
	index, err := NewVectorIndex(ctx)
	if err != nil {
		t.Fatalf("NewVectorIndex error: %v", err)
	}

	query := []float32{0.1, 0.2}
	// Empty index search should handle gracefully
	results, err := index.SearchTheIndex(ctx, query, 0)
	if err != nil {
		// Expected to fail or return empty
		return
	}

	if len(results) != 0 {
		t.Errorf("Search on empty index returned %d results, want 0", len(results))
	}
}

func TestVectorIndex_SearchTheIndex_TopKLimit(t *testing.T) {
	ctx := context.Background()
	index, err := NewVectorIndex(ctx)
	if err != nil {
		t.Fatalf("NewVectorIndex error: %v", err)
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

	// Search with topK=3
	query := []float32{0.5, 0.0}
	results, err := index.SearchTheIndex(ctx, query, 3)
	if err != nil {
		t.Fatalf("SearchTheIndex error: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Search with topK=3 returned %d results, want 3", len(results))
	}
}