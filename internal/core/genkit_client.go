package core

import (
	"context"

	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/anthropic"
	"github.com/firebase/genkit/go/plugins/compat_oai/openai"
	"github.com/firebase/genkit/go/plugins/googlegenai"
)

func providerToGenkitPlugin(provider LLMProvider, apiKey string) genkit.GenkitOption {
	switch provider {
	case Gemini:
		return genkit.WithPlugins(&googlegenai.GoogleAI{APIKey: apiKey})
	case OpenAi:
		return genkit.WithPlugins(&openai.OpenAI{APIKey: apiKey})
	case Anthropic:
		return genkit.WithPlugins(&anthropic.Anthropic{APIKey: apiKey})
	}

	return nil
}

type GenkitClient struct {
	G      *genkit.Genkit
	Config ModelConfig
}

func NewGenkitClient(config ModelConfig) *GenkitClient {
	g := genkit.Init(context.Background(), providerToGenkitPlugin(config.Provider, config.ApiKey),
		genkit.WithDefaultModel(config.ModelName))

	return &GenkitClient{
		G:      g,
		Config: config,
	}
}
