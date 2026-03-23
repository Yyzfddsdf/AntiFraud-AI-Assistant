package multi_agent_test

import (
	"testing"
)

func TestBuildVideoDataURL_DataURLPassthrough(t *testing.T) {
	input := "data:video/mp4;base64,AAAA"
	got, err := buildVideoDataURL(input)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got != input {
		t.Fatalf("expected passthrough data url, got %q", got)
	}
}

func TestBuildVideoDataURL_InvalidDataURL(t *testing.T) {
	_, err := buildVideoDataURL("data:video/mp4,AAAA")
	if err == nil {
		t.Fatalf("expected error for invalid data url")
	}
}

func TestBuildVideoDataURL_EmptyInput(t *testing.T) {
	_, err := buildVideoDataURL("   ")
	if err == nil {
		t.Fatalf("expected error for empty input")
	}
}

func TestBuildVideoDataURL_Base64Wrap(t *testing.T) {
	got, err := buildVideoDataURL("  QUJD  ")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	want := "data:video/mp4;base64,QUJD"
	if got != want {
		t.Fatalf("unexpected data url: want %q, got %q", want, got)
	}
}
