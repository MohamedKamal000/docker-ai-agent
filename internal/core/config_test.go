package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStringToProviderInfo(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ProviderInfo
	}{
		{
			name:     "Gemini provider",
			input:    "Gemini",
			expected: ProviderInfo{Provider: Gemini, EnvName: "GEMINI_API_KEY", PrefixName: "googleai"},
		},
		{
			name:     "OpenAI provider",
			input:    "OpenAi",
			expected: ProviderInfo{Provider: OpenAi, EnvName: "OPENAI_API_KEY", PrefixName: "openai"},
		},
		{
			name:     "Anthropic provider",
			input:    "Anthropic",
			expected: ProviderInfo{Provider: Anthropic, EnvName: "ANTHROPIC_API_KEY", PrefixName: "anthropic"},
		},
		{
			name:     "Ollama provider",
			input:    "Ollama",
			expected: ProviderInfo{Provider: Ollama, EnvName: "SERVER_ADDRESS", PrefixName: "ollama"},
		},
		{
			name:     "DeepSeek provider",
			input:    "DeepSeek",
			expected: ProviderInfo{Provider: Deepseek, EnvName: "DEEPSEEK_API_KEY", PrefixName: "deepseek"},
		},
		{
			name:     "Kimi provider",
			input:    "Kimi",
			expected: ProviderInfo{Provider: Kimi, EnvName: "KIMI_API_KEY", PrefixName: "moonshotai"},
		},
		{
			name:     "Qwen provider",
			input:    "Qwen",
			expected: ProviderInfo{Provider: Qwen, EnvName: "QWEN_API_KEY", PrefixName: "alibaba"},
		},
		{
			name:     "Grok provider",
			input:    "Grok",
			expected: ProviderInfo{Provider: Grok, EnvName: "GROK_API_KEY", PrefixName: "xai"},
		},
		{
			name:     "Unknown provider",
			input:    "UnknownProvider",
			expected: ProviderInfo{Provider: Unknown},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stringToProviderInfo(tt.input)
			if result.Provider != tt.expected.Provider {
				t.Errorf("stringToProviderInfo(%q) = %v, want %v", tt.input, result.Provider, tt.expected.Provider)
			}
			if result.EnvName != tt.expected.EnvName {
				t.Errorf("stringToProviderInfo(%q) EnvName = %q, want %q", tt.input, result.EnvName, tt.expected.EnvName)
			}
			if result.PrefixName != tt.expected.PrefixName {
				t.Errorf("stringToProviderInfo(%q) PrefixName = %q, want %q", tt.input, result.PrefixName, tt.expected.PrefixName)
			}
		})
	}
}

