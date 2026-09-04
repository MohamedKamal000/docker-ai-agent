package rag

import (
	"context"

	models "docker-cli/internal/models"
	rag "docker-cli/internal/rag/storage"
)

type Retriever struct {
	index             *rag.VectorIndex
	embedder          Embedder
	topKWhenSearching int // default to 10, might add an option pattern here for user configuration choice later
}

func buildTheIndex(ctx context.Context, dbClient *rag.BoltClient) (*rag.VectorIndex, error) {
	index, err := rag.NewVectorIndex(ctx)
	if err != nil {
		return nil, err
	}

	ch := make(chan models.VectorDocument, 4)
	go dbClient.ReadAllDocuments(ch)

	err = index.BuildVectorIndex(ctx, ch)
	if err != nil {
		return nil, err
	}

	return index, nil
}

func newRetriever(ctx context.Context, embedder Embedder, dbClient *rag.BoltClient) (*Retriever, error) {
	index, err := buildTheIndex(ctx, dbClient)
	if err != nil {
		return nil, err
	}

	return &Retriever{
		index:             index,
		embedder:          embedder,
		topKWhenSearching: 10,
	}, nil
}

func (r *Retriever) SearchUserReqeust(request string) ([]models.SearchResult, error) {
	ctx := context.Background()
	embededQuery, err := r.embedder.EmbedQuery(ctx, request)
	if err != nil {
		return nil, err
	}

	result, err := r.index.SearchTheIndex(ctx, embededQuery, r.topKWhenSearching)
	if err != nil {
		return nil, err
	}

	return result, nil
}
