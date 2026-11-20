// Package server 封装 HTTP 服务器的生命周期与中间件。
package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/zsdili/go-starter/internal/config"
)

// Server 包装 http.Server，提供优雅关闭。
type Server struct {
	httpServer *http.Server
	log        *slog.Logger
	cfg        config.Config
}

// New 创建一个 Server。
func New(cfg config.Config, handler http.Handler, log *slog.Logger) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:              cfg.Addr,
			Handler:           chain(handler, requestIDMiddleware, loggingMiddleware(log), recoverMiddleware(log)),
			ReadHeaderTimeout: cfg.ReadTimeout,
			ReadTimeout:       cfg.ReadTimeout,
			WriteTimeout:      cfg.WriteTimeout,
			IdleTimeout:       60 * time.Second,
		},
		log: log,
		cfg: cfg,
	}
}

// Run 启动服务器，阻塞直到关闭。
func (s *Server) Run() error {
	s.log.Info("HTTP 服务器启动", "addr", s.cfg.Addr)
	err := s.httpServer.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

// Shutdown 优雅关闭服务器。
func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
	defer cancel()
	s.log.Info("正在优雅关闭服务器")
	return s.httpServer.Shutdown(ctx)
}
