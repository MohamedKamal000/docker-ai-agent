package core

import (
	"context"
	"log"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

type Tool interface {
	Name() string

	Description() string

	Call(ctx context.Context, input any) (string, error) // tools must return a structured output to use with the llm immediately

	GetInputSchema() map[string]any

	ShouldRaiseWarning(input any) (string, bool)
}

type ToolRegistry interface {
	Register(t Tool, g *genkit.Genkit)
	Get(name string) (Tool, bool)
	List() []Tool
}

type GenkitToolRegistry struct {
	registeredTools map[string]Tool
}

func NewGenkitToolRegistry() *GenkitToolRegistry {
	return &GenkitToolRegistry{registeredTools: make(map[string]Tool)}
}

func (gtr GenkitToolRegistry) Register(t Tool, g *genkit.Genkit) {
	genkit.DefineTool(g, t.Name(), t.Description(), func(ctx *ai.ToolContext, input any) (any, error) {
		result, err := t.Call(ctx, input)
		if err != nil {
			log.Fatal(err)
		}
		return result, nil
	}, ai.WithInputSchema(t.GetInputSchema()))
	gtr.registeredTools[t.Name()] = t
}

func (gtr GenkitToolRegistry) Get(name string) (Tool, bool) {
	tool, ok := gtr.registeredTools[name]
	return tool, ok
}

func (gtr GenkitToolRegistry) List() []Tool {
	result := make([]Tool, 0)
	for _, v := range gtr.registeredTools {
		result = append(result, v)
	}
	return result
}
