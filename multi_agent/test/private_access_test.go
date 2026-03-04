package multi_agent_test

import _ "unsafe"

//go:linkname buildImageDataURL antifraud/multi_agent.buildImageDataURL
func buildImageDataURL(imageBase64 string) (string, error)

//go:linkname buildVideoDataURL antifraud/multi_agent.buildVideoDataURL
func buildVideoDataURL(videoBase64 string) (string, error)

//go:linkname normalizeBase64List antifraud/multi_agent.normalizeBase64List
func normalizeBase64List(items []string) []string
