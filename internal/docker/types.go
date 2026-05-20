package docker

import "time"

type Context struct {
	Containers []ContainerSummary `json:"containers"`
	Images     []ImageSummary     `json:"images"`
	Volumes    []VolumeSummary    `json:"volumes"`
	Networks   []NetworkSummary   `json:"networks"`
	CapturedAt time.Time          `json:"captured_at"`
}

type ContainerSummary struct {
	ID      string        `json:"id"`
	Name    string        `json:"name"`
	Image   string        `json:"image"`
	State   string        `json:"state"` 
	Status  string        `json:"status"` 
	Ports   []PortMapping `json:"ports,omitempty"`
	Created time.Time     `json:"created"`
}

type PortMapping struct {
	HostPort      string `json:"host_port"`
	ContainerPort string `json:"container_port"`
	Protocol      string `json:"protocol"` 
}

type ImageSummary struct {
	ID      string    `json:"id"`
	Tags    []string  `json:"tags"`
	Size    int64     `json:"size"` 
	Created time.Time `json:"created"`
}

type VolumeSummary struct {
	Name   string `json:"name"`
	Driver string `json:"driver"`
}

type NetworkSummary struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Driver string `json:"driver"`
	Scope  string `json:"scope"`
}

type ExecResult struct {
	Command  string
	Stdout   string
	Stderr   string
	ExitCode int
	Duration time.Duration
}

func (r ExecResult) Succeeded() bool { return r.ExitCode == 0 }