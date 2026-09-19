package rag

import (
	"os"
	"path/filepath"
	"testing"

	"docker-cli/internal/models"
)

func TestCleanFile(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Remove Hugo shortcodes",
			input:    `Some text {{< shortcode >}} more text`,
			expected: `Some text  more text`,
		},
		{
			name:     "Remove multiple Hugo shortcodes",
			input:    `Start {{< code >}} middle {{< highlight >}} end`,
			expected: `Start  middle  end`,
		},
		{
			name:     "Remove empty lines",
			input:    "Line 1\n\n\nLine 2\n   \nLine 3",
			expected: "Line 1\nLine 2\nLine 3",
		},
		{
			name:     "Combined cleaning",
			input:    "Title\n\n{{< note >}}\nContent\n\nEnd",
			expected: "Title\nContent\nEnd",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "No changes needed",
			input:    "Simple text without special chars",
			expected: "Simple text without special chars",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanFile(tt.input)
			if result != tt.expected {
				t.Errorf("cleanFile(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseDocument(t *testing.T) {
	tests := []struct {
		name          string
		file          string
		content       string
		expectError   bool
		expectedMeta  models.MetaData
		expectedContent string
	}{
		{
			name: "Valid document with all metadata",
			file: "test.md",
			content: "---\ntitle: Test Title\ndescription: Test Description\nkeywords: keyword1, keyword2\n---\nDocument content here",
			expectError: false,
			expectedMeta: models.MetaData{
				Title:       "Test Title",
				Description: "Test Description",
				Keywords:    []string{"keyword1", "keyword2"},
			},
			expectedContent: "Document content here",
		},
		{
			name: "Valid document with partial metadata",
			file: "test.md",
			content: "---\ntitle: Only Title\n---\nContent only",
			expectError: false,
			expectedMeta: models.MetaData{
				Title: "Only Title",
			},
			expectedContent: "Content only",
		},
		{
			name: "Missing header",
			file: "test.md",
			content: "No header here",
			expectError: true,
		},
		{
			name: "Incomplete splits",
			file: "test.md",
			content: "---\ntitle: Test\n---",
			expectError: true,
		},
		{
			name: "Empty metadata value",
			file: "test.md",
			content: "---\ntitle: \ndescription: \nkeywords: \n---\nContent",
			expectError: false,
			expectedMeta: models.MetaData{
				Title:       "",
				Description: "",
				Keywords:    []string{""},
			},
			expectedContent: "Content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err, doc := ParseDocument(tt.file, tt.content)
			if tt.expectError {
				if err == nil {
					t.Errorf("ParseDocument expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseDocument unexpected error: %v", err)
			}

			if doc == nil {
				t.Fatal("ParseDocument returned nil document")
			}

			if doc.MetaData.Title != tt.expectedMeta.Title {
				t.Errorf("MetaData.Title = %q, want %q", doc.MetaData.Title, tt.expectedMeta.Title)
			}
			if doc.MetaData.Description != tt.expectedMeta.Description {
				t.Errorf("MetaData.Description = %q, want %q", doc.MetaData.Description, tt.expectedMeta.Description)
			}
			if len(doc.MetaData.Keywords) != len(tt.expectedMeta.Keywords) {
				t.Errorf("MetaData.Keywords length = %d, want %d", len(doc.MetaData.Keywords), len(tt.expectedMeta.Keywords))
			}
			if doc.Content != tt.expectedContent {
				t.Errorf("Content = %q, want %q", doc.Content, tt.expectedContent)
			}
			if doc.Path != tt.file {
				t.Errorf("Path = %q, want %q", doc.Path, tt.file)
			}
		})
	}
}

func TestNewChunkGenerator(t *testing.T) {
	cg := NewChunkGenerator(1000, 100, "/test/path")

	if cg.chunkSize != 1000 {
		t.Errorf("chunkSize = %d, want 1000", cg.chunkSize)
	}
	if cg.overlapSize != 100 {
		t.Errorf("overlapSize = %d, want 100", cg.overlapSize)
	}
	if cg.documentsDirectoryPath != "/test/path" {
		t.Errorf("documentsDirectoryPath = %q, want %q", cg.documentsDirectoryPath, "/test/path")
	}
}

func TestChunkGenerator_ProduceChunks(t *testing.T) {
	// Create a temporary directory with test markdown files
	tmpDir := t.TempDir()

	// Create test markdown files with proper front matter format
	testFiles := map[string]string{
		"doc1.md": "---\ntitle: Document 1\ndescription: First test document\nkeywords: test, doc1\n---\n# Header 1\n\nThis is the first document content.\n\n## Subsection\n\nMore content here.",
		"doc2.md": "---\ntitle: Document 2\ndescription: Second test document\nkeywords: test, doc2\n---\n# Header A\n\nSecond document content.",
		"not-md.txt": "This is not a markdown file",
	}

	for name, content := range testFiles {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}
	}

	cg := NewChunkGenerator(500, 50, tmpDir)
	ch := make(chan models.Chunk, 10)

	cg.ProduceChunks(ch)

	chunks := make([]models.Chunk, 0)
	for chunk := range ch {
		chunks = append(chunks, chunk)
	}

	if len(chunks) == 0 {
		t.Error("ProduceChunks produced no chunks")
	}

	// Check that chunks have correct metadata
	for _, chunk := range chunks {
		if chunk.SourcePath == "" {
			t.Error("Chunk missing SourcePath")
		}
		if chunk.Content == "" {
			t.Error("Chunk missing Content")
		}
		if chunk.Index == 0 && chunk.MetaData.Title == "" {
			t.Error("First chunk should have metadata")
		}
	}
}

func TestChunkGenerator_ProduceChunks_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	cg := NewChunkGenerator(500, 50, tmpDir)
	ch := make(chan models.Chunk, 10)

	cg.ProduceChunks(ch)

	chunks := make([]models.Chunk, 0)
	for chunk := range ch {
		chunks = append(chunks, chunk)
	}

	if len(chunks) != 0 {
		t.Errorf("ProduceChunks on empty dir produced %d chunks, want 0", len(chunks))
	}
}

func TestChunkGenerator_ProduceChunks_NoMdFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create non-markdown files
	testFiles := map[string]string{
		"test.txt": "Text file",
		"data.json": "{}",
	}

	for name, content := range testFiles {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}
	}

	cg := NewChunkGenerator(500, 50, tmpDir)
	ch := make(chan models.Chunk, 10)

	cg.ProduceChunks(ch)

	chunks := make([]models.Chunk, 0)
	for chunk := range ch {
		chunks = append(chunks, chunk)
	}

	if len(chunks) != 0 {
		t.Errorf("ProduceChunks with no .md files produced %d chunks, want 0", len(chunks))
	}
}