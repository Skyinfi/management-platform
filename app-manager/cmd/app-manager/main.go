package main

import (
	"log"
	"net/http"

	"github.com/Skyinfi/app-manager/internal/app"
	"github.com/Skyinfi/app-manager/internal/config"
	"github.com/Skyinfi/app-manager/internal/httpapi"
	"github.com/Skyinfi/app-manager/internal/middleware"
)

func main() {
	cfg := config.Load()
	server := app.NewServer(
		app.DefaultStore(),
		app.NewAuthService(cfg),
		httpapi.WithJWTValidator(middleware.NewJWTService(cfg.JWTSecret)),
		httpapi.WithCORS(cfg.EnableCORS, cfg.AllowOrigin),
	)

	log.Printf("app-manager listening on %s env=%s log_level=%s cors=%t", cfg.Addr, cfg.Env, cfg.LogLevel, cfg.EnableCORS)
	if err := http.ListenAndServe(cfg.Addr, server.Routes()); err != nil {
		log.Fatal(err)
	}
}
