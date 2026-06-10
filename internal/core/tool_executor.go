package core

import (
	"context"
	"fmt"

	"github.com/firebase/genkit/go/ai"
)

type ToolExecutor struct {
	registry ToolRegistry
}

func NewToolExecutor(registry ToolRegistry) *ToolExecutor {
	return &ToolExecutor{registry: registry}
}

func (te *ToolExecutor) RaiseWarnsIfExists(modelResponse *ai.ModelResponse) ([]string, bool) {
	warns := make([]string, 0)
	var ok bool
	for _, req := range modelResponse.ToolRequests() {
		tool, ok := te.registry.Get(req.Name)
		if !ok {
			continue
		}

		warn, ok := tool.ShouldRaiseWarning(req.Input)
		if ok {
			warns = append(warns, warn)
		}
	}

	return warns, ok
}

func (te *ToolExecutor) ExecuteGenkitTool(ctx context.Context, modelResponse *ai.ModelResponse) (map[string]string, error) {
	toolsOutput := make(map[string]string)
	for _, req := range modelResponse.ToolRequests() {
		tool, ok := te.registry.Get(req.Name)
		if !ok {
			return nil, fmt.Errorf("tool %s not found", req.Name)
		}

		output, err := tool.Call(ctx, req.Input)

		if err != nil {
			return nil, err
		}

		toolsOutput[tool.Name()] = output
	}

	return toolsOutput, nil
}
