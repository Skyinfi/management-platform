package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Addr        string         `yaml:"addr"`
	Env         string         `yaml:"env"`
	LogLevel    string         `yaml:"log_level"`
	EnableCORS  bool           `yaml:"enable_cors"`
	JWTSecret   string         `yaml:"jwt_secret"`
	AllowOrigin string         `yaml:"allow_origin"`
	Docker      DockerConfig   `yaml:"docker"`
	Auth        AuthConfig     `yaml:"auth"`
	Scanner     ScannerConfig  `yaml:"scanner"`
	Services    []ServiceDef   `yaml:"services"`
}

type ScannerConfig struct {
	Addr string `yaml:"addr"`
}

type DockerConfig struct {
	Enabled bool   `yaml:"enabled"`
	Host    string `yaml:"host"`
}

type AuthConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	TokenTTL string `yaml:"token_ttl"`
}

type ServiceDef struct {
	Name     string `yaml:"name"`
	Display  string `yaml:"display"`
	Type     string `yaml:"type"`
	Unit     string `yaml:"unit"`
	LogPath  string `yaml:"log_path"`
	Endpoint string `yaml:"endpoint"`
	Owner    string `yaml:"owner"`
}

func Load() Config {
	cfg := Config{
		Addr:        ":8080",
		Env:         "development",
		LogLevel:    "info",
		EnableCORS:  true,
		JWTSecret:   "dev-secret",
		AllowOrigin: "*",
		Docker: DockerConfig{
			Enabled: true,
			Host:    "unix:///var/run/docker.sock",
		},
		Auth: AuthConfig{
			Username: "admin",
			Password: "admin123",
			TokenTTL: "24h",
		},
		Scanner: ScannerConfig{
			Addr: "http://localhost:8081",
		},
		Services: []ServiceDef{
			{Name: "order-service", Display: "订单服务", Type: "process", Unit: "order-service", Endpoint: "10.0.1.12:9001", Owner: "交易团队"},
			{Name: "report-worker", Display: "报表Worker", Type: "process", Unit: "report-worker", LogPath: "/var/log/report-worker/app.log", Endpoint: "batch-only", Owner: "数据团队"},
		},
	}

	configPath := getString("APP_MANAGER_CONFIG", "")
	if configPath == "" {
		exe, err := os.Executable()
		if err == nil {
			candidate := filepath.Join(filepath.Dir(exe), "config.yaml")
			if _, err := os.Stat(candidate); err == nil {
				configPath = candidate
			}
		}
	}

	if configPath == "" {
		candidate := "config.yaml"
		if _, err := os.Stat(candidate); err == nil {
			configPath = candidate
		}
	}

	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to read config %s: %v\n", configPath, err)
		} else if err := yaml.Unmarshal(data, &cfg); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to parse config %s: %v\n", configPath, err)
		}
	}

	overrideFromEnv(&cfg)
	return cfg
}

func overrideFromEnv(cfg *Config) {
	if v := getString("APP_MANAGER_ADDR", ""); v != "" {
		cfg.Addr = v
	}
	if v := getString("APP_MANAGER_ENV", ""); v != "" {
		cfg.Env = v
	}
	if v := getString("APP_MANAGER_LOG_LEVEL", ""); v != "" {
		cfg.LogLevel = v
	}
	if v, err := strconv.ParseBool(getString("APP_MANAGER_ENABLE_CORS", "")); err == nil {
		cfg.EnableCORS = v
	}
	if v := getString("APP_MANAGER_JWT_SECRET", ""); v != "" {
		cfg.JWTSecret = v
	}
	if v := getString("APP_MANAGER_ALLOW_ORIGIN", ""); v != "" {
		cfg.AllowOrigin = v
	}
	if v := getString("APP_MANAGER_DOCKER_HOST", ""); v != "" {
		cfg.Docker.Host = v
	}
	if v := getString("APP_MANAGER_AUTH_USERNAME", ""); v != "" {
		cfg.Auth.Username = v
	}
	if v := getString("APP_MANAGER_AUTH_PASSWORD", ""); v != "" {
		cfg.Auth.Password = v
	}
	if v := getString("APP_MANAGER_SCANNER_ADDR", ""); v != "" {
		cfg.Scanner.Addr = v
	}
}

func getString(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func getBool(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}
