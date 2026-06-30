package core

import "docker-cli/internal/models"

// to be continue, some one has to research on how to make sessions persistance
// may require refactor to current code if you want, mainly in the main loop i think

// made this for now, if someone going to work on it, make a new folder (package) for memory related stuff
// and move the interface and the implementation to it, we can also add more implementations later on if needed

type MemoryStore interface {
	Save(userGoal string, toolsExecuted map[string]string, steps []models.AgentResult) error
	Load() ([]models.ChatInteraction, error)
	Clear() error
}

type StaticMemoryStore struct {
	Memory []models.ChatInteraction
}

func NewStaticMemoryStore() *StaticMemoryStore {
	return &StaticMemoryStore{Memory: []models.ChatInteraction{}}
}

func (sms *StaticMemoryStore) Save(userGoal string, toolsExecuted map[string]string, steps []models.AgentResult) error {
	interaction := models.ChatInteraction{
		UserRequest:        userGoal,
		LLMThoughts:        []string{},
		UnstructuredOutput: []string{},
		ToolsExecuted:      toolsExecuted,
	}

	for _, step := range steps {
		if step.IsStructured {
			interaction.LLMThoughts = append(interaction.LLMThoughts, step.Structured.Thought)
			if step.Structured.Done {
				interaction.FinalResponse = step.Structured.FinalResponse
				break
			}
		} else {
			interaction.UnstructuredOutput = append(interaction.UnstructuredOutput, step.Raw)
		}
	}

	sms.Memory = append(sms.Memory, interaction)
	return nil
}

func (sms *StaticMemoryStore) Load() ([]models.ChatInteraction, error) {
	return sms.Memory, nil
}

func (sms *StaticMemoryStore) Clear() error {
	sms.Memory = []models.ChatInteraction{}
	return nil
}
