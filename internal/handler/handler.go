// Package handler 提供 HTTP 路由与处理器。
package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sync/atomic"
	"time"
)

// Handler 持有应用状态与依赖。
type Handler struct {
	startedAt time.Time
	requests  atomic.Int64
	version   string
	appName   string
}

// New 创建一个 Handler。
func New(appName, version string) *Handler {
	return &Handler{
		startedAt: time.Now(),
		version:   version,
		appName:   appName,
	}
}

// Routes 返回配置好所有路由的 http.Handler。
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", h.index)
	mux.HandleFunc("GET /healthz", h.healthz)
	mux.HandleFunc("GET /readyz", h.readyz)
	mux.HandleFunc("GET /api/info", h.info)
	mux.HandleFunc("GET /metrics", h.metrics)
	return mux
}

func (h *Handler) index(w http.ResponseWriter, r *http.Request) {
	h.requests.Add(1)
	writeJSON(w, http.StatusOK, map[string]any{
		"message": "Hello from " + h.appName,
		"version": h.version,
		"path":    r.URL.Path,
	})
}

// healthz 存活探针：进程活着就返回 200。
func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "uptime": time.Since(h.startedAt).String()})
}

// readyz 就绪探针：可在此加入依赖检查（DB/缓存等）。
func (h *Handler) readyz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ready"})
}

func (h *Handler) info(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	writeJSON(w, http.StatusOK, map[string]any{
		"app":        h.appName,
		"version":    h.version,
		"go":         runtime.Version(),
		"num_cpu":    runtime.NumCPU(),
		"goroutines": runtime.NumGoroutine(),
		"requests":   h.requests.Load(),
		"uptime":     time.Since(h.startedAt).String(),
		"alloc_mb":   fmt.Sprintf("%.1f", float64(m.Alloc)/1024/1024),
	})
}

// metrics 输出 Prometheus 文本格式的简易指标。
func (h *Handler) metrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	fmt.Fprintln(w, "# HELP app_requests_total Total HTTP requests handled.")
	fmt.Fprintln(w, "# TYPE app_requests_total counter")
	fmt.Fprintf(w, "app_requests_total %d\n", h.requests.Load())
	fmt.Fprintln(w, "# HELP app_uptime_seconds Seconds since process start.")
	fmt.Fprintln(w, "# TYPE app_uptime_seconds gauge")
	fmt.Fprintf(w, "app_uptime_seconds %.0f\n", time.Since(h.startedAt).Seconds())
	fmt.Fprintln(w, "# HELP app_goroutines Current number of goroutines.")
	fmt.Fprintln(w, "# TYPE app_goroutines gauge")
	fmt.Fprintf(w, "app_goroutines %d\n", runtime.NumGoroutine())
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
