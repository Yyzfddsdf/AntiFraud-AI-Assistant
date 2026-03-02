package tool

import (
	"context"
	"strings"
)

type userIDContextKey struct{}
type taskIDContextKey struct{}
type taskPayloadContextKey struct{}
type taskInsightContextKey struct{}
type finalReportContextKey struct{}

type TaskPayloadContext struct {
	Text   string
	Videos []string
	Audios []string
	Images []string
}

type TaskInsightContext struct {
	VideoInsights []string
	AudioInsights []string
	ImageInsights []string
}

// BindUserID 将当前请求关联的用户 ID 绑定到 ctx。
// 工具侧通过 CurrentUserID 读取，用于按用户维度查询/写入数据。
func BindUserID(ctx context.Context, userID string) context.Context {
	uid := strings.TrimSpace(userID)
	if uid == "" {
		uid = "demo-user"
	}
	return context.WithValue(ctx, userIDContextKey{}, uid)
}

// CurrentUserID 从 ctx 读取用户 ID。
// 若缺失则返回 demo-user，保证工具调用时总能拿到稳定用户标识。
func CurrentUserID(ctx context.Context) string {
	if ctx == nil {
		return "demo-user"
	}
	value := ctx.Value(userIDContextKey{})
	uid, ok := value.(string)
	if !ok {
		return "demo-user"
	}
	uid = strings.TrimSpace(uid)
	if uid == "" {
		return "demo-user"
	}
	return uid
}

// BindTaskID 将当前任务 ID 绑定到 ctx。
// 归档类工具会用它把分析结果和任务记录对齐。
func BindTaskID(ctx context.Context, taskID string) context.Context {
	tid := strings.TrimSpace(taskID)
	return context.WithValue(ctx, taskIDContextKey{}, tid)
}

// CurrentTaskID 从 ctx 读取任务 ID。
// 若调用链中未绑定，返回空字符串。
func CurrentTaskID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	value := ctx.Value(taskIDContextKey{})
	tid, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(tid)
}

// BindTaskPayload 将任务原始输入写入 ctx（文本 + 各模态 base64 列表）。
// write_user_history_case 在归档时会读取该 payload 并持久化到 history_cases。
func BindTaskPayload(ctx context.Context, text string, videos []string, audios []string, images []string) context.Context {
	payload := TaskPayloadContext{
		Text:   strings.TrimSpace(text),
		Videos: append([]string{}, videos...),
		Audios: append([]string{}, audios...),
		Images: append([]string{}, images...),
	}
	return context.WithValue(ctx, taskPayloadContextKey{}, payload)
}

// CurrentTaskPayload 从 ctx 读取原始任务输入。
// 返回拷贝，避免调用方意外修改底层切片。
func CurrentTaskPayload(ctx context.Context) TaskPayloadContext {
	if ctx == nil {
		return TaskPayloadContext{}
	}
	value := ctx.Value(taskPayloadContextKey{})
	payload, ok := value.(TaskPayloadContext)
	if !ok {
		return TaskPayloadContext{}
	}
	return TaskPayloadContext{
		Text:   strings.TrimSpace(payload.Text),
		Videos: append([]string{}, payload.Videos...),
		Audios: append([]string{}, payload.Audios...),
		Images: append([]string{}, payload.Images...),
	}
}

// BindTaskInsights 将子模态分析结果（image/video/audio insights）写入 ctx。
// 归档时会与原始 payload 一并写入历史记录。
func BindTaskInsights(ctx context.Context, videoInsights []string, audioInsights []string, imageInsights []string) context.Context {
	insight := TaskInsightContext{
		VideoInsights: append([]string{}, videoInsights...),
		AudioInsights: append([]string{}, audioInsights...),
		ImageInsights: append([]string{}, imageInsights...),
	}
	return context.WithValue(ctx, taskInsightContextKey{}, insight)
}

// CurrentTaskInsights 从 ctx 读取子模态洞察结果。
// 返回拷贝，避免共享切片被后续流程污染。
func CurrentTaskInsights(ctx context.Context) TaskInsightContext {
	if ctx == nil {
		return TaskInsightContext{}
	}
	value := ctx.Value(taskInsightContextKey{})
	insight, ok := value.(TaskInsightContext)
	if !ok {
		return TaskInsightContext{}
	}
	return TaskInsightContext{
		VideoInsights: append([]string{}, insight.VideoInsights...),
		AudioInsights: append([]string{}, insight.AudioInsights...),
		ImageInsights: append([]string{}, insight.ImageInsights...),
	}
}

// BindFinalReport 将 submit_final_report 产生的最终报告文本写入 ctx。
// write_user_history_case 在归档时读取该字段并落库。
func BindFinalReport(ctx context.Context, report string) context.Context {
	return context.WithValue(ctx, finalReportContextKey{}, strings.TrimSpace(report))
}

// CurrentFinalReport 从 ctx 读取最终报告文本。
// 若上游尚未生成报告，则返回空字符串。
func CurrentFinalReport(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	value := ctx.Value(finalReportContextKey{})
	report, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(report)
}
