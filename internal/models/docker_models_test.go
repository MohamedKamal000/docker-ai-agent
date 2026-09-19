package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestContainer_JSON(t *testing.T) {
	container := Container{
		Id:      "abc123",
		Name:    "web-server",
		Image:   "nginx:latest",
		Status:  "Up 2 hours",
		State:   "running",
		Created: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	jsonData, err := json.Marshal(container)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var parsed Container
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if parsed.Id != container.Id {
		t.Errorf("Id = %q, want %q", parsed.Id, container.Id)
	}
	if parsed.Name != container.Name {
		t.Errorf("Name = %q, want %q", parsed.Name, container.Name)
	}
	if parsed.Image != container.Image {
		t.Errorf("Image = %q, want %q", parsed.Image, container.Image)
	}
	if parsed.Status != container.Status {
		t.Errorf("Status = %q, want %q", parsed.Status, container.Status)
	}
	if parsed.State != container.State {
		t.Errorf("State = %q, want %q", parsed.State, container.State)
	}
	if !parsed.Created.Equal(container.Created) {
		t.Errorf("Created = %v, want %v", parsed.Created, container.Created)
	}
}

func TestPortMapping_JSON(t *testing.T) {
	pm := PortMapping{
		HostIP:        "0.0.0.0",
		HostPort:      "8080",
		ContainerPort: "80",
		Protocol:      "tcp",
	}

	jsonData, err := json.Marshal(pm)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var parsed PortMapping
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if parsed.HostIP != pm.HostIP {
		t.Errorf("HostIP = %q, want %q", parsed.HostIP, pm.HostIP)
	}
	if parsed.HostPort != pm.HostPort {
		t.Errorf("HostPort = %q, want %q", parsed.HostPort, pm.HostPort)
	}
	if parsed.ContainerPort != pm.ContainerPort {
		t.Errorf("ContainerPort = %q, want %q", parsed.ContainerPort, pm.ContainerPort)
	}
	if parsed.Protocol != pm.Protocol {
		t.Errorf("Protocol = %q, want %q", parsed.Protocol, pm.Protocol)
	}
}

func TestPortMapping_JSON_OmitEmpty(t *testing.T) {
	pm := PortMapping{
		HostPort:      "8080",
		ContainerPort: "80",
		Protocol:      "tcp",
		// HostIP omitted
	}

	jsonData, err := json.Marshal(pm)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	// HostIP should be omitted due to omitempty
	jsonStr := string(jsonData)
	if containsString(jsonStr, "host_ip") {
		t.Errorf("HostIP should be omitted: %s", jsonStr)
	}
}

func TestContextSnapshot_JSON(t *testing.T) {
	snapshot := ContextSnapshot{
		Containers: []Container{
			{Id: "1", Name: "c1", Image: "nginx", State: "running"},
			{Id: "2", Name: "c2", Image: "redis", State: "stopped"},
		},
		Images: []Image{
			{ID: "img1", Tags: []string{"nginx:latest"}, Size: 100},
			{ID: "img2", Tags: []string{"redis:7"}, Size: 200},
		},
		Volumes: []Volume{
			{Name: "vol1", Driver: "local"},
		},
		Networks: []Network{
			{ID: "net1", Name: "bridge", Mode: "bridge"},
		},
		CapturedAt: time.Now(),
	}

	jsonData, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var parsed ContextSnapshot
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if len(parsed.Containers) != 2 {
		t.Errorf("Containers = %d, want 2", len(parsed.Containers))
	}
	if len(parsed.Images) != 2 {
		t.Errorf("Images = %d, want 2", len(parsed.Images))
	}
	if len(parsed.Volumes) != 1 {
		t.Errorf("Volumes = %d, want 1", len(parsed.Volumes))
	}
	if len(parsed.Networks) != 1 {
		t.Errorf("Networks = %d, want 1", len(parsed.Networks))
	}
}

func TestImage_JSON(t *testing.T) {
	img := Image{
		ID:      "sha256:abc123",
		Tags:    []string{"nginx:latest", "nginx:1.24"},
		Size:    1024 * 1024 * 50,
		Created: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	jsonData, err := json.Marshal(img)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var parsed Image
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if parsed.ID != img.ID {
		t.Errorf("ID = %q, want %q", parsed.ID, img.ID)
	}
	if len(parsed.Tags) != 2 {
		t.Errorf("Tags = %d, want 2", len(parsed.Tags))
	}
	if parsed.Size != img.Size {
		t.Errorf("Size = %d, want %d", parsed.Size, img.Size)
	}
	if !parsed.Created.Equal(img.Created) {
		t.Errorf("Created = %v, want %v", parsed.Created, img.Created)
	}
}

func TestVolume_JSON(t *testing.T) {
	vol := Volume{
		Name:   "data-volume",
		Driver: "local",
	}

	jsonData, err := json.Marshal(vol)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var parsed Volume
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if parsed.Name != vol.Name {
		t.Errorf("Name = %q, want %q", parsed.Name, vol.Name)
	}
	if parsed.Driver != vol.Driver {
		t.Errorf("Driver = %q, want %q", parsed.Driver, vol.Driver)
	}
}

func TestNetwork_JSON(t *testing.T) {
	net := Network{
		ID:   "net1",
		Name: "bridge",
		Mode: "bridge",
	}

	jsonData, err := json.Marshal(net)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var parsed Network
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if parsed.ID != net.ID {
		t.Errorf("ID = %q, want %q", parsed.ID, net.ID)
	}
	if parsed.Name != net.Name {
		t.Errorf("Name = %q, want %q", parsed.Name, net.Name)
	}
	if parsed.Mode != net.Mode {
		t.Errorf("Mode = %q, want %q", parsed.Mode, net.Mode)
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