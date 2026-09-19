package docker

import (
	"testing"
	"time"
)

func TestIsDestructive(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		expected bool
	}{
		{
			name:     "rm command",
			command:  "docker rm container1",
			expected: true,
		},
		{
			name:     "rmi command",
			command:  "docker rmi image1",
			expected: true,
		},
		{
			name:     "prune command",
			command:  "docker system prune",
			expected: true,
		},
		{
			name:     "stop command",
			command:  "docker stop container1",
			expected: true,
		},
		{
			name:     "kill command",
			command:  "docker kill container1",
			expected: true,
		},
		{
			name:     "down command",
			command:  "docker compose down",
			expected: true,
		},
		{
			name:     "restart command",
			command:  "docker restart container1",
			expected: true,
		},
		{
			name:     "pause command",
			command:  "docker pause container1",
			expected: true,
		},
		{
			name:     "remove command",
			command:  "docker remove container1",
			expected: true,
		},
		{
			name:     "ps command (read-only)",
			command:  "docker ps",
			expected: false,
		},
		{
			name:     "images command (read-only)",
			command:  "docker images",
			expected: false,
		},
		{
			name:     "logs command (read-only)",
			command:  "docker logs container1",
			expected: false,
		},
		{
			name:     "inspect command (read-only)",
			command:  "docker inspect container1",
			expected: false,
		},
		{
			name:     "network ls command (read-only)",
			command:  "docker network ls",
			expected: false,
		},
		{
			name:     "volume ls command (read-only)",
			command:  "docker volume ls",
			expected: false,
		},
		{
			name:     "Uppercase destructive command",
			command:  "DOCKER RM CONTAINER1",
			expected: true,
		},
		{
			name:     "Destructive command with tabs",
			command:  "docker\trm\tcontainer1",
			expected: true,
		},
		{
			name:     "Partial match should not trigger (e.g., 'form')",
			command:  "docker info",
			expected: false,
		},
		{
			name:     "Empty command",
			command:  "",
			expected: false,
		},
		{
			name:     "Command with rm in image name",
			command:  "docker run my-image-with-rm",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsDestructive(tt.command)
			if result != tt.expected {
				t.Errorf("IsDestructive(%q) = %v, want %v", tt.command, result, tt.expected)
			}
		})
	}
}

func TestParseCommands(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Simple commands",
			input:    "docker ps\ndocker images",
			expected: []string{"docker ps", "docker images"},
		},
		{
			name:     "With empty lines",
			input:    "docker ps\n\ndocker images\n",
			expected: []string{"docker ps", "docker images"},
		},
		{
			name:     "With comments",
			input:    "# This is a comment\ndocker ps\n# Another comment\ndocker images",
			expected: []string{"docker ps", "docker images"},
		},
		{
			name:     "Inline comments",
			input:    "docker ps # list containers\ndocker images",
			expected: []string{"docker ps # list containers", "docker images"},
		},
		{
			name:     "Only comments",
			input:    "# comment 1\n# comment 2",
			expected: []string{},
		},
		{
			name:     "Empty input",
			input:    "",
			expected: []string{},
		},
		{
			name:     "Whitespace only",
			input:    "   \n\t\n  ",
			expected: []string{},
		},
		{
			name:     "Commands with spaces",
			input:    "  docker ps  \n  docker images  ",
			expected: []string{"docker ps", "docker images"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseCommands(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("ParseCommands(%q) returned %d commands, want %d", tt.input, len(result), len(tt.expected))
				return
			}
			for i, cmd := range result {
				if cmd != tt.expected[i] {
					t.Errorf("ParseCommands(%q)[%d] = %q, want %q", tt.input, i, cmd, tt.expected[i])
				}
			}
		})
	}
}

func TestNormalizeCommand(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "No duplicate",
			input:    []string{"docker", "ps"},
			expected: []string{"docker", "ps"},
		},
		{
			name:     "Non-docker duplicates preserved",
			input:    []string{"docker", "ps", "ps"},
			expected: []string{"docker", "ps", "ps"},
		},
		{
			name:     "Empty args",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "Single arg",
			input:    []string{"docker"},
			expected: []string{"docker"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeCommand(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("NormalizeCommand(%v) returned %d args, want %d", tt.input, len(result), len(tt.expected))
				return
			}
			for i, arg := range result {
				if arg != tt.expected[i] {
					t.Errorf("NormalizeCommand(%v)[%d] = %q, want %q", tt.input, i, arg, tt.expected[i])
				}
			}
		})
	}
}

