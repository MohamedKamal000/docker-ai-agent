package core

import (
	"docker-cli/internal/models"
	"testing"
)

type testContainerSummary struct {
	Name   string
	Image  string
	State  string
	Status string
	ID     string
}

type testImageSummary struct {
	ID   string
	Tags []string
	Size int64
}

type testVolumeSummary struct {
	Name   string
	Driver string
}

type testNetworkSummary struct {
	Name   string
	Driver string
	Scope  string
}

func TestParsePrompt_SystemPrompt(t *testing.T) {
	data := map[string]any{
		"Containers": []testContainerSummary{
			{Name: "web", Image: "nginx:latest", State: "running", Status: "Up 1 hour", ID: "abc123"},
		},
		"Images": []testImageSummary{
			{ID: "sha256:abc123", Tags: []string{"nginx:latest"}, Size: 1024 * 1024 * 50},
		},
		"Volumes": []testVolumeSummary{
			{Name: "data", Driver: "local"},
		},
		"Networks": []testNetworkSummary{
			{Name: "bridge", Driver: "bridge", Scope: "local"},
		},
	}

	result, err := ParsePrompt(System_Prompt_Template, data)
	if err != nil {
		t.Fatalf("ParsePrompt error: %v", err)
	}

	expectedSections := []string{
		"ROLE",
		"STRICT OUTPUT FORMAT",
		"OPERATIONAL GUIDELINES",
		"SAFETY & DESTRUCTIVE ACTIONS",
		"WHEN NOT TO RESPOND WITH AN ACTION",
		"CURRENT DOCKER STATE",
		"web",
		"nginx:latest",
		"running",
		"data",
		"bridge",
	}

	for _, section := range expectedSections {
		if !containsString(result, section) {
			t.Errorf("System prompt missing section %q", section)
		}
	}
}

func TestParsePrompt_UserPrompt(t *testing.T) {
	data := models.UserInputPrompt{
		Goal: "test goal",
		History: []models.HistoryEntry{
			{
				Run:  1,
				Goal: "previous goal",
				ToolCalls: []models.ToolCallInfo{
					{ToolName: "docker_command_tool", Command: "docker ps", Status: "completed", Result: "container list"},
				},
				Feedback: "Found 3 containers",
			},
		},
	}

	result, err := ParsePrompt(User_Prompt_Template, data)
	if err != nil {
		t.Fatalf("ParsePrompt error: %v", err)
	}

	expectedSections := []string{
		"GOAL",
		"test goal",
		"HISTORY",
		"Run 1",
		"previous goal",
		"docker_command_tool",
		"docker ps",
		"completed",
		"container list",
		"Found 3 containers",
	}

	for _, section := range expectedSections {
		if !containsString(result, section) {
			t.Errorf("User prompt missing section %q", section)
		}
	}
}

func TestParsePrompt_UserPrompt_EmptyHistory(t *testing.T) {
	data := models.UserInputPrompt{
		Goal:    "test goal",
		History: []models.HistoryEntry{},
	}

	result, err := ParsePrompt(User_Prompt_Template, data)
	if err != nil {
		t.Fatalf("ParsePrompt error: %v", err)
	}

	if !containsString(result, "No history yet") {
		t.Errorf("User prompt should show 'No history yet' for empty history")
	}
}

func TestParsePrompt_DockerQueryPrompt(t *testing.T) {
	data := DockerQueryInput{
		Goal: "list containers",
		History: []models.HistoryEntry{
			{
				Run:  1,
				Goal: "list containers",
				ToolCalls: []models.ToolCallInfo{
					{ToolName: "docker_command_tool", Command: "docker ps", Status: "completed", Result: "container list"},
				},
			},
		},
	}

	result, err := ParsePrompt(DockerQuery_Prompt_Template, data)
	if err != nil {
		t.Fatalf("ParsePrompt error: %v", err)
	}

	expectedSections := []string{
		"GOAL",
		"list containers",
		"HISTORY",
		"Run 1",
		"docker_command_tool",
		"docker ps",
	}

	for _, section := range expectedSections {
		if !containsString(result, section) {
			t.Errorf("Docker query prompt missing section %q", section)
		}
	}
}

func TestParsePrompt_DockerQueryPrompt_EmptyHistory(t *testing.T) {
	data := DockerQueryInput{
		Goal:    "list containers",
		History: []models.HistoryEntry{},
	}

	result, err := ParsePrompt(DockerQuery_Prompt_Template, data)
	if err != nil {
		t.Fatalf("ParsePrompt error: %v", err)
	}

	if !containsString(result, "No history yet") {
		t.Errorf("Docker query prompt should show 'No history yet' for empty history")
	}
}

