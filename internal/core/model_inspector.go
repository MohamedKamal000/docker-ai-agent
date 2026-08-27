package core

import (
	"encoding/json"
	"io"
	"net/http"
	"slices"
	"strings"
)

type ModelInfo struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Family      string     `json:"family"`
	Attachment  bool       `json:"attachment"`
	Reasoning   bool       `json:"reasoning"`
	ToolCall    bool       `json:"tool_call"`
	Temperature bool       `json:"temperature"`
	ReleaseDate string     `json:"release_date"`
	LastUpdated string     `json:"last_updated"`
	Modalities  Modalities `json:"modalities"`
	Limit       Limit      `json:"limit"`
}

type Modalities struct {
	Input  []string `json:"input"`
	Output []string `json:"output"`
}

type Limit struct {
	Context int `json:"context"`
	Output  int `json:"output"`
}

type ModelInspector struct {
	supportedModels map[string]ModelInfo
}

func hasTextAsInputAndOutput(m Modalities) bool {
	ok := slices.Contains(m.Input, "text")

	if slices.Contains(m.Output, "text") {
		ok = ok && true
	}

	return ok
}

func (m *ModelInspector) buildSupportedModelsList() error {
	m.supportedModels = map[string]ModelInfo{}
	resp, err := http.Get("https://models.dev/models.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var models map[string]ModelInfo

	if err = json.Unmarshal(data, &models); err != nil {
		return err
	}

	modelPrefixes := make(map[string]bool, 0)
	for _, v := range ProviderMap {
		// genkit has this weird prefix instead of just writing google, no idea why
		// the api i call has google as prefix instead of googleai
		if v.PrefixName == "googleai" {
			modelPrefixes["google"] = true
			continue
		}
		modelPrefixes[v.PrefixName] = true
	}

	for k, v := range models {
		prefix, _, _ := strings.Cut(k, "/")

		if modelPrefixes[prefix] && hasTextAsInputAndOutput(v.Modalities) {
			m.supportedModels[v.ID] = v
		}
	}

	return nil
}

func (m *ModelInspector) GetModelsByProvider() map[string][]ModelInfo {
	bucket := map[string][]ModelInfo{}

	for k, v := range m.supportedModels {
		prefix, _, _ := strings.Cut(k, "/")
		bucket[prefix] = append(bucket[prefix], v)
	}
	return bucket
}

func (m *ModelInspector) InspectModel(modelId string) (ModelInfo, bool) {
	result, ok := m.supportedModels[modelId]
	return result, ok
}

func NewModelInspector() (ModelInspector, error) {
	var result ModelInspector
	err := result.buildSupportedModelsList()
	if err != nil {
		return ModelInspector{}, err
	}
	return result, nil
}