func TestShellSplit(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		expected    []string
	}{
		{
			name:     "Simple command",
			input:    "docker ps",
			expected: []string{"docker", "ps"},
		},
		{
			name:     "Command with quoted args",
			input:    `docker run "my container"`,
			expected: []string{"docker", "run", "my container"},
		},
		{
			name:     "Command with single quotes",
			input:    "docker run 'my container'",
			expected: []string{"docker", "run", "my container"},
		},
		{
			name:     "Command with escaped quotes",
			input:    `docker run "my \"container\""`,
			expected: []string{"docker", "run", `my "container"`},
		},
		{
			name:     "Command with escaped backslash",
			input:    `docker run path\\to\\file`,
			expected: []string{"docker", "run", "path\\to\\file"},
		},
		{
			name:     "Mixed quotes",
			input:    `docker run "double" 'single'`,
			expected: []string{"docker", "run", "double", "single"},
		},
		{
			name:     "Tabs and spaces",
			input:    "docker\tps\t-a",
			expected: []string{"docker", "ps", "-a"},
		},
		{
			name:     "Unterminated single quote",
			input:    `docker run 'unterminated`,
			expectError: true,
		},
		{
			name:     "Unterminated double quote",
			input:    `docker run "unterminated`,
			expectError: true,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "Whitespace only",
			input:    "   ",
			expected: []string{},
		},
		{
			name:     "Complex command",
			input:    `docker run -d --name "my-app" -p 8080:80 nginx:latest`,
			expected: []string{"docker", "run", "-d", "--name", "my-app", "-p", "8080:80", "nginx:latest"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := shellSplit(tt.input)
			if tt.expectError {
				if err == nil {
					t.Errorf("shellSplit(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("shellSplit(%q) unexpected error: %v", tt.input, err)
			}
			if len(result) != len(tt.expected) {
				t.Errorf("shellSplit(%q) returned %d args, want %d", tt.input, len(result), len(tt.expected))
				return
			}
			for i, arg := range result {
				if arg != tt.expected[i] {
					t.Errorf("shellSplit(%q)[%d] = %q, want %q", tt.input, i, arg, tt.expected[i])
				}
			}
		})
	}
}

func TestFormatPorts(t *testing.T) {
	tests := []struct {
		name     string
		ports    []PortMapping
		expected string
	}{
		{
			name:     "Empty ports",
			ports:    []PortMapping{},
			expected: "",
		},
		{
			name: "Single port with host",
			ports: []PortMapping{
				{HostPort: "8080", ContainerPort: "80", Protocol: "tcp"},
			},
			expected: "8080→80/tcp",
		},
		{
			name: "Single port without host",
			ports: []PortMapping{
				{HostPort: "", ContainerPort: "80", Protocol: "tcp"},
			},
			expected: "80/tcp",
		},
		{
			name: "Single port with host port 0",
			ports: []PortMapping{
				{HostPort: "0", ContainerPort: "80", Protocol: "tcp"},
			},
			expected: "80/tcp",
		},
		{
			name: "Multiple ports",
			ports: []PortMapping{
				{HostPort: "8080", ContainerPort: "80", Protocol: "tcp"},
				{HostPort: "8443", ContainerPort: "443", Protocol: "tcp"},
				{HostPort: "", ContainerPort: "53", Protocol: "udp"},
			},
			expected: "8080→80/tcp, 8443→443/tcp, 53/udp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatPorts(tt.ports)
			if result != tt.expected {
				t.Errorf("formatPorts(%v) = %q, want %q", tt.ports, result, tt.expected)
			}
		})
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{
			name:     "Bytes",
			input:    500,
			expected: "500 B",
		},
		{
			name:     "Exactly 1KB",
			input:    1024,
			expected: "1.0 KB",
		},
		{
			name:     "KB",
			input:    1536,
			expected: "1.5 KB",
		},
		{
			name:     "MB",
			input:    1024 * 1024 * 5,
			expected: "5.0 MB",
		},
		{
			name:     "GB",
			input:    1024 * 1024 * 1024 * 2,
			expected: "2.0 GB",
		},
		{
			name:     "TB",
			input:    1024 * 1024 * 1024 * 1024 * 3,
			expected: "3.0 TB",
		},
		{
			name:     "Zero",
			input:    0,
			expected: "0 B",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatBytes(tt.input)
			if result != tt.expected {
				t.Errorf("formatBytes(%d) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatContextPrompt(t *testing.T) {
	now := time.Now()

	ctx := &Context{
		Containers: []ContainerSummary{
			{
				ID:      "abc123",
				Name:    "web-server",
				Image:   "nginx:latest",
				State:   "running",
				Status:  "Up 2 hours",
				Ports: []PortMapping{
					{HostPort: "8080", ContainerPort: "80", Protocol: "tcp"},
				},
				Created: now,
			},
			{
				ID:      "def456",
				Name:    "database",
				Image:   "postgres:15",
				State:   "running",
				Status:  "Up 1 hour",
				Ports: []PortMapping{
					{HostPort: "", ContainerPort: "5432", Protocol: "tcp"},
				},
				Created: now,
			},
		},
		Images: []ImageSummary{
			{
				ID:      "sha256:abc123",
				Tags:    []string{"nginx:latest", "nginx:1.24"},
				Size:    1024 * 1024 * 50,
				Created: now,
			},
			{
				ID:      "sha256:def456",
				Tags:    []string{"postgres:15"},
				Size:    1024 * 1024 * 200,
				Created: now,
			},
			{
				ID:      "sha256:ghi789",
				Tags:    []string{}, // untagged image
				Size:    1024 * 1024 * 10,
				Created: now,
			},
		},
		Volumes: []VolumeSummary{
			{Name: "data-volume", Driver: "local"},
			{Name: "logs-volume", Driver: "local"},
		},
		Networks: []NetworkSummary{
			{ID: "net1", Name: "bridge", Driver: "bridge", Scope: "local"},
			{ID: "net2", Name: "custom-net", Driver: "bridge", Scope: "local"},
		},
		CapturedAt: now,
	}

	result := FormatContextPrompt(ctx)

	// Check that all sections are present
	expectedSections := []string{
		"=== Docker Environment",
		"## Containers (2)",
		"web-server",
		"nginx:latest",
		"running",
		"8080→80/tcp",
		"database",
		"postgres:15",
		"5432/tcp",
		"## Images (3)",
		"nginx:latest, nginx:1.24",
		"postgres:15",
		"<untagged>",
		"50.0 MB",
		"200.0 MB",
		"10.0 MB",
		"## Volumes (2)",
		"data-volume",
		"logs-volume",
		"## Networks (2)",
		"bridge",
		"custom-net",
		"=== End Docker Environment ===",
	}

	for _, section := range expectedSections {
		if !containsString(result, section) {
			t.Errorf("FormatContextPrompt() missing expected section %q", section)
		}
	}

	// Test nil context
	result = FormatContextPrompt(nil)
	if result != "" {
		t.Errorf("FormatContextPrompt(nil) = %q, want empty string", result)
	}
}

func TestFormatContextPrompt_Empty(t *testing.T) {
	ctx := &Context{
		Containers: []ContainerSummary{},
		Images:     []ImageSummary{},
		Volumes:    []VolumeSummary{},
		Networks:   []NetworkSummary{},
		CapturedAt: time.Now(),
	}

	result := FormatContextPrompt(ctx)

	if !containsString(result, "## Containers (0)") {
		t.Errorf("FormatContextPrompt() missing empty containers section")
	}
	if !containsString(result, "## Images (0)") {
		t.Errorf("FormatContextPrompt() missing empty images section")
	}
	if !containsString(result, "## Volumes (0)") {
		t.Errorf("FormatContextPrompt() missing empty volumes section")
	}
	if !containsString(result, "## Networks (0)") {
		t.Errorf("FormatContextPrompt() missing empty networks section")
	}
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}