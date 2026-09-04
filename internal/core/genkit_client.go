package core

import (
	"context"

	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/anthropic"
	"github.com/firebase/genkit/go/plugins/compat_oai/dashscope"
	"github.com/firebase/genkit/go/plugins/compat_oai/deepseek"
	"github.com/firebase/genkit/go/plugins/compat_oai/kimi"
	"github.com/firebase/genkit/go/plugins/compat_oai/openai"
	"github.com/firebase/genkit/go/plugins/compat_oai/xai"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	"github.com/firebase/genkit/go/plugins/ollama"
)

var providerPlugins = map[LLMProvider]func(string) genkit.GenkitOption{
	Gemini: func(apiKey string) genkit.GenkitOption {
		return genkit.WithPlugins(&googlegenai.GoogleAI{
			APIKey: apiKey,
		})
	},

	OpenAi: func(apiKey string) genkit.GenkitOption {
		return genkit.WithPlugins(&openai.OpenAI{
			APIKey: apiKey,
		})
	},

	Anthropic: func(apiKey string) genkit.GenkitOption {
		return genkit.WithPlugins(&anthropic.Anthropic{
			APIKey: apiKey,
		})
	},

	Ollama: func(serverAddress string) genkit.GenkitOption {
		return genkit.WithPlugins(&ollama.Ollama{
			ServerAddress: serverAddress,
			Timeout:       60,
		})
	},

	Deepseek: func(apiKey string) genkit.GenkitOption {
		return genkit.WithPlugins(&deepseek.DeepSeek{
			APIKey: apiKey,
		})
	},

	Kimi: func(apiKey string) genkit.GenkitOption {
		return genkit.WithPlugins(&kimi.Kimi{
			APIKey: apiKey,
		})
	},

	Qwen: func(apiKey string) genkit.GenkitOption {
		return genkit.WithPlugins(&dashscope.DashScope{
			APIKey: apiKey,
		})
	},

	Grok: func(apiKey string) genkit.GenkitOption {
		return genkit.WithPlugins(&xai.XAI{
			APIKey: apiKey,
		})
	},
}

func providerToGenkitPlugin(config AppConfig) genkit.GenkitOption {
	provider := stringToProviderInfo(config.Provider).Provider
	providerFunc, ok := providerPlugins[provider]
	if !ok { // likely not possible since we check this already in the config parsing step
		return nil
	}

	if provider == Ollama {
		return providerFunc(config.ServerAdress)
	}

	return providerFunc(config.ApiKey)
}

type GenkitClient struct {
	G      *genkit.Genkit
	Config AppConfig
}

func NewGenkitClient(config AppConfig) *GenkitClient {
	g := genkit.Init(context.Background(), providerToGenkitPlugin(config),
		genkit.WithDefaultModel(config.ModelName))
	return &GenkitClient{
		G:      g,
		Config: config,
	}
}