func TestCheckFileExists(t *testing.T) {
	tmpDir := t.TempDir()

	existingFile := filepath.Join(tmpDir, "exists.txt")
	if err := os.WriteFile(existingFile, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	nonExistingFile := filepath.Join(tmpDir, "notexists.txt")

	if !checkFileExists(existingFile) {
		t.Error("checkFileExists(existingFile) = false, want true")
	}

	if checkFileExists(nonExistingFile) {
		t.Error("checkFileExists(nonExistingFile) = true, want false")
	}
}

func TestFallBackToCurrentDir(t *testing.T) {
	result, err := fallBackToCurrentDir()
	if err != nil {
		t.Fatalf("fallBackToCurrentDir() returned error: %v", err)
	}

	if result == "" {
		t.Error("fallBackToCurrentDir() returned empty string")
	}

	expectedSuffix := "config.json"
	if len(result) < len(expectedSuffix) || result[len(result)-len(expectedSuffix):] != expectedSuffix {
		t.Errorf("fallBackToCurrentDir() = %q, want suffix %q", result, expectedSuffix)
	}
}

func TestRAGConfigValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      *RAGConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid local CPU config",
			config: &RAGConfig{
				EmbeddingType:   EmbeddingLocal,
				InferenceType:   CPU,
				ChunkSize:       1000,
				OverlapSize:     100,
				WorkersNumber:   4,
				ModelName:       "test-model",
				EmbeddingApiKey: "",
			},
			expectError: false,
		},
		{
			name: "Valid local GPU config",
			config: &RAGConfig{
				EmbeddingType:   EmbeddingLocal,
				InferenceType:   GPU,
				ChunkSize:       1000,
				OverlapSize:     100,
				WorkersNumber:   4,
				ModelName:       "test-model",
				EmbeddingApiKey: "",
			},
			expectError: false,
		},
		{
			name: "Invalid inference type for local",
			config: &RAGConfig{
				EmbeddingType:   EmbeddingLocal,
				InferenceType:   "invalid",
				ChunkSize:       1000,
				OverlapSize:     100,
				WorkersNumber:   4,
				ModelName:       "test-model",
				EmbeddingApiKey: "",
			},
			expectError: true,
			errorMsg:    "invalid inference type",
		},
		{
			name: "Local with embedding api key should fail",
			config: &RAGConfig{
				EmbeddingType:   EmbeddingLocal,
				InferenceType:   CPU,
				ChunkSize:       1000,
				OverlapSize:     100,
				WorkersNumber:   4,
				ModelName:       "test-model",
				EmbeddingApiKey: "should-not-be-set",
			},
			expectError: true,
			errorMsg:    "embedding api key cannot be set when using local embeddings",
		},
		{
			name: "Valid remote config with api key",
			config: &RAGConfig{
				EmbeddingType:   EmbeddingRemote,
				InferenceType:   "",
				ChunkSize:       1000,
				OverlapSize:     100,
				WorkersNumber:   4,
				ModelName:       "test-model",
				EmbeddingApiKey: "test-key",
			},
			expectError: false,
		},
		{
			name: "Remote config with inference type should fail",
			config: &RAGConfig{
				EmbeddingType:   EmbeddingRemote,
				InferenceType:   CPU,
				ChunkSize:       1000,
				OverlapSize:     100,
				WorkersNumber:   4,
				ModelName:       "test-model",
				EmbeddingApiKey: "test-key",
			},
			expectError: true,
			errorMsg:    "inference type is only valid for local embeddings",
		},
		{
			name: "Invalid embedding type",
			config: &RAGConfig{
				EmbeddingType:   "invalid",
				InferenceType:   CPU,
				ChunkSize:       1000,
				OverlapSize:     100,
				WorkersNumber:   4,
				ModelName:       "test-model",
				EmbeddingApiKey: "",
			},
			expectError: true,
			errorMsg:    "invalid embedding type",
		},
		{
			name: "Chunk size <= 0 should fail",
			config: &RAGConfig{
				EmbeddingType:   EmbeddingLocal,
				InferenceType:   CPU,
				ChunkSize:       0,
				OverlapSize:     100,
				WorkersNumber:   4,
				ModelName:       "test-model",
				EmbeddingApiKey: "",
			},
			expectError: true,
			errorMsg:    "chunk size must be greater than 0",
		},
		{
			name: "Negative overlap size should fail",
			config: &RAGConfig{
				EmbeddingType:   EmbeddingLocal,
				InferenceType:   CPU,
				ChunkSize:       1000,
				OverlapSize:     -1,
				WorkersNumber:   4,
				ModelName:       "test-model",
				EmbeddingApiKey: "",
			},
			expectError: true,
			errorMsg:    "overlap size cannot be negative",
		},
		{
			name: "Overlap >= chunk size should fail",
			config: &RAGConfig{
				EmbeddingType:   EmbeddingLocal,
				InferenceType:   CPU,
				ChunkSize:       1000,
				OverlapSize:     1000,
				WorkersNumber:   4,
				ModelName:       "test-model",
				EmbeddingApiKey: "",
			},
			expectError: true,
			errorMsg:    "overlap size must be smaller than chunk size",
		},
		{
			name: "Workers number <= 0 should fail",
			config: &RAGConfig{
				EmbeddingType:   EmbeddingLocal,
				InferenceType:   CPU,
				ChunkSize:       1000,
				OverlapSize:     100,
				WorkersNumber:   0,
				ModelName:       "test-model",
				EmbeddingApiKey: "",
			},
			expectError: true,
			errorMsg:    "workers number must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				if err == nil {
					t.Errorf("Validate() = nil, want error containing %q", tt.errorMsg)
				} else if !containsString(err.Error(), tt.errorMsg) {
					t.Errorf("Validate() error = %q, want error containing %q", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() = %v, want nil", err)
				}
			}
		})
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestModelConfigFromJsonFile(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		configJSON  string
		expectError bool
		errorMsg    string
		checkConfig func(t *testing.T, config AppConfig)
		setupEnv    func() func()
	}{
		{
			name: "Valid Gemini config",
			configJSON: `{
				"provider": "Gemini",
				"model-name": "gemini-2.5-flash-lite",
				"temperature": 0.7,
				"max-tokens": 1024,
				"max-iterations": 10
			}`,
			expectError: false,
			checkConfig: func(t *testing.T, config AppConfig) {
				if config.Provider != "Gemini" {
					t.Errorf("Provider = %q, want %q", config.Provider, "Gemini")
				}
				if config.ModelName != "googleai/gemini-2.5-flash-lite" {
					t.Errorf("ModelName = %q, want %q", config.ModelName, "googleai/gemini-2.5-flash-lite")
				}
				if config.Temperature != 0.7 {
					t.Errorf("Temperature = %v, want %v", config.Temperature, 0.7)
				}
				if config.MaxTokens != 1024 {
					t.Errorf("MaxTokens = %v, want %v", config.MaxTokens, 1024)
				}
				if config.MaxIterations != 10 {
					t.Errorf("MaxIterations = %v, want %v", config.MaxIterations, 10)
				}
			},
			setupEnv: func() func() {
				os.Setenv("GEMINI_API_KEY", "test-gemini-key")
				return func() { os.Unsetenv("GEMINI_API_KEY") }
			},
		},
		{
			name: "Valid OpenAI config with prefix",
			configJSON: `{
				"provider": "OpenAi",
				"model-name": "gpt-4o-mini",
				"temperature": 0.5,
				"max-tokens": 2048,
				"max-iterations": 5
			}`,
			expectError: false,
			checkConfig: func(t *testing.T, config AppConfig) {
				if config.ModelName != "openai/gpt-4o-mini" {
					t.Errorf("ModelName = %q, want %q", config.ModelName, "openai/gpt-4o-mini")
				}
			},
			setupEnv: func() func() {
				os.Setenv("OPENAI_API_KEY", "test-openai-key")
				return func() { os.Unsetenv("OPENAI_API_KEY") }
			},
		},
		{
			name: "Valid Ollama config (no api key needed)",
			configJSON: `{
				"provider": "Ollama",
				"model-name": "llama3",
				"server-address": "http://localhost:11434",
				"temperature": 0.7,
				"max-tokens": 1024,
				"max-iterations": 10
			}`,
			expectError: false,
			checkConfig: func(t *testing.T, config AppConfig) {
				if config.Provider != "Ollama" {
					t.Errorf("Provider = %q, want %q", config.Provider, "Ollama")
				}
				if config.ServerAdress != "http://localhost:11434" {
					t.Errorf("ServerAdress = %q, want %q", config.ServerAdress, "http://localhost:11434")
				}
				if config.ModelName != "ollama/llama3" {
					t.Errorf("ModelName = %q, want %q", config.ModelName, "ollama/llama3")
				}
			},
		},
		{
			name: "Missing model name should fail",
			configJSON: `{
				"provider": "Gemini",
				"temperature": 0.7,
				"max-tokens": 1024,
				"max-iterations": 10
			}`,
			expectError: true,
			errorMsg:    "model name can't be empty",
		},
		{
			name: "Unknown provider should fail",
			configJSON: `{
				"provider": "UnknownProvider",
				"model-name": "some-model",
				"temperature": 0.7,
				"max-tokens": 1024,
				"max-iterations": 10
			}`,
			expectError: true,
			errorMsg:    "provider UnknownProvider is not supported",
		},
		{
			name: "Negative max iterations should fail",
			configJSON: `{
				"provider": "Gemini",
				"model-name": "gemini-2.5-flash-lite",
				"temperature": 0.7,
				"max-tokens": 1024,
				"max-iterations": -1
			}`,
			expectError: true,
			errorMsg:    "max iterations can't be negative",
		},
		{
			name: "Valid config with RAG",
			configJSON: `{
				"provider": "Gemini",
				"model-name": "gemini-2.5-flash-lite",
				"temperature": 0.7,
				"max-tokens": 1024,
				"max-iterations": 10,
				"rag": {
					"embedding-type": "local",
					"inference-type": "cpu",
					"chunk-size": 1000,
					"overlap-size": 100,
					"workers-number": 4,
					"model-name": "embedding-model",
					"embedding-api-key": ""
				}
			}`,
			expectError: false,
			checkConfig: func(t *testing.T, config AppConfig) {
				if config.RagConfig == nil {
					t.Fatal("RagConfig is nil")
				}
				if config.RagConfig.EmbeddingType != EmbeddingLocal {
					t.Errorf("RagConfig.EmbeddingType = %v, want %v", config.RagConfig.EmbeddingType, EmbeddingLocal)
				}
				if config.RagConfig.InferenceType != CPU {
					t.Errorf("RagConfig.InferenceType = %v, want %v", config.RagConfig.InferenceType, CPU)
				}
			},
			setupEnv: func() func() {
				os.Setenv("GEMINI_API_KEY", "test-gemini-key")
				return func() { os.Unsetenv("GEMINI_API_KEY") }
			},
		},
		{
			name: "Invalid RAG config should fail",
			configJSON: `{
				"provider": "Gemini",
				"model-name": "gemini-2.5-flash-lite",
				"temperature": 0.7,
				"max-tokens": 1024,
				"max-iterations": 10,
				"rag": {
					"embedding-type": "local",
					"inference-type": "cpu",
					"chunk-size": 0,
					"overlap-size": 100,
					"workers-number": 4,
					"model-name": "embedding-model",
					"embedding-api-key": ""
				}
			}`,
			expectError: true,
			errorMsg:    "chunk size must be greater than 0",
			setupEnv: func() func() {
				os.Setenv("GEMINI_API_KEY", "test-gemini-key")
				return func() { os.Unsetenv("GEMINI_API_KEY") }
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cleanup func()
			if tt.setupEnv != nil {
				cleanup = tt.setupEnv()
				defer cleanup()
			}
			configFile := filepath.Join(tmpDir, "config.json")
			if err := os.WriteFile(configFile, []byte(tt.configJSON), 0644); err != nil {
				t.Fatalf("failed to write config file: %v", err)
			}

			config, err := ModelConfigFromJsonFile(configFile)
			if tt.expectError {
				if err == nil {
					t.Errorf("ModelConfigFromJsonFile() = nil, want error containing %q", tt.errorMsg)
				} else if !containsString(err.Error(), tt.errorMsg) {
					t.Errorf("ModelConfigFromJsonFile() error = %q, want error containing %q", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Fatalf("ModelConfigFromJsonFile() = %v, want nil", err)
				}
				if tt.checkConfig != nil {
					tt.checkConfig(t, config)
				}
			}
		})
	}
}

