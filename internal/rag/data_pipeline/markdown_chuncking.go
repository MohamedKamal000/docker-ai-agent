package rag

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

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

type MetaData struct {
	Title       string
	Description string
	Keywords    []string
}

type ParsedFile struct {
	FileMetaData MetaData
	Content      string
	FilePath     string
}

type Chunck struct {
	ChunckMetaData MetaData
	ChunckContent  string
	FilePath       string
	Index          uint64
}

func NewChunck(chunckContent string, parsedFile *ParsedFile, index uint64) Chunck {
	return Chunck{
		ChunckContent:  chunckContent,
		Index:          index,
		FilePath:       parsedFile.FilePath,
		ChunckMetaData: parsedFile.FileMetaData,
	}
}

func (c *Chunck) ToString() string {
	return fmt.Sprintf(
		`Title: %s
Description: %s
%s`,
		c.ChunckMetaData.Title,
		c.ChunckMetaData.Description,
		c.ChunckContent,
	)
}

func ParseFile(file string, content string) (error, *ParsedFile) {
	if !strings.HasPrefix(content, "---\n") {
		return fmt.Errorf("Failed to parse file header,file %s might not have a header", file), nil
	}

	splits := strings.SplitN(content, "---\n", 3)
	metadataNonParsed := splits[1]
	fileContent := splits[2]
	var metadata MetaData
	lines := strings.Split(metadataNonParsed, "\n")
	for _, line := range lines {
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

	return nil, &ParsedFile{
		FileMetaData: metadata,
		Content:      fileContent,
		FilePath:     file,
	}
}

type ChunckGenerator struct {
	chunckSize  uint
	overlapSize uint
	folderPath  string
}

func NewChunckGenerator(chunckSize uint, overlapSize uint, folderPath string) ChunckGenerator {
	return ChunckGenerator{
		chunckSize:  chunckSize,
		overlapSize: overlapSize,
		folderPath:  folderPath,
	}
}

func (cg *ChunckGenerator) ProduceChuncks(ch chan<- Chunck) {
	markDownSplitter := textsplitter.NewMarkdownTextSplitter(textsplitter.WithChunkSize(int(cg.chunckSize)),
		textsplitter.WithChunkOverlap(int(cg.overlapSize)))

	err := filepath.WalkDir(cg.folderPath, func(path string, d fs.DirEntry, err error) error {
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
		err, fileContent := ParseFile(path, cleanedContent)
		if err != nil {
			return err
		}
		result, err := markDownSplitter.SplitText(fileContent.Content)
		if err != nil {
			return err
		}
		for index, text := range result {
			ch <- NewChunck(text, fileContent, uint64(index))
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
