package football

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ClientOptions struct {
	BaseURL         string
	Token           []string
	CompetitionCode string
	Season          string
}

type Client struct {
	baseURL         string
	tokens          []string
	competitionCode string
	season          string
	httpClient      *http.Client
	activeToken     string
	mu              sync.Mutex
}

func (c *Client) rotateToken() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, token := range c.tokens {
		if token == c.activeToken {
			c.activeToken = c.tokens[(i+1)%len(c.tokens)]
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
	if len(tokens) != 0 {
		tokenUse = tokens[0]
	}

	return &Client{
		baseURL:         baseURL,
		tokens:          tokens,
		competitionCode: competitionCode,
		season:          season,
		httpClient: &http.Client{
			Timeout: 12 * time.Second,
		},
		activeToken: tokenUse,
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
	var raw RawState
	for _, req := range requests {
		if err := c.callAPI(
			ctx,
			req.Endpoint.URL(c),
			req.Target(&raw),
		); err != nil {
			return RawState{}, err
		}
	}
	raw.FetchedAt = time.Now().UTC()
	if raw.Source == "" {
		raw.Source = "football-data"
	}
	return raw, nil
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

	if len(c.tokens) == 0 {
		return c.tryRequest(ctx, requestURL, target, "")
	}

	for i := 0; i < len(c.tokens); i++ {
		c.mu.Lock()
		token := c.activeToken
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
