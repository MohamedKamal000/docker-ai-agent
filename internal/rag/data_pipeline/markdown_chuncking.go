package rag

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	models "docker-cli/internal/rag/models"

	"github.com/tmc/langchaingo/textsplitter"
)

// title, description, keywords

var (
	hugoRegex = regexp.MustCompile(`\{\{[\s\S]*?\}\}`)
	emptyLine = regexp.MustCompile(`(?m)^\s*$\r?\n`)
)

func cleanFile(content string) string {
	result := hugoRegex.ReplaceAllString(content, "")
	result = emptyLine.ReplaceAllString(result, "")
	return result
}

func ParseDocument(file string, content string) (error, *models.SourceDocument) {
	if !strings.HasPrefix(content, "---\n") {
		return fmt.Errorf("Failed to parse file header,file %s might not have a header", file), nil
	}

	splits := strings.SplitN(content, "---\n", 3)

	if len(splits) != 3 {
		return fmt.Errorf("file has no content %s", file), nil
	}

	metadataNonParsed := splits[1]
	fileContent := splits[2]
	var metadata models.MetaData
	lines := strings.SplitSeq(metadataNonParsed, "\n")
	for line := range lines {
		kv := strings.SplitN(line, ":", 2)
		if len(kv) != 2 {
			continue
		}

		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])
		switch key {
		case "title":
			metadata.Title = value
		case "description":
			metadata.Description = value
		case "keywords":
			metadata.Keywords = strings.Split(value, ", ")
		}
	}

	return nil, &models.SourceDocument{
		MetaData: metadata,
		Content:  fileContent,
		Path:     file,
	}
}

type ChunkGenerator struct {
	chunkSize              uint
	overlapSize            uint
	documentsDirectoryPath string
}

func NewChunkGenerator(chunkSize uint, overlapSize uint, folderPath string) ChunkGenerator {
	return ChunkGenerator{
		chunkSize:              chunkSize,
		overlapSize:            overlapSize,
		documentsDirectoryPath: folderPath,
	}
}

func (cg *ChunkGenerator) ProduceChunks(ch chan<- models.Chunk) {
	defer close(ch)
	markDownSplitter := textsplitter.NewMarkdownTextSplitter(
		textsplitter.WithHeadingHierarchy(true),
		textsplitter.WithKeepSeparator(true),
		textsplitter.WithCodeBlocks(true),
		textsplitter.WithChunkSize(int(cg.chunkSize)),
		textsplitter.WithChunkOverlap(int(cg.overlapSize)))

	err := filepath.WalkDir(cg.documentsDirectoryPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			log.Fatal(err)
		}

		cleanedContent := cleanFile(string(content))
		err, sourceDocument := ParseDocument(path, cleanedContent)
		if err != nil {
			return nil // skip files with no header
		}
		result, err := markDownSplitter.SplitText(sourceDocument.Content)
		if err != nil {
			return err
		}
		for index, text := range result {
			chunk := sourceDocument.TransformToChunk(text, uint64(index))
			ch <- chunk
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
