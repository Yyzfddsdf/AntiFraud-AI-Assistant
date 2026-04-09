package application

import (
	"context"
	"encoding/base64"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"antifraud/internal/modules/multi_agent/adapters/outbound/state"
)

const (
	maxEncodedMediaBytes        = 10 * 1024 * 1024
	encodedPayloadSafetyReserve = 256 * 1024
	ffmpegRunTimeout            = 2 * time.Minute
	maxVideoFrameCount          = 5000
)

var (
	ffmpegFallbackPaths = []string{
		`D:\ffmpeg\bin\ffmpeg.exe`,
		`D:/ffmpeg/bin/ffmpeg.exe`,
	}
	ffprobeFallbackPaths = []string{
		`D:\ffmpeg\bin\ffprobe.exe`,
		`D:/ffmpeg/bin/ffprobe.exe`,
	}
)

type dataURLPayload struct {
	MIME string
	Raw  []byte
}

type videoPreset struct {
	MaxWidth  int
	MaxFPS    float64
	CRF       int
	AudioRate string
}

type audioPreset struct {
	Bitrate string
	Rate    int
}

// 预设按“优先贴近预算上限，再逐级回退”的顺序排列，避免一上来就过度压缩。
var videoPresets = []videoPreset{
	{MaxWidth: 1600, MaxFPS: 24, CRF: 22, AudioRate: "128k"},
	{MaxWidth: 1280, MaxFPS: 30, CRF: 20, AudioRate: "128k"},
	{MaxWidth: 1280, MaxFPS: 24, CRF: 20, AudioRate: "128k"},
	{MaxWidth: 1280, MaxFPS: 24, CRF: 24, AudioRate: "96k"},
	{MaxWidth: 1280, MaxFPS: 18, CRF: 26, AudioRate: "96k"},
	{MaxWidth: 960, MaxFPS: 18, CRF: 24, AudioRate: "96k"},
	{MaxWidth: 854, MaxFPS: 16, CRF: 24, AudioRate: "96k"},
	{MaxWidth: 960, MaxFPS: 12, CRF: 31, AudioRate: "64k"},
	{MaxWidth: 640, MaxFPS: 8, CRF: 36, AudioRate: "32k"},
	{MaxWidth: 512, MaxFPS: 6, CRF: 38, AudioRate: "24k"},
	{MaxWidth: 426, MaxFPS: 4, CRF: 40, AudioRate: "16k"},
}

var audioPresets = []audioPreset{
	{Bitrate: "128k", Rate: 32000},
	{Bitrate: "96k", Rate: 22050},
	{Bitrate: "64k", Rate: 16000},
	{Bitrate: "48k", Rate: 12000},
	{Bitrate: "32k", Rate: 8000},
	{Bitrate: "16k", Rate: 8000},
	{Bitrate: "8k", Rate: 8000},
}

// NormalizeTaskPayload 在任务入队前对多媒体进行规范化，避免超长、超大输入直接进入后续分析链路。
func NormalizeTaskPayload(payload state.TaskPayload) (state.TaskPayload, error) {
	normalized := state.TaskPayload{
		Text:          strings.TrimSpace(payload.Text),
		Images:        append([]string{}, payload.Images...),
		VideoInsights: append([]string{}, payload.VideoInsights...),
		AudioInsights: append([]string{}, payload.AudioInsights...),
		ImageInsights: append([]string{}, payload.ImageInsights...),
	}

	normalized.Videos = make([]string, 0, len(payload.Videos))
	for idx, item := range payload.Videos {
		video, err := normalizeVideoInput(item)
		if err != nil {
			return state.TaskPayload{}, fmt.Errorf("视频 %d 预处理失败: %w", idx+1, err)
		}
		normalized.Videos = append(normalized.Videos, video)
	}

	normalized.Audios = make([]string, 0, len(payload.Audios))
	for idx, item := range payload.Audios {
		audio, err := normalizeAudioData(item)
		if err != nil {
			return state.TaskPayload{}, fmt.Errorf("音频 %d 预处理失败: %w", idx+1, err)
		}
		normalized.Audios = append(normalized.Audios, audio)
	}

	return normalized, nil
}

func normalizeVideoInput(input string) (string, error) {
	payload, err := decodeMediaInput(input, "video/mp4")
	if err != nil {
		return "", err
	}
	return encodeDataURL(payload.MIME, payload.Raw), nil
}

func normalizeAudioData(input string) (string, error) {
	payload, err := decodeMediaInput(input, "audio/mpeg")
	if err != nil {
		return "", err
	}

	output, err := transcodeWithFFmpeg(payload.Raw, mimeToExtension(payload.MIME, ".mp3"), ".mp3", func(inPath string, outPath string) error {
		durationSeconds, probeErr := probeMediaDurationSeconds(inPath)
		if probeErr != nil {
			return fmt.Errorf("读取音频时长失败: %w", probeErr)
		}
		targetBitrateKbps := calculateTargetAudioBitrateKbps(durationSeconds)
		for _, preset := range candidateAudioPresets(targetBitrateKbps) {
			args := []string{
				"-y",
				"-i", inPath,
				"-map", "0:a:0",
				"-vn",
				"-ac", "1",
				"-ar", fmt.Sprintf("%d", preset.Rate),
				"-c:a", "libmp3lame",
				"-b:a", preset.Bitrate,
				outPath,
			}
			ok, runErr := runFFmpegAttempt(args, outPath)
			if runErr != nil {
				return runErr
			}
			if ok {
				return nil
			}
		}
		return fmt.Errorf("压缩后仍超过大小限制 %d 字节", maxRawMediaBytes())
	})
	if err != nil {
		return "", err
	}

	return encodeDataURL("audio/mpeg", output), nil
}

func transcodeWithFFmpeg(raw []byte, inputExt string, outputExt string, process func(inPath string, outPath string) error) ([]byte, error) {
	tempDir, err := os.MkdirTemp("", "sentinel-media-*")
	if err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer os.RemoveAll(tempDir)

	inputPath := filepath.Join(tempDir, "input"+inputExt)
	if err := os.WriteFile(inputPath, raw, 0o600); err != nil {
		return nil, fmt.Errorf("写入输入媒体失败: %w", err)
	}

	outputPath := filepath.Join(tempDir, "output"+outputExt)
	if err := process(inputPath, outputPath); err != nil {
		return nil, err
	}

	bytes, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, fmt.Errorf("读取处理结果失败: %w", err)
	}
	if len(bytes) == 0 {
		return nil, fmt.Errorf("ffmpeg 输出为空")
	}
	if len(bytes) > maxRawMediaBytes() {
		return nil, fmt.Errorf("处理结果仍超过大小限制 %d 字节", maxRawMediaBytes())
	}
	return bytes, nil
}

func runFFmpegAttempt(args []string, outPath string) (bool, error) {
	ffmpegPath, err := lookupFFmpegPath()
	if err != nil {
		return false, err
	}

	cmdCtx, cancel := context.WithTimeout(context.Background(), ffmpegRunTimeout)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, ffmpegPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if cmdCtx.Err() == context.DeadlineExceeded {
			return false, fmt.Errorf("ffmpeg 处理超时: %w", err)
		}
		return false, fmt.Errorf("ffmpeg 处理失败: %s", strings.TrimSpace(string(output)))
	}

	info, err := os.Stat(outPath)
	if err != nil {
		return false, fmt.Errorf("读取 ffmpeg 输出文件失败: %w", err)
	}
	if info.Size() <= 0 {
		return false, fmt.Errorf("ffmpeg 输出文件为空")
	}
	return info.Size() <= int64(maxRawMediaBytes()), nil
}

