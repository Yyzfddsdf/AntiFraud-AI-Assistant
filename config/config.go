package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type ModelConfig struct {
	Model       string  `json:"model"`
	APIKey      string  `json:"api_key"`
	BaseURL     string  `json:"base_url"`
	MaxTokens   int     `json:"max_tokens"`
	TopP        float64 `json:"top_p"`
	Temperature float64 `json:"temperature"`
}

type EmbeddingConfig struct {
	Model   string `json:"model"`
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url"`
}

// RetryConfig 定义通用重试策略。
type RetryConfig struct {
	MaxRetries   int `json:"max_retries"`
	RetryDelayMS int `json:"retry_delay_ms"`
}

// AgentModelConfig 按智能体拆分模型与调用参数，便于后续扩展新 provider/model。
type AgentModelConfig struct {
	Main  ModelConfig `json:"main"`
	Image ModelConfig `json:"image"`
	Video ModelConfig `json:"video"`
	Audio ModelConfig `json:"audio"`
}

// PromptConfig 统一托管各智能体提示词，避免硬编码散落在代码里。
type PromptConfig struct {
	Main  string `json:"main"`
	Image string `json:"image"`
	Video string `json:"video"`
	Audio string `json:"audio"`
}

// Config 是项目总配置对象。
type Config struct {
	Agents    AgentModelConfig `json:"agents"`
	Embedding EmbeddingConfig  `json:"embedding"`
	Prompts   PromptConfig     `json:"prompts"`
	Retry     RetryConfig      `json:"retry"`
}

// LoadConfig 负责读取、标准化并校验配置文件。
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

	cfg.normalize()
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// normalize 统一裁剪字符串空白，减少运行期参数格式问题。
func (c *Config) normalize() {
	c.Agents.Main = normalizeModel(c.Agents.Main)
	c.Agents.Image = normalizeModel(c.Agents.Image)
	c.Agents.Video = normalizeModel(c.Agents.Video)
	c.Agents.Audio = normalizeModel(c.Agents.Audio)
	c.Embedding = normalizeEmbedding(c.Embedding)
	c.Prompts.Main = strings.TrimSpace(c.Prompts.Main)
	c.Prompts.Image = strings.TrimSpace(c.Prompts.Image)
	c.Prompts.Video = strings.TrimSpace(c.Prompts.Video)
	c.Prompts.Audio = strings.TrimSpace(c.Prompts.Audio)
}

// normalizeModel 处理单个模型配置的字符串规范化。
func normalizeModel(modelCfg ModelConfig) ModelConfig {
	modelCfg.APIKey = strings.TrimSpace(modelCfg.APIKey)
	modelCfg.BaseURL = strings.TrimSpace(modelCfg.BaseURL)
	modelCfg.Model = strings.TrimSpace(modelCfg.Model)
	return modelCfg
}

func normalizeEmbedding(embeddingCfg EmbeddingConfig) EmbeddingConfig {
	embeddingCfg.APIKey = strings.TrimSpace(embeddingCfg.APIKey)
	embeddingCfg.BaseURL = strings.TrimSpace(embeddingCfg.BaseURL)
	embeddingCfg.Model = strings.TrimSpace(embeddingCfg.Model)
	return embeddingCfg
}

// validate 校验整体配置完整性。
func (c Config) validate() error {
	if c.Retry.MaxRetries <= 0 {
		return fmt.Errorf("invalid retry.max_retries: must be > 0")
	}
	if c.Retry.RetryDelayMS <= 0 {
		return fmt.Errorf("invalid retry.retry_delay_ms: must be > 0")
	}

	if err := validateModel("agents.main", c.Agents.Main); err != nil {
		return err
	}
	if err := validateModel("agents.image", c.Agents.Image); err != nil {
		return err
	}
	if err := validateModel("agents.video", c.Agents.Video); err != nil {
		return err
	}
	if err := validateModel("agents.audio", c.Agents.Audio); err != nil {
		return err
	}
	if err := validateEmbedding("embedding", c.Embedding); err != nil {
		return err
	}
	if err := validatePrompt("prompts.main", c.Prompts.Main); err != nil {
		return err
	}
	if err := validatePrompt("prompts.image", c.Prompts.Image); err != nil {
		return err
	}
	if err := validatePrompt("prompts.video", c.Prompts.Video); err != nil {
		return err
	}
	if err := validatePrompt("prompts.audio", c.Prompts.Audio); err != nil {
		return err
	}
	return nil
}

// validateModel 校验单个模型配置字段是否合法。
func validateModel(name string, modelCfg ModelConfig) error {
	if modelCfg.Model == "" {
		return fmt.Errorf("%s.model is required", name)
	}
	if modelCfg.APIKey == "" {
		return fmt.Errorf("%s.api_key is required", name)
	}
	if modelCfg.BaseURL == "" {
		return fmt.Errorf("%s.base_url is required", name)
	}
	if modelCfg.MaxTokens <= 0 {
		return fmt.Errorf("%s.max_tokens must be > 0", name)
	}
	if modelCfg.TopP <= 0 {
		return fmt.Errorf("%s.top_p must be > 0", name)
	}
	if modelCfg.Temperature < 0 {
		return fmt.Errorf("%s.temperature must be >= 0", name)
	}
	return nil
}

func validateEmbedding(name string, embeddingCfg EmbeddingConfig) error {
	if embeddingCfg.Model == "" {
		return fmt.Errorf("%s.model is required", name)
	}
	if embeddingCfg.APIKey == "" {
		return fmt.Errorf("%s.api_key is required", name)
	}
	if embeddingCfg.BaseURL == "" {
		return fmt.Errorf("%s.base_url is required", name)
	}
	return nil
}

// validatePrompt 确保提示词非空。
func validatePrompt(name string, prompt string) error {
	if strings.TrimSpace(prompt) == "" {
		return fmt.Errorf("%s is required", name)
	}
	return nil
}

// resolveConfigPath 支持相对路径与项目根目录兜底查找。
func resolveConfigPath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}

	if _, err := os.Stat(path); err == nil {
		return path
	}

	_, currentFile, _, ok := runtime.Caller(0)
	if ok {
		projectRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), ".."))
		candidate := filepath.Join(projectRoot, path)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return path
}
