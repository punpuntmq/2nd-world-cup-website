package server

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
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
	mux.HandleFunc("/api/team/", s.handleTeam)
	mux.HandleFunc("/api/match/", s.handleMatch)
	mux.HandleFunc("/api/matchs/", s.handleMatch)
	mux.HandleFunc("/api/matches", s.handleMatches)
	mux.HandleFunc("/api/matches/", s.handleMatch)
	mux.HandleFunc("/team/", s.handleApp)
	mux.HandleFunc("/match/", s.handleApp)
	mux.HandleFunc("/server/main_view.html", s.handleApp)
	mux.HandleFunc("/server/team/", s.handleApp)
	mux.HandleFunc("/server/match/", s.handleApp)
	mux.HandleFunc("/server/matchs/", s.handleApp)
	mux.HandleFunc("/server/matches/", s.handleApp)
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

func (s *Server) handleTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id, ok := parseID(r.URL.Path, "/api/team/")
	if !ok {
		http.NotFound(w, r)
		return
	}
	for _, team := range s.store.Snapshot().Teams {
		if team.ID == id {
			writeJSON(w, http.StatusOK, team)
			return
		}
	}
	http.NotFound(w, r)
}

func (s *Server) handleMatches(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if r.URL.Path != "/api/matches" {
		s.handleMatch(w, r)
		return
	}
	view := s.store.Snapshot()
	all := make([]store.MatchView, 0, len(view.LiveMatches)+len(view.UpcomingMatches)+len(view.FinishedMatches)+len(view.SpecialMatches))
	all = append(all, view.LiveMatches...)
	all = append(all, view.UpcomingMatches...)
	all = append(all, view.FinishedMatches...)
	all = append(all, view.SpecialMatches...)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"live":     view.LiveMatches,
		"upcoming": view.UpcomingMatches,
		"finished": view.FinishedMatches,
		"special":  view.SpecialMatches,
		"all":      all,
	})
}

func (s *Server) handleMatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id, ok := parseID(r.URL.Path, "/api/match/")
	if !ok {
		id, ok = parseID(r.URL.Path, "/api/matchs/")
	}
	if !ok {
		id, ok = parseID(r.URL.Path, "/api/matches/")
	}
	if !ok {
		http.NotFound(w, r)
		return
	}
	if match, ok := findMatch(s.store.Snapshot(), id); ok {
		writeJSON(w, http.StatusOK, match)
		return
	}
	http.NotFound(w, r)
}

func (s *Server) handleApp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, filepath.Join(s.cfg.WebDir, "index.html"))
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

func parseID(requestPath, prefix string) (int, bool) {
	value := strings.Trim(strings.TrimPrefix(requestPath, prefix), "/")
	if value == requestPath || value == "" {
		return 0, false
	}
	id, err := strconv.Atoi(value)
	return id, err == nil
}

func findMatch(view store.ViewState, id int) (store.MatchView, bool) {
	if view.CurrentMatch != nil && view.CurrentMatch.ID == id {
		return *view.CurrentMatch, true
	}
	if view.NextMatch != nil && view.NextMatch.ID == id {
		return *view.NextMatch, true
	}
	for _, matches := range [][]store.MatchView{
		view.LiveMatches,
		view.UpcomingMatches,
		view.FinishedMatches,
		view.SpecialMatches,
	} {
		for _, match := range matches {
			if match.ID == id {
				return match, true
			}
		}
	}
	return store.MatchView{}, false
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
