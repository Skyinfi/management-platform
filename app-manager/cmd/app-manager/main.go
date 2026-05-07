package main

import (
	"log"
	"net/http"

	"github.com/Skyinfi/management-platform/app-manager/internal/app"
	"github.com/Skyinfi/management-platform/app-manager/internal/config"
	"github.com/Skyinfi/management-platform/app-manager/internal/httpapi"
	"github.com/Skyinfi/management-platform/app-manager/internal/middleware"
)

func main() {
	cfg := config.Load()
	deps := app.Bootstrap(cfg)

	server := app.NewServer(deps,
		httpapi.WithJWTValidator(middleware.NewJWTService(cfg.JWTSecret)),
		httpapi.WithCORS(cfg.EnableCORS, cfg.AllowOrigin),
	)

	log.Printf("app-manager listening on %s env=%s log_level=%s cors=%t docker=%t",
		cfg.Addr, cfg.Env, cfg.LogLevel, cfg.EnableCORS, deps.Docker.Enabled())
	if err := http.ListenAndServe(cfg.Addr, server.Routes()); err != nil {
		log.Fatal(err)
	}
}
