package rag

import (
	"testing"

	models "docker-cli/internal/models"
)

func TestToBytes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []byte
	}{
		{
			name:     "Simple string",
			input:    "hello",
			expected: []byte("hello"),
		},
		{
			name:     "Empty string",
			input:    "",
			expected: []byte{},
		},
		{
			name:     "String with special chars",
			input:    "test\n\r\t",
			expected: []byte("test\n\r\t"),
		},
		{
			name:     "Unicode",
			input:    "测试",
			expected: []byte("测试"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toBytes(tt.input)
			if string(result) != string(tt.expected) {
				t.Errorf("toBytes(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNewBoltClient(t *testing.T) {
	tmpDir := t.TempDir()

	client, err := NewBoltClient(tmpDir)
	if err != nil {
		t.Fatalf("NewBoltClient error: %v", err)
	}

	if client == nil {
		t.Fatal("NewBoltClient returned nil")
	}
	if client.db == nil {
		t.Fatal("client.db is nil")
	}

	client.Close()
}

func TestNewBoltClient_InvalidPath(t *testing.T) {
	_, err := NewBoltClient("/invalid/path/that/does/not/exist")
	if err == nil {
		t.Error("NewBoltClient expected error for invalid path")
	}
}

func TestBoltClient_ConcurrentInsert(t *testing.T) {
	tmpDir := t.TempDir()

	client, err := NewBoltClient(tmpDir)
	if err != nil {
		t.Fatalf("NewBoltClient error: %v", err)
	}
	defer client.Close()

	doc := models.VectorDocument{
		Id:         "test-doc-1",
		Content:    "Test content",
		MetaData: models.MetaData{Title: "Test Doc"},
		Embeddings: []float32{0.1, 0.2, 0.3},
	}

	insertErr := client.ConcurrentInsert(doc)
	if insertErr != nil {
		t.Fatalf("ConcurrentInsert error: %v", insertErr)
	}

	// Verify we can read it back
	readDoc, readErr := client.ReadDocument("test-doc-1")
	if readErr != nil {
		t.Fatalf("ReadDocument error: %v", readErr)
	}

	if readDoc.Id != doc.Id {
		t.Errorf("ReadDocument Id = %q, want %q", readDoc.Id, doc.Id)
	}
	if readDoc.Content != doc.Content {
		t.Errorf("ReadDocument Content = %q, want %q", readDoc.Content, doc.Content)
	}
	if readDoc.MetaData.Title != doc.MetaData.Title {
		t.Errorf("ReadDocument MetaData.Title = %q, want %q", readDoc.MetaData.Title, doc.MetaData.Title)
	}
	if len(readDoc.Embeddings) != len(doc.Embeddings) {
		t.Errorf("ReadDocument Embeddings length = %d, want %d", len(readDoc.Embeddings), len(doc.Embeddings))
	}
}

func TestBoltClient_ConcurrentInsert_Multiple(t *testing.T) {
	tmpDir := t.TempDir()

	client, err := NewBoltClient(tmpDir)
	if err != nil {
		t.Fatalf("NewBoltClient error: %v", err)
	}
	defer client.Close()

	docs := []models.VectorDocument{
		{Id: "doc1", Content: "Content 1", MetaData: models.MetaData{Title: "Doc 1"}, Embeddings: []float32{0.1}},
		{Id: "doc2", Content: "Content 2", MetaData: models.MetaData{Title: "Doc 2"}, Embeddings: []float32{0.2}},
		{Id: "doc3", Content: "Content 3", MetaData: models.MetaData{Title: "Doc 3"}, Embeddings: []float32{0.3}},
	}

	for _, doc := range docs {
		insertErr := client.ConcurrentInsert(doc)
		if insertErr != nil {
			t.Fatalf("ConcurrentInsert error for %s: %v", doc.Id, insertErr)
		}
	}

	// Verify all can be read back
	for _, doc := range docs {
		readDoc, readErr := client.ReadDocument(doc.Id)
		if readErr != nil {
			t.Fatalf("ReadDocument error for %s: %v", doc.Id, readErr)
		}
		if readDoc.Content != doc.Content {
			t.Errorf("Content mismatch for %s: got %q, want %q", doc.Id, readDoc.Content, doc.Content)
		}
	}
}

func TestBoltClient_ReadDocument_NotFound(t *testing.T) {
	tmpDir := t.TempDir()

	client, err := NewBoltClient(tmpDir)
	if err != nil {
		t.Fatalf("NewBoltClient error: %v", err)
	}
	defer client.Close()

	_, readErr := client.ReadDocument("nonexistent")
	if readErr == nil {
		t.Error("ReadDocument expected error for nonexistent document")
	}
	if readErr.Error() != "document with Id nonexistent not found" {
		t.Errorf("ReadDocument error = %q, want %q", readErr.Error(), "document with Id nonexistent not found")
	}
}

func TestBoltClient_ReadAllDocuments(t *testing.T) {
	tmpDir := t.TempDir()

	client, err := NewBoltClient(tmpDir)
	if err != nil {
		t.Fatalf("NewBoltClient error: %v", err)
	}
	defer client.Close()

	docs := []models.VectorDocument{
		{Id: "doc1", Content: "Content 1", MetaData: models.MetaData{Title: "Doc 1"}, Embeddings: []float32{0.1}},
		{Id: "doc2", Content: "Content 2", MetaData: models.MetaData{Title: "Doc 2"}, Embeddings: []float32{0.2}},
		{Id: "doc3", Content: "Content 3", MetaData: models.MetaData{Title: "Doc 3"}, Embeddings: []float32{0.3}},
	}

	for _, doc := range docs {
		insertErr := client.ConcurrentInsert(doc)
		if insertErr != nil {
			t.Fatalf("ConcurrentInsert error: %v", insertErr)
		}
	}

	ch := make(chan models.VectorDocument, 3)
	readErr := client.ReadAllDocuments(ch)
	if readErr != nil {
		t.Fatalf("ReadAllDocuments error: %v", readErr)
	}

	readDocs := make([]models.VectorDocument, 0)
	for doc := range ch {
		readDocs = append(readDocs, doc)
	}

	if len(readDocs) != 3 {
		t.Errorf("ReadAllDocuments returned %d docs, want 3", len(readDocs))
	}

	// Verify all docs present
	found := make(map[string]bool)
	for _, doc := range readDocs {
		found[doc.Id] = true
	}
	for _, doc := range docs {
		if !found[doc.Id] {
			t.Errorf("Missing document %s", doc.Id)
		}
	}
}

func TestBoltClient_ReadAllDocuments_Empty(t *testing.T) {
	tmpDir := t.TempDir()

	client, err := NewBoltClient(tmpDir)
	if err != nil {
		t.Fatalf("NewBoltClient error: %v", err)
	}
	defer client.Close()

	ch := make(chan models.VectorDocument, 1)
	readErr := client.ReadAllDocuments(ch)
	if readErr != nil {
		t.Fatalf("ReadAllDocuments error: %v", readErr)
	}

	readDocs := make([]models.VectorDocument, 0)
	for doc := range ch {
		readDocs = append(readDocs, doc)
	}

	if len(readDocs) != 0 {
		t.Errorf("ReadAllDocuments on empty db returned %d docs, want 0", len(readDocs))
	}
}

func TestBoltClient_Close(t *testing.T) {
	tmpDir := t.TempDir()

	client, err := NewBoltClient(tmpDir)
	if err != nil {
		t.Fatalf("NewBoltClient error: %v", err)
	}

	// Close should not panic
	client.Close()

	// Double close should not panic
	client.Close()
}