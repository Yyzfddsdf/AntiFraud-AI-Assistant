package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"antifraud/internal/modules/multi_agent/adapters/outbound/state"
	"antifraud/internal/modules/multi_agent/application"
	multiagent "antifraud/internal/modules/multi_agent/core"
	appcfg "antifraud/internal/platform/config"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: go run ./cmd/video_api_probe <video-path>")
		os.Exit(2)
	}

	videoPath := os.Args[1]
	raw, err := os.ReadFile(videoPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read file failed: %v\n", err)
		os.Exit(1)
	}

	encoded := base64.StdEncoding.EncodeToString(raw)
	fmt.Printf("VIDEO_PATH=%s\n", videoPath)
	fmt.Printf("RAW_BYTES=%d\n", len(raw))
	fmt.Printf("BASE64_LEN=%d\n", len(encoded))
	if len(encoded) > 64 {
		fmt.Printf("BASE64_PREFIX=%s\n", encoded[:64])
	} else {
		fmt.Printf("BASE64_PREFIX=%s\n", encoded)
	}

	normalizedPayload, err := application.NormalizeTaskPayload(state.TaskPayload{
		Videos: []string{encoded},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "normalize payload failed: %v\n", err)
		os.Exit(1)
	}
	if len(normalizedPayload.Videos) != 1 {
		fmt.Fprintf(os.Stderr, "unexpected normalized video count: %d\n", len(normalizedPayload.Videos))
		os.Exit(1)
	}

	normalized := normalizedPayload.Videos[0]
	normalizedBase64 := normalized
	if strings.HasPrefix(normalizedBase64, "data:") {
		parts := strings.SplitN(normalizedBase64, ",", 2)
		if len(parts) != 2 {
			fmt.Fprintln(os.Stderr, "normalized data url missing payload")
			os.Exit(1)
		}
		normalizedBase64 = parts[1]
	}
	fmt.Printf("NORMALIZED_BASE64_LEN=%d\n", len(normalizedBase64))
	normalizedRaw, err := base64.StdEncoding.DecodeString(normalizedBase64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "decode normalized payload failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("NORMALIZED_RAW_BYTES=%d\n", len(normalizedRaw))

	cfg, err := appcfg.LoadConfig("internal/platform/config/config.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("MODEL=%s\n", strings.TrimSpace(cfg.Agents.Video.Model))
	fmt.Printf("BASE_URL=%s\n", strings.TrimSpace(cfg.Agents.Video.BaseURL))

	agent := multiagent.NewVideoAgent(cfg.Agents.Video, cfg.Retry, cfg.Prompts.Video)
	result, err := agent.Analyze(context.Background(), normalized, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ANALYZE_ERROR=%v\n", err)
		os.Exit(1)
	}

	fmt.Printf("ANALYZE_RESULT=%s\n", result)
}
