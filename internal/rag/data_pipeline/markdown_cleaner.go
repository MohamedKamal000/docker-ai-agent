package rag

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func cleanDocs(path string) error {
	return filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(d.Name(), ".md") {
			bytes, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			content := string(bytes)
			result := cleanFile(content)

			err = os.WriteFile(path, []byte(result), 0o644)
			if err != nil {
				return err
			}
		}

		return nil
	})
}
