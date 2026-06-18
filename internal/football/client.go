package football

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Mode string

const (
	Fake Mode = "Fake"
	Real Mode = "Real"
)

type ClientOptions struct {
	BaseURL         string
	Token           []string
	CompetitionCode string
	Season          string
	FakeDir         string
	ForceFake       bool
}

type Client struct {
	baseURL         string
	tokens          []string
	competitionCode string
	season          string
	fakeDir         string
	httpClient      *http.Client
	token_use       string
	mu              sync.Mutex
	mode            Mode
}

func (c *Client) currentToken() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.token_use != "" {
		return c.token_use
	}

	c.token_use = c.tokens[0]
	return c.token_use
}

func (c *Client) rotateToken() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, token := range c.tokens {
		if token == c.token_use {
			c.token_use = c.tokens[(i+1)%len(c.tokens)]
			return
		}
	}
}

func NewClient(options ClientOptions) *Client {

	baseURL := strings.TrimRight(options.BaseURL, "/")
	if baseURL == "" {
		baseURL = "https://api.football-data.org/v4"
	}
	competitionCode := strings.TrimSpace(options.CompetitionCode)
	if competitionCode == "" {
		competitionCode = "WC"
	}

	season := strings.TrimSpace(options.Season)
	if season == "" {
		season = strconv.Itoa(time.Now().Year())
	}

	tokens := options.Token
	tokenUse := ""
	var mode Mode = Real
	if len(tokens) != 0 && !options.ForceFake {
		tokenUse = tokens[0]
	} else {
		mode = Fake
	}

	return &Client{
		baseURL:         baseURL,
		tokens:          tokens,
		competitionCode: competitionCode,
		season:          season,
		fakeDir:         options.FakeDir,
		httpClient: &http.Client{
			Timeout: 12 * time.Second,
		},
		token_use: tokenUse,
		mode:      mode,
	}
}

type apiError struct {
	StatusCode int
	Body       string
}

func (e *apiError) Error() string {
	if e.Body != "" {
		return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Body)
	}
	return fmt.Sprintf("HTTP %d", e.StatusCode)
}

type Request struct {
	Endpoint Endpoint
	Target   func(*RawState) any
}

func (c *Client) Fetch(ctx context.Context, requests []Request) (RawState, error) {
	var (
		raw RawState
		err error
	)
	switch c.mode {
	case Fake:
		{
			raw, err = c.FetchFake(ctx)
		}
	case Real:
		{
			for _, req := range requests {
				if err := c.callAPI(
					ctx,
					req.Endpoint.URL(c),
					req.Target(&raw),
				); err != nil {
					return RawState{}, err
				}
			}
		}
	}
	if err == nil {
		raw.FetchedAt = time.Now().UTC()
	}
	return raw, err
}

func (c *Client) FetchFake(ctx context.Context) (RawState, error) {
	select {
	case <-ctx.Done():
		return RawState{}, ctx.Err()
	default:
	}

	var raw RawState
	if err := readJSON(filepath.Join(c.fakeDir, "matches.json"), &raw.Matches); err != nil {
		return RawState{}, err
	}
	if err := readJSON(filepath.Join(c.fakeDir, "teams.json"), &raw.Teams); err != nil {
		return RawState{}, err
	}
	raw.Source = "fake-data"
	return raw, nil
}

func readJSON(path string, target interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	return nil
}

func (c *Client) tryRequest(ctx context.Context, requestURL string, target interface{}, token string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	if token != "" {
		req.Header.Set("X-Auth-Token", token)
	}
	req.Header.Set("X-Unfold-Lineups", "true")
	req.Header.Set("X-Unfold-Bookings", "true")
	req.Header.Set("X-Unfold-Subs", "true")
	req.Header.Set("X-Unfold-Goals", "true")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("%s returned HTTP %d: %s", requestURL, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("decode %s: %w", requestURL, err)
	}

	return nil
}

func (c *Client) callAPI(ctx context.Context, endpoint string, target interface{}) error {
	endpoint = strings.TrimPrefix(endpoint, "/")
	requestURL := c.baseURL + "/" + endpoint

	for i := 0; i < len(c.tokens); i++ {
		c.mu.Lock()
		token := c.token_use
		c.mu.Unlock()

		result := c.tryRequest(ctx, requestURL, target, token)
		if result == nil {
			return nil
		}

		if strings.Contains(result.Error(), "HTTP 429") {
			c.rotateToken()
			continue
		}

		return result
	}

	return fmt.Errorf("%s: all tokens rate-limited", requestURL)
}
