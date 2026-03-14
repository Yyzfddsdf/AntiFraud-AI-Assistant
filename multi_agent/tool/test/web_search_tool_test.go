package tool_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	appcfg "antifraud/config"
	agenttool "antifraud/multi_agent/tool"
	"antifraud/web_search_system"
)

type stubSearcher struct {
	result web_search_system.SearchResponse
	err    error
}

func (s stubSearcher) Search(ctx context.Context, query string, maxResults int) (web_search_system.SearchResponse, error) {
	return s.result, s.err
}

func TestParseWebSearchInput(t *testing.T) {
	input, err := agenttool.ParseWebSearchInput(`{"query":"risk event","max_results":3}`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if input.Query != "risk event" || input.MaxResults != 3 {
		t.Fatalf("unexpected input: %+v", input)
	}
}

func TestExecuteWebSearchSuccess(t *testing.T) {
	payload, err := agenttool.ExecuteWebSearch(context.Background(), stubSearcher{
		result: web_search_system.SearchResponse{
			Query:  "risk event",
			Answer: "summary",
			Results: []web_search_system.SearchResult{
				{Title: "A", URL: "https://example.com/a", Content: "snippet"},
			},
		},
	}, agenttool.WebSearchInput{
		Query:      "risk event",
		MaxResults: 3,
	})
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	if payload["status"] != "success" {
		t.Fatalf("unexpected status: %#v", payload["status"])
	}
	if payload["query"] != "risk event" {
		t.Fatalf("unexpected query: %#v", payload["query"])
	}
	if payload["answer"] != "summary" {
		t.Fatalf("unexpected answer: %#v", payload["answer"])
	}
	if payload["max_results"] != 3 {
		t.Fatalf("unexpected max_results: %#v", payload["max_results"])
	}
}

func TestExecuteWebSearchRejectsEmptyQuery(t *testing.T) {
	_, err := agenttool.ExecuteWebSearch(context.Background(), stubSearcher{}, agenttool.WebSearchInput{Query: "   "})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "query is empty") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecuteWebSearchUsesDefaultResultCount(t *testing.T) {
	payload, err := agenttool.ExecuteWebSearch(context.Background(), stubSearcher{
		result: web_search_system.SearchResponse{
			Query:   "risk event",
			Answer:  "summary",
			Results: []web_search_system.SearchResult{},
		},
	}, agenttool.WebSearchInput{Query: "risk event"})
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if payload["max_results"] != 10 {
		t.Fatalf("expected default max_results 10, got %#v", payload["max_results"])
	}
}

func TestExecuteWebSearchRejectsOutOfRangeMaxResults(t *testing.T) {
	_, err := agenttool.ExecuteWebSearch(context.Background(), stubSearcher{}, agenttool.WebSearchInput{
		Query:      "risk event",
		MaxResults: 21,
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "max_results must be between 1 and 20") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWebSearchHandlerReturnsFailurePayload(t *testing.T) {
	originalLoad := loadWebSearchConfig
	originalSearcherFactory := newWebSearcher
	t.Cleanup(func() {
		loadWebSearchConfig = originalLoad
		newWebSearcher = originalSearcherFactory
	})

	loadWebSearchConfig = func(path string) (*appcfg.Config, error) {
		return &appcfg.Config{
			Tavily: appcfg.TavilyConfig{APIKey: "key"},
		}, nil
	}
	newWebSearcher = func(cfg appcfg.TavilyConfig) web_search_system.Searcher {
		return stubSearcher{err: errors.New("upstream failed")}
	}

	resp, err := (&agenttool.WebSearchHandler{}).Handle(context.Background(), `{"query":"risk event"}`)
	if err != nil {
		t.Fatalf("handle failed: %v", err)
	}
	if resp.Payload["status"] != "failed" {
		t.Fatalf("unexpected status: %#v", resp.Payload["status"])
	}
	if !strings.Contains(resp.Payload["error"].(string), "upstream failed") {
		t.Fatalf("unexpected error payload: %#v", resp.Payload["error"])
	}
}

func TestWebSearchHandlerSuccess(t *testing.T) {
	originalLoad := loadWebSearchConfig
	originalSearcherFactory := newWebSearcher
	t.Cleanup(func() {
		loadWebSearchConfig = originalLoad
		newWebSearcher = originalSearcherFactory
	})

	loadWebSearchConfig = func(path string) (*appcfg.Config, error) {
		return &appcfg.Config{
			Tavily: appcfg.TavilyConfig{APIKey: "key"},
		}, nil
	}
	newWebSearcher = func(cfg appcfg.TavilyConfig) web_search_system.Searcher {
		return stubSearcher{
			result: web_search_system.SearchResponse{
				Query:  "risk event",
				Answer: "summary",
				Results: []web_search_system.SearchResult{
					{Title: "A", URL: "https://example.com/a", Content: "snippet"},
				},
			},
		}
	}

	resp, err := (&agenttool.WebSearchHandler{}).Handle(context.Background(), `{"query":"risk event","max_results":2}`)
	if err != nil {
		t.Fatalf("handle failed: %v", err)
	}
	if resp.Payload["status"] != "success" {
		t.Fatalf("unexpected status: %#v", resp.Payload["status"])
	}
	if resp.Payload["query"] != "risk event" {
		t.Fatalf("unexpected query: %#v", resp.Payload["query"])
	}
}

func TestWebSearchToolNotRegisteredByDefault(t *testing.T) {
	if handler := agenttool.GetToolHandler(agenttool.WebSearchToolName); handler != nil {
		t.Fatalf("expected web search handler to stay unregistered, got %#v", handler)
	}

	for _, tool := range agenttool.MainAgentTools() {
		if tool.Function != nil && tool.Function.Name == agenttool.WebSearchToolName {
			t.Fatalf("expected %s to stay out of the default registry", agenttool.WebSearchToolName)
		}
	}
}
