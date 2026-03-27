package auth_test

import (
	"errors"
	"testing"

	auth "antifraud/internal/modules/login/domain/auth"
)

func TestIssueAndParseToken(t *testing.T) {
	token, err := auth.IssueToken(123, "alice@example.com", "alice")
	if err != nil {
		t.Fatalf("IssueToken returned error: %v", err)
	}
	if token == "" {
		t.Fatalf("IssueToken returned empty token")
	}

	claims, err := auth.ParseToken(token)
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
	_, err := auth.ParseToken(" ")
	if !errors.Is(err, auth.ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got=%v", err)
	}
}

func TestIssueTokenGeneratesUniqueTokenIDs(t *testing.T) {
	first, err := auth.IssueToken(123, "alice@example.com", "alice")
	if err != nil {
		t.Fatalf("IssueToken first returned error: %v", err)
	}
	second, err := auth.IssueToken(123, "alice@example.com", "alice")
	if err != nil {
		t.Fatalf("IssueToken second returned error: %v", err)
	}
	if first == second {
		t.Fatalf("expected distinct tokens for repeated issue calls")
	}

	firstClaims, err := auth.ParseToken(first)
	if err != nil {
		t.Fatalf("ParseToken first returned error: %v", err)
	}
	secondClaims, err := auth.ParseToken(second)
	if err != nil {
		t.Fatalf("ParseToken second returned error: %v", err)
	}
	if firstClaims.ID == "" || secondClaims.ID == "" {
		t.Fatalf("expected non-empty token ids: first=%q second=%q", firstClaims.ID, secondClaims.ID)
	}
	if firstClaims.ID == secondClaims.ID {
		t.Fatalf("expected distinct token ids")
	}
}
