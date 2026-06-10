package app

import (
	"context"
	"docker-cli/internal/core"
	"docker-cli/internal/docker"
	"docker-cli/internal/tools"
	"fmt"
	"log"

	"github.com/firebase/genkit/go/genkit"
)

type Agent struct {
	AgentLoop      core.AgentLoop
	SessionContext *core.LoopContext
}

var availableTools = map[string]func() core.Tool{
	"docker_command_tool": func() core.Tool {
		return tools.NewDockerCommandsTool()
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

func NewAgent(config core.ModelConfig, ctx context.Context, toolsToRegister []string) *Agent {
	genkitClient := core.NewGenkitClient(config)
	chatSession := core.NewStaticMemoryStore()
	toolRegistry := core.NewGenkitToolRegistry()
	err := initalizeRegistery(genkitClient.G, toolRegistry, toolsToRegister)
	if err != nil {
		log.Fatal(err)
	}
	sessionContext := &core.LoopContext{
		Memory: chatSession,
		Tools:  toolRegistry}
	err = docker.Init()
	if err != nil {
		log.Fatal(err)
	}

	dockerContext, err := docker.GetContext(ctx)
	if err != nil {
		log.Fatal(err)
	}

	systemPrompt, err := core.ParsePrompt(core.System_Prompt_Template, dockerContext)

	if err != nil {
		log.Fatal(err)
	}

	agentLoop := core.NewGenkitAgentLoop(*genkitClient, sessionContext, systemPrompt)

	return &Agent{
		AgentLoop:      agentLoop,
		SessionContext: sessionContext,
	}
}
