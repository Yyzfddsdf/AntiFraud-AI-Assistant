package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"antifraud/login_system/middleware"

	"github.com/gin-gonic/gin"
)

type stubActiveTokenLimiter struct {
	touchedUserID uint
	touchedToken  string
	allowed       bool
	err           error
}

func (s *stubActiveTokenLimiter) RegisterToken(_ context.Context, userID uint, tokenString string) error {
	s.touchedUserID = userID
	s.touchedToken = tokenString
	return s.err
}

func (s *stubActiveTokenLimiter) AllowRequestToken(_ context.Context, userID uint, tokenString string) (bool, error) {
	s.touchedUserID = userID
	s.touchedToken = tokenString
	if s.err != nil {
		return false, s.err
	}
	return s.allowed, nil
}

func TestActiveTokenLimitMiddlewareTouchesLimiter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter := &stubActiveTokenLimiter{allowed: true}
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", uint(42))
		c.Set("authToken", "token-a")
		c.Next()
	})
	router.Use(middleware.ActiveTokenLimitMiddleware(limiter))
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("unexpected status: got=%d want=%d", resp.Code, http.StatusOK)
	}
	if limiter.touchedUserID != 42 {
		t.Fatalf("unexpected touched user id: got=%d want=%d", limiter.touchedUserID, 42)
	}
	if limiter.touchedToken != "token-a" {
		t.Fatalf("unexpected touched token: got=%q want=%q", limiter.touchedToken, "token-a")
	}
}

func TestActiveTokenLimitMiddlewareFailsOpenOnLimiterError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter := &stubActiveTokenLimiter{allowed: true, err: errors.New("redis down")}
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", uint(7))
		c.Set("authToken", "token-b")
		c.Next()
	})
	router.Use(middleware.ActiveTokenLimitMiddleware(limiter))
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("unexpected status: got=%d want=%d", resp.Code, http.StatusOK)
	}
}

func TestActiveTokenLimitMiddlewareRejectsMissingTokenContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", uint(7))
		c.Next()
	})
	router.Use(middleware.ActiveTokenLimitMiddleware(&stubActiveTokenLimiter{}))
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status: got=%d want=%d", resp.Code, http.StatusUnauthorized)
	}
}

func TestActiveTokenLimitMiddlewareRejectsEvictedToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", uint(7))
		c.Set("authToken", "token-c")
		c.Next()
	})
	router.Use(middleware.ActiveTokenLimitMiddleware(&stubActiveTokenLimiter{allowed: false}))
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status: got=%d want=%d", resp.Code, http.StatusUnauthorized)
	}
}
