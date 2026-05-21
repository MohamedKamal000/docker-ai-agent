package docker

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
)

var (
	once      sync.Once
	cli       *client.Client
	dockerBin string
	initErr   error
)

func Init(opts ...client.Opt) error {
	once.Do(func() {
		defaultOpts := []client.Opt{
			client.FromEnv,
			client.WithAPIVersionNegotiation(),
		}
		defaultOpts = append(defaultOpts, opts...)

		cli, initErr = client.NewClientWithOpts(defaultOpts...)
		if initErr != nil {
			return
		}

		if _, initErr = cli.Ping(context.Background()); initErr != nil {
			initErr = fmt.Errorf("docker: cannot reach daemon: %w", initErr)
			return
		}

		dockerBin, initErr = exec.LookPath("docker")
		if initErr != nil {
			initErr = fmt.Errorf("docker: binary not found in PATH: %w", initErr)
		}
	})
	return initErr
}

func must() {
	if cli == nil {
		panic("docker: call docker.Init() before using any docker package functions")
	}
}

func GetContext(ctx context.Context) (*Context, error) {
	must()

	containers, err := ListContainers(ctx)
	if err != nil {
		return nil, err
	}

	imgs, err := ListImages(ctx)
	if err != nil {
		return nil, err
	}

	vols, err := ListVolumes(ctx)
	if err != nil {
		return nil, err
	}

	nets, err := ListNetworks(ctx)
	if err != nil {
		return nil, err
	}

	return &Context{
		Containers: containers,
		Images:     imgs,
		Volumes:    vols,
		Networks:   nets,
		CapturedAt: time.Now(),
	}, nil
}

func ListContainers(ctx context.Context, opts ...ListOption) ([]ContainerSummary, error) {
	must()

	o := listOptions{}
	for _, opt := range opts {
		opt(&o)
	}

	raw, err := cli.ContainerList(ctx, container.ListOptions{
		All: o.all,
	})
	if err != nil {
		return nil, fmt.Errorf("docker: list containers: %w", err)
	}

	out := make([]ContainerSummary, 0, len(raw))
	for _, c := range raw {
		name := ""
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}

		var ports []PortMapping
		for _, p := range c.Ports {
			ports = append(ports, PortMapping{
				HostPort:      fmt.Sprintf("%d", p.PublicPort),
				ContainerPort: fmt.Sprintf("%d", p.PrivatePort),
				Protocol:      p.Type,
			})
		}

		out = append(out, ContainerSummary{
			ID:      c.ID[:min(12, len(c.ID))],
			Name:    name,
			Image:   c.Image,
			State:   c.State,
			Status:  c.Status,
			Ports:   ports,
			Created: time.Unix(c.Created, 0),
		})
	}
	return out, nil
}

func ListImages(ctx context.Context) ([]ImageSummary, error) {
	must()

	raw, err := cli.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("docker: list images: %w", err)
	}

	out := make([]ImageSummary, 0, len(raw))
	for _, img := range raw {
		tags := make([]string, 0, len(img.RepoTags))
		for _, t := range img.RepoTags {
			if t != "<none>:<none>" {
				tags = append(tags, t)
			}
		}

		out = append(out, ImageSummary{
			ID:      img.ID[:min(12+7, len(img.ID))],
			Tags:    tags,
			Size:    img.Size,
			Created: time.Unix(img.Created, 0),
		})
	}
	return out, nil
}

func ListVolumes(ctx context.Context) ([]VolumeSummary, error) {
	must()

	raw, err := cli.VolumeList(ctx, volume.ListOptions{
		Filters: filters.Args{},
	})
	if err != nil {
		return nil, fmt.Errorf("docker: list volumes: %w", err)
	}

	out := make([]VolumeSummary, 0, len(raw.Volumes))
	for _, v := range raw.Volumes {
		out = append(out, VolumeSummary{
			Name:   v.Name,
			Driver: v.Driver,
		})
	}
	return out, nil
}

func ListNetworks(ctx context.Context) ([]NetworkSummary, error) {
	must()

	raw, err := cli.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("docker: list networks: %w", err)
	}

	out := make([]NetworkSummary, 0, len(raw))
	for _, n := range raw {
		out = append(out, NetworkSummary{
			ID:     n.ID[:min(12, len(n.ID))],
			Name:   n.Name,
			Driver: n.Driver,
			Scope:  n.Scope,
		})
	}
	return out, nil
}

var DestructiveVerbs = []string{
	"rm", "rmi", "prune", "stop", "kill", "down", "restart", "pause", "remove",
}

func IsDestructive(command string) bool {
	lower := strings.ToLower(command)
	for _, v := range DestructiveVerbs {
		if strings.Contains(lower, " "+v) || strings.Contains(lower, "\t"+v) {
			return true
		}
	}
	return false
}

