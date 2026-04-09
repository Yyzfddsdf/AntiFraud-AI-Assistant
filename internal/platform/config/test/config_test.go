package config_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	appcfg "antifraud/internal/platform/config"
)

func validConfig() appcfg.Config {
	model := appcfg.ModelConfig{
		Model:       "gpt-4.1",
		APIKey:      "key",
		BaseURL:     "https://example.com/v1",
		MaxTokens:   1024,
		TopP:        1,
		Temperature: 0.2,
	}
	return appcfg.Config{
		Agents: appcfg.AgentModelConfig{
			Main:           model,
			Image:          model,
			ImageQuick:     model,
			Video:          model,
			Audio:          model,
			ASR:            model,
			CaseCollection: model,
		},
		Embedding: appcfg.EmbeddingConfig{
			Model:   "text-embedding-3-large",
			APIKey:  "ekey",
			BaseURL: "https://example.com/v1",
		},
		Chat: appcfg.ChatConfig{
			Prompt:  "chat prompt",
			Model:   "chat-model",
			APIKey:  "ckey",
			BaseURL: "https://chat.example.com/v1",
		},
		Prompts: appcfg.PromptConfig{
			Main:           "m",
			Image:          "i",
			ImageQuick:     "iq",
			Video:          "v",
			Audio:          "a",
			CaseCollection: "c",
		},
		Retry: appcfg.RetryConfig{
			MaxRetries:   3,
			RetryDelayMS: 10,
		},
	}
}

func writeConfigFile(t *testing.T, cfg appcfg.Config) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.json")
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal config failed: %v", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write config failed: %v", err)
	}
	return path
}

func TestConfigNormalize(t *testing.T) {
	cfg := validConfig()
	cfg.Agents.Main.APIKey = "  key  "
	cfg.Agents.Video.BaseURL = "  https://video.example.com  "
	cfg.Embedding.BaseURL = "  https://example.com/v1  "
	cfg.Prompts.Main = "  prompt  "

	file := writeConfigFile(t, cfg)
	loaded, err := appcfg.LoadConfig(file)
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}

	if loaded.Agents.Main.APIKey != "key" {
		t.Fatalf("expected trimmed APIKey, got %q", loaded.Agents.Main.APIKey)
	}
	if loaded.Embedding.BaseURL != "https://example.com/v1" {
		t.Fatalf("expected trimmed embedding base url, got %q", loaded.Embedding.BaseURL)
	}
	if loaded.Agents.Video.BaseURL != "https://video.example.com" {
		t.Fatalf("expected trimmed video base url, got %q", loaded.Agents.Video.BaseURL)
	}
	if loaded.Prompts.Main != "prompt" {
		t.Fatalf("expected trimmed prompt, got %q", loaded.Prompts.Main)
	}
}

func TestLoadConfigSupportsUTF8BOM(t *testing.T) {
	cfg := validConfig()
	path := filepath.Join(t.TempDir(), "config-with-bom.json")

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal config failed: %v", err)
	}

	withBOM := append([]byte{0xEF, 0xBB, 0xBF}, data...)
	if err := os.WriteFile(path, withBOM, 0o600); err != nil {
		t.Fatalf("write config with bom failed: %v", err)
	}

	loaded, err := appcfg.LoadConfig(path)
	if err != nil {
		t.Fatalf("load config with bom failed: %v", err)
	}

	if loaded.Agents.Main.Model != cfg.Agents.Main.Model {
		t.Fatalf("expected main model %q, got %q", cfg.Agents.Main.Model, loaded.Agents.Main.Model)
	}
}

func TestConfigChatFallbackToMainModel(t *testing.T) {
	cfg := validConfig()
	cfg.Chat = appcfg.ChatConfig{Prompt: "chat prompt"}

	file := writeConfigFile(t, cfg)
	loaded, err := appcfg.LoadConfig(file)
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}

	if loaded.Chat.Model != cfg.Agents.Main.Model {
		t.Fatalf("expected chat model fallback to agents.main.model, got %q", loaded.Chat.Model)
	}
	if loaded.Chat.APIKey != cfg.Agents.Main.APIKey {
		t.Fatalf("expected chat api_key fallback to agents.main.api_key, got %q", loaded.Chat.APIKey)
	}
	if loaded.Chat.BaseURL != cfg.Agents.Main.BaseURL {
		t.Fatalf("expected chat base_url fallback to agents.main.base_url, got %q", loaded.Chat.BaseURL)
	}
}

func TestConfigEnvOverridesAPIKeys(t *testing.T) {
	t.Setenv("AGENT_MAIN_API_KEY", "env-main-key")
	t.Setenv("CHAT_API_KEY", "env-chat-key")
	t.Setenv("EMBEDDING_API_KEY", "env-embedding-key")
	t.Setenv("TAVILY_API_KEY", "env-tavily-key")

	cfg := validConfig()
	cfg.Tavily = appcfg.TavilyConfig{
		APIKey: "config-tavily-key",
	}

	file := writeConfigFile(t, cfg)
	loaded, err := appcfg.LoadConfig(file)
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}

	if loaded.Agents.Main.APIKey != "env-main-key" {
		t.Fatalf("expected env override for agents.main.api_key, got %q", loaded.Agents.Main.APIKey)
	}
	if loaded.Chat.APIKey != "env-chat-key" {
		t.Fatalf("expected env override for chat.api_key, got %q", loaded.Chat.APIKey)
	}
	if loaded.Embedding.APIKey != "env-embedding-key" {
		t.Fatalf("expected env override for embedding.api_key, got %q", loaded.Embedding.APIKey)
	}
	if loaded.Tavily.APIKey != "env-tavily-key" {
		t.Fatalf("expected env override for tavily.api_key, got %q", loaded.Tavily.APIKey)
	}
}

