package llm

import (
	"net/http"

	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type ResponsesClientConfig struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

func NewResponsesClient(cfg ResponsesClientConfig) openai.Client {
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Transport: strippedHeadersTransport{base: http.DefaultTransport}}
	}

	return openai.NewClient(
		option.WithAPIKey(cfg.APIKey),
		option.WithBaseURL(cfg.BaseURL),
		option.WithHTTPClient(httpClient),
		option.WithHeaderDel("User-Agent"),
		option.WithHeaderDel("X-Stainless-Lang"),
		option.WithHeaderDel("X-Stainless-Package-Version"),
		option.WithHeaderDel("X-Stainless-OS"),
		option.WithHeaderDel("X-Stainless-Arch"),
		option.WithHeaderDel("X-Stainless-Runtime"),
		option.WithHeaderDel("X-Stainless-Runtime-Version"),
		option.WithHeaderDel("X-Stainless-Retry-Count"),
		option.WithHeaderDel("X-Stainless-Timeout"),
	)
}

type strippedHeadersTransport struct {
	base http.RoundTripper
}

func (t strippedHeadersTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.Header = req.Header.Clone()
	clone.Header.Del("User-Agent")
	clone.Header.Del("X-Stainless-Lang")
	clone.Header.Del("X-Stainless-Package-Version")
	clone.Header.Del("X-Stainless-OS")
	clone.Header.Del("X-Stainless-Arch")
	clone.Header.Del("X-Stainless-Runtime")
	clone.Header.Del("X-Stainless-Runtime-Version")
	clone.Header.Del("X-Stainless-Retry-Count")
	clone.Header.Del("X-Stainless-Timeout")
	if t.base == nil {
		return http.DefaultTransport.RoundTrip(clone)
	}
	return t.base.RoundTrip(clone)
}
