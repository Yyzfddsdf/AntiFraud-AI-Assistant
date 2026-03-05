package auth

import (
	"errors"
	"testing"
)

func TestIssueAndParseToken(t *testing.T) {
	token, err := IssueToken(123, "alice@example.com", "alice")
	if err != nil {
		t.Fatalf("IssueToken returned error: %v", err)
	}
	if token == "" {
		t.Fatalf("IssueToken returned empty token")
	}

	claims, err := ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken returned error: %v", err)
	}
	if claims == nil {
		t.Fatalf("ParseToken returned nil claims")
	}
	if claims.UserID != 123 {
		t.Fatalf("unexpected user id: got=%d want=%d", claims.UserID, 123)
	}
	if claims.Email != "alice@example.com" {
		t.Fatalf("unexpected email: got=%s", claims.Email)
	}
	if claims.Username != "alice" {
		t.Fatalf("unexpected username: got=%s", claims.Username)
	}
}

func TestParseTokenRejectsEmptyToken(t *testing.T) {
	_, err := ParseToken(" ")
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got=%v", err)
	}
}
