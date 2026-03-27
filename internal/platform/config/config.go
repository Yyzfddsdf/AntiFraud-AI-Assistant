package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
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

// ChatConfig 定义聊天系统模型配置。
type ChatConfig struct {
	Prompt  string `json:"prompt"`
	Model   string `json:"model"`
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url"`
}

// TavilyConfig 定义联网搜索服务配置。
type TavilyConfig struct {
	APIKey        string `json:"api_key"`
	BaseURL       string `json:"base_url"`
	IncludeAnswer string `json:"include_answer"`
	SearchDepth   string `json:"search_depth"`
	TimeoutMS     int    `json:"timeout_ms"`
}

// RedisConfig 定义统一 Redis 连接配置。
type RedisConfig struct {
	Addr     string `json:"addr"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

// RetryConfig 定义通用重试策略。
type RetryConfig struct {
	MaxRetries   int `json:"max_retries"`
	RetryDelayMS int `json:"retry_delay_ms"`
}

// AlertWSConfig 定义实时告警 WebSocket 轮询配置。
type AlertWSConfig struct {
	PollIntervalSeconds int `json:"poll_interval_seconds"`
	RecentWindowMinutes int `json:"recent_window_minutes"`
}

// AgentModelConfig 按智能体拆分模型与调用参数，便于后续扩展新 provider/model。
type AgentModelConfig struct {
	Main           ModelConfig `json:"main"`
	Image          ModelConfig `json:"image"`
	ImageQuick     ModelConfig `json:"image_quick"`
	Video          ModelConfig `json:"video"`
	Audio          ModelConfig `json:"audio"`
	CaseCollection ModelConfig `json:"case_collection"`
	SimulationQuiz ModelConfig `json:"simulation_quiz"`
}

// PromptConfig 统一托管各智能体提示词，避免硬编码散落在代码里。
type PromptConfig struct {
	Main           string `json:"main"`
	Image          string `json:"image"`
	ImageQuick     string `json:"image_quick"`
	Video          string `json:"video"`
	Audio          string `json:"audio"`
	CaseCollection string `json:"case_collection"`
	SimulationQuiz string `json:"simulation_quiz"`
}

// Config 是项目总配置对象。
type Config struct {
	Agents        AgentModelConfig `json:"agents"`
	Embedding     EmbeddingConfig  `json:"embedding"`
	Chat          ChatConfig       `json:"chat"`
	AdminChat     ChatConfig       `json:"admin_chat"`
	Tavily        TavilyConfig     `json:"tavily"`
	Redis         RedisConfig      `json:"redis"`
	Prompts       PromptConfig     `json:"prompts"`
	Retry         RetryConfig      `json:"retry"`
	AlertWS       AlertWSConfig    `json:"alert_ws"`
	FamilyAlertWS AlertWSConfig    `json:"family_alert_ws"`
}

var (
	configCacheMu sync.RWMutex
	configCache   = map[string]*Config{}
	utf8BOM       = []byte{0xEF, 0xBB, 0xBF}
)

// LoadConfig 负责读取、标准化并校验配置文件。
// 缓存策略：
// 1) 以“解析后的绝对路径”作为缓存键；
// 2) 首次读取并校验后写入缓存；
// 3) 后续调用直接返回缓存副本，避免请求路径重复读盘与反序列化。
func LoadConfig(path string) (*Config, error) {
	resolvedPath := resolveConfigPath(path)

	configCacheMu.RLock()
	if cached, ok := configCache[resolvedPath]; ok {
		configCacheMu.RUnlock()
		return cloneConfig(cached), nil
	}
	configCacheMu.RUnlock()

	data, err := os.ReadFile(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file (%s): %w", resolvedPath, err)
	}
	data = bytes.TrimPrefix(data, utf8BOM)
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, fmt.Errorf("config file is empty: %s", resolvedPath)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	cfg.applyEnvOverrides()
	cfg.normalize()
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	configCacheMu.Lock()
	if cached, ok := configCache[resolvedPath]; ok {
		configCacheMu.Unlock()
		return cloneConfig(cached), nil
	}
	configCache[resolvedPath] = &cfg
	configCacheMu.Unlock()

	return cloneConfig(&cfg), nil
}

func cloneConfig(cfg *Config) *Config {
	if cfg == nil {
		return nil
	}
	cloned := *cfg
	return &cloned
}

// applyEnvOverrides 允许通过环境变量覆盖 config.json 中的敏感配置，优先用于 API Key。
func (c *Config) applyEnvOverrides() {
	c.Agents.Main.APIKey = firstNonEmptyEnv("AGENT_MAIN_API_KEY", c.Agents.Main.APIKey)
	c.Agents.Image.APIKey = firstNonEmptyEnv("AGENT_IMAGE_API_KEY", c.Agents.Image.APIKey)
	c.Agents.ImageQuick.APIKey = firstNonEmptyEnv("AGENT_IMAGE_QUICK_API_KEY", c.Agents.ImageQuick.APIKey)
	c.Agents.Video.APIKey = firstNonEmptyEnv("AGENT_VIDEO_API_KEY", c.Agents.Video.APIKey)
	c.Agents.Audio.APIKey = firstNonEmptyEnv("AGENT_AUDIO_API_KEY", c.Agents.Audio.APIKey)
	c.Agents.CaseCollection.APIKey = firstNonEmptyEnv("AGENT_CASE_COLLECTION_API_KEY", c.Agents.CaseCollection.APIKey)
	c.Agents.SimulationQuiz.APIKey = firstNonEmptyEnv("AGENT_SIMULATION_QUIZ_API_KEY", c.Agents.SimulationQuiz.APIKey)

	c.Embedding.APIKey = firstNonEmptyEnv("EMBEDDING_API_KEY", c.Embedding.APIKey)
	c.Chat.APIKey = firstNonEmptyEnv("CHAT_API_KEY", c.Chat.APIKey)
	c.AdminChat.APIKey = firstNonEmptyEnv("ADMIN_CHAT_API_KEY", c.AdminChat.APIKey)
	c.Tavily.APIKey = firstNonEmptyEnv("TAVILY_API_KEY", c.Tavily.APIKey)
}

func firstNonEmptyEnv(envName string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(envName)); value != "" {
		return value
	}
	return strings.TrimSpace(fallback)
}

// normalize 统一裁剪字符串空白，减少运行期参数格式问题。
func (c *Config) normalize() {
	c.Agents.Main = normalizeModel(c.Agents.Main)
	c.Agents.Image = normalizeModel(c.Agents.Image)
	c.Agents.ImageQuick = normalizeModel(c.Agents.ImageQuick)
	c.Agents.Video = normalizeModel(c.Agents.Video)
	c.Agents.Audio = normalizeModel(c.Agents.Audio)
	c.Agents.CaseCollection = normalizeModel(c.Agents.CaseCollection)
	c.Agents.SimulationQuiz = normalizeModel(c.Agents.SimulationQuiz)
	if c.Agents.SimulationQuiz.Model == "" {
		c.Agents.SimulationQuiz = c.Agents.Main
	}
	c.Embedding = normalizeEmbedding(c.Embedding)
	c.Chat = normalizeChatFromModel(c.Chat, c.Agents.Main)
	c.AdminChat = normalizeChatFromChat(c.AdminChat, c.Chat)
	c.Tavily = normalizeTavily(c.Tavily)
	c.Redis = normalizeRedis(c.Redis)
	c.Prompts.Main = strings.TrimSpace(c.Prompts.Main)
	c.Prompts.Image = strings.TrimSpace(c.Prompts.Image)
	c.Prompts.ImageQuick = strings.TrimSpace(c.Prompts.ImageQuick)
	c.Prompts.Video = strings.TrimSpace(c.Prompts.Video)
	c.Prompts.Audio = strings.TrimSpace(c.Prompts.Audio)
	c.Prompts.CaseCollection = strings.TrimSpace(c.Prompts.CaseCollection)
	c.Prompts.SimulationQuiz = strings.TrimSpace(c.Prompts.SimulationQuiz)
	if c.Prompts.SimulationQuiz == "" {
		c.Prompts.SimulationQuiz = "你是反诈模拟题目生成智能体。你必须调用 submit_simulation_quiz_pack 工具提交固定10步结构的题包，不允许输出工具外文本。每一道题的正确选项分布必须有变化，不允许所有题目都使用同一个答案字母作为正确答案。"
	}
	c.AlertWS = normalizeAlertWS(c.AlertWS)
	c.FamilyAlertWS = normalizeAlertWS(c.FamilyAlertWS)
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

func normalizeChatFromModel(chatCfg ChatConfig, mainModelCfg ModelConfig) ChatConfig {
	chatCfg.Prompt = strings.TrimSpace(chatCfg.Prompt)
	chatCfg.APIKey = strings.TrimSpace(chatCfg.APIKey)
	chatCfg.BaseURL = strings.TrimSpace(chatCfg.BaseURL)
	chatCfg.Model = strings.TrimSpace(chatCfg.Model)

	if chatCfg.APIKey == "" {
		chatCfg.APIKey = strings.TrimSpace(mainModelCfg.APIKey)
	}
	if chatCfg.BaseURL == "" {
		chatCfg.BaseURL = strings.TrimSpace(mainModelCfg.BaseURL)
	}
	if chatCfg.Model == "" {
		chatCfg.Model = strings.TrimSpace(mainModelCfg.Model)
	}
	return chatCfg
}

func normalizeChatFromChat(chatCfg ChatConfig, baseChatCfg ChatConfig) ChatConfig {
	chatCfg.Prompt = strings.TrimSpace(chatCfg.Prompt)
	chatCfg.APIKey = strings.TrimSpace(chatCfg.APIKey)
	chatCfg.BaseURL = strings.TrimSpace(chatCfg.BaseURL)
	chatCfg.Model = strings.TrimSpace(chatCfg.Model)

	if chatCfg.Prompt == "" {
		chatCfg.Prompt = strings.TrimSpace(baseChatCfg.Prompt)
	}
	if chatCfg.APIKey == "" {
		chatCfg.APIKey = strings.TrimSpace(baseChatCfg.APIKey)
	}
	if chatCfg.BaseURL == "" {
		chatCfg.BaseURL = strings.TrimSpace(baseChatCfg.BaseURL)
	}
	if chatCfg.Model == "" {
		chatCfg.Model = strings.TrimSpace(baseChatCfg.Model)
	}
	return chatCfg
}

func normalizeTavily(tavilyCfg TavilyConfig) TavilyConfig {
	tavilyCfg.APIKey = strings.TrimSpace(tavilyCfg.APIKey)
	tavilyCfg.BaseURL = strings.TrimRight(strings.TrimSpace(tavilyCfg.BaseURL), "/")
	tavilyCfg.IncludeAnswer = strings.TrimSpace(tavilyCfg.IncludeAnswer)
	tavilyCfg.SearchDepth = strings.TrimSpace(tavilyCfg.SearchDepth)
	if tavilyCfg.BaseURL == "" {
		tavilyCfg.BaseURL = "https://api.tavily.com"
	}
	if tavilyCfg.IncludeAnswer == "" {
		tavilyCfg.IncludeAnswer = "advanced"
	}
	if tavilyCfg.SearchDepth == "" {
		tavilyCfg.SearchDepth = "advanced"
	}
	if tavilyCfg.TimeoutMS <= 0 {
		tavilyCfg.TimeoutMS = 15000
	}
	return tavilyCfg
}

func normalizeRedis(redisCfg RedisConfig) RedisConfig {
	redisCfg.Addr = strings.TrimSpace(redisCfg.Addr)
	redisCfg.Password = strings.TrimSpace(redisCfg.Password)
	if redisCfg.Addr == "" {
		redisCfg.Addr = "127.0.0.1:6379"
	}
	if redisCfg.DB < 0 {
		redisCfg.DB = 0
	}
	return redisCfg
}

func normalizeAlertWS(alertCfg AlertWSConfig) AlertWSConfig {
	if alertCfg.PollIntervalSeconds <= 0 {
		alertCfg.PollIntervalSeconds = 30
	}
	if alertCfg.RecentWindowMinutes <= 0 {
		alertCfg.RecentWindowMinutes = 60
	}
	return alertCfg
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
	if err := validateModel("agents.image_quick", c.Agents.ImageQuick); err != nil {
		return err
	}
	if err := validateModel("agents.video", c.Agents.Video); err != nil {
		return err
	}
	if err := validateModel("agents.audio", c.Agents.Audio); err != nil {
		return err
	}
	if err := validateModel("agents.case_collection", c.Agents.CaseCollection); err != nil {
		return err
	}
	if err := validateModel("agents.simulation_quiz", c.Agents.SimulationQuiz); err != nil {
		return err
	}
	if err := validateEmbedding("embedding", c.Embedding); err != nil {
		return err
	}
	if err := validateChat("chat", c.Chat); err != nil {
		return err
	}
	if err := validateChat("admin_chat", c.AdminChat); err != nil {
		return err
	}
	if err := validateTavily("tavily", c.Tavily); err != nil {
		return err
	}
	if err := validatePrompt("prompts.main", c.Prompts.Main); err != nil {
		return err
	}
	if err := validatePrompt("prompts.image", c.Prompts.Image); err != nil {
		return err
	}
	if err := validatePrompt("prompts.image_quick", c.Prompts.ImageQuick); err != nil {
		return err
	}
	if err := validatePrompt("prompts.video", c.Prompts.Video); err != nil {
		return err
	}
	if err := validatePrompt("prompts.audio", c.Prompts.Audio); err != nil {
		return err
	}
	if err := validatePrompt("prompts.case_collection", c.Prompts.CaseCollection); err != nil {
		return err
	}
	if err := validatePrompt("prompts.simulation_quiz", c.Prompts.SimulationQuiz); err != nil {
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

func validateChat(name string, chatCfg ChatConfig) error {
	if chatCfg.Prompt == "" {
		return fmt.Errorf("%s.prompt is required", name)
	}
	if chatCfg.Model == "" {
		return fmt.Errorf("%s.model is required", name)
	}
	if chatCfg.APIKey == "" {
		return fmt.Errorf("%s.api_key is required", name)
	}
	if chatCfg.BaseURL == "" {
		return fmt.Errorf("%s.base_url is required", name)
	}
	return nil
}

func validateTavily(name string, tavilyCfg TavilyConfig) error {
	if tavilyCfg.APIKey == "" &&
		tavilyCfg.BaseURL == "https://api.tavily.com" &&
		tavilyCfg.IncludeAnswer == "advanced" &&
		tavilyCfg.SearchDepth == "advanced" &&
		tavilyCfg.TimeoutMS == 15000 {
		return nil
	}
	if tavilyCfg.APIKey == "" {
		return fmt.Errorf("%s.api_key is required", name)
	}
	if tavilyCfg.BaseURL == "" {
		return fmt.Errorf("%s.base_url is required", name)
	}
	if tavilyCfg.TimeoutMS <= 0 {
		return fmt.Errorf("%s.timeout_ms must be > 0", name)
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
		currentDir := filepath.Dir(currentFile)
		baseName := filepath.Base(path)
		for i := 0; i < 10; i++ {
			candidate := filepath.Join(currentDir, path)
			if _, err := os.Stat(candidate); err == nil {
				return candidate
			}
			if baseName != "" {
				fallback := filepath.Join(currentDir, baseName)
				if _, err := os.Stat(fallback); err == nil {
					return fallback
				}
			}
			parent := filepath.Dir(currentDir)
			if parent == currentDir {
				break
			}
			currentDir = parent
		}
	}

	return path
}
