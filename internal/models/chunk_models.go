package models

import (
	"fmt"
	"strconv"
)

type MetaData struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Keywords    []string `json:"keywords"`
}

type SourceDocument struct {
	MetaData MetaData
	Content  string
	Path     string
}

type Chunk struct {
	MetaData   MetaData
	Content    string
	SourcePath string
	Index      uint64
}

type SearchResult struct {
	Document   VectorDocument
	Similarity float32
}

func (s *SourceDocument) TransformToChunk(content string, index uint64) Chunk {
	return Chunk{
		Content:    content,
		Index:      index,
		SourcePath: s.Path,
		MetaData:   s.MetaData,
	}
}

func (c *Chunk) String() string {
	return fmt.Sprintf(
		`Title: %s
Description: %s
%s`,
		c.MetaData.Title,
		c.MetaData.Description,
		c.Content,
	)
}

type VectorDocument struct {
	Id         string    `json:"id"`
	Content    string    `json:"content"`
	MetaData   MetaData  `json:"metadata"`
	Embeddings []float32 `json:"embeddings"`
}

func NewVectorDocument(chunk Chunk, data []float32) VectorDocument {
	return VectorDocument{
		Id:         chunk.SourcePath + ":" + strconv.Itoa(int(chunk.Index)),
		MetaData:   chunk.MetaData,
		Content:    chunk.Content,
		Embeddings: data,
	}
}
