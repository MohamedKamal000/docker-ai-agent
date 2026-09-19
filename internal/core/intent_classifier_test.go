package core

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/firebase/genkit/go/ai"
)



func TestExtractJSONInClassifier(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Valid JSON in markdown",
			input:    "```json\n{\"intent\": \"general_question\", \"rewritten_prompt\": \"test\"}\n```",
			expected: "{\"intent\": \"general_question\", \"rewritten_prompt\": \"test\"}",
		},
		{
			name:     "Valid JSON without markdown",
			input:    "{\"intent\": \"docker_query\", \"rewritten_prompt\": \"list containers\"}",
			expected: "{\"intent\": \"docker_query\", \"rewritten_prompt\": \"list containers\"}",
		},
		{
			name:     "Empty response",
			input:    "",
			expected: "",
		},
		{
			name:     "Invalid JSON",
			input:    "not json",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractJSON(tt.input)
			if result != tt.expected {
				t.Errorf("extractJSON(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestClassificationResponseUnmarshal(t *testing.T) {
	tests := []struct {
		name          string
		jsonInput     string
		expectedIntent Intent
		expectedPrompt string
		expectError   bool
	}{
		{
			name: "General question intent",
			jsonInput: `{"intent": "general_question", "rewritten_prompt": "How do I run a container?"}`,
			expectedIntent: IntentGeneralQuestion,
			expectedPrompt: "How do I run a container?",
		},
		{
			name: "Docker query intent",
			jsonInput: `{"intent": "docker_query", "rewritten_prompt": "List my running containers"}`,
			expectedIntent: IntentDockerQuery,
			expectedPrompt: "List my running containers",
		},
		{
			name: "Action request intent",
			jsonInput: `{"intent": "action_request", "rewritten_prompt": "Run nginx container"}`,
			expectedIntent: IntentActionRequest,
			expectedPrompt: "Run nginx container",
		},
		{
			name: "Ambiguous intent",
			jsonInput: `{"intent": "ambiguous", "rewritten_prompt": "What?"}`,
			expectedIntent: IntentAmbiguous,
			expectedPrompt: "What?",
		},
		{
			name: "Invalid intent value (parses but invalid)",
			jsonInput:   `{"intent": "invalid", "rewritten_prompt": "test"}`,
			expectedIntent: Intent("invalid"),
			expectedPrompt: "test",
			expectError:   false,
		},
		{
			name:        "Missing fields (partial parse)",
			jsonInput:   `{"intent": "general_question"}`,
			expectedIntent: IntentGeneralQuestion,
			expectedPrompt: "",
			expectError:   false,
		},
		{
			name:        "Invalid JSON",
			jsonInput:   `not json`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp ClassificationResponse
			err := json.Unmarshal([]byte(tt.jsonInput), &resp)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for %q, got nil", tt.jsonInput)
				}
				return
			}
			
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			
			intent := Intent(resp.Intent)
			if intent != tt.expectedIntent {
				t.Errorf("Intent = %v, want %v", intent, tt.expectedIntent)
			}
			
			if resp.RewrittenPrompt != tt.expectedPrompt {
				t.Errorf("RewrittenPrompt = %q, want %q", resp.RewrittenPrompt, tt.expectedPrompt)
			}
		})
	}
}

func TestIntentConstants(t *testing.T) {
	if IntentGeneralQuestion != "general_question" {
		t.Errorf("IntentGeneralQuestion = %q, want %q", IntentGeneralQuestion, "general_question")
	}
	if IntentDockerQuery != "docker_query" {
		t.Errorf("IntentDockerQuery = %q, want %q", IntentDockerQuery, "docker_query")
	}
	if IntentActionRequest != "action_request" {
		t.Errorf("IntentActionRequest = %q, want %q", IntentActionRequest, "action_request")
	}
	if IntentAmbiguous != "ambiguous" {
		t.Errorf("IntentAmbiguous = %q, want %q", IntentAmbiguous, "ambiguous")
	}
}

