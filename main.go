package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"worldcup-realtime/internal/config"
	"worldcup-realtime/internal/football"
	"worldcup-realtime/internal/server"
	"worldcup-realtime/internal/store"
)

func main() {
	rootDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	cfg := config.Load(rootDir)
	client := football.NewClient(football.ClientOptions{
		BaseURL:         cfg.BaseURL,
		Token:           cfg.Token,
		CompetitionCode: cfg.CompetitionCode,
		Season:          cfg.Season,
		FakeDir:         cfg.FakeDir,
		ForceFake:       cfg.ForceFake,
	})

	memory := store.NewMemoryStore(client, store.StoreOptions{
		RefreshInterval: cfg.RefreshInterval,
		Quota:           store.NewFixedWindowQuota(cfg.QuotaLimit, cfg.QuotaWindow),
	})
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := memory.InitialRefresh(ctx); err != nil {
		log.Printf("initial refresh used fallback or failed: %v", err)
	}
	go memory.Run(ctx)

	app := server.New(cfg, memory)
	addr := ":" + cfg.Port
	log.Printf("World Cup realtime dashboard listening on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, app.Router()); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
