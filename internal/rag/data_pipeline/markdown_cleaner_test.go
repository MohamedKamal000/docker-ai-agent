package rag

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCleanDocs(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test markdown files
	testFiles := map[string]string{
		"doc1.md": `---\ntitle: Test\n---\nContent {{< shortcode >}} here`,
		"doc2.md": `---\ntitle: Test2\n---\nMore content {{< note >}} there`,
		"not-md.txt": "Should not be cleaned",
	}

	for name, content := range testFiles {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}
	}

	err := cleanDocs(tmpDir)
	if err != nil {
		t.Fatalf("cleanDocs error: %v", err)
	}

	// Check that markdown files were cleaned
	for name, originalContent := range testFiles {
		path := filepath.Join(tmpDir, name)
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}

		if name == "doc1.md" {
			if containsString(string(content), "shortcode") {
				t.Errorf("doc1.md should have shortcode removed: %s", string(content))
			}
			if !containsString(string(content), "Content  here") {
				t.Errorf("doc1.md should have cleaned content: %s", string(content))
			}
		}
		if name == "doc2.md" {
			if containsString(string(content), "note") {
				t.Errorf("doc2.md should have note removed: %s", string(content))
			}
			if !containsString(string(content), "More content  there") {
				t.Errorf("doc2.md should have cleaned content: %s", string(content))
			}
		}
		if name == "not-md.txt" {
			if string(content) != originalContent {
				t.Errorf("not-md.txt should not be modified: %s", string(content))
			}
		}
	}
}

func TestCleanDocs_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	err := cleanDocs(tmpDir)
	if err != nil {
		t.Fatalf("cleanDocs on empty dir error: %v", err)
	}
}

func TestCleanDocs_NoMdFiles(t *testing.T) {
	tmpDir := t.TempDir()

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

	err := cleanDocs(tmpDir)
	if err != nil {
		t.Fatalf("cleanDocs error: %v", err)
	}

	// Verify files unchanged
	for name, expected := range testFiles {
		path := filepath.Join(tmpDir, name)
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}
		if string(content) != expected {
			t.Errorf("%s was modified unexpectedly: %s", name, string(content))
		}
	}
}

func TestCleanDocs_FileReadError(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a markdown file
	path := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(path, []byte("---\ntitle: Test\n---\nContent"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Remove read permissions to cause read error
	if err := os.Chmod(path, 0000); err != nil {
		t.Fatalf("Failed to chmod: %v", err)
	}
	defer os.Chmod(path, 0644) // restore for cleanup

	err := cleanDocs(tmpDir)
	if err == nil {
		t.Error("cleanDocs expected error for unreadable file")
	}
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}