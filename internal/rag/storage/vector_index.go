package rag

import (
	"context"
	"strings"

	"github.com/philippgille/chromem-go"

	models "docker-cli/internal/models"
)

/*
In-Memory chromem database for reading embedded data from disk
and doing search on them while the app is working
*/

const COLLECTION_NAME string = "docker_vector_index"

type VectorIndex struct {
	collection *chromem.Collection
}

func NewVectorIndex(ctx context.Context) (*VectorIndex, error) {
	db := chromem.NewDB()

	collection, err := db.CreateCollection(COLLECTION_NAME, nil, nil)
	if err != nil {
		return nil, err
	}

	return &VectorIndex{
		collection: collection,
	}, nil
}

func ToChromemDocument(doc models.VectorDocument) chromem.Document {
	return chromem.Document{
		ID:        doc.Id,
		Embedding: doc.Embeddings,
		Content:   doc.Content,
		Metadata: map[string]string{
			"Title":       doc.MetaData.Title,
			"Description": doc.MetaData.Description,
			"Keywords":    strings.Join(doc.MetaData.Keywords, ","),
		},
	}
}

func ToSearchResult(doc chromem.Result) models.SearchResult {
	return models.SearchResult{
		Document: models.VectorDocument{
			Id:         doc.ID,
			Embeddings: doc.Embedding,
			Content:    doc.Content,
			MetaData: models.MetaData{
				Title:       doc.Metadata["Title"],
				Description: doc.Metadata["Description"],
				Keywords:    strings.Split(doc.Metadata["Keywords"], ","),
			},
		},
		Similarity: doc.Similarity,
	}
}

func (vi *VectorIndex) BuildVectorIndex(ctx context.Context, ch <-chan models.VectorDocument) error {
	for document := range ch {
		err := vi.collection.AddDocument(ctx, ToChromemDocument(document))
		if err != nil {
			return err
		}
	}
	return nil
}

func (vi *VectorIndex) SearchTheIndex(ctx context.Context, userQuery []float32, topK int) ([]models.SearchResult, error) {
	docsResult, err := vi.collection.QueryEmbedding(ctx, userQuery, topK, nil, nil)
	if err != nil {
		return nil, err
	}

	result := make([]models.SearchResult, 0, topK)
	for _, doc := range docsResult {
		result = append(result, ToSearchResult(doc))
	}
	return result, nil
}
