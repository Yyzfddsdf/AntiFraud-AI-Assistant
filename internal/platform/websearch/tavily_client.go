package web_search_system

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"antifraud/internal/platform/config"
)

const (
	defaultBaseURL       = "https://api.tavily.com"
	defaultIncludeAnswer = "advanced"
	defaultSearchDepth   = "advanced"
	defaultMaxResults    = 3
	maxAllowedResults    = 5
	defaultTimeoutMS     = 15000
	maxAnswerRunes       = 1200
	maxContentRunes      = 600
)

// Searcher 定义联网搜索能力接口，便于上层工具复用与测试替换。
type Searcher interface {
	Search(ctx context.Context, query string, maxResults int) (SearchResponse, error)
}

// TavilyClient 封装 Tavily Search API 调用。
type TavilyClient struct {
	apiKey        string
	baseURL       string
	includeAnswer string
	searchDepth   string
	maxResults    int
	httpClient    *http.Client
}

// SearchResponse 是搜索后的结构化结果。
type SearchResponse struct {
	Query   string         `json:"query"`
	Answer  string         `json:"answer"`
	Results []SearchResult `json:"results"`
}

// SearchResult 是单条搜索结果的精简表示。
type SearchResult struct {
	Title         string  `json:"title"`
	URL           string  `json:"url"`
	Content       string  `json:"content"`
	Score         float64 `json:"score"`
	PublishedDate string  `json:"published_date,omitempty"`
}

type tavilySearchRequest struct {
	Query         string `json:"query"`
	IncludeAnswer string `json:"include_answer,omitempty"`
	SearchDepth   string `json:"search_depth,omitempty"`
	MaxResults    int    `json:"max_results,omitempty"`
}

type tavilySearchResponse struct {
	Query   string               `json:"query"`
	Answer  string               `json:"answer"`
	Results []tavilySearchResult `json:"results"`
}

type tavilySearchResult struct {
	Title         string  `json:"title"`
	URL           string  `json:"url"`
	Content       string  `json:"content"`
	Score         float64 `json:"score"`
	PublishedDate string  `json:"published_date"`
}

// NewTavilyClient 根据配置创建默认 HTTP 客户端。
func NewTavilyClient(cfg config.TavilyConfig) *TavilyClient {
	timeoutMS := cfg.TimeoutMS
	if timeoutMS <= 0 {
		timeoutMS = defaultTimeoutMS
	}

	return NewTavilyClientWithHTTPClient(cfg, &http.Client{
		Timeout: time.Duration(timeoutMS) * time.Millisecond,
	})
}

// NewTavilyClientWithHTTPClient 允许测试注入自定义 HTTP 客户端。
func NewTavilyClientWithHTTPClient(cfg config.TavilyConfig, httpClient *http.Client) *TavilyClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultTimeoutMS * time.Millisecond}
	}

	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	includeAnswer := strings.TrimSpace(cfg.IncludeAnswer)
	if includeAnswer == "" {
		includeAnswer = defaultIncludeAnswer
	}

	searchDepth := strings.TrimSpace(cfg.SearchDepth)
	if searchDepth == "" {
		searchDepth = defaultSearchDepth
	}

	return &TavilyClient{
		apiKey:        strings.TrimSpace(cfg.APIKey),
		baseURL:       baseURL,
		includeAnswer: includeAnswer,
		searchDepth:   searchDepth,
		maxResults:    defaultMaxResults,
		httpClient:    httpClient,
	}
}

// Search 执行单次 Tavily 搜索并返回裁剪后的结果。
func (c *TavilyClient) Search(ctx context.Context, query string, maxResults int) (SearchResponse, error) {
	trimmedQuery := strings.TrimSpace(query)
	if trimmedQuery == "" {
		return SearchResponse{}, fmt.Errorf("query is empty")
	}
	if strings.TrimSpace(c.apiKey) == "" {
		return SearchResponse{}, fmt.Errorf("tavily api key is empty")
	}
	if c.httpClient == nil {
		return SearchResponse{}, fmt.Errorf("http client is nil")
	}

	resolvedMaxResults := maxResults
	if resolvedMaxResults <= 0 {
		resolvedMaxResults = c.maxResults
	}
	if resolvedMaxResults > maxAllowedResults {
		resolvedMaxResults = maxAllowedResults
	}
	if resolvedMaxResults <= 0 {
		resolvedMaxResults = defaultMaxResults
	}

	reqBody := tavilySearchRequest{
		Query:         trimmedQuery,
		IncludeAnswer: c.includeAnswer,
		SearchDepth:   c.searchDepth,
		MaxResults:    resolvedMaxResults,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return SearchResponse{}, fmt.Errorf("marshal tavily request failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/search", bytes.NewReader(payload))
	if err != nil {
		return SearchResponse{}, fmt.Errorf("build tavily request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return SearchResponse{}, fmt.Errorf("send tavily request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return SearchResponse{}, fmt.Errorf("read tavily response failed: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return SearchResponse{}, fmt.Errorf("tavily search failed, status=%d, body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var decoded tavilySearchResponse
	if err := json.Unmarshal(body, &decoded); err != nil {
		return SearchResponse{}, fmt.Errorf("decode tavily response failed: %w", err)
	}

	results := make([]SearchResult, 0, len(decoded.Results))
	for _, item := range decoded.Results {
		results = append(results, SearchResult{
			Title:         strings.TrimSpace(item.Title),
			URL:           strings.TrimSpace(item.URL),
			Content:       truncateRunes(strings.TrimSpace(item.Content), maxContentRunes),
			Score:         item.Score,
			PublishedDate: strings.TrimSpace(item.PublishedDate),
		})
	}

	return SearchResponse{
		Query:   firstNonEmpty(strings.TrimSpace(decoded.Query), trimmedQuery),
		Answer:  truncateRunes(strings.TrimSpace(decoded.Answer), maxAnswerRunes),
		Results: results,
	}, nil
}

func truncateRunes(input string, maxRunes int) string {
	if maxRunes <= 0 {
		return strings.TrimSpace(input)
	}
	runes := []rune(strings.TrimSpace(input))
	if len(runes) <= maxRunes {
		return string(runes)
	}
	return string(runes[:maxRunes]) + "..."
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
