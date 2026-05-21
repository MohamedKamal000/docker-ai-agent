package core

import "docker-cli/internal/models"

// to be continue, some one has to research on how to make sessions persistance
// may require refactor to current code if you want, mainly in the main loop i think

// made this for now, if someone going to work on it, make a new folder (package) for memory related stuff
// and move the interface and the implementation to it, we can also add more implementations later on if needed

type MemoryStore interface {
	Save(step models.ContextStep) error
	Load() ([]models.ContextStep, error)
	Clear() error
}

type StaticMemoryStore struct {
	Memory []models.ContextStep
}

func NewStaticMemoryStore() *StaticMemoryStore {
	return &StaticMemoryStore{Memory: []models.ContextStep{}}
}

func (sms *StaticMemoryStore) Save(step models.ContextStep) error {
	sms.Memory = append(sms.Memory, step)
	return nil
}

func (sms *StaticMemoryStore) Load() ([]models.ContextStep, error) {
	return sms.Memory, nil
}

func (sms *StaticMemoryStore) Clear() error {
	sms.Memory = []models.ContextStep{}
	return nil
}
