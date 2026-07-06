package app

import (
	"context"
	"net/http"

	"worldcup-realtime/internal/config"
	"worldcup-realtime/internal/football"
	"worldcup-realtime/internal/handler"
	"worldcup-realtime/internal/router"
	"worldcup-realtime/internal/scheduler"
	"worldcup-realtime/internal/service"
	"worldcup-realtime/internal/sse"
	"worldcup-realtime/internal/store"
)

type App struct {
	cfg       config.Config
	service   *service.WorldCupService
	hub       *sse.Hub
	scheduler *scheduler.Scheduler
	router    http.Handler
}

func New(cfg config.Config) *App {
	client := football.NewClient(football.ClientOptions{
		BaseURL:         cfg.BaseURL,
		Token:           cfg.Token,
		CompetitionCode: cfg.CompetitionCode,
		Season:          cfg.Season,
	})
	memoryStore := store.NewMemoryStore()
	quota := store.NewFixedWindowQuota(cfg.QuotaLimit, cfg.QuotaWindow)
	worldCupService := service.NewWorldCupService(client, memoryStore, service.Options{
		LiveRefreshInterval: cfg.LiveRefreshInterval,
		IdleRefreshInterval: cfg.IdleRefreshInterval,
		Quota:               quota,
	})
	hub := sse.NewHub()
	refreshScheduler := scheduler.New(worldCupService, hub, cfg.LiveRefreshInterval, cfg.IdleRefreshInterval, cfg.RefreshTimeout, cfg.RefreshTimeoutBuffer)
	httpHandler := handler.New(cfg, worldCupService, hub)

	return &App{
		cfg:       cfg,
		service:   worldCupService,
		hub:       hub,
		scheduler: refreshScheduler,
		router:    router.New(cfg, httpHandler),
	}
}

func (a *App) StartWorkers(ctx context.Context) {
	go a.hub.Run(ctx)
	go a.scheduler.Run(ctx)
}

func (a *App) InitialRefresh(ctx context.Context) error {
	changed, err := a.service.InitialRefresh(ctx)
	if changed {
		a.hub.BroadcastState(a.service.State())
	}
	return err
}

func (a *App) Handler() http.Handler {
	return a.router
}

func (a *App) Config() config.Config {
	return a.cfg
}
