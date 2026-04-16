package multi_agent

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

	appcfg "antifraud/internal/platform/config"
)

const (
	videoAgentEncodedLimitBytes = 10 * 1024 * 1024
	videoAgentSafetyReserve     = 256 * 1024
	videoAgentMaxFrameCount     = 5000
	videoAgentASRMaxSeconds     = 300
	videoAgentFFmpegTimeout     = 2 * time.Minute
)

type videoCompressionPreset struct {
	MaxWidth  int
	MaxFPS    float64
	CRF       int
	AudioRate string
}

type audioCompressionPreset struct {
	Bitrate string
	Rate    int
}

var (
	videoAgentVideoPresets = []videoCompressionPreset{
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
	videoAgentAudioPresets = []audioCompressionPreset{
		{Bitrate: "128k", Rate: 32000},
		{Bitrate: "96k", Rate: 22050},
		{Bitrate: "64k", Rate: 16000},
		{Bitrate: "48k", Rate: 12000},
		{Bitrate: "32k", Rate: 8000},
		{Bitrate: "16k", Rate: 8000},
		{Bitrate: "8k", Rate: 8000},
	}
)

func compressVideoForAnalysis(videoInput string) (string, error) {
	raw, err := decodeVideoInputRaw(videoInput)
	if err != nil {
		return "", err
	}

	output, err := transcodeVideoBranch(raw)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("data:video/mp4;base64,%s", base64.StdEncoding.EncodeToString(output)), nil
}

func extractAudioTrackForASR(videoInput string) (string, error) {
	raw, err := decodeVideoInputRaw(videoInput)
	if err != nil {
		return "", err
	}

	output, err := transcodeASRAudioBranch(raw)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("data:audio/mpeg;base64,%s", base64.StdEncoding.EncodeToString(output)), nil
}

func transcodeVideoBranch(raw []byte) ([]byte, error) {
	return transcodeVideoMedia(raw, ".mp4", func(inPath string, outPath string) error {
		durationSeconds, probeErr := probeCoreMediaDurationSeconds(inPath)
		if probeErr != nil {
			return fmt.Errorf("读取视频时长失败: %w", probeErr)
		}
		targetFPS := calculateCoreTargetFPS(durationSeconds)
		for _, preset := range videoAgentVideoPresets {
			fps := math.Min(preset.MaxFPS, targetFPS)
			args := []string{
				"-y",
				"-i", inPath,
				"-map", "0:v:0",
				"-map", "0:a?",
				"-vf", fmt.Sprintf("scale=trunc(min(%d\\,iw)/2)*2:-2,fps=%s", preset.MaxWidth, formatCoreFPS(fps)),
				"-c:v", "libx264",
				"-preset", "veryfast",
				"-crf", fmt.Sprintf("%d", preset.CRF),
				"-pix_fmt", "yuv420p",
				"-movflags", "+faststart",
				"-c:a", "aac",
				"-ac", "1",
				"-ar", "16000",
				"-b:a", preset.AudioRate,
				outPath,
			}
			ok, runErr := runCoreFFmpegAttempt(args, outPath)
			if runErr != nil {
				return runErr
			}
			if ok {
				return nil
			}
		}
		return fmt.Errorf("压缩后仍超过大小限制 %d 字节", coreMaxRawMediaBytes())
	})
}

func transcodeASRAudioBranch(raw []byte) ([]byte, error) {
	return transcodeAudioMedia(raw, ".mp4", ".mp3", func(inPath string, outPath string) error {
		durationSeconds, probeErr := probeCoreMediaDurationSeconds(inPath)
		if probeErr != nil {
			return fmt.Errorf("读取视频时长失败: %w", probeErr)
		}

		limitedDuration := math.Min(durationSeconds, float64(videoAgentASRMaxSeconds))
		targetBitrateKbps := calculateCoreTargetAudioBitrateKbps(limitedDuration)
		for _, preset := range candidateCoreAudioPresets(targetBitrateKbps) {
			args := []string{
				"-y",
				"-i", inPath,
				"-map", "0:a:0?",
				"-vn",
				"-t", fmt.Sprintf("%d", videoAgentASRMaxSeconds),
				"-ac", "1",
				"-ar", fmt.Sprintf("%d", preset.Rate),
				"-c:a", "libmp3lame",
				"-b:a", preset.Bitrate,
				outPath,
			}
			ok, runErr := runCoreFFmpegAttempt(args, outPath)
			if runErr != nil {
				return runErr
			}
			if ok {
				return nil
			}
		}
		return fmt.Errorf("提取音轨后仍超过大小限制 %d 字节", coreMaxRawMediaBytes())
	})
}

func transcodeVideoMedia(raw []byte, inputExt string, process func(inPath string, outPath string) error) ([]byte, error) {
	tempDir, err := os.MkdirTemp("", "sentinel-video-branch-*")
	if err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer os.RemoveAll(tempDir)

	inputPath := filepath.Join(tempDir, "input"+inputExt)
	if err := os.WriteFile(inputPath, raw, 0o600); err != nil {
		return nil, fmt.Errorf("写入视频临时文件失败: %w", err)
	}
	outputPath := filepath.Join(tempDir, "output.mp4")

	if err := process(inputPath, outputPath); err != nil {
		return nil, err
	}
	return readCoreOutputWithinLimit(outputPath)
}

func transcodeAudioMedia(raw []byte, inputExt string, outputExt string, process func(inPath string, outPath string) error) ([]byte, error) {
	tempDir, err := os.MkdirTemp("", "sentinel-audio-branch-*")
	if err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer os.RemoveAll(tempDir)

	inputPath := filepath.Join(tempDir, "input"+inputExt)
	if err := os.WriteFile(inputPath, raw, 0o600); err != nil {
		return nil, fmt.Errorf("写入媒体临时文件失败: %w", err)
	}
	outputPath := filepath.Join(tempDir, "output"+outputExt)

	if err := process(inputPath, outputPath); err != nil {
		return nil, err
	}
	return readCoreOutputWithinLimit(outputPath)
}

func readCoreOutputWithinLimit(path string) ([]byte, error) {
	output, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取处理结果失败: %w", err)
	}
	if len(output) == 0 {
		return nil, fmt.Errorf("ffmpeg 输出为空")
	}
	if len(output) > coreMaxRawMediaBytes() {
		return nil, fmt.Errorf("处理结果仍超过大小限制 %d 字节", coreMaxRawMediaBytes())
	}
	return output, nil
}

