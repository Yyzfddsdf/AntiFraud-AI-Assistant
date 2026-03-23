package web_search_system_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	appcfg "antifraud/internal/platform/config"
	"antifraud/internal/platform/websearch"
)

func TestTavilyClientSearchSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer tavily-key" {
			t.Fatalf("unexpected auth header: %q", got)
		}

		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode body failed: %v", err)
		}
		if payload["query"] != "fraud hotline" {
			t.Fatalf("unexpected query: %#v", payload["query"])
		}
		if payload["include_answer"] != "advanced" {
			t.Fatalf("unexpected include_answer: %#v", payload["include_answer"])
		}
		if payload["search_depth"] != "advanced" {
			t.Fatalf("unexpected search_depth: %#v", payload["search_depth"])
		}
		if payload["max_results"] != float64(5) {
			t.Fatalf("unexpected max_results: %#v", payload["max_results"])
		}

		response := map[string]interface{}{
			"query":  "fraud hotline",
			"answer": strings.Repeat("A", 20),
			"results": []map[string]interface{}{
				{
					"title":          "Result 1",
					"url":            "https://example.com/1",
					"content":        "content 1",
					"score":          0.91,
					"published_date": "2026-03-14",
				},
			},
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Fatalf("encode response failed: %v", err)
		}
	}))
	defer server.Close()

	client := web_search_system.NewTavilyClientWithHTTPClient(appcfg.TavilyConfig{
		APIKey:        "tavily-key",
		BaseURL:       server.URL,
		IncludeAnswer: "advanced",
		SearchDepth:   "advanced",
	}, server.Client())

	result, err := client.Search(context.Background(), "fraud hotline", 5)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}

	if result.Query != "fraud hotline" {
		t.Fatalf("unexpected query: %q", result.Query)
	}
	if result.Answer != strings.Repeat("A", 20) {
		t.Fatalf("unexpected answer: %q", result.Answer)
	}
	if len(result.Results) != 1 {
		t.Fatalf("unexpected result count: %d", len(result.Results))
	}
	if result.Results[0].Title != "Result 1" || result.Results[0].URL != "https://example.com/1" {
		t.Fatalf("unexpected first result: %+v", result.Results[0])
	}
}

func TestTavilyClientSearchReturnsReadableError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"rate limited"}`, http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := web_search_system.NewTavilyClientWithHTTPClient(appcfg.TavilyConfig{
		APIKey:  "tavily-key",
		BaseURL: server.URL,
	}, server.Client())

	_, err := client.Search(context.Background(), "fraud hotline", 1)
	if err == nil {
		t.Fatal("expected search error")
	}
	if !strings.Contains(err.Error(), "status=429") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTavilyClientSearchRejectsEmptyQuery(t *testing.T) {
	client := web_search_system.NewTavilyClient(appcfg.TavilyConfig{
		APIKey: "tavily-key",
	})

	_, err := client.Search(context.Background(), "   ", 1)
	if err == nil {
		t.Fatal("expected empty query error")
	}
	if !strings.Contains(err.Error(), "query is empty") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTavilyClientSearchCapsResultCountAtTwenty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode body failed: %v", err)
		}
		if payload["max_results"] != float64(5) {
			t.Fatalf("expected max_results to be capped at 5, got %#v", payload["max_results"])
		}

		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"query":   "fraud hotline",
			"answer":  "ok",
			"results": []map[string]interface{}{},
		}); err != nil {
			t.Fatalf("encode response failed: %v", err)
		}
	}))
	defer server.Close()

	client := web_search_system.NewTavilyClientWithHTTPClient(appcfg.TavilyConfig{
		APIKey:  "tavily-key",
		BaseURL: server.URL,
	}, server.Client())

	if _, err := client.Search(context.Background(), "fraud hotline", 99); err != nil {
		t.Fatalf("search failed: %v", err)
	}
}
