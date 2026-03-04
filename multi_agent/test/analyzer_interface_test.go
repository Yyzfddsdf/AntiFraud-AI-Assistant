package multi_agent_test

import (
	"errors"
	"strings"
	"testing"

	appcfg "antifraud/config"
	multiagent "antifraud/multi_agent"
)

func testCommonAgent() multiagent.CommonAgent {
	return multiagent.NewCommonAgent(
		"test-agent",
		appcfg.ModelConfig{
			Model:       "gpt-test",
			APIKey:      "k",
			BaseURL:     "https://example.com/v1",
			MaxTokens:   64,
			TopP:        1,
			Temperature: 0.1,
		},
		appcfg.RetryConfig{
			MaxRetries:   3,
			RetryDelayMS: 1,
		},
	)
}

func TestCommonAgentRetry_SuccessFirstAttempt(t *testing.T) {
	agent := testCommonAgent()
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
	agent := testCommonAgent()
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
	agent := testCommonAgent()
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
