package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	RootDir         string
	WebDir          string
	FakeDir         string
	BaseURL         string
	Token           string
	CompetitionCode string
	Season          string
	Port            string
	RefreshInterval time.Duration
	ForceFake       bool
}

func Load(rootDir string) Config {
	values := readTokenEnv(filepath.Join(rootDir, "token.env"))

	cfg := Config{
		RootDir:         rootDir,
		WebDir:          filepath.Join(rootDir, "web"),
		FakeDir:         filepath.Join(rootDir, "fake-data"),
		BaseURL:         getValue(values, "FOOTBALL_DATA_BASE_URL", "https://api.football-data.org/v4"),
		Token:           getValue(values, "FOOTBALL_DATA_TOKEN", ""),
		CompetitionCode: getValue(values, "FOOTBALL_DATA_COMPETITION", "WC"),
		Season:          getValue(values, "FOOTBALL_DATA_SEASON", "2026"),
		Port:            getValue(values, "PORT", "8080"),
		RefreshInterval: time.Duration(getInt(values, "REFRESH_SECONDS", 45)) * time.Second,
		ForceFake:       getBool(values, "USE_FAKE_DATA", false),
	}

	if cfg.Token == "" {
		cfg.Token = getValue(values, "API_TOKEN", "")
	}
	return cfg
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

func getBool(values map[string]string, key string, fallback bool) bool {
	value := strings.ToLower(getValue(values, key, ""))
	switch value {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}
