package router

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"worldcup-realtime/internal/config"
	"worldcup-realtime/internal/handler"
)

func New(cfg config.Config, h *handler.Handler) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery(), cors(cfg.AllowedOrigins), noCache())

	api := engine.Group("/api")
	{
		api.GET("/health", h.Health)
		api.GET("/state", h.State)
		api.GET("/events", h.Events)

		api.GET("/team/:id", h.Team)
		api.GET("/match/:id", h.Match)
		api.GET("/matches", h.Matches)
	}

	engine.GET("/healthz", h.Health)
	engine.NoRoute(staticFallback(cfg.WebDir))

	return engine
}

func cors(allowedOrigins []string) gin.HandlerFunc {
	allowAny := len(allowedOrigins) == 0
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		if origin == "*" {
			allowAny = true
			continue
		}
		allowed[origin] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			if allowAny {
				c.Header("Access-Control-Allow-Origin", "*")
			} else if _, ok := allowed[origin]; ok {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Vary", "Origin")
			}
			c.Header("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")
			c.Header("Access-Control-Max-Age", "86400")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func noCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.Header("Cache-Control", "no-store")
		}
		c.Next()
	}
}

func staticFallback(webDir string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
			c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
			return
		}

		cleanPath := strings.TrimPrefix(path.Clean(c.Request.URL.Path), "/")
		if cleanPath == "." || cleanPath == "" {
			c.File(filepath.Join(webDir, "index.html"))
			return
		}

		candidate := filepath.Join(webDir, filepath.FromSlash(cleanPath))
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			c.File(candidate)
			return
		}
		c.File(filepath.Join(webDir, "index.html"))
	}
}
