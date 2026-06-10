package server

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"worldcup-realtime/internal/config"
	"worldcup-realtime/internal/store"
)

type Store interface {
	Snapshot() store.ViewState
	Refresh(context.Context) error
}

type Server struct {
	cfg   config.Config
	store Store
}

func New(cfg config.Config, store Store) *Server {
	return &Server{cfg: cfg, store: store}
}

func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/api/state", s.handleState)
	mux.HandleFunc("/api/refresh", s.handleRefresh)
	mux.HandleFunc("/", s.handleStatic)
	return noCache(mux)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, http.StatusOK, s.store.Snapshot())
}

func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	err := s.store.Refresh(ctx)
	status := http.StatusOK
	if err != nil {
		status = http.StatusAccepted
	}
	writeJSON(w, status, map[string]interface{}{
		"ok":    err == nil,
		"error": errorString(err),
		"state": s.store.Snapshot(),
	})
}

func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cleanPath := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	if cleanPath == "." || cleanPath == "" {
		http.ServeFile(w, r, filepath.Join(s.cfg.WebDir, "index.html"))
		return
	}

	candidate := filepath.Join(s.cfg.WebDir, cleanPath)
	if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
		http.ServeFile(w, r, candidate)
		return
	}
	http.ServeFile(w, r, filepath.Join(s.cfg.WebDir, "index.html"))
}

func writeJSON(w http.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func noCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Cache-Control", "no-store")
		}
		next.ServeHTTP(w, r)
	})
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
