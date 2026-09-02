package rag

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

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

const (
	DOWNLOAD_URL        = "https://huggingface.co/BAAI/bge-small-en-v1.5/resolve/main/"
	DOCS_MAIN_REPO      = "https://github.com/docker/docs/archive/refs/heads/main.zip"
	DEPENDENCY_DIR_NAME = ".docker_agent"
)

func GetDependencyPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	savingPath := filepath.Join(home, DEPENDENCY_DIR_NAME)
	return savingPath, nil
}

// we need to compare with the hash that comes from the url with the first computed one
func (dd *DependencyDownloader) ensureModelIsInstalled() bool {
	model := filepath.Join(
		dd.savingPath,
		"local_model",
		"onnx",
		"model.onnx",
	)

	_, err := os.Stat(model)
	return err == nil
}

type DependencyDownloader struct {
	savingPath string
}

func NewDependencyDownloader() (DependencyDownloader, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return DependencyDownloader{}, err
	}
	savingPath := filepath.Join(home, DEPENDENCY_DIR_NAME)
	err = os.MkdirAll(savingPath, 0o755)
	if err != nil {
		return DependencyDownloader{}, err
	}

	return DependencyDownloader{
		savingPath: savingPath,
	}, nil
}

func GetDownloadBar(total int64, downloadable string, current int, totalDownloadableItems int) *progressbar.ProgressBar {
	if total <= 0 {
		total = -1
	}
	return progressbar.NewOptions64(
		total,
		progressbar.OptionSetDescription(
			fmt.Sprintf("[%d/%d] %-20s", current, totalDownloadableItems, downloadable),
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

func (dd *DependencyDownloader) DownloadModel() {
	if dd.ensureModelIsInstalled() {
		log.Println("Model is already installed")
		return
	}

	installPath := filepath.Join(dd.savingPath, "local_model")
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
		bar := GetDownloadBar(total, downloadable, index+1, len(downloadables))

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

func (dd *DependencyDownloader) doesExist() bool {
	docsPath := filepath.Join(dd.savingPath, "content")
	info, err := os.Stat(docsPath)
	return err == nil && info.IsDir() && info.Size() > 0
}

func (dd *DependencyDownloader) FetchDockerDocs() {
	if dd.doesExist() {
		log.Println("docs directory already exist")
		return
	}

	zipPath := filepath.Join(dd.savingPath, "docker_docs.zip")
	client := http.Client{
		Timeout: 2 * time.Minute,
	}
	resp, err := client.Get(DOCS_MAIN_REPO)
	total := resp.ContentLength

	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatal("failed to fetch the docs, no 200 OK")
		return
	}

	if total <= 0 {
		total = resolveDownloadSize(&client, DOCS_MAIN_REPO, -1)
	}

	bar := GetDownloadBar(total, "Docker Docs", 1, 1)

	zipFile, err := os.Create(zipPath)
	if err != nil {
		log.Fatal(err)
	}

	defer zipFile.Close()
	if _, err := io.Copy(io.MultiWriter(bar, zipFile), resp.Body); err != nil {
		log.Fatal(err)
	}

	if err := unzipContent(zipPath, dd.savingPath); err != nil {
		log.Fatal(err)
	}

	os.Remove(zipPath)
}

func unzipContent(zipPath, dst string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	const prefix = "docs-main/content/"

	for _, f := range r.File {
		if !filepath.HasPrefix(f.Name, prefix) {
			continue
		}

		rel := f.Name[len(prefix):]
		target := filepath.Join(dst, "content", rel)

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}

		src, err := f.Open()
		if err != nil {
			return err
		}

		dstFile, err := os.Create(target)
		if err != nil {
			src.Close()
			return err
		}

		_, err = io.Copy(dstFile, src)
		src.Close()
		dstFile.Close()

		if err != nil {
			return err
		}
	}

	return nil
}
