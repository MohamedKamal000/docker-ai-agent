package core

import "docker-cli/internal/models"

type ResponseType int

const (
	Thoughts ResponseType = iota
	FinalResponse
	Warning
)

type AiResponse struct {
	Type    ResponseType
	Message string
}

func NewThought(result models.AgentResult) AiResponse {
	if !result.IsStructured {
		return AiResponse{
			Type:    Thoughts,
			Message: result.Raw,
		}
	}

	return AiResponse{
		Type:    Thoughts,
		Message: result.Structured.Thought,
	}
}

func NewFinal(result models.AgentResult) AiResponse {
	if !result.IsStructured {
		return AiResponse{
			Type:    FinalResponse,
			Message: result.Raw,
		}
	}

	return AiResponse{
		Type:    FinalResponse,
		Message: result.Structured.FinalResponse,
	}
}

func NewWarning(msg string) AiResponse {
	return AiResponse{
		Type:    Warning,
		Message: msg,
	}
}

type UserCommand struct {
	ShouldContinue bool
}

type AgentCommunication struct {
	ToUser   chan AiResponse
	FromUser chan UserCommand
}
