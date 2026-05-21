package models

import "time"

type Container struct {
	Id      string    `json:"id"`
	Name    string    `json:"name"`
	Image   string    `json:"image"`
	Status  string    `json:"status"`
	State   string    `json:"state"`
	Created time.Time `json:"created"`
}

type PortMapping struct {
	HostIP        string `json:"host_ip,omitempty"`
	HostPort      string `json:"host_port"`
	ContainerPort string `json:"container_port"`
	Protocol      string `json:"protocol"`
}

type ContextSnapshot struct {
	Containers []Container `json:"containers"`
	Images     []Image     `json:"images"`
	Volumes    []Volume    `json:"volumes"`
	Networks   []Network   `json:"networks"`
	CapturedAt time.Time   `json:"captured_at"`
}

type Image struct {
	ID      string    `json:"id"`
	Tags    []string  `json:"tags"`
	Size    int64     `json:"size"`
	Created time.Time `json:"created"`
}

type Volume struct {
	Name   string `json:"name"`
	Driver string `json:"driver"`
}

type Network struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Mode string `json:"mode"`
}
