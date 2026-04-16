package realtime

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/websocket"
)

const (
	WebSocketHeartbeatTypePing = "ping"
	WebSocketHeartbeatTypePong = "pong"
	WebSocketHeartbeatInterval = 25 * time.Second
	WebSocketHeartbeatTimeout  = 90 * time.Second
)

type WebSocketHeartbeatMessage struct {
	Type       string `json:"type"`
	SentAt     string `json:"sent_at,omitempty"`
	ReceivedAt string `json:"received_at,omitempty"`
}

type SafeWebSocketConnection struct {
	ws      *websocket.Conn
	writeMu sync.Mutex
}

type WebSocketHeartbeatTracker struct {
	lastSeenUnixNano atomic.Int64
}

func NewSafeWebSocketConnection(ws *websocket.Conn) *SafeWebSocketConnection {
	return &SafeWebSocketConnection{ws: ws}
}

func NewWebSocketHeartbeatTracker(now time.Time) *WebSocketHeartbeatTracker {
	tracker := &WebSocketHeartbeatTracker{}
	tracker.TouchAt(now)
	return tracker
}

func (c *SafeWebSocketConnection) Raw() *websocket.Conn {
	if c == nil {
		return nil
	}
	return c.ws
}

func (c *SafeWebSocketConnection) Close() error {
	if c == nil || c.ws == nil {
		return nil
	}
	return c.ws.Close()
}

func (c *SafeWebSocketConnection) SendJSON(payload any) error {
	if c == nil || c.ws == nil {
		return fmt.Errorf("websocket connection is nil")
	}
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return websocket.JSON.Send(c.ws, payload)
}

func (c *SafeWebSocketConnection) SendPing() error {
	return c.SendJSON(WebSocketHeartbeatMessage{
		Type:   WebSocketHeartbeatTypePing,
		SentAt: time.Now().UTC().Format(time.RFC3339),
	})
}

func (c *SafeWebSocketConnection) SendPong(sentAt string) error {
	return c.SendJSON(WebSocketHeartbeatMessage{
		Type:       WebSocketHeartbeatTypePong,
		SentAt:     strings.TrimSpace(sentAt),
		ReceivedAt: time.Now().UTC().Format(time.RFC3339),
	})
}

func (c *SafeWebSocketConnection) HandleHeartbeatMessage(incoming string) (bool, error) {
	if c == nil {
		return false, nil
	}
	var msg WebSocketHeartbeatMessage
	if err := json.Unmarshal([]byte(incoming), &msg); err != nil {
		return false, nil
	}
	switch NormalizeWebSocketHeartbeatType(msg.Type) {
	case WebSocketHeartbeatTypePing:
		return true, c.SendPong(msg.SentAt)
	case WebSocketHeartbeatTypePong:
		return true, nil
	default:
		return false, nil
	}
}

func NormalizeWebSocketHeartbeatType(raw string) string {
	switch strings.TrimSpace(raw) {
	case WebSocketHeartbeatTypePing:
		return WebSocketHeartbeatTypePing
	case WebSocketHeartbeatTypePong:
		return WebSocketHeartbeatTypePong
	default:
		return ""
	}
}

func (t *WebSocketHeartbeatTracker) Touch() {
	t.TouchAt(time.Now())
}

func (t *WebSocketHeartbeatTracker) TouchAt(now time.Time) {
	if t == nil {
		return
	}
	t.lastSeenUnixNano.Store(now.UTC().UnixNano())
}

func (t *WebSocketHeartbeatTracker) LastSeenAt() time.Time {
	if t == nil {
		return time.Time{}
	}
	nanos := t.lastSeenUnixNano.Load()
	if nanos <= 0 {
		return time.Time{}
	}
	return time.Unix(0, nanos).UTC()
}

func (t *WebSocketHeartbeatTracker) IsExpired(now time.Time, timeout time.Duration) bool {
	if t == nil {
		return false
	}
	if timeout <= 0 {
		timeout = WebSocketHeartbeatTimeout
	}
	lastSeenAt := t.LastSeenAt()
	if lastSeenAt.IsZero() {
		return false
	}
	return now.UTC().Sub(lastSeenAt) > timeout
}
