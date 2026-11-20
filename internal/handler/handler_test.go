package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRoutes(t *testing.T) {
	h := New("test-app", "v1.2.3")
	router := h.Routes()

	tests := []struct {
		path       string
		wantStatus int
		contains   string
	}{
		{"/", http.StatusOK, "test-app"},
		{"/healthz", http.StatusOK, "ok"},
		{"/readyz", http.StatusOK, "ready"},
		{"/api/info", http.StatusOK, "v1.2.3"},
		{"/metrics", http.StatusOK, "app_requests_total"},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			if !strings.Contains(rec.Body.String(), tc.contains) {
				t.Errorf("body %q does not contain %q", rec.Body.String(), tc.contains)
			}
		})
	}
}

func TestIndexJSON(t *testing.T) {
	h := New("app", "v1")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, req)

	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if got["version"] != "v1" {
		t.Errorf("version = %v, want v1", got["version"])
	}
}

func BenchmarkIndex(b *testing.B) {
	h := New("app", "v1")
	router := h.Routes()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
	}
}
