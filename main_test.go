package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckOriginAllowsNativeClientWithoutOrigin(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)

	if !checkOrigin(nil)(req) {
		t.Fatal("native WebSocket clients without Origin should be allowed")
	}
}

func TestCheckOriginRejectsUnlistedBrowserOrigin(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	req.Header.Set("Origin", "https://example.invalid")

	if checkOrigin(map[string]struct{}{"https://sygy.example": {}})(req) {
		t.Fatal("unlisted browser Origin should be rejected")
	}
}

func TestCheckOriginAllowsListedBrowserOrigin(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	req.Header.Set("Origin", "https://sygy.example")

	if !checkOrigin(map[string]struct{}{"https://sygy.example": {}})(req) {
		t.Fatal("listed browser Origin should be allowed")
	}
}
