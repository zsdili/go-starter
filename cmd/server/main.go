// Command server 是 go-starter 服务的入口。
package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/zsdili/go-starter/internal/config"
	"github.com/zsdili/go-starter/internal/handler"
	"github.com/zsdili/go-starter/internal/server"
)

var (
	version = "dev" // 由构建脚本通过 -ldflags 注入
	commit  = "none"
	buildAt = "unknown"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "配置错误:", err)
		os.Exit(1)
	}
	cfg.Version = version
	log := newLogger(cfg)

	h := handler.New(cfg.AppName, version)
	srv := server.New(cfg, h.Routes(), log)

	errCh := make(chan error, 1)
	go func() { errCh <- srv.Run() }()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		log.Info("收到信号，准备退出", "signal", sig.String())
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("服务器异常退出", "error", err)
			os.Exit(1)
		}
	}

	if err := srv.Shutdown(); err != nil {
		log.Error("优雅关闭失败", "error", err)
		os.Exit(1)
	}
	log.Info("服务已停止", "commit", commit, "built", buildAt)
}

func newLogger(cfg config.Config) *slog.Logger {
	var level slog.Level
	switch cfg.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	opts := &slog.HandlerOptions{Level: level}
	var h slog.Handler
	if cfg.LogFormat == "text" {
		h = slog.NewTextHandler(os.Stdout, opts)
	} else {
		h = slog.NewJSONHandler(os.Stdout, opts)
	}
	return slog.New(h).With("app", cfg.AppName, "version", version)
}
