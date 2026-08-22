package rag

import (
	"context"
	"log"
	"strconv"
	"time"

	models "docker-cli/internal/rag/models"
	storage "docker-cli/internal/rag/storage"

	"github.com/schollz/progressbar/v3"
	"golang.org/x/sync/errgroup"
)

type WorkerRunner struct {
	embedder Embedder
	dbClient *storage.BoltClient
}

func NewWorkerRunner(emd Embedder, dbClient *storage.BoltClient) *WorkerRunner {
	return &WorkerRunner{
		embedder: emd,
		dbClient: dbClient,
	}
}

func GetEmbeddingBar() *progressbar.ProgressBar {
	return progressbar.NewOptions64(
		-1,
		progressbar.OptionSetDescription("Embedding"),
		progressbar.OptionSetWidth(30),
		progressbar.OptionShowCount(),
		progressbar.OptionThrottle(100*time.Millisecond),
		progressbar.OptionSpinnerType(14),
	)
}

func MakeId(chunk models.Chunk) string {
	return chunk.SourcePath + ":" + strconv.Itoa(int(chunk.Index))
}

func work(ctx context.Context, chunk <-chan models.Chunk, embedder Embedder,
	dbClient *storage.BoltClient, workerNumber int, bar *progressbar.ProgressBar,
) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case b, ok := <-chunk:
			if !ok {
				return nil
			}

			_, lookup := dbClient.ReadDocument(MakeId(b))

			if lookup == nil {

				bar.Add(1)
				continue
			}

			doc, err := embedder.Embed(ctx, b)
			if err != nil {
				log.Printf("worker number %d failed", workerNumber)
				return err
			}

			if err := dbClient.ConcurrentInsert(doc); err != nil {
				return err
			}
			bar.Add(1)
		}
	}
}

func (wr *WorkerRunner) RunWorkers(ctx context.Context, ch <-chan models.Chunk, workersNumber int, batchSize int) error {
	chunksQueue := make(chan models.Chunk, batchSize)

	g, ctx := errgroup.WithContext(ctx)
	b := GetEmbeddingBar()
	for wn := 0; wn < workersNumber; wn++ {
		g.Go(func() error {
			return work(ctx, chunksQueue, wr.embedder, wr.dbClient, wn, b)
		})
	}

	g.Go(func() error {
		defer close(chunksQueue)

		for chunk := range ch {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			select {
			case chunksQueue <- chunk:
			case <-ctx.Done():
				return ctx.Err()
			}

		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return err
	}
	return nil
}
