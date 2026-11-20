// Package config 负责从环境变量加载配置，并提供合理默认值。
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config 是应用配置的集中定义。
type Config struct {
	Addr            string        // 监听地址，如 :8080
	ReadTimeout     time.Duration // 读超时
	WriteTimeout    time.Duration // 写超时
	ShutdownTimeout time.Duration // 优雅关闭等待超时
	LogLevel        string        // debug/info/warn/error
	LogFormat       string        // json/text
	AppName         string
	Version         string
	EnableMetrics   bool
}

// Load 从环境变量读取配置并校验。
func Load() (Config, error) {
	c := Config{
		Addr:            getEnv("APP_ADDR", ":8080"),
		ReadTimeout:     getDuration("APP_READ_TIMEOUT", 5*time.Second),
		WriteTimeout:    getDuration("APP_WRITE_TIMEOUT", 10*time.Second),
		ShutdownTimeout: getDuration("APP_SHUTDOWN_TIMEOUT", 15*time.Second),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		LogFormat:       getEnv("LOG_FORMAT", "json"),
		AppName:         getEnv("APP_NAME", "go-starter"),
		Version:         getEnv("APP_VERSION", "dev"),
		EnableMetrics:   getBool("ENABLE_METRICS", true),
	}
	if err := c.validate(); err != nil {
		return c, err
	}
	return c, nil
}

func (c Config) validate() error {
	if c.Addr == "" {
		return fmt.Errorf("APP_ADDR 不能为空")
	}
	switch c.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("非法 LOG_LEVEL: %q", c.LogLevel)
	}
	if c.LogFormat != "json" && c.LogFormat != "text" {
		return fmt.Errorf("非法 LOG_FORMAT: %q", c.LogFormat)
	}
	return nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getDuration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}

func getBool(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}
