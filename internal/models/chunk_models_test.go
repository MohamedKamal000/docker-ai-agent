package models

import (
	"encoding/json"
	"testing"
)

func TestMetaData_JSON(t *testing.T) {
	meta := MetaData{
		Title:       "Test Title",
		Description: "Test Description",
		Keywords:    []string{"keyword1", "keyword2"},
	}

	jsonData, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var parsed MetaData
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if parsed.Title != meta.Title {
		t.Errorf("Title = %q, want %q", parsed.Title, meta.Title)
	}
	if parsed.Description != meta.Description {
		t.Errorf("Description = %q, want %q", parsed.Description, meta.Description)
	}
	if len(parsed.Keywords) != 2 {
		t.Errorf("Keywords = %d, want 2", len(parsed.Keywords))
	}
}

func TestSourceDocument_TransformToChunk(t *testing.T) {
	doc := SourceDocument{
		MetaData: MetaData{
			Title:       "Doc Title",
			Description: "Doc Description",
			Keywords:    []string{"doc", "test"},
		},
		Content: "Full document content",
		Path:    "/path/to/doc.md",
	}

	chunk := doc.TransformToChunk("Chunk content", 1)

	if chunk.Content != "Chunk content" {
		t.Errorf("Content = %q, want %q", chunk.Content, "Chunk content")
	}
	if chunk.Index != 1 {
		t.Errorf("Index = %d, want 1", chunk.Index)
	}
	if chunk.SourcePath != "/path/to/doc.md" {
		t.Errorf("SourcePath = %q, want %q", chunk.SourcePath, "/path/to/doc.md")
	}
	if chunk.MetaData.Title != "Doc Title" {
		t.Errorf("MetaData.Title = %q, want %q", chunk.MetaData.Title, "Doc Title")
	}
}

func TestChunk_String(t *testing.T) {
	chunk := Chunk{
		MetaData: MetaData{
			Title:       "Chunk Title",
			Description: "Chunk Description",
		},
		Content: "Chunk content here",
	}

	result := chunk.String()

	expected := "Title: Chunk Title\nDescription: Chunk Description\nChunk content here"
	if result != expected {
		t.Errorf("String() = %q, want %q", result, expected)
	}
}

func TestVectorDocument_JSON(t *testing.T) {
	vd := VectorDocument{
		Id:         "doc:1",
		Content:    "Vector document content",
		MetaData: MetaData{
			Title: "Vector Doc",
		},
		Embeddings: []float32{0.1, 0.2, 0.3},
	}

	jsonData, err := json.Marshal(vd)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var parsed VectorDocument
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if parsed.Id != vd.Id {
		t.Errorf("Id = %q, want %q", parsed.Id, vd.Id)
	}
	if parsed.Content != vd.Content {
		t.Errorf("Content = %q, want %q", parsed.Content, vd.Content)
	}
	if parsed.MetaData.Title != vd.MetaData.Title {
		t.Errorf("MetaData.Title = %q, want %q", parsed.MetaData.Title, vd.MetaData.Title)
	}
	if len(parsed.Embeddings) != 3 {
		t.Errorf("Embeddings = %d, want 3", len(parsed.Embeddings))
	}
}

func TestNewVectorDocument(t *testing.T) {
	chunk := Chunk{
		SourcePath: "/path/to/doc.md",
		Index:      5,
		MetaData: MetaData{
			Title: "Chunk Title",
		},
		Content: "Chunk content",
	}

	embeddings := []float32{0.1, 0.2, 0.3, 0.4}
	vd := NewVectorDocument(chunk, embeddings)

	if vd.Id != "/path/to/doc.md:5" {
		t.Errorf("Id = %q, want %q", vd.Id, "/path/to/doc.md:5")
	}
	if vd.Content != chunk.Content {
		t.Errorf("Content = %q, want %q", vd.Content, chunk.Content)
	}
	if vd.MetaData.Title != chunk.MetaData.Title {
		t.Errorf("MetaData.Title = %q, want %q", vd.MetaData.Title, chunk.MetaData.Title)
	}
	if len(vd.Embeddings) != 4 {
		t.Errorf("Embeddings = %d, want 4", len(vd.Embeddings))
	}
	if vd.Embeddings[0] != 0.1 {
		t.Errorf("Embeddings[0] = %v, want 0.1", vd.Embeddings[0])
	}
}

func TestSearchResult_JSON(t *testing.T) {
	sr := SearchResult{
		Document: VectorDocument{
			Id:       "doc:1",
			Content:  "Search result content",
			MetaData: MetaData{Title: "Search Doc"},
			Embeddings: []float32{0.5, 0.5},
		},
		Similarity: 0.95,
	}

	jsonData, err := json.Marshal(sr)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var parsed SearchResult
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if parsed.Document.Id != sr.Document.Id {
		t.Errorf("Document.Id = %q, want %q", parsed.Document.Id, sr.Document.Id)
	}
	if parsed.Similarity != sr.Similarity {
		t.Errorf("Similarity = %v, want %v", parsed.Similarity, sr.Similarity)
	}
}