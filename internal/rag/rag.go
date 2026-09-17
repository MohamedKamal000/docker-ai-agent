package rag

import (
	"context"
	"fmt"
	"path/filepath"

	"docker-cli/internal/core"
	models "docker-cli/internal/models"

	pipeLine "docker-cli/internal/rag/data_pipeline"
	embedder "docker-cli/internal/rag/embeddings"

	storage "docker-cli/internal/rag/storage"

	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
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
we might need to extend it to work on gpu for this work later in the future (gpu is done)
2- you can choose a remote model via an api key if you have tokens to do it will be much faster and easier
3- you can choose chunk size, overlap size, model configuration that is used (either local or remote and its name)
4- choosing which local model to download from hugging face
5- number of workers when doing the embedding, you should try a reasonable number to not heavily consume your resources



NOTE:
make sure to choose a chunk size that will work with the number of tokens the model expect
for example, while embedding with bge-small-en-v1 with 1000 chunk size, it stopped in the middle
since it does not expect more than 512 token

*/

func downloadDependency() (string, error) {
	dd, err := NewDependencyDownloader()
	if err != nil {
		return "", err
	}

	dd.FetchDockerDocs()
	fmt.Println("FINISHED DOCS DOWNLOADING")

	dd.DownloadModel()

	fmt.Println("FINISHED MODEL DOWNLOADING")

	return GetDependencyPath()
}

func getCorrectEmbedder(ctx context.Context, rc *core.RAGConfig, path string) (Embedder, error) {
	if rc.EmbeddingType == core.EmbeddingLocal {
		return embedder.NewLocalEmbedder(ctx, filepath.Join(path, "local_model"), rc.InferenceType == core.GPU)
	} else {
		return embedder.NewGenkitEmbedder(ctx,
			genkit.WithPlugins(&googlegenai.GoogleAI{APIKey: rc.EmbeddingApiKey}),
			rc.ModelName), nil
	}
}

func InitalizeRag(ctx context.Context, config core.AppConfig) error {
	rc := config.RagConfig
	path, err := downloadDependency()
	if err != nil {
		return err
	}

	ch := make(chan models.Chunk, 64)
	generator := pipeLine.NewChunkGenerator(uint(rc.ChunkSize), uint(rc.OverlapSize), filepath.Join(path, "content"))

	client, err := storage.NewBoltClient(path)
	if err != nil {
		return err
	}

	em, err := getCorrectEmbedder(ctx, rc, path)
	if err != nil {
		return err
	}

	defer em.Close()
	runner := NewWorkerRunner(em, client)

	go generator.ProduceChunks(ch)

	err = runner.RunWorkers(ctx, ch, rc.WorkersNumber, rc.WorkersNumber*2)
	if err != nil {
		return err
	}
	return nil
}

// TODO: find a way to check if the rag is initalized in the first place
func NewRetriever(ctx context.Context, config core.AppConfig) (*Retriever, error) {
	path, err := GetDependencyPath()
	if err != nil {
		return nil, err
	}

	client, err := storage.NewBoltClient(path)
	if err != nil {
		return nil, err
	}

	em, err := getCorrectEmbedder(ctx, config.RagConfig, path)

	return newRetriever(ctx, em, client)
}