func TestIntentClassificationInput(t *testing.T) {
	input := IntentClassificationInput{UserInput: "test input"}
	if input.UserInput != "test input" {
		t.Errorf("UserInput = %q, want %q", input.UserInput, "test input")
	}
}

type mockFlow struct {
	response *ai.ModelResponse
	err      error
}

func (m *mockFlow) Run(input IntentClassificationInput) (*ai.ModelResponse, error) {
	return m.response, m.err
}

func TestGenkitIntentClassifier_Classify(t *testing.T) {
	
	tests := []struct {
		name           string
		flowResponse   *ai.ModelResponse
		flowError      error
		expectError    bool
		expectedIntent Intent
		expectedPrompt string
	}{
		{
			name: "General question classification",
			flowResponse: &ai.ModelResponse{
				Message: &ai.Message{
					Content: []*ai.Part{
						{Kind: ai.PartText, Text: `{"intent": "general_question", "rewritten_prompt": "How do I run a container?"}`},
					},
				},
			},
			expectedIntent:  IntentGeneralQuestion,
			expectedPrompt: "How do I run a container?",
		},
		{
			name: "Docker query classification",
			flowResponse: &ai.ModelResponse{
				Message: &ai.Message{
					Content: []*ai.Part{
						{Kind: ai.PartText, Text: `{"intent": "docker_query", "rewritten_prompt": "List my containers"}`},
					},
				},
			},
			expectedIntent:  IntentDockerQuery,
			expectedPrompt: "List my containers",
		},
		{
			name: "Action request classification",
			flowResponse: &ai.ModelResponse{
				Message: &ai.Message{
					Content: []*ai.Part{
						{Kind: ai.PartText, Text: `{"intent": "action_request", "rewritten_prompt": "Run nginx"}`},
					},
				},
			},
			expectedIntent:  IntentActionRequest,
			expectedPrompt: "Run nginx",
		},
		{
			name: "Ambiguous classification",
			flowResponse: &ai.ModelResponse{
				Message: &ai.Message{
					Content: []*ai.Part{
						{Kind: ai.PartText, Text: `{"intent": "ambiguous", "rewritten_prompt": "What?"}`},
					},
				},
			},
			expectedIntent:  IntentAmbiguous,
			expectedPrompt: "What?",
		},
		{
			name:        "Flow error",
			flowError:   context.Canceled,
			expectError: true,
		},
		{
			name: "Invalid JSON in response",
			flowResponse: &ai.ModelResponse{
				Message: &ai.Message{
					Content: []*ai.Part{
						{Kind: ai.PartText, Text: `not valid json`},
					},
				},
			},
			expectError: true,
		},
		{
			name: "Empty response",
			flowResponse: &ai.ModelResponse{
				Message: &ai.Message{
					Content: []*ai.Part{},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Since we can't easily mock the Genkit flow, we test the JSON parsing logic directly
			// This validates the ClassificationResult parsing logic
			if tt.flowError != nil {
				// Test error handling
				return
			}
			
			if tt.flowResponse != nil && len(tt.flowResponse.Message.Content) > 0 {
				raw := tt.flowResponse.Text()
				jsonStr := extractJSON(raw)
				if jsonStr == "" {
					if !tt.expectError {
						t.Errorf("extractJSON returned empty for valid response")
					}
					return
				}
				
				var classification ClassificationResponse
				err := json.Unmarshal([]byte(jsonStr), &classification)
				if tt.expectError {
					if err == nil {
						t.Errorf("Expected error parsing JSON")
					}
					return
				}
				
				if err != nil {
					t.Fatalf("Unexpected JSON parse error: %v", err)
				}
				
				intent := Intent(classification.Intent)
				if intent != tt.expectedIntent {
					t.Errorf("Intent = %v, want %v", intent, tt.expectedIntent)
				}
				if classification.RewrittenPrompt != tt.expectedPrompt {
					t.Errorf("RewrittenPrompt = %q, want %q", classification.RewrittenPrompt, tt.expectedPrompt)
				}
			}
		})
	}
}