package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	openai "antifraud/llm"
)

const (
	defaultAPIKey          = "sk-28b1d19d210ce836717a2ee225cb0daa9bce75215ee33cfe710c9b4a4dc6bbd6"
	defaultBaseURL         = "https://gmn.chuangzuoli.com/v1"
	defaultModel           = "gpt-5.2"
	defaultTimeoutSeconds  = 30.0
	defaultUseWebSearch    = true
	defaultDeveloperPrompt = "你是一个简洁、友好的中文助手。遇到实时信息时优先使用 web_search。"
)

type appConfig struct {
	APIKey          string
	BaseURL         string
	Model           string
	Timeout         time.Duration
	UseWebSearch    bool
	DeveloperPrompt string
}

func main() {
	config, err := loadConfig()
	if err != nil {
		fmt.Printf("启动失败: %v\n", err)
		os.Exit(1)
	}

	client := openai.NewClientWithConfig(openai.Config{
		APIKey:  config.APIKey,
		BaseURL: config.BaseURL,
		HTTPClient: &http.Client{
			Timeout: config.Timeout,
		},
	})

	if len(os.Args) > 1 {
		if err := runOneShotImageQuestion(context.Background(), client, config, os.Args[1:]); err != nil {
			fmt.Printf("请求失败: %v\n", err)
			os.Exit(1)
		}
		return
	}

	history := []openai.ResponsesMessage{
		openai.MakeResponsesMessageText("developer", config.DeveloperPrompt),
	}

	fmt.Printf("base_url=%s\n", config.BaseURL)
	fmt.Printf("model=%s\n", config.Model)
	fmt.Printf("timeout=%gs\n", config.Timeout.Seconds())
	fmt.Println("输入 exit 退出，输入 clear 清空历史。")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("你: ")
		userText, readErr := reader.ReadString('\n')
		if readErr != nil && !errors.Is(readErr, io.EOF) {
			fmt.Printf("读取输入失败: %v\n", readErr)
			os.Exit(1)
		}

		userText = strings.TrimSpace(userText)
		if userText == "" {
			if errors.Is(readErr, io.EOF) {
				fmt.Println("已退出。")
				return
			}
			continue
		}

		switch strings.ToLower(userText) {
		case "exit", "quit":
			fmt.Println("已退出。")
			return
		case "clear":
			history = []openai.ResponsesMessage{
				openai.MakeResponsesMessageText("developer", config.DeveloperPrompt),
			}
			fmt.Println("历史已清空。")
			fmt.Println()
			if errors.Is(readErr, io.EOF) {
				fmt.Println("已退出。")
				return
			}
			continue
		}

		history = append(history, openai.MakeResponsesMessageText("user", userText))

		request := openai.ResponsesRequest{
			Model:    config.Model,
			Messages: history,
		}
		if config.UseWebSearch {
			request.Tools = []openai.Tool{{Type: "web_search"}}
		}

		answer, outputTypes, err := collectStreamText(context.Background(), client, request)
		if err != nil {
			fmt.Printf("请求失败: %v\n\n", err)
			history = history[:len(history)-1]
			if errors.Is(readErr, io.EOF) {
				fmt.Println("已退出。")
				return
			}
			continue
		}

		if len(outputTypes) > 0 {
			fmt.Printf("输出类型: %v\n", outputTypes)
		}

		history = append(history, openai.MakeResponsesMessageOutputText("assistant", answer))
		fmt.Println()

		if errors.Is(readErr, io.EOF) {
			fmt.Println("已退出。")
			return
		}
	}
}

func runOneShotImageQuestion(ctx context.Context, client *openai.Client, config *appConfig, args []string) error {
	imagePath := strings.TrimSpace(args[0])
	question := "请描述这张图片。"
	if len(args) > 1 {
		question = strings.TrimSpace(strings.Join(args[1:], " "))
	}
	if question == "" {
		question = "请描述这张图片。"
	}

	imageDataURL, err := makeImageDataURL(imagePath)
	if err != nil {
		return err
	}

	request := openai.ResponsesRequest{
		Model: config.Model,
		Messages: []openai.ResponsesMessage{
			openai.MakeResponsesMessageText("developer", config.DeveloperPrompt),
			{
				Type: "message",
				Role: "user",
				Content: []openai.ResponsesMessagePart{
					{
						Type: "input_text",
						Text: question,
					},
					{
						Type:     "input_image",
						ImageURL: imageDataURL,
						Detail:   "auto",
					},
				},
			},
		},
	}
	if config.UseWebSearch {
		request.Tools = []openai.Tool{{Type: "web_search"}}
	}

	fmt.Printf("image=%s\n", imagePath)
	fmt.Printf("question=%s\n", question)
	answer, outputTypes, err := collectStreamText(ctx, client, request)
	if err != nil {
		return err
	}
	if len(outputTypes) > 0 {
		fmt.Printf("输出类型: %v\n", outputTypes)
	}
	if strings.TrimSpace(answer) == "" {
		fmt.Println("回答为空。")
	}
	return nil
}

func collectStreamText(ctx context.Context, client *openai.Client, request openai.ResponsesRequest) (string, []string, error) {
	stream, err := client.CreateResponsesStream(ctx, request)
	if err != nil {
		return "", nil, err
	}
	defer stream.Close()

	var answerBuilder strings.Builder
	outputTypes := make([]string, 0, 4)

	fmt.Print("助手: ")
	for {
		event, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			fmt.Println()
			return "", outputTypes, err
		}

		switch event.Type {
		case "response.output_text.delta":
			if event.Delta == "" {
				continue
			}
			answerBuilder.WriteString(event.Delta)
			fmt.Print(event.Delta)
		case "response.output_item.added":
			if event.Item != nil && event.Item.Type != "" {
				outputTypes = append(outputTypes, event.Item.Type)
			}
		}
	}

	fmt.Println()
	return strings.TrimSpace(answerBuilder.String()), outputTypes, nil
}

func loadConfig() (*appConfig, error) {
	apiKey := firstNonEmpty(
		os.Getenv("MANUAL_OPENAI_API_KEY"),
		os.Getenv("OPENAI_API_KEY"),
		os.Getenv("API_KEY"),
		defaultAPIKey,
	)

	baseURL := firstNonEmpty(
		os.Getenv("MANUAL_OPENAI_BASE_URL"),
		os.Getenv("BASE_URL"),
		defaultBaseURL,
	)
	model := firstNonEmpty(
		os.Getenv("MANUAL_OPENAI_MODEL"),
		os.Getenv("MODEL"),
		defaultModel,
	)
	developerPrompt := firstNonEmpty(
		os.Getenv("MANUAL_OPENAI_DEVELOPER_PROMPT"),
		os.Getenv("DEVELOPER_PROMPT"),
		defaultDeveloperPrompt,
	)

	timeoutSeconds := defaultTimeoutSeconds
	if value := firstNonEmpty(os.Getenv("MANUAL_OPENAI_TIMEOUT"), os.Getenv("TIMEOUT")); value != "" {
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, fmt.Errorf("解析超时配置失败: %w", err)
		}
		timeoutSeconds = parsed
	}

	useWebSearch := defaultUseWebSearch
	if value := firstNonEmpty(os.Getenv("MANUAL_OPENAI_USE_WEB_SEARCH"), os.Getenv("USE_WEB_SEARCH")); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return nil, fmt.Errorf("解析 USE_WEB_SEARCH 失败: %w", err)
		}
		useWebSearch = parsed
	}

	return &appConfig{
		APIKey:          apiKey,
		BaseURL:         baseURL,
		Model:           model,
		Timeout:         time.Duration(timeoutSeconds * float64(time.Second)),
		UseWebSearch:    useWebSearch,
		DeveloperPrompt: developerPrompt,
	}, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func makeImageDataURL(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("读取图片失败: %w", err)
	}

	contentType := detectImageContentType(path)
	encoded := base64.StdEncoding.EncodeToString(data)
	return "data:" + contentType + ";base64," + encoded, nil
}

func detectImageContentType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	if contentType := mime.TypeByExtension(ext); contentType != "" {
		return contentType
	}

	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	case ".bmp":
		return "image/bmp"
	default:
		return "application/octet-stream"
	}
}
