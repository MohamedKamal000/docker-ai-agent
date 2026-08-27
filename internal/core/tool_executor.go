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

func (te *ToolExecutor) ExecuteGenkitTool(ctx context.Context, modelResponse *ai.ModelResponse, comm *AgentCommunication) (map[string]string, error) {
	toolsOutput := make(map[string]string)
	for _, req := range modelResponse.ToolRequests() {
		tool, ok := te.registry.Get(req.ToolRequest.Name)
		if !ok {
			return nil, fmt.Errorf("tool %s not found", req.ToolRequest.Name)
		}
		warn, ok := tool.ShouldRaiseWarning(req.ToolRequest.Input)
		if ok {
			comm.ToUser <- NewWarning(warn)
			var cmd UserCommand
			select {
			case cmd = <-comm.FromUser:
			case <-ctx.Done():
				return nil, ctx.Err()
			}

			if !cmd.ShouldContinue {
				return nil, fmt.Errorf("execution is canceled")
			}
		}
		output, err := tool.Call(ctx, req.ToolRequest.Input)
		if err != nil {
			return nil, err
		}

		toolsOutput[tool.Name()] = output
	}

	return toolsOutput, nil
}