func decodeVideoInputRaw(input string) ([]byte, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return nil, fmt.Errorf("视频内容为空")
	}

	if strings.HasPrefix(trimmed, "data:") {
		parts := strings.SplitN(trimmed, ",", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("video data url 格式不正确")
		}
		payload := strings.TrimSpace(parts[1])
		if payload == "" {
			return nil, fmt.Errorf("视频内容为空")
		}
		raw, err := base64.StdEncoding.DecodeString(payload)
		if err != nil {
			return nil, fmt.Errorf("video base64 解码失败: %w", err)
		}
		if len(raw) == 0 {
			return nil, fmt.Errorf("视频内容为空")
		}
		return raw, nil
	}

	raw, err := base64.StdEncoding.DecodeString(trimmed)
	if err != nil {
		return nil, fmt.Errorf("video base64 解码失败: %w", err)
	}
	if len(raw) == 0 {
		return nil, fmt.Errorf("视频内容为空")
	}
	return raw, nil
}

func runCoreFFmpegAttempt(args []string, outPath string) (bool, error) {
	ffmpegPath, err := lookupCoreFFmpegPath()
	if err != nil {
		return false, err
	}

	cmdCtx, cancel := context.WithTimeout(context.Background(), videoAgentFFmpegTimeout)
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
	return info.Size() <= int64(coreMaxRawMediaBytes()), nil
}