func TestModelConfigFromJsonFile_Fallback(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	configJSON := `{
		"provider": "Gemini",
		"model-name": "gemini-2.5-flash-lite",
		"temperature": 0.7,
		"max-tokens": 1024,
		"max-iterations": 10
	}`

	if err := os.WriteFile("config.json", []byte(configJSON), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	os.Setenv("GEMINI_API_KEY", "test-gemini-key")
	defer os.Unsetenv("GEMINI_API_KEY")

	config, err := ModelConfigFromJsonFile("nonexistent.json")
	if err != nil {
		t.Fatalf("ModelConfigFromJsonFile() with fallback = %v, want nil", err)
	}
	if config.ModelName != "googleai/gemini-2.5-flash-lite" {
		t.Errorf("ModelName = %q, want %q", config.ModelName, "googleai/gemini-2.5-flash-lite")
	}
}

func TestModelConfigFromJsonFile_ApiKeyFromEnv(t *testing.T) {
	tmpDir := t.TempDir()

	configJSON := `{
		"provider": "Gemini",
		"model-name": "gemini-2.5-flash-lite",
		"temperature": 0.7,
		"max-tokens": 1024,
		"max-iterations": 10
	}`

	configFile := filepath.Join(tmpDir, "config.json")
	if err := os.WriteFile(configFile, []byte(configJSON), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	os.Setenv("GEMINI_API_KEY", "test-api-key-from-env")
	defer os.Unsetenv("GEMINI_API_KEY")

	config, err := ModelConfigFromJsonFile(configFile)
	if err != nil {
		t.Fatalf("ModelConfigFromJsonFile() = %v, want nil", err)
	}
	if config.ApiKey != "test-api-key-from-env" {
		t.Errorf("ApiKey = %q, want %q", config.ApiKey, "test-api-key-from-env")
	}
}

func TestModelConfigFromJsonFile_RemoteRAGApiKeyFromEnv(t *testing.T) {
	tmpDir := t.TempDir()

	configJSON := `{
		"provider": "Gemini",
		"model-name": "gemini-2.5-flash-lite",
		"temperature": 0.7,
		"max-tokens": 1024,
		"max-iterations": 10,
		"rag": {
			"embedding-type": "remote",
			"inference-type": "",
			"chunk-size": 1000,
			"overlap-size": 100,
			"workers-number": 4,
			"model-name": "embedding-model",
			"embedding-api-key": ""
		}
	}`

	configFile := filepath.Join(tmpDir, "config.json")
	if err := os.WriteFile(configFile, []byte(configJSON), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	os.Setenv("GEMINI_API_KEY", "test-gemini-key")
	defer os.Unsetenv("GEMINI_API_KEY")
	os.Setenv("EMBEDDING_API_KEY", "test-embedding-key-from-env")
	defer os.Unsetenv("EMBEDDING_API_KEY")

	config, err := ModelConfigFromJsonFile(configFile)
	if err != nil {
		t.Fatalf("ModelConfigFromJsonFile() = %v, want nil", err)
	}
	if config.RagConfig == nil {
		t.Fatal("RagConfig is nil")
	}
	if config.RagConfig.EmbeddingApiKey != "test-embedding-key-from-env" {
		t.Errorf("RagConfig.EmbeddingApiKey = %q, want %q", config.RagConfig.EmbeddingApiKey, "test-embedding-key-from-env")
	}
}

func TestModelConfigFromJsonFile_MissingRemoteRAGApiKey(t *testing.T) {
	tmpDir := t.TempDir()

	configJSON := `{
		"provider": "Gemini",
		"model-name": "gemini-2.5-flash-lite",
		"temperature": 0.7,
		"max-tokens": 1024,
		"max-iterations": 10,
		"rag": {
			"embedding-type": "remote",
			"inference-type": "",
			"chunk-size": 1000,
			"overlap-size": 100,
			"workers-number": 4,
			"model-name": "embedding-model",
			"embedding-api-key": ""
		}
	}`

	configFile := filepath.Join(tmpDir, "config.json")
	if err := os.WriteFile(configFile, []byte(configJSON), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	os.Setenv("GEMINI_API_KEY", "test-gemini-key")
	defer os.Unsetenv("GEMINI_API_KEY")
	os.Unsetenv("EMBEDDING_API_KEY")

	_, err := ModelConfigFromJsonFile(configFile)
	if err == nil {
		t.Error("ModelConfigFromJsonFile() = nil, want error about missing embedding api key")
	} else if !containsString(err.Error(), "embedding api key not found") {
		t.Errorf("ModelConfigFromJsonFile() error = %q, want error containing %q", err.Error(), "embedding api key not found")
	}
}