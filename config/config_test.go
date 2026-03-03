package config

import (
	"strings"
	"testing"
)

func validConfig() Config {
	model := ModelConfig{
		Model:       "gpt-4.1",
		APIKey:      "key",
		BaseURL:     "https://example.com/v1",
		MaxTokens:   1024,
		TopP:        1,
		Temperature: 0.2,
	}
	return Config{
		Agents: AgentModelConfig{
			Main:  model,
			Image: model,
			Video: model,
			Audio: model,
		},
		Embedding: EmbeddingConfig{
			Model:   "text-embedding-3-large",
			APIKey:  "ekey",
			BaseURL: "https://example.com/v1",
		},
		Prompts: PromptConfig{
			Main:  "m",
			Image: "i",
			Video: "v",
			Audio: "a",
		},
		Retry: RetryConfig{
			MaxRetries:   3,
			RetryDelayMS: 10,
		},
	}
}

func TestConfigNormalize(t *testing.T) {
	cfg := validConfig()
	cfg.Agents.Main.APIKey = "  key  "
	cfg.Embedding.BaseURL = "  https://example.com/v1  "
	cfg.Prompts.Main = "  prompt  "
	cfg.normalize()

	if cfg.Agents.Main.APIKey != "key" {
		t.Fatalf("expected trimmed APIKey, got %q", cfg.Agents.Main.APIKey)
	}
	if cfg.Embedding.BaseURL != "https://example.com/v1" {
		t.Fatalf("expected trimmed embedding base url, got %q", cfg.Embedding.BaseURL)
	}
	if cfg.Prompts.Main != "prompt" {
		t.Fatalf("expected trimmed prompt, got %q", cfg.Prompts.Main)
	}
}

func TestConfigValidate(t *testing.T) {
	cases := []struct {
		name      string
		modify    func(*Config)
		wantInErr string
	}{
		{
			name: "invalid retry max",
			modify: func(c *Config) {
				c.Retry.MaxRetries = 0
			},
			wantInErr: "retry.max_retries",
		},
		{
			name: "missing main api key",
			modify: func(c *Config) {
				c.Agents.Main.APIKey = ""
			},
			wantInErr: "agents.main.api_key",
		},
		{
			name: "missing embedding base url",
			modify: func(c *Config) {
				c.Embedding.BaseURL = ""
			},
			wantInErr: "embedding.base_url",
		},
		{
			name: "empty prompt",
			modify: func(c *Config) {
				c.Prompts.Audio = "   "
			},
			wantInErr: "prompts.audio",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := validConfig()
			tc.modify(&cfg)
			err := cfg.validate()
			if err == nil {
				t.Fatalf("expected validation error")
			}
			if !strings.Contains(err.Error(), tc.wantInErr) {
				t.Fatalf("unexpected validation error: %v", err)
			}
		})
	}
}