func lookupCoreFFmpegPath() (string, error) {
	configured := ""
	if cfg, err := appcfg.LoadConfig("internal/platform/config/config.json"); err == nil && cfg != nil {
		configured = strings.TrimSpace(cfg.MediaTools.FFmpegPath)
	}
	return resolveCoreBinaryPath("ffmpeg", configured)
}

func lookupCoreFFprobePath() (string, error) {
	configured := ""
	if cfg, err := appcfg.LoadConfig("internal/platform/config/config.json"); err == nil && cfg != nil {
		configured = strings.TrimSpace(cfg.MediaTools.FFprobePath)
	}
	return resolveCoreBinaryPath("ffprobe", configured)
}

func resolveCoreBinaryPath(toolName string, configured string) (string, error) {
	trimmed := strings.TrimSpace(configured)
	if trimmed == "" {
		trimmed = toolName
	}

	hasPathHint := filepath.IsAbs(trimmed) || strings.Contains(trimmed, "/") || strings.Contains(trimmed, "\\")
	if hasPathHint {
		info, err := os.Stat(trimmed)
		if err == nil && !info.IsDir() {
			return trimmed, nil
		}
		return "", fmt.Errorf("未找到 %s，可执行文件路径无效: %s。请先安装 ffmpeg，并在 config.json 的 media_tools 中配置路径", toolName, trimmed)
	}

	if path, err := exec.LookPath(trimmed); err == nil && strings.TrimSpace(path) != "" {
		return path, nil
	}

	return "", fmt.Errorf("未找到 %s。请先安装 ffmpeg，并在 config.json 的 media_tools 中配置路径，或将 %s 加入 PATH", toolName, trimmed)
}

func probeCoreMediaDurationSeconds(path string) (float64, error) {
	ffprobePath, err := lookupCoreFFprobePath()
	if err != nil {
		return 0, err
	}

	cmdCtx, cancel := context.WithTimeout(context.Background(), videoAgentFFmpegTimeout)
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

func coreMaxRawMediaBytes() int {
	rawBudget := (videoAgentEncodedLimitBytes * 3 / 4) - videoAgentSafetyReserve
	if rawBudget < 1 {
		return 1
	}
	return rawBudget
}

func calculateCoreTargetFPS(durationSeconds float64) float64 {
	if durationSeconds <= 0 {
		return 1
	}
	fps := float64(videoAgentMaxFrameCount) / durationSeconds
	if fps < 1 {
		return 1
	}
	return fps
}

func calculateCoreTargetAudioBitrateKbps(durationSeconds float64) int {
	if durationSeconds <= 0 {
		return 8
	}
	rawBudgetBits := float64(coreMaxRawMediaBytes() * 8)
	bitrate := int(math.Floor(rawBudgetBits / durationSeconds / 1000))
	if bitrate < 8 {
		return 8
	}
	return bitrate
}

func candidateCoreAudioPresets(targetBitrateKbps int) []audioCompressionPreset {
	filtered := make([]audioCompressionPreset, 0, len(videoAgentAudioPresets))
	for _, preset := range videoAgentAudioPresets {
		if parseCoreBitrateKbps(preset.Bitrate) <= targetBitrateKbps {
			filtered = append(filtered, preset)
		}
	}
	if len(filtered) > 0 {
		return filtered
	}
	return []audioCompressionPreset{videoAgentAudioPresets[len(videoAgentAudioPresets)-1]}
}

func parseCoreBitrateKbps(value string) int {
	trimmed := strings.ToLower(strings.TrimSpace(value))
	trimmed = strings.TrimSuffix(trimmed, "k")
	n, err := strconv.Atoi(trimmed)
	if err != nil || n <= 0 {
		return 0
	}
	return n
}

func formatCoreFPS(value float64) string {
	rounded := math.Round(value*1000) / 1000
	return strconv.FormatFloat(rounded, 'f', -1, 64)
}
