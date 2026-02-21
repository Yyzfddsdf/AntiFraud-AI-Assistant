package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type Config struct {
	APIKey    string `json:"api_key"`
	BaseURL   string `json:"base_url"`
	ChatModel string `json:"chat_model"`
	RedisAddr string `json:"redis_addr"`
	RedisPwd  string `json:"redis_password"`
	RedisDB   int    `json:"redis_db"`
}

func LoadConfig(path string) (*Config, error) {
	resolvedPath := resolveConfigPath(path)
	file, err := os.Open(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file (%s): %w", resolvedPath, err)
	}
	defer file.Close()

	var cfg Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	if cfg.RedisAddr == "" {
		cfg.RedisAddr = "127.0.0.1:6379"
	}

	return &cfg, nil
}

func resolveConfigPath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}

	if _, err := os.Stat(path); err == nil {
		return path
	}

	_, currentFile, _, ok := runtime.Caller(0)
	if ok {
		projectRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", ".."))
		candidate := filepath.Join(projectRoot, path)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return path
}
