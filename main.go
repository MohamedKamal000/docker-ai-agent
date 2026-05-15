package main

import (
	"context"
	"docker-cli/internal/core"
	"docker-cli/internal/models"
	"fmt"
	"log"

	"github.com/moby/moby/client"
)

func main() {
	// Create a new client with "client.FromEnv" (configuring the client
	// from commonly used environment variables such as DOCKER_HOST and
	// DOCKER_API_VERSION) and set a custom User-Agent.
	//
	// API-version negotiation is enabled by default to allow downgrading
	// the API version when connecting with an older daemon version.
	apiClient, err := client.New(
		client.FromEnv,
		client.WithUserAgent("my-application/1.0.0"),
	)
	if err != nil {
		panic(err)
	}
	defer apiClient.Close()

	// List all containers (both stopped and running).
	result, err := apiClient.ContainerList(context.Background(), client.ContainerListOptions{All: true})
	if err != nil {
		panic(err)
	}

	containers := make([]models.Container, 0)

	for _, ctr := range result.Items {

		containers = append(containers, models.Container{
			Name:   ctr.Names[0],
			Image:  ctr.Image,
			Status: ctr.Status,
			Id:     ctr.ID,
		})
	}

	userInput := models.UserInputPrompt{
		Goal:    "list all the containers for me",
		Context: make([]models.ContextStep, 0),
	}

	userInput.Context = append(userInput.Context, *models.NewContextStep(
		models.ContextAction{Action: "executed containers list", Thought: "i need to execute the containers list"},
		"got all containers",
	))

	parsedRes, err := core.ParsePrompt(core.User_Prompt, userInput)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(parsedRes)
}
