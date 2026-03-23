package httpapi

import (
	"log"
	"strings"
	"time"

	"antifraud/internal/modules/multi_agent/adapters/outbound/state"
	appcfg "antifraud/internal/platform/config"
)

// AlertHistoryReader 定义告警 websocket 依赖的历史读取端口。
type AlertHistoryReader interface {
	GetCaseHistory(userID string) []state.CaseHistoryRecord
}

// AlertConfigProvider 定义告警 websocket 运行时配置端口。
type AlertConfigProvider interface {
	LoadAlertWSConfig() appcfg.AlertWSConfig
}

type alertService struct {
	historyReader AlertHistoryReader
	config        AlertConfigProvider
}

func newAlertService(historyReader AlertHistoryReader, configProvider AlertConfigProvider) *alertService {
	if historyReader == nil {
		historyReader = defaultAlertHistoryReader{}
	}
	if configProvider == nil {
		configProvider = defaultAlertConfigProvider{}
	}
	return &alertService{
		historyReader: historyReader,
		config:        configProvider,
	}
}

func (s *alertService) recentHistory(userID string) []state.CaseHistoryRecord {
	if s == nil || s.historyReader == nil {
		return []state.CaseHistoryRecord{}
	}
	return s.historyReader.GetCaseHistory(strings.TrimSpace(userID))
}

func (s *alertService) runtimeConfig() alertWSRuntimeConfig {
	result := alertWSRuntimeConfig{
		pollInterval: defaultAlertPollInterval,
		recentWindow: defaultAlertRecentWindow,
	}
	if s == nil || s.config == nil {
		return result
	}
	cfg := s.config.LoadAlertWSConfig()
	pollInterval := time.Duration(cfg.PollIntervalSeconds) * time.Second
	if pollInterval > 0 {
		result.pollInterval = pollInterval
	}
	recentWindow := time.Duration(cfg.RecentWindowMinutes) * time.Minute
	if recentWindow > 0 {
		result.recentWindow = recentWindow
	}
	return result
}

type defaultAlertHistoryReader struct{}

func (defaultAlertHistoryReader) GetCaseHistory(userID string) []state.CaseHistoryRecord {
	return state.GetCaseHistory(userID)
}

type defaultAlertConfigProvider struct{}

func (defaultAlertConfigProvider) LoadAlertWSConfig() appcfg.AlertWSConfig {
	cfg, err := appcfg.LoadConfig("internal/platform/config/config.json")
	if err != nil {
		log.Printf("[alert_ws] load config failed, fallback to defaults: err=%v", err)
		return appcfg.AlertWSConfig{}
	}
	if cfg == nil {
		return appcfg.AlertWSConfig{}
	}
	return cfg.AlertWS
}
