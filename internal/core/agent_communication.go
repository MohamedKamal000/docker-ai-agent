package core

import (
	"context"
	"time"

	"docker-cli/internal/models"
)

// SleepCancellable waits for d, returning early with the context error if ctx
// is canceled first.
func SleepCancellable(ctx context.Context, d time.Duration) error {
	select {
	case <-time.After(d):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

type ResponseType int

const (
	Thoughts ResponseType = iota
	FinalResponse
	Warning
	Retrying
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

func NewRetryingMessage(msg string) AiResponse {
	return AiResponse{
		Type:    Retrying,
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
