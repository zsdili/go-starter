package config

import (
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("APP_ADDR", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("LOG_FORMAT", "")
	c, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if c.Addr != ":8080" {
		t.Errorf("Addr = %q, want :8080", c.Addr)
	}
	if c.LogLevel != "info" {
		t.Errorf("LogLevel = %q, want info", c.LogLevel)
	}
	if c.ReadTimeout != 5*time.Second {
		t.Errorf("ReadTimeout = %v, want 5s", c.ReadTimeout)
	}
	if !c.EnableMetrics {
		t.Error("EnableMetrics default should be true")
	}
}

func TestLoadCustom(t *testing.T) {
	t.Setenv("APP_ADDR", ":9090")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("LOG_FORMAT", "text")
	t.Setenv("APP_READ_TIMEOUT", "2s")
	t.Setenv("ENABLE_METRICS", "false")
	c, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if c.Addr != ":9090" {
		t.Errorf("Addr = %q, want :9090", c.Addr)
	}
	if c.ReadTimeout != 2*time.Second {
		t.Errorf("ReadTimeout = %v, want 2s", c.ReadTimeout)
	}
	if c.EnableMetrics {
		t.Error("EnableMetrics should be false")
	}
}

func TestLoadInvalidLogLevel(t *testing.T) {
	t.Setenv("LOG_LEVEL", "verbose")
	if _, err := Load(); err == nil {
		t.Error("expected error for invalid LOG_LEVEL")
	}
}
