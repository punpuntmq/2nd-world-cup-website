package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// parse cho api token
func parseTokens(s string) []string {
	parts := strings.Split(s, "|")

	tokens := make([]string, 0, len(parts))

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			tokens = append(tokens, p)
		}
	}

	return tokens
}

type Config struct {
	RootDir string
	WebDir  string

	BaseURL              string
	Token                []string
	CompetitionCode      string
	Season               string
	Port                 string
	LiveRefreshInterval  time.Duration
	IdleRefreshInterval  time.Duration
	RefreshTimeout       time.Duration
	RefreshTimeoutBuffer time.Duration
	QuotaLimit           int
	QuotaWindow          time.Duration

	AllowedOrigins []string
}

func Load(rootDir string) Config {
	values := readTokenEnv(filepath.Join(rootDir, "token.env"))
	tokens := parseTokens(getValue(values, "FOOTBALL_DATA_TOKEN", ""))

	cfg := Config{
		RootDir:              rootDir,
		WebDir:               filepath.Join(rootDir, "web"),
		BaseURL:              getValue(values, "FOOTBALL_API_BASE_URL", ""),
		Token:                tokens,
		CompetitionCode:      getValue(values, "FOOTBALL_DATA_COMPETITION", ""),
		Season:               getValue(values, "FOOTBALL_DATA_SEASON", ""),
		Port:                 getValue(values, "PORT", "8080"),
		LiveRefreshInterval:  time.Duration(getInt(values, "LIVE_REFRESH_SECONDS", 10)) * time.Second,
		IdleRefreshInterval:  time.Duration(getInt(values, "IDLE_REFRESH_SECONDS", 120)) * time.Second,
		RefreshTimeout:       time.Duration(getInt(values, "REFRESH_TIMEOUT_SECONDS", 30)) * time.Second,
		RefreshTimeoutBuffer: time.Duration(getInt(values, "REFRESH_TIMEOUT_BUFFER_MS", 500)) * time.Millisecond,
		QuotaLimit:           effectiveQuotaLimit(tokens),
		QuotaWindow:          time.Duration(getInt(values, "FOOTBALL_API_QUOTA_WINDOW_SECONDS", 60)) * time.Second,
		AllowedOrigins:       parseList(getValue(values, "CORS_ALLOWED_ORIGINS", "*")),
	}
	return cfg
}

func parseList(s string) []string {
	parts := strings.Split(s, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value != "" {
			values = append(values, value)
		}
	}
	return values
}

func effectiveQuotaLimit(tokens []string) int {
	if len(tokens) == 0 {
		return 1000
	}
	return len(tokens) * 10
}

func readTokenEnv(path string) map[string]string {
	values := make(map[string]string)
	data, err := os.ReadFile(path)
	if err != nil {
		return values
	}

	for _, rawLine := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if key != "" {
			values[key] = value
		}
	}
	return values
}

func getValue(values map[string]string, key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	if value := strings.TrimSpace(values[key]); value != "" {
		return value
	}
	return fallback
}

func getInt(values map[string]string, key string, fallback int) int {
	value := getValue(values, key, "")
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