func ParseCommands(block string) []string {
	var out []string
	for _, line := range strings.Split(block, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out = append(out, line)
	}
	return out
}

func Exec(ctx context.Context, command string) (ExecResult, error) {
	must()

	args, err := shellSplit(command)
	if err != nil {
		return ExecResult{}, fmt.Errorf("docker: cannot parse %q: %w", command, err)
	}

	if len(args) > 0 && args[0] == "docker" {
		args = args[1:]
	}
	if len(args) == 0 {
		return ExecResult{}, fmt.Errorf("docker: empty command")
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, dockerBin, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	runErr := cmd.Run()
	elapsed := time.Since(start)

	exitCode := 0
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return ExecResult{}, fmt.Errorf("docker: exec %v: %w", args, runErr)
		}
	}

	return ExecResult{
		Command:  dockerBin + " " + strings.Join(args, " "),
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
		Duration: elapsed,
	}, nil
}

func ExecMany(ctx context.Context, commands []string, continueOnError bool) ([]ExecResult, error) {
	results := make([]ExecResult, 0, len(commands))
	for _, cmd := range commands {
		result, err := Exec(ctx, cmd)
		if err != nil {
			return results, err
		}
		results = append(results, result)
		if !result.Succeeded() && !continueOnError {
			return results, fmt.Errorf("docker: command exited %d: %s",
				result.ExitCode, strings.TrimSpace(result.Stderr))
		}
	}
	return results, nil
}

func FormatContextPrompt(dc *Context) string {
	if dc == nil {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("=== Docker Environment (captured %s) ===\n",
		dc.CapturedAt.Format(time.RFC3339)))

	sb.WriteString(fmt.Sprintf("\n## Containers (%d)\n", len(dc.Containers)))
	for _, c := range dc.Containers {
		ports := formatPorts(c.Ports)
		if ports != "" {
			ports = " ports=[" + ports + "]"
		}
		sb.WriteString(fmt.Sprintf("  %-20s  image=%-30s  state=%s%s\n",
			c.Name, c.Image, c.State, ports))
	}

	sb.WriteString(fmt.Sprintf("\n## Images (%d)\n", len(dc.Images)))
	for _, img := range dc.Images {
		tags := strings.Join(img.Tags, ", ")
		if tags == "" {
			tags = "<untagged>"
		}
		sb.WriteString(fmt.Sprintf("  %-40s  size=%s\n", tags, formatBytes(img.Size)))
	}

	sb.WriteString(fmt.Sprintf("\n## Volumes (%d)\n", len(dc.Volumes)))
	for _, v := range dc.Volumes {
		sb.WriteString(fmt.Sprintf("  %s  driver=%s\n", v.Name, v.Driver))
	}

	sb.WriteString(fmt.Sprintf("\n## Networks (%d)\n", len(dc.Networks)))
	for _, n := range dc.Networks {
		sb.WriteString(fmt.Sprintf("  %-20s  driver=%-10s  scope=%s\n", n.Name, n.Driver, n.Scope))
	}

	sb.WriteString("\n=== End Docker Environment ===\n")
	return sb.String()
}

type listOptions struct{ all bool }
type ListOption func(*listOptions)

func WithAll() ListOption {
	return func(o *listOptions) { o.all = true }
}

func formatPorts(ports []PortMapping) string {
	parts := make([]string, 0, len(ports))
	for _, p := range ports {
		if p.HostPort == "0" || p.HostPort == "" {
			parts = append(parts, fmt.Sprintf("%s/%s", p.ContainerPort, p.Protocol))
		} else {
			parts = append(parts, fmt.Sprintf("%s→%s/%s", p.HostPort, p.ContainerPort, p.Protocol))
		}
	}
	return strings.Join(parts, ", ")
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func shellSplit(s string) ([]string, error) {
	var tokens []string
	var current strings.Builder
	inSingle, inDouble, escaped := false, false, false

	for i := 0; i < len(s); i++ {
		ch := s[i]
		if escaped {
			current.WriteByte(ch)
			escaped = false
			continue
		}
		switch {
		case ch == '\\' && !inSingle:
			escaped = true
		case ch == '\'' && !inDouble:
			inSingle = !inSingle
		case ch == '"' && !inSingle:
			inDouble = !inDouble
		case (ch == ' ' || ch == '\t') && !inSingle && !inDouble:
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		default:
			current.WriteByte(ch)
		}
	}
	if inSingle {
		return nil, fmt.Errorf("unterminated single quote")
	}
	if inDouble {
		return nil, fmt.Errorf("unterminated double quote")
	}
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}
	return tokens, nil
}
