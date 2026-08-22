package rag

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	rag "docker-cli/internal/rag/data_pipeline"

	"github.com/schollz/progressbar/v3"
)

var downloadables []string = []string{
	"config.json",
	"tokenizer.json",
	"tokenizer_config.json",
	"special_tokens_map.json",
	"vocab.txt",
	"model.safetensors",
	"onnx/model.onnx",
}

var downloadablesMapFileSizes map[string]int64 = map[string]int64{
	"config.json":             743,
	"tokenizer.json":          711396,
	"tokenizer_config.json":   366,
	"special_tokens_map.json": 125,
	"vocab.txt":               231508,
	"model.safetensors":       133466304,
	"onnx/model.onnx":         133093490,
}

var DOWNLOAD_URL string = "https://huggingface.co/BAAI/bge-small-en-v1.5/resolve/main/"

// we need to compare with the hash that comes from the url with the first computed one
func ensureModelIsInstalled() bool {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Fatal(err)
	}

	model := filepath.Join(
		cacheDir,
		rag.DOCS_DIR_NAME,
		"local_model",
		"onnx",
		"model.onnx",
	)

	_, err = os.Stat(model)
	return err == nil
}

func GetDownloadBar(total int64, downloadable string, current int) *progressbar.ProgressBar {
	if total <= 0 {
		total = -1
	}
	return progressbar.NewOptions64(
		total,
		progressbar.OptionSetDescription(
			fmt.Sprintf("[%d/%d] %-20s", current, len(downloadables), downloadable),
		),
		progressbar.OptionSetWidth(30),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionThrottle(100*time.Millisecond),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerHead:    "█",
			SaucerPadding: "░",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)
}

func resolveDownloadSize(client *http.Client, url string, fallback int64) int64 {
	resp, err := client.Head(url)
	if err == nil {
		defer resp.Body.Close()
		if resp.ContentLength > 0 {
			return resp.ContentLength
		}
	}
	return fallback
}

func createParentDirectoriesIfPossible(itemPath string) {
	dirPath := filepath.Dir(itemPath)

	err := os.MkdirAll(dirPath, 0o755)
	if err != nil {
		log.Fatalf("Failed to create directories: %v", err)
	}
}

func DownloadModel() {
	if ensureModelIsInstalled() {
		return
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Fatal(err)
	}

	installPath := filepath.Join(cacheDir, rag.DOCS_DIR_NAME, "local_model")
	os.MkdirAll(installPath, 0o755)
	client := http.Client{
		Timeout: 10 * time.Minute,
	}
	for index, downloadable := range downloadables {
		itemName := filepath.Join(installPath, downloadable)
		createParentDirectoriesIfPossible(itemName)

		resp, err := client.Get(DOWNLOAD_URL + downloadable)
		if err != nil {
			log.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			log.Fatal(fmt.Errorf("Failed to download %s", downloadable))
		}

		file, err := os.Create(itemName + ".tmp")
		if err != nil {
			resp.Body.Close()
			log.Fatal(err)
		}

		total := resp.ContentLength
		if total <= 0 {
			total = resolveDownloadSize(&client, DOWNLOAD_URL+downloadable, downloadablesMapFileSizes[downloadable])
		}
		bar := GetDownloadBar(total, downloadable, index+1)

		_, err = io.Copy(io.MultiWriter(bar, file), resp.Body)
		if err != nil {
			bar.Finish()
			file.Close()
			resp.Body.Close()
			os.Remove(itemName + ".tmp")
			log.Fatal(err)
		}
		bar.Finish()

		if err := file.Close(); err != nil {
			resp.Body.Close()
			os.Remove(itemName + ".tmp")
			log.Fatal(err)
		}

		if err := resp.Body.Close(); err != nil {
			os.Remove(itemName + ".tmp")
			log.Fatal(err)
		}

		if err := os.Rename(itemName+".tmp", itemName); err != nil {
			os.Remove(itemName + ".tmp")
			log.Fatal(err)
		}
	}
}
