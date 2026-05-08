package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Addr         string
	ProcRoot     string
	ScanInterval time.Duration
}

func Load() Config {
	cfg := Config{
		Addr:         ":8081",
		ProcRoot:     "/proc",
		ScanInterval: 60 * time.Second,
	}
	overrideFromEnv(&cfg)
	return cfg
}

func overrideFromEnv(cfg *Config) {
	if v := strings.TrimSpace(os.Getenv("SCANNER_ADDR")); v != "" {
		cfg.Addr = v
	}
	if v := strings.TrimSpace(os.Getenv("PROC_ROOT")); v != "" {
		cfg.ProcRoot = v
	}
	if v := strings.TrimSpace(os.Getenv("SCANNER_INTERVAL")); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.ScanInterval = d
		} else if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.ScanInterval = time.Duration(n) * time.Second
		}
	}
}
