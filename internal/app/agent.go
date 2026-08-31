package app

import (
	"context"
	"docker-cli/internal/core"
	"docker-cli/internal/docker"
	"docker-cli/internal/tools"
	"fmt"

	"github.com/firebase/genkit/go/genkit"
)

type Agent struct {
	AgentLoop      core.AgentLoop
	SessionContext *core.LoopContext
}

var availableTools = map[string]func() core.Tool{
	"docker_command_tool": func() core.Tool {
		return tools.NewDockerCommandsTool(nil)
	},
}

func initalizeRegistery(g *genkit.Genkit, toolRegistry core.ToolRegistry, toolsToRegister []string) error {
	for _, toolName := range toolsToRegister {
		t, ok := availableTools[toolName]
		if !ok {
			return fmt.Errorf("tool Name %s not found", toolName)
		}
		toolRegistry.Register(t(), g)
	}
	return nil
}

func NewAgent(config core.ModelConfig, ctx context.Context, toolsToRegister []string) (*Agent, error) {
	genkitClient := core.NewGenkitClient(config)
	chatSession := core.NewStaticMemoryStore()
	toolRegistry := core.NewGenkitToolRegistry()
	taskRegistry := core.NewTaskRegistry()

	err := initalizeRegistery(genkitClient.G, toolRegistry, toolsToRegister)
	if err != nil {
		return nil, err
	}

	for _, tool := range toolRegistry.List() {
		if dt, ok := tool.(*tools.DockerCommandsTool); ok {
			dt.Tasks = taskRegistry
		}
	}

	sessionContext := &core.LoopContext{
		Memory: chatSession,
		Tools:  toolRegistry,
		Tasks:  taskRegistry,
	}
	err = docker.Init()
	if err != nil {
		return nil, err
	}

	dockerContext, err := docker.GetContext(ctx)
	if err != nil {
		return nil, err
	}

	systemPrompt, err := core.ParsePrompt(core.System_Prompt_Template, dockerContext)

	if err != nil {
		return nil, err
	}

	classifier := core.NewIntentClassifier(*genkitClient)
	agentLoop := core.NewGenkitAgentLoop(*genkitClient, sessionContext, systemPrompt, classifier)

	return &Agent{
		AgentLoop:      agentLoop,
		SessionContext: sessionContext,
	}, nil
}
