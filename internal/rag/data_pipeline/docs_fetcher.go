package rag

import (
	"archive/zip"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const DOCS_DIR_NAME = "docker_agent_rag_docs"

func GetDocsFolderPath() string {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Fatal(err)
	}

	return filepath.Join(cacheDir, DOCS_DIR_NAME, "content")
}

func doesExist() bool {
	info, err := os.Stat(GetDocsFolderPath())
	return err == nil && info.IsDir() && info.Size() > 0
}

func FetchDockerDocs() {
	if doesExist() {
		log.Println("docs directory already exist")
		return
	}
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Fatal(err)
	}

	installPath := filepath.Join(cacheDir, DOCS_DIR_NAME)
	zipPath := filepath.Join(cacheDir, "docker_docs.zip")

	if err := os.Mkdir(installPath, 0o755); err != nil {
		log.Fatal(err)
	}

	resp, err := http.Get("https://github.com/docker/docs/archive/refs/heads/main.zip")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	zipFile, err := os.Create(zipPath)
	defer zipFile.Close()
	if err != nil {
		log.Fatal(err)
	}
	if _, err := io.Copy(zipFile, resp.Body); err != nil {
		log.Fatal(err)
	}

	if err := unzipContent(zipPath, installPath); err != nil {
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
