package multi_agent_test

import _ "unsafe"

//go:linkname buildImageDataURL antifraud/internal/modules/multi_agent/core.buildImageDataURL
func buildImageDataURL(imageBase64 string) (string, error)

//go:linkname buildVideoDataURL antifraud/internal/modules/multi_agent/core.buildVideoDataURL
func buildVideoDataURL(videoBase64 string) (string, error)

//go:linkname normalizeBase64List antifraud/internal/modules/multi_agent/core.normalizeBase64List
func normalizeBase64List(items []string) []string
