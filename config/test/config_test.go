package config_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	appcfg "antifraud/config"
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
			Main:  model,
			Image: model,
			Video: model,
			Audio: model,
		},
		Embedding: appcfg.EmbeddingConfig{
			Model:   "text-embedding-3-large",
			APIKey:  "ekey",
			BaseURL: "https://example.com/v1",
		},
		Prompts: appcfg.PromptConfig{
			Main:  "m",
			Image: "i",
			Video: "v",
			Audio: "a",
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
			name: "missing embedding base url",
			modify: func(c *appcfg.Config) {
				c.Embedding.BaseURL = ""
			},
			wantInErr: "embedding.base_url",
		},
		{
			name: "empty prompt",
			modify: func(c *appcfg.Config) {
				c.Prompts.Audio = "   "
			},
			wantInErr: "prompts.audio",
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
