package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
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

type ModelConfig struct {
	Provider      string  `json:"provider"`
	ModelName     string  `json:"model-name"`
	Temperature   float32 `json:"temperature,omitempty"`
	MaxTokens     uint32  `json:"max-tokens,omitempty"`
	ApiKey        string  `json:"api-key"`
	ServerAdress  string  `json:"server-address,omitempty"`
	MaxIterations int     `json:"max-iterations"`
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

func ModelConfigFromJsonFile(filepath string) (ModelConfig, error) {
	if !checkFileExists(filepath) {
		currFilePath, err := fallBackToCurrentDir()
		if err != nil {
			return ModelConfig{}, fmt.Errorf("config file not found")
		}
		filepath = currFilePath
	}
	fileBytes, err := os.ReadFile(filepath)
	if err != nil {
		return ModelConfig{}, err
	}

	var result ModelConfig
	err = json.Unmarshal(fileBytes, &result)
	if err != nil {
		return ModelConfig{}, err
	}

	if result.ModelName == "" {
		return ModelConfig{}, fmt.Errorf("model name can't be empty")
	}
	proivder := stringToProviderInfo(result.Provider)

	if proivder.Provider != Ollama && !strings.HasPrefix(result.ModelName, proivder.PrefixName) {
		result.ModelName = proivder.PrefixName + "/" + result.ModelName
	}

	if proivder.Provider == Unknown {
		return ModelConfig{}, fmt.Errorf("provider %s is not supported", result.Provider)
	}

	if result.MaxIterations < 0 {
		return ModelConfig{}, fmt.Errorf("max iterations can't be negative")
	}

	if result.ApiKey == "" && proivder.Provider != Ollama {
		apiKey := os.Getenv(proivder.EnvName)
		if apiKey == "" {
			return ModelConfig{}, fmt.Errorf("api key not found")
		}
		result.ApiKey = apiKey
	}

	return result, nil
}
