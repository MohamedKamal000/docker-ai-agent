package app

import (
	"context"
	"docker-cli/internal/core"
	"docker-cli/internal/docker"
	"log"
)

type Agent struct {
	AgentLoop   core.AgentLoop
	ChatSession core.MemoryStore
	// tool regestry will be added here as well
}

func NewAgent(config core.ModelConfig, ctx context.Context) *Agent {

	genkitClient := core.NewGenkitClient(config)

	chatSession := core.NewStaticMemoryStore()

	err := docker.Init()
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

	agentLoop := &core.GenkitAgentLoop{
		Client:       *genkitClient,
		SystemPrompt: systemPrompt,
	}

	return &Agent{
		AgentLoop:   agentLoop,
		ChatSession: chatSession,
	}
}
