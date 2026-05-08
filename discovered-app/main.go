package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Skyinfi/management-platform/discovered-app/internal/config"
	"github.com/Skyinfi/management-platform/discovered-app/internal/httpapi"
	"github.com/Skyinfi/management-platform/discovered-app/internal/scanner"
)

func main() {
	cfg := config.Load()

	log.Printf("discovered-app starting: addr=%s proc_root=%s interval=%s",
		cfg.Addr, cfg.ProcRoot, cfg.ScanInterval)

	agent := scanner.NewAgent(cfg.ProcRoot)

	handler := httpapi.NewHandler(agent)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go agent.StartScheduler(ctx, cfg.ScanInterval)

	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: corsMiddleware(handler),
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("shutting down...")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	fmt.Printf("discovered-app listening on %s\n", cfg.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
