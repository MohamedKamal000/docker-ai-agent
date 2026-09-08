package tools

import (
	"context"
	"fmt"

	"docker-cli/internal/core"
	"docker-cli/internal/docker"
)

type DockerCommandsTool struct {
	InputSchema map[string]any
	Tasks       *core.TaskRegistry
}

func NewDockerCommandsTool(tasks *core.TaskRegistry) *DockerCommandsTool {
	return &DockerCommandsTool{
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"command": map[string]any{
					"type":        "string",
					"description": "A valid Docker CLI command to execute.",
				},
			},
			"required":             []string{"command"},
			"additionalProperties": false,
		},
		Tasks: tasks,
	}
}

func (d *DockerCommandsTool) Name() string {
	return "docker_command_tool"
}

func (d *DockerCommandsTool) Description() string {
	return `Executes Docker a CLI commands on the host system.
Use this tool to interact with Docker containers, images, networks, volumes, and other Docker resources. The input should be a valid Docker command.
This tool can inspect, create, modify, start, stop, restart, and remove Docker resources. Commands executed through this tool may alter the state of the Docker environment and should be used with caution.
Use this tool only when direct interaction with Docker is required.`
}

func (d *DockerCommandsTool) Call(ctx context.Context, input any) (string, error) {
	m, ok := input.(map[string]any)
	if !ok {
		return "", fmt.Errorf("failed to cast input to docker command input")
	}

	cmd, ok := m["command"].(string)
	if !ok {
		return "", fmt.Errorf("command must be a string")
	}

	id, err := docker.Exec(ctx, cmd, d.Tasks)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("task id: %s", id), nil
}

func (d *DockerCommandsTool) GetInputSchema() map[string]any {
	return d.InputSchema
}

func (d *DockerCommandsTool) ShouldRaiseWarning(input any) (string, bool) {
	m, ok := input.(map[string]any)
	if !ok {
		return "unknown input to docker exec", true
	}
	cmd := m["command"].(string)
	res := docker.IsDestructive(cmd)
	if res {
		return fmt.Sprintf("docker command \"%s\" contains a destructive command\n", cmd), true
	}
	return "", false
}
