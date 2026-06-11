package tools

import (
	"context"
	"docker-cli/internal/docker"
	"encoding/json"
	"fmt"
)

type DockerCommandsTool struct {
	InputSchema map[string]any // JSON schema, see: https://json-schema.org/
}

func NewDockerCommandsTool() *DockerCommandsTool {
	return &DockerCommandsTool{InputSchema: map[string]any{
		"type": "object",
		"properties": map[string]any{
			"command": map[string]any{
				"type":        "string",
				"description": "A valid Docker CLI command to execute.",
			},
		},
		"required":             []string{"command"},
		"additionalProperties": false,
	}}
}

type DockerCommandInput struct {
	Command string `json:"command"`
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
		return "", fmt.Errorf("failed to cast input to docker Command input")
	}
	cmd := m["command"].(string)

	res, err := docker.Exec(ctx, cmd)

	if err != nil {
		return "", err
	}

	b, err := json.MarshalIndent(res, "", "  ")

	if err != nil {
		return "", err
	}

	return string(b), nil
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
		return fmt.Sprintf("docker command %s contains a destructive command", cmd), true
	}

	return "", false
}
