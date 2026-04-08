package application_test

import (
	"strings"
	"testing"

	"antifraud/internal/modules/multi_agent/adapters/outbound/state"
	"antifraud/internal/modules/multi_agent/application"
)

func TestNormalizeTaskPayload_TextAndImagesPassthrough(t *testing.T) {
	input := state.TaskPayload{
		Text:   "  hello  ",
		Images: []string{"data:image/png;base64,AAAA"},
	}

	got, err := application.NormalizeTaskPayload(input)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if got.Text != "hello" {
		t.Fatalf("unexpected trimmed text: %q", got.Text)
	}
	if len(got.Images) != 1 || got.Images[0] != input.Images[0] {
		t.Fatalf("unexpected images passthrough: %#v", got.Images)
	}
	if len(got.Videos) != 0 || len(got.Audios) != 0 {
		t.Fatalf("expected empty media outputs, got videos=%d audios=%d", len(got.Videos), len(got.Audios))
	}
}

func TestNormalizeTaskPayload_InvalidVideoBase64(t *testing.T) {
	_, err := application.NormalizeTaskPayload(state.TaskPayload{
		Videos: []string{"%%%"},
	})
	if err == nil {
		t.Fatal("expected invalid video error")
	}
	if !strings.Contains(err.Error(), "视频 1 预处理失败") {
		t.Fatalf("unexpected error: %v", err)
	}
}
