package core

import "docker-cli/internal/models"

type MemoryStore interface {
	Save(history []models.HistoryEntry) error
	Load() ([]models.HistoryEntry, error)
	Clear() error
}

type StaticMemoryStore struct {
	Entries []models.HistoryEntry
}

func NewStaticMemoryStore() *StaticMemoryStore {
	return &StaticMemoryStore{Entries: []models.HistoryEntry{}}
}

func (sms *StaticMemoryStore) Save(history []models.HistoryEntry) error {
	sms.Entries = append(sms.Entries, history...)
	return nil
}

func (sms *StaticMemoryStore) Load() ([]models.HistoryEntry, error) {
	return sms.Entries, nil
}

func (sms *StaticMemoryStore) Clear() error {
	sms.Entries = []models.HistoryEntry{}
	return nil
}
