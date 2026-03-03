package multi_agent

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestCommonAgentRetry_SuccessFirstAttempt(t *testing.T) {
	agent := CommonAgent{name: "test-agent", RetryMax: 3, RetryDelay: time.Millisecond}
	calls := 0
	err := agent.Retry("do-something", func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestCommonAgentRetry_SuccessAfterRetry(t *testing.T) {
	agent := CommonAgent{name: "test-agent", RetryMax: 3, RetryDelay: time.Millisecond}
	calls := 0
	err := agent.Retry("do-something", func() error {
		calls++
		if calls < 2 {
			return errors.New("temporary")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if calls != 2 {
		t.Fatalf("expected 2 calls, got %d", calls)
	}
}

func TestCommonAgentRetry_FailedAfterMaxAttempts(t *testing.T) {
	agent := CommonAgent{name: "test-agent", RetryMax: 3, RetryDelay: time.Millisecond}
	calls := 0
	err := agent.Retry("do-something", func() error {
		calls++
		return errors.New("always-fail")
	})
	if err == nil {
		t.Fatalf("expected retry error")
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
	if !strings.Contains(err.Error(), "do-something failed after 3 attempts") {
		t.Fatalf("unexpected error text: %v", err)
	}
}