func TestConfigFallsBackToFileAPIKeysWhenEnvMissing(t *testing.T) {
	cfg := validConfig()
	cfg.Agents.Image.APIKey = "config-image-key"
	cfg.Chat.APIKey = "config-chat-key"

	file := writeConfigFile(t, cfg)
	loaded, err := appcfg.LoadConfig(file)
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}

	if loaded.Agents.Image.APIKey != "config-image-key" {
		t.Fatalf("expected config fallback for agents.image.api_key, got %q", loaded.Agents.Image.APIKey)
	}
	if loaded.Chat.APIKey != "config-chat-key" {
		t.Fatalf("expected config fallback for chat.api_key, got %q", loaded.Chat.APIKey)
	}
}

func TestConfigValidate(t *testing.T) {
	cases := []struct {
		name      string
		modify    func(*appcfg.Config)
		wantInErr string
	}{
		{
			name: "invalid retry max",
			modify: func(c *appcfg.Config) {
				c.Retry.MaxRetries = 0
			},
			wantInErr: "retry.max_retries",
		},
		{
			name: "invalid retry delay",
			modify: func(c *appcfg.Config) {
				c.Retry.RetryDelayMS = 0
			},
			wantInErr: "retry.retry_delay_ms",
		},
		{
			name: "missing main api key",
			modify: func(c *appcfg.Config) {
				c.Agents.Main.APIKey = ""
			},
			wantInErr: "agents.main.api_key",
		},
		{
			name: "missing image api key",
			modify: func(c *appcfg.Config) {
				c.Agents.Image.APIKey = ""
			},
			wantInErr: "agents.image.api_key",
		},
		{
			name: "missing image quick api key",
			modify: func(c *appcfg.Config) {
				c.Agents.ImageQuick.APIKey = ""
			},
			wantInErr: "agents.image_quick.api_key",
		},
		{
			name: "missing embedding base url",
			modify: func(c *appcfg.Config) {
				c.Embedding.BaseURL = ""
			},
			wantInErr: "embedding.base_url",
		},
		{
			name: "missing case collection api key",
			modify: func(c *appcfg.Config) {
				c.Agents.CaseCollection.APIKey = ""
			},
			wantInErr: "agents.case_collection.api_key",
		},
		{
			name: "empty image quick prompt",
			modify: func(c *appcfg.Config) {
				c.Prompts.ImageQuick = "   "
			},
			wantInErr: "prompts.image_quick",
		},
		{
			name: "empty prompt",
			modify: func(c *appcfg.Config) {
				c.Prompts.Audio = "   "
			},
			wantInErr: "prompts.audio",
		},
		{
			name: "empty case collection prompt",
			modify: func(c *appcfg.Config) {
				c.Prompts.CaseCollection = "   "
			},
			wantInErr: "prompts.case_collection",
		},
		{
			name: "empty chat prompt",
			modify: func(c *appcfg.Config) {
				c.Chat.Prompt = "   "
			},
			wantInErr: "chat.prompt",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := validConfig()
			tc.modify(&cfg)
			file := writeConfigFile(t, cfg)

			_, err := appcfg.LoadConfig(file)
			if err == nil {
				t.Fatalf("expected validation error")
			}
			if !strings.Contains(err.Error(), tc.wantInErr) {
				t.Fatalf("unexpected validation error: %v", err)
			}
		})
	}
}

func TestConfigNormalizeTavilyDefaults(t *testing.T) {
	cfg := validConfig()
	cfg.Tavily = appcfg.TavilyConfig{
		APIKey: "  tavily-key  ",
	}

	file := writeConfigFile(t, cfg)
	loaded, err := appcfg.LoadConfig(file)
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}

	if loaded.Tavily.APIKey != "tavily-key" {
		t.Fatalf("expected trimmed tavily api key, got %q", loaded.Tavily.APIKey)
	}
	if loaded.Tavily.BaseURL != "https://api.tavily.com" {
		t.Fatalf("expected default tavily base url, got %q", loaded.Tavily.BaseURL)
	}
	if loaded.Tavily.IncludeAnswer != "advanced" {
		t.Fatalf("expected default include_answer, got %q", loaded.Tavily.IncludeAnswer)
	}
	if loaded.Tavily.SearchDepth != "advanced" {
		t.Fatalf("expected default search_depth, got %q", loaded.Tavily.SearchDepth)
	}
	if loaded.Tavily.TimeoutMS != 15000 {
		t.Fatalf("expected default timeout_ms, got %d", loaded.Tavily.TimeoutMS)
	}
}

func TestConfigValidateTavilyRequiresAPIKeyWhenConfigured(t *testing.T) {
	cfg := validConfig()
	cfg.Tavily = appcfg.TavilyConfig{
		BaseURL: "https://proxy.example.com",
	}

	file := writeConfigFile(t, cfg)
	_, err := appcfg.LoadConfig(file)
	if err == nil {
		t.Fatal("expected tavily validation error")
	}
	if !strings.Contains(err.Error(), "tavily.api_key") {
		t.Fatalf("unexpected validation error: %v", err)
	}
}
