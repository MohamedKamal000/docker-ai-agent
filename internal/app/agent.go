package app

import (
	"context"
	"fmt"

	"docker-cli/internal/core"
	"docker-cli/internal/docker"
	"docker-cli/internal/rag"
	"docker-cli/internal/tools"

	"github.com/firebase/genkit/go/genkit"
)

type Agent struct {
	AgentLoop      core.AgentLoop
	SessionContext *core.LoopContext
	ModelName      string
}

var availableTools = map[string]func(*core.TaskRegistry) core.Tool{
	"docker_command_tool": func(tasks *core.TaskRegistry) core.Tool {
		return tools.NewDockerCommandsTool(tasks)
	},
	"task_status_tool": func(tasks *core.TaskRegistry) core.Tool {
		return tools.NewTaskStatusTool(tasks)
	},
}

func initalizeRegistery(g *genkit.Genkit, toolRegistry core.ToolRegistry, toolsToRegister []string, taskRegistry *core.TaskRegistry) error {
	for _, toolName := range toolsToRegister {
		t, ok := availableTools[toolName]
		if !ok {
			return fmt.Errorf("tool Name %s not found", toolName)
		}
		toolRegistry.Register(t(taskRegistry), g)
	}
	return nil
}

type AgentOptions func(*Agent)

func WithRagOption(retreiver *rag.Retriever) AgentOptions {
	return func(agent *Agent) {
		agent.SessionContext.Search = retreiver.SearchUserReqeust
	}
}

func NewAgent(config core.AppConfig, ctx context.Context, toolsToRegister []string, options ...AgentOptions) (*Agent, error) {
	genkitClient := core.NewGenkitClient(config)
	chatSession := core.NewStaticMemoryStore()
	toolRegistry := core.NewGenkitToolRegistry()
	taskRegistry := core.NewTaskRegistry()

	err := initalizeRegistery(genkitClient.G, toolRegistry, toolsToRegister, taskRegistry)
	if err != nil {
		return nil, err
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
	evaluator := core.NewGoalEvaluator(*genkitClient, toolRegistry, core.Evaluator_System_Prompt)
	queryFlow := core.NewDockerQueryFlow(*genkitClient, toolRegistry, core.DockerQuery_System_Prompt)
	agentLoop := core.NewGenkitAgentLoop(*genkitClient, sessionContext, systemPrompt, classifier, evaluator, queryFlow)

	agent := Agent{
		AgentLoop:      agentLoop,
		SessionContext: sessionContext,
		ModelName:      config.ModelName,
	}

	for _, o := range options {
		o(&agent)
	}

	return &agent, nil
}
