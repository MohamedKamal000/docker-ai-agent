package screens

import "docker-cli/internal/core"

type SwitchToNewSessionMessage struct{}

type SendMessageRequest struct {
	UserMessage string
}

type ReceiveMessageResponse struct {
	Response core.AiResponse
}

type ErrorMessage struct {
	Message string
}

type ConfirmMessage struct {
	ShouldContinue bool
}
