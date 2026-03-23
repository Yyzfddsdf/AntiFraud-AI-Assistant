package httpapi

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/websocket"
)

const (
	defaultAlertPollInterval = 30 * time.Second
	defaultAlertRecentWindow = 1 * time.Hour
)

type alertWSRuntimeConfig struct {
	pollInterval time.Duration
	recentWindow time.Duration
}

type alertWSMessage struct {
	Type        string `json:"type"`
	UserID      string `json:"user_id"`
	RecordID    string `json:"record_id"`
	Title       string `json:"title"`
	CaseSummary string `json:"case_summary"`
	ScamType    string `json:"scam_type"`
	RiskLevel   string `json:"risk_level"`
	CreatedAt   string `json:"created_at"`
	SentAt      string `json:"sent_at"`
}

type AlertWSHandler struct {
	service *alertService
}

func NewAlertWSHandler(service *alertService) *AlertWSHandler {
	if service == nil {
		service = newAlertService(nil, nil)
	}
	return &AlertWSHandler{service: service}
}

var defaultAlertWSHandler = NewAlertWSHandler(nil)

// AlertWebSocketHandle 提供中高风险预警推送连接：
// 1) 建立连接后按配置轮询 history_cases；
// 2) 命中“中/高风险 + 告警窗口内创建”的记录时主动推送；
// 3) 连接断开后轮询协程自动退出。
func AlertWebSocketHandle(c *gin.Context) {
	defaultAlertWSHandler.Handle(c)
}

func (h *AlertWSHandler) Handle(c *gin.Context) {
	if c.Request.Method != http.MethodGet {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
		return
	}

	userID := getCurrentUserID(c)
	runtimeCfg := h.service.runtimeConfig()
	wsServer := websocket.Server{
		Handshake: func(cfg *websocket.Config, req *http.Request) error {
			_ = cfg
			_ = req
			return nil
		},
		Handler: func(ws *websocket.Conn) {
			h.runSession(ws, userID, runtimeCfg)
		},
	}
	wsServer.ServeHTTP(c.Writer, c.Request)
}

func (h *AlertWSHandler) runSession(ws *websocket.Conn, userID string, runtimeCfg alertWSRuntimeConfig) {
	if ws == nil {
		return
	}
	defer ws.Close()

	done := make(chan struct{})
	var once sync.Once
	stop := func() {
		once.Do(func() {
			close(done)
		})
	}

	go func() {
		defer stop()
		for {
			var incoming string
			if err := websocket.Message.Receive(ws, &incoming); err != nil {
				return
			}
		}
	}()

	ticker := time.NewTicker(runtimeCfg.pollInterval)
	defer ticker.Stop()

	sentRecordIDs := make(map[string]struct{})
	if err := h.pushRecentRiskAlerts(ws, strings.TrimSpace(userID), sentRecordIDs, runtimeCfg.recentWindow); err != nil {
		stop()
	}

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			if err := h.pushRecentRiskAlerts(ws, strings.TrimSpace(userID), sentRecordIDs, runtimeCfg.recentWindow); err != nil {
				stop()
				return
			}
		}
	}
}

func (h *AlertWSHandler) pushRecentRiskAlerts(ws *websocket.Conn, userID string, sentRecordIDs map[string]struct{}, recentWindow time.Duration) error {
	if ws == nil {
		return fmt.Errorf("websocket connection is nil")
	}

	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		normalizedUserID = "demo-user"
	}
	if sentRecordIDs == nil {
		sentRecordIDs = map[string]struct{}{}
	}
	if recentWindow <= 0 {
		recentWindow = defaultAlertRecentWindow
	}

	cutoff := time.Now().Add(-recentWindow)
	history := h.service.recentHistory(normalizedUserID)
	for _, item := range history {
		recordID := strings.TrimSpace(item.RecordID)
		if recordID == "" {
			continue
		}
		if _, exists := sentRecordIDs[recordID]; exists {
			continue
		}
		riskLevel := normalizeAlertRiskLevel(item.RiskLevel)
		if riskLevel == "" {
			continue
		}
		if item.CreatedAt.Before(cutoff) {
			continue
		}

		msg := alertWSMessage{
			Type:        "risk_alert",
			UserID:      normalizedUserID,
			RecordID:    recordID,
			Title:       strings.TrimSpace(item.Title),
			CaseSummary: strings.TrimSpace(item.CaseSummary),
			ScamType:    strings.TrimSpace(item.ScamType),
			RiskLevel:   riskLevel,
			CreatedAt:   item.CreatedAt.UTC().Format(time.RFC3339),
			SentAt:      time.Now().UTC().Format(time.RFC3339),
		}
		if err := websocket.JSON.Send(ws, msg); err != nil {
			return fmt.Errorf("send risk alert failed: %w", err)
		}
		sentRecordIDs[recordID] = struct{}{}
	}

	return nil
}

func normalizeAlertRiskLevel(value string) string {
	switch strings.TrimSpace(value) {
	case "高":
		return "高"
	case "中":
		return "中"
	default:
		return ""
	}
}
