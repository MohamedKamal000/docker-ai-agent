package rag

import (
	"context"
	"os"
	"path/filepath"

	models "docker-cli/internal/rag/models"

	pipeLine "docker-cli/internal/rag/data_pipeline"
	embedder "docker-cli/internal/rag/embeddings"

	storage "docker-cli/internal/rag/storage"
)

/*
this is the entry point of the rag pipeline
it has two calls specifically to be used
1- initalizeRag => it does initalize the required data and embeddings before starting the agent
exact steps are: docs fetching => cleaning and chuncking => embedding => storing to an embedded database

this process can be stopped and resumed as the user wants
specifically for the docs fetching and embedding, unless the user
changed the embedding configurations

embedding configurations:
1- you can choose a local model that will get downloaded (default to: bge-small-en-v1.5)
we use go hugot which has a default cpu inference that does not need any dependencies to setup before working
but it may take a lot of time based on how good your cpu and memory is
we might need to extend it to work on gpu for this work later in the future
2- you can choose a remote model via an api key if you have tokens to do it will be much faster and easier
3- you can choose chunk size, overlap size, model configuration that is used (either local or remote and its name)
4- choosing which local model to download from hugging face
5- number of workers when doing the embedding, you should try a reasonable number to not heavily consume your resources

NOTE:
make sure to choose a chunk size that will work with the number of tokens the model expect
for example, while embedding with bge-small-en-v1 with 1000 chunk size, it stopped in the middle
since it does not expect more than 512 token

*/

// TODO: extend the app config and pass it here, include the docs and model fetching steps
func InitalizeRag() error {
	ch := make(chan models.Chunk, 64)
	generator := pipeLine.NewChunkGenerator(256, 50, pipeLine.GetDocsFolderPath())
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return err
	}

	path := filepath.Join(
		cacheDir,
		pipeLine.DOCS_DIR_NAME,
	)

	ctx := context.Background()
	client, err := storage.NewBoltClient(path)
	if err != nil {
		return err
	}

	embedder, err := embedder.NewLocalEmbedder(ctx, filepath.Join(path, "local_model"))
	if err != nil {
		return err
	}

	defer embedder.Close()
	runner := NewWorkerRunner(embedder, client)

	go generator.ProduceChunks(ch)
	const batchSize = 4
	const numberOfWorkers = 4

	err = runner.RunWorkers(ctx, ch, numberOfWorkers, batchSize)
	if err != nil {
		return err
	}
	return nil
}

// TODO: extend the app config and pass it here, do nessesry checks before calling actual function
func NewRetriever() (*Retriever, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(
		cacheDir,
		pipeLine.DOCS_DIR_NAME,
	)

	ctx := context.Background()
	client, err := storage.NewBoltClient(path)
	if err != nil {
		return nil, err
	}

	embedder, err := embedder.NewLocalEmbedder(ctx, filepath.Join(path, "local_model"))

	return newRetriever(ctx, embedder, client)
}