func lookupFFmpegPath() (string, error) {
	if path, err := exec.LookPath("ffmpeg"); err == nil && strings.TrimSpace(path) != "" {
		return path, nil
	}
	for _, candidate := range ffmpegFallbackPaths {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("未找到 ffmpeg，可执行文件需在 PATH 中或位于 D:\\ffmpeg\\bin\\ffmpeg.exe")
}

func lookupFFprobePath() (string, error) {
	if path, err := exec.LookPath("ffprobe"); err == nil && strings.TrimSpace(path) != "" {
		return path, nil
	}
	for _, candidate := range ffprobeFallbackPaths {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("未找到 ffprobe，可执行文件需在 PATH 中或位于 D:\\ffmpeg\\bin\\ffprobe.exe")
}

func maxRawMediaBytes() int {
	rawBudget := (maxEncodedMediaBytes * 3 / 4) - encodedPayloadSafetyReserve
	if rawBudget < 1 {
		return 1
	}
	return rawBudget
}

func probeMediaDurationSeconds(path string) (float64, error) {
	ffprobePath, err := lookupFFprobePath()
	if err != nil {
		return 0, err
	}

	cmdCtx, cancel := context.WithTimeout(context.Background(), ffmpegRunTimeout)
	defer cancel()

	args := []string{
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path,
	}
	cmd := exec.CommandContext(cmdCtx, ffprobePath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if cmdCtx.Err() == context.DeadlineExceeded {
			return 0, fmt.Errorf("ffprobe 处理超时: %w", err)
		}
		return 0, fmt.Errorf("ffprobe 执行失败: %s", strings.TrimSpace(string(output)))
	}

	durationSeconds, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	if err != nil {
		return 0, fmt.Errorf("解析媒体时长失败: %w", err)
	}
	if durationSeconds <= 0 {
		return 0, fmt.Errorf("媒体时长无效: %.4f", durationSeconds)
	}
	return durationSeconds, nil
}

func calculateTargetFPS(durationSeconds float64) float64 {
	if durationSeconds <= 0 {
		return 1
	}
	fps := float64(maxVideoFrameCount) / durationSeconds
	if fps < 1 {
		return 1
	}
	return fps
}

func calculateTargetAudioBitrateKbps(durationSeconds float64) int {
	if durationSeconds <= 0 {
		return 8
	}
	rawBudgetBits := float64(maxRawMediaBytes() * 8)
	bitrate := int(math.Floor(rawBudgetBits / durationSeconds / 1000))
	if bitrate < 8 {
		return 8
	}
	return bitrate
}

func candidateAudioPresets(targetBitrateKbps int) []audioPreset {
	filtered := make([]audioPreset, 0, len(audioPresets))
	for _, preset := range audioPresets {
		presetBitrate := parseBitrateKbps(preset.Bitrate)
		if presetBitrate <= targetBitrateKbps {
			filtered = append(filtered, preset)
		}
	}
	if len(filtered) > 0 {
		return filtered
	}
	return []audioPreset{audioPresets[len(audioPresets)-1]}
}

func parseBitrateKbps(value string) int {
	trimmed := strings.ToLower(strings.TrimSpace(value))
	trimmed = strings.TrimSuffix(trimmed, "k")
	n, err := strconv.Atoi(trimmed)
	if err != nil || n <= 0 {
		return 0
	}
	return n
}

func formatFPS(value float64) string {
	rounded := math.Round(value*1000) / 1000
	return strconv.FormatFloat(rounded, 'f', -1, 64)
}

func decodeMediaInput(input string, fallbackMIME string) (dataURLPayload, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return dataURLPayload{}, fmt.Errorf("媒体内容为空")
	}

	if strings.HasPrefix(trimmed, "data:") {
		parts := strings.SplitN(trimmed, ",", 2)
		if len(parts) != 2 {
			return dataURLPayload{}, fmt.Errorf("data url 格式不正确")
		}

		header := strings.TrimPrefix(parts[0], "data:")
		if !strings.Contains(header, ";base64") {
			return dataURLPayload{}, fmt.Errorf("仅支持 base64 data url")
		}

		mimeType := strings.TrimSpace(strings.TrimSuffix(header, ";base64"))
		if mimeType == "" {
			mimeType = fallbackMIME
		}

		raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(parts[1]))
		if err != nil {
			return dataURLPayload{}, fmt.Errorf("base64 解码失败: %w", err)
		}
		if len(raw) == 0 {
			return dataURLPayload{}, fmt.Errorf("媒体内容为空")
		}
		return dataURLPayload{MIME: mimeType, Raw: raw}, nil
	}

	raw, err := base64.StdEncoding.DecodeString(trimmed)
	if err != nil {
		return dataURLPayload{}, fmt.Errorf("base64 解码失败: %w", err)
	}
	if len(raw) == 0 {
		return dataURLPayload{}, fmt.Errorf("媒体内容为空")
	}
	return dataURLPayload{MIME: fallbackMIME, Raw: raw}, nil
}

func encodeDataURL(mimeType string, raw []byte) string {
	return fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(raw))
}

func mimeToExtension(mimeType string, fallback string) string {
	lower := strings.ToLower(strings.TrimSpace(mimeType))
	switch lower {
	case "video/mp4":
		return ".mp4"
	case "video/quicktime":
		return ".mov"
	case "video/x-matroska":
		return ".mkv"
	case "video/webm":
		return ".webm"
	case "audio/mpeg", "audio/mp3":
		return ".mp3"
	case "audio/mp4", "audio/aac":
		return ".m4a"
	case "audio/wav", "audio/x-wav":
		return ".wav"
	case "audio/ogg":
		return ".ogg"
	default:
		return fallback
	}
}
