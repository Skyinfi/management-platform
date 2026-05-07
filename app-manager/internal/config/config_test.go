package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	t.Setenv("APP_MANAGER_ADDR", "")
	t.Setenv("APP_MANAGER_ENV", "")
	t.Setenv("APP_MANAGER_LOG_LEVEL", "")
	t.Setenv("APP_MANAGER_ENABLE_CORS", "")
	t.Setenv("APP_MANAGER_JWT_SECRET", "")
	t.Setenv("APP_MANAGER_ALLOW_ORIGIN", "")

	cfg := Load()

	if cfg.Addr != ":8080" {
		t.Fatalf("expected default addr, got %q", cfg.Addr)
	}
	if cfg.Env != "development" {
		t.Fatalf("expected default env, got %q", cfg.Env)
	}
	if cfg.LogLevel != "info" {
		t.Fatalf("expected default log level, got %q", cfg.LogLevel)
	}
	if !cfg.EnableCORS {
		t.Fatalf("expected default cors enabled")
	}
	if cfg.JWTSecret != "dev-secret" {
		t.Fatalf("expected default jwt secret, got %q", cfg.JWTSecret)
	}
	if cfg.AllowOrigin != "*" {
		t.Fatalf("expected default allow origin, got %q", cfg.AllowOrigin)
	}
	if cfg.Auth.Username != "admin" {
		t.Fatalf("expected default username, got %q", cfg.Auth.Username)
	}
	if cfg.Auth.Password != "admin123" {
		t.Fatalf("expected default password, got %q", cfg.Auth.Password)
	}
	if !cfg.Docker.Enabled {
		t.Fatalf("expected docker enabled by default")
	}
}