func TestParsePrompt_EvaluatorPrompt(t *testing.T) {
	data := EvaluatorInput{
		Goal: "test goal",
		History: []models.HistoryEntry{
			{
				Run:  1,
				Goal: "test goal",
				ToolCalls: []models.ToolCallInfo{
					{ToolName: "docker_command_tool", Command: "docker run nginx", Status: "completed", Result: "started"},
				},
				Feedback: "Container started",
			},
		},
	}

	result, err := ParsePrompt(Evaluator_Prompt_Template, data)
	if err != nil {
		t.Fatalf("ParsePrompt error: %v", err)
	}

	expectedSections := []string{
		"GOAL",
		"test goal",
		"HISTORY",
		"Run 1",
		"docker_command_tool",
		"docker run nginx",
		"completed",
		"started",
		"Container started",
	}

	for _, section := range expectedSections {
		if !containsString(result, section) {
			t.Errorf("Evaluator prompt missing section %q: %s", section, result[:min(200, len(result))])
		}
	}
}

func TestParsePrompt_IntentClassification(t *testing.T) {
	data := IntentClassificationInput{
		UserInput: "run nginx container",
	}

	result, err := ParsePrompt(Intent_Classification_Template, data)
	if err != nil {
		t.Fatalf("ParsePrompt error: %v", err)
	}

	expectedSections := []string{
		"ROLE",
		"CATEGORIES",
		"general_question",
		"docker_query",
		"action_request",
		"ambiguous",
		"OUTPUT JSON EXAMPLE",
	}

	for _, section := range expectedSections {
		if !containsString(result, section) {
			t.Errorf("Intent classification prompt missing section %q", section)
		}
	}
}

func TestParsePrompt_GeneralQuestionWithRAG(t *testing.T) {
	data := map[string]any{
		"Goal": "How do I use volumes?",
		"RagResult": []models.SearchResult{
			{
				Document: models.VectorDocument{
					Id: "doc1",
					MetaData: models.MetaData{
						Title: "Docker Volumes Guide",
					},
					Content: "Volumes are the preferred mechanism for persisting data...",
				},
			},
		},
	}

	result, err := ParsePrompt(GENERAL_QUESTION_WITH_RAG_PROMPT, data)
	if err != nil {
		t.Fatalf("ParsePrompt error: %v", err)
	}

	expectedSections := []string{
		"RETRIEVED KNOWLEDGE",
		"Docker Volumes Guide",
		"Volumes are the preferred mechanism",
		"USER QUESTION",
		"How do I use volumes?",
	}

	for _, section := range expectedSections {
		if !containsString(result, section) {
			t.Errorf("General question with RAG prompt missing section %q", section)
		}
	}
}

func TestParsePrompt_GeneralQuestionWithRAG_EmptyResults(t *testing.T) {
	data := map[string]any{
		"Goal":      "How do I use volumes?",
		"RagResult": []models.SearchResult{},
	}

	result, err := ParsePrompt(GENERAL_QUESTION_WITH_RAG_PROMPT, data)
	if err != nil {
		t.Fatalf("ParsePrompt error: %v", err)
	}

	if !containsString(result, "No relevant documents retrieved") {
		t.Errorf("Should show 'No relevant documents retrieved' for empty results")
	}
}

func TestParsePrompt_ErrorCases(t *testing.T) {
	// Test invalid template syntax
	_, err := ParsePrompt("{{invalid", map[string]any{})
	if err == nil {
		t.Error("ParsePrompt should error for invalid template syntax")
	}

	// Test template with undefined variable - Go templates don't error on undefined fields
	// They just render as empty string
	result, err := ParsePrompt("{{.UndefinedField}}", map[string]any{})
	if err != nil {
		t.Errorf("ParsePrompt should not error for undefined field: %v", err)
	}
	if result != "<no value>" {
		t.Errorf("Undefined field should render as '<no value>', got %q", result)
	}
}

func TestPromptConstantsExist(t *testing.T) {
	// Ensure all prompt templates are defined and non-empty
	templates := []string{
		System_Prompt_Template,
		User_Prompt_Template,
		DockerQuery_Prompt_Template,
		DockerQuery_System_Prompt,
		Evaluator_System_Prompt,
		Evaluator_Prompt_Template,
		Intent_Classification_Template,
		GENERAL_QUESTION_SYSTEM_PROMPT,
		GENERAL_QUESTION_WITH_RAG_PROMPT,
	}

	for i, tmpl := range templates {
		if tmpl == "" {
			t.Errorf("Template %d is empty", i)
		}
		if len(tmpl) < 10 {
			t.Errorf("Template %d is too short: %s", i, tmpl)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}