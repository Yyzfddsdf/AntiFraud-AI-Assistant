package multi_agent

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestBuildImageDataURL_DataURLPassthrough(t *testing.T) {
	input := "data:image/png;base64,AAAA"
	got, err := buildImageDataURL(input)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got != input {
		t.Fatalf("expected passthrough data url, got %q", got)
	}
}

func TestBuildImageDataURL_InvalidDataURL(t *testing.T) {
	_, err := buildImageDataURL("data:image/png,AAAA")
	if err == nil {
		t.Fatalf("expected error for invalid data url")
	}
}

func TestBuildImageDataURL_EmptyInput(t *testing.T) {
	_, err := buildImageDataURL("   ")
	if err == nil {
		t.Fatalf("expected error for empty input")
	}
}

func TestBuildImageDataURL_InvalidBase64(t *testing.T) {
	_, err := buildImageDataURL("%%%")
	if err == nil {
		t.Fatalf("expected error for invalid base64")
	}
}

func TestBuildImageDataURL_NormalBase64(t *testing.T) {
	got, err := buildImageDataURL("aGVsbG8=") // "hello"
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !strings.HasPrefix(got, "data:text/plain") {
		t.Fatalf("unexpected mime prefix: %q", got)
	}
	if !strings.Contains(got, ";base64,aGVsbG8=") {
		t.Fatalf("expected normalized base64 payload, got %q", got)
	}
}

func TestBuildImageDataURL_OctetStreamFallbackToJpeg(t *testing.T) {
	raw := make([]byte, 32)
	for i := range raw {
		raw[i] = byte(i + 1)
	}
	encoded := base64.StdEncoding.EncodeToString(raw)

	got, err := buildImageDataURL(encoded)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !strings.HasPrefix(got, "data:image/jpeg") {
		t.Fatalf("expected jpeg fallback, got %q", got)
	}
}
