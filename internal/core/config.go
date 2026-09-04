package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

type EmbeddingType string

const (
	EmbeddingLocal  EmbeddingType = "local"
	EmbeddingRemote EmbeddingType = "remote"
)

type EmbeddingDevice string

const (
	CPU EmbeddingDevice = "cpu"
	GPU EmbeddingDevice = "gpu"
)

type LLMProvider int

const (
	Gemini LLMProvider = iota

	OpenAi

	Anthropic

	Ollama

	Deepseek

	Kimi

	Qwen

	Grok

	Unknown
)

var ProviderMap map[string]ProviderInfo = map[string]ProviderInfo{
	"Gemini": {
		Provider:   Gemini,
		EnvName:    "GEMINI_API_KEY",
		PrefixName: "googleai",
	},
	"OpenAi": {
		Provider:   OpenAi,
		EnvName:    "OPENAI_API_KEY",
		PrefixName: "openai",
	},
	"Anthropic": {
		Provider:   Anthropic,
		EnvName:    "ANTHROPIC_API_KEY",
		PrefixName: "anthropic",
	},
	"Ollama": {
		Provider:   Ollama,
		EnvName:    "SERVER_ADDRESS",
		PrefixName: "",
	},
	"DeepSeek": {
		Provider:   Deepseek,
		EnvName:    "DEEPSEEK_API_KEY",
		PrefixName: "deepseek",
	},
	"Kimi": {
		Provider:   Kimi,
		EnvName:    "KIMI_API_KEY",
		PrefixName: "moonshotai",
	},
	"Qwen": {
		Provider:   Qwen,
		EnvName:    "QWEN_API_KEY",
		PrefixName: "alibaba",
	},
	"Grok": {
		Provider:   Grok,
		EnvName:    "GROK_API_KEY",
		PrefixName: "xai",
	},
}

// Api key will be => EMBEDDING_API_KEY
type RAGConfig struct {
	EmbeddingType   EmbeddingType   `json:"embedding-type"` // either local or remote
	InferenceType   EmbeddingDevice `json:"inference-type"` // either cpu or gpu (gpu needs cuda run time)
	ChunkSize       int             `json:"chunk-size"`
	OverlapSize     int             `json:"overlap-size"`
	WorkersNumber   int             `json:"workers-number"` // won't really make a deal when running on gpu
	ModelName       string          `json:"model-name"`
	EmbeddingApiKey string          `json:"embedding-api-key"` // defaults to gemini always, genkit lacks different providers for embeddings
}

type AppConfig struct {
	Provider      string     `json:"provider"`
	ModelName     string     `json:"model-name"`
	Temperature   float32    `json:"temperature,omitempty"`
	MaxTokens     uint32     `json:"max-tokens,omitempty"`
	ApiKey        string     `json:"api-key"`
	ServerAdress  string     `json:"server-address,omitempty"`
	MaxIterations int        `json:"max-iterations"`
	RagConfig     *RAGConfig `json:"rag,omitempty"`
}

type ProviderInfo struct {
	Provider   LLMProvider
	EnvName    string
	PrefixName string
}

func stringToProviderInfo(provider string) ProviderInfo {
	prov, ok := ProviderMap[provider]
	if !ok {
		return ProviderInfo{
			Provider: Unknown,
		}
	}
	return prov
}

func checkFileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !errors.Is(err, os.ErrNotExist)
}

func fallBackToCurrentDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return strings.Join([]string{cwd, "/config.json"}, ""), nil
}

func ModelConfigFromJsonFile(filepath string) (AppConfig, error) {
	if !checkFileExists(filepath) {
		currFilePath, err := fallBackToCurrentDir()
		if err != nil {
			return AppConfig{}, fmt.Errorf("config file not found")
		}
		filepath = currFilePath
	}
	fileBytes, err := os.ReadFile(filepath)
	if err != nil {
		return AppConfig{}, err
	}

	var result AppConfig
	err = json.Unmarshal(fileBytes, &result)
	if err != nil {
		return AppConfig{}, err
	}

	if result.ModelName == "" {
		return AppConfig{}, fmt.Errorf("model name can't be empty")
	}
	proivder := stringToProviderInfo(result.Provider)

	if proivder.Provider != Ollama && !strings.HasPrefix(result.ModelName, proivder.PrefixName) {
		result.ModelName = proivder.PrefixName + "/" + result.ModelName
	}

	if proivder.Provider == Unknown {
		return AppConfig{}, fmt.Errorf("provider %s is not supported", result.Provider)
	}

	if result.MaxIterations < 0 {
		return AppConfig{}, fmt.Errorf("max iterations can't be negative")
	}

	if result.ApiKey == "" && proivder.Provider != Ollama {
		apiKey := os.Getenv(proivder.EnvName)
		if apiKey == "" {
			return AppConfig{}, fmt.Errorf("api key not found")
		}
		result.ApiKey = apiKey
	}

	if result.RagConfig != nil {
		err = result.RagConfig.Validate()
		if err != nil {
			return AppConfig{}, err
		}
	}

	return result, nil
}

func (r *RAGConfig) Validate() error {
	switch r.EmbeddingType {
	case EmbeddingLocal:
		switch r.InferenceType {
		case CPU, GPU:
		default:
			return fmt.Errorf("invalid inference type %q", r.InferenceType)
		}

		if r.EmbeddingApiKey != "" {
			return fmt.Errorf("embedding api key cannot be set when using local embeddings")
		}

	case EmbeddingRemote:

		if r.InferenceType != "" {
			return fmt.Errorf("inference type is only valid for local embeddings")
		}

		if r.EmbeddingApiKey == "" {
			apiKey := os.Getenv("EMBEDDING_API_KEY")
			if apiKey == "" {
				return fmt.Errorf("embedding api key not found")
			}
			r.EmbeddingApiKey = apiKey
		}

	default:
		return fmt.Errorf("invalid embedding type %q", r.EmbeddingType)
	}

	if r.ChunkSize <= 0 {
		return fmt.Errorf("chunk size must be greater than 0")
	}

	if r.OverlapSize < 0 {
		return fmt.Errorf("overlap size cannot be negative")
	}

	if r.OverlapSize >= r.ChunkSize {
		return fmt.Errorf("overlap size must be smaller than chunk size")
	}

	if r.WorkersNumber <= 0 {
		return fmt.Errorf("workers number must be greater than 0")
	}

	return nil
}
