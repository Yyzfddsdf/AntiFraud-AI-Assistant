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

func BindUserID(ctx context.Context, userID string) context.Context {
	uid := strings.TrimSpace(userID)
	if uid == "" {
		uid = "demo-user"
	}
	return context.WithValue(ctx, userIDContextKey{}, uid)
}

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

func BindTaskID(ctx context.Context, taskID string) context.Context {
	tid := strings.TrimSpace(taskID)
	return context.WithValue(ctx, taskIDContextKey{}, tid)
}

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

func BindTaskPayload(ctx context.Context, text string, videos []string, audios []string, images []string) context.Context {
	payload := TaskPayloadContext{
		Text:   strings.TrimSpace(text),
		Videos: append([]string{}, videos...),
		Audios: append([]string{}, audios...),
		Images: append([]string{}, images...),
	}
	return context.WithValue(ctx, taskPayloadContextKey{}, payload)
}

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

func BindTaskInsights(ctx context.Context, videoInsights []string, audioInsights []string, imageInsights []string) context.Context {
	insight := TaskInsightContext{
		VideoInsights: append([]string{}, videoInsights...),
		AudioInsights: append([]string{}, audioInsights...),
		ImageInsights: append([]string{}, imageInsights...),
	}
	return context.WithValue(ctx, taskInsightContextKey{}, insight)
}

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

func BindFinalReport(ctx context.Context, report string) context.Context {
	return context.WithValue(ctx, finalReportContextKey{}, strings.TrimSpace(report))
}

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
