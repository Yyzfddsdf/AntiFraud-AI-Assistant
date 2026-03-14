package tool_test

import (
	appcfg "antifraud/config"
	"antifraud/web_search_system"
	_ "unsafe"
)

//go:linkname loadWebSearchConfig antifraud/multi_agent/tool.loadWebSearchConfig
var loadWebSearchConfig func(string) (*appcfg.Config, error)

//go:linkname newWebSearcher antifraud/multi_agent/tool.newWebSearcher
var newWebSearcher func(appcfg.TavilyConfig) web_search_system.Searcher
