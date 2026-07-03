package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"worldcup-realtime/internal/app"
	"worldcup-realtime/internal/config"
)

func main() {
	rootDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	cfg := config.Load(rootDir)
	application := app.New(cfg)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	initialCtx, cancel := context.WithTimeout(ctx, cfg.RefreshTimeout)
	if err := application.InitialRefresh(initialCtx); err != nil {
		log.Printf("initial refresh used fallback or failed: %v", err)
	}
	cancel()

	application.StartWorkers(ctx)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: application.Handler(),
	}

	go func() {
		<-ctx.Done()
		log.Println("received termination signal, shutting down server")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("server shutdown error: %v", err)
		}
	}()

	log.Printf("World Cup realtime dashboard listening on http://localhost:%s", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
