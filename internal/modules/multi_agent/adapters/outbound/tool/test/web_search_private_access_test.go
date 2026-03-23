package tool_test

import (
	appcfg "antifraud/internal/platform/config"
	"antifraud/internal/platform/websearch"
	_ "unsafe"
)

//go:linkname loadWebSearchConfig antifraud/internal/modules/multi_agent/adapters/outbound/tool.loadWebSearchConfig
var loadWebSearchConfig func(string) (*appcfg.Config, error)

//go:linkname newWebSearcher antifraud/internal/modules/multi_agent/adapters/outbound/tool.newWebSearcher
var newWebSearcher func(appcfg.TavilyConfig) web_search_system.Searcher
