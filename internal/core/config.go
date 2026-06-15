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

	Unknown
)

type ModelConfig struct {
	Provider    string  `json:"provider"`
	ModelName   string  `json:"model-name"`
	Temperature float32 `json:"temperature,omitempty"`
	MaxTokens   uint32  `json:"max-tokens,omitempty"`
	ApiKey      string  `json:"api-key"`
}

func stringToProvider(provider string) LLMProvider {
	switch provider {
	case "Gemini":
		return Gemini
	case "OpenAi":
		return OpenAi
	case "Anthropic":
		return Anthropic
	default:
		return Unknown
	}
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
	if stringToProvider(result.Provider) == Unknown {
		return ModelConfig{}, fmt.Errorf("provider %s is not supported", result.Provider)
	}

	if result.ApiKey == "" {
		var apiKey string
		switch stringToProvider(result.Provider) {
		case Gemini:
			apiKey = os.Getenv("GEMINI_API_KEY")
		case Anthropic:
			apiKey = os.Getenv("ANTHROPIC_API_KEY")
		case OpenAi:
			apiKey = os.Getenv("OPENAI_API_KEY")
		}
		if apiKey == "" {
			return ModelConfig{}, fmt.Errorf("api key not found")
		}
		result.ApiKey = apiKey
	}

	return result, nil
}
