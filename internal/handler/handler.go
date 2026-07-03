package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"worldcup-realtime/internal/config"
	"worldcup-realtime/internal/service"
	"worldcup-realtime/internal/sse"
	"worldcup-realtime/internal/store"
)

type WorldCupService interface {
	State() service.ViewState
	Refresh(context.Context) (bool, error)
	Team(int) (service.TeamView, bool)
	Match(int) (service.MatchView, bool)
	Matches() service.MatchesView
}

type Handler struct {
	cfg     config.Config
	service WorldCupService
	hub     *sse.Hub
}

func New(cfg config.Config, service WorldCupService, hub *sse.Hub) *Handler {
	return &Handler{
		cfg:     cfg,
		service: service,
		hub:     hub,
	}
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) State(c *gin.Context) {
	c.JSON(http.StatusOK, h.service.State())
}

func (h *Handler) Events(c *gin.Context) {
	h.hub.ServeHTTP(c.Writer, c.Request, h.service.State())
}

func (h *Handler) Refresh(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.cfg.RefreshTimeout)
	defer cancel()

	changed, err := h.service.Refresh(ctx)
	state := h.service.State()
	if changed {
		h.hub.BroadcastState(state)
	}

	if err == nil {
		c.JSON(http.StatusOK, gin.H{
			"ok":      true,
			"changed": changed,
			"error":   "",
			"state":   state,
		})
		return
	}

	c.JSON(refreshStatusCode(err), gin.H{
		"ok":      false,
		"changed": changed,
		"error":   err.Error(),
		"state":   state,
	})
}

func (h *Handler) Team(c *gin.Context) {
	id, ok := parseID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team id"})
		return
	}
	team, found := h.service.Team(id)
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "team not found"})
		return
	}
	c.JSON(http.StatusOK, team)
}

func (h *Handler) Match(c *gin.Context) {
	id, ok := parseID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match id"})
		return
	}
	match, found := h.service.Match(id)
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "match not found"})
		return
	}
	c.JSON(http.StatusOK, match)
}

func (h *Handler) Matches(c *gin.Context) {
	c.JSON(http.StatusOK, h.service.Matches())
}

func refreshStatusCode(err error) int {
	switch {
	case errors.Is(err, store.ErrRefreshInProgress):
		return http.StatusConflict
	case errors.Is(err, store.ErrRateLimited):
		return http.StatusTooManyRequests
	case errors.Is(err, context.DeadlineExceeded):
		return http.StatusGatewayTimeout
	case errors.Is(err, context.Canceled):
		return http.StatusRequestTimeout
	default:
		return http.StatusInternalServerError
	}
}

func parseID(value string) (int, bool) {
	id, err := strconv.Atoi(value)
	return id, err == nil
}
