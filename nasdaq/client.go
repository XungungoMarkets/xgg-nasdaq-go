package nasdaq

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	defaultBaseAPIURL = "https://api.nasdaq.com/api"
	defaultBaseWWWURL = "https://www.nasdaq.com/api"

	// Rate limiting configuration
	defaultRequestsPerSecond = 2
	defaultMaxRetries        = 3
	defaultTimeout           = 30 * time.Second
)

// Client represents the NASDAQ API client
type Client struct {
	httpClient       *http.Client
	rateLimiter      *rateLimiter
	baseAPIURL       string
	baseWWWURL       string
	watchlistType    string
	userAgents       []string
	currentUserAgent int
	uaMutex          sync.Mutex
	maxRetries       int
	retryDelay       time.Duration
}

// NewClient creates a new NASDAQ API client with rate limiting
func NewClient(opts ...ClientOption) *Client {
	// Create HTTP client with automatic decompression
	transport := &http.Transport{
		DisableCompression: false,
	}

	client := &Client{
		httpClient: &http.Client{
			Timeout:   defaultTimeout,
			Transport: transport,
		},
		rateLimiter: newRateLimiter(defaultRequestsPerSecond),
		baseAPIURL:  defaultBaseAPIURL,
		baseWWWURL:  defaultBaseWWWURL,
		userAgents:  defaultUserAgents,
		maxRetries:  defaultMaxRetries,
		retryDelay:  time.Second * 2,
	}

	// Apply custom options
	for _, opt := range opts {
		opt(client)
	}

	return client
}

// ClientOption represents a function that configures the Client
type ClientOption func(*Client)

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = hc
	}
}

// WithRateLimit sets the maximum requests per second
func WithRateLimit(rps int) ClientOption {
	return func(c *Client) {
		if c.rateLimiter != nil {
			c.rateLimiter.Stop()
		}
		c.rateLimiter = newRateLimiter(rps)
	}
}

// WithMaxRetries sets the maximum number of retries
func WithMaxRetries(max int) ClientOption {
	return func(c *Client) {
		c.maxRetries = max
	}
}

// WithRetryDelay sets the initial retry delay
func WithRetryDelay(delay time.Duration) ClientOption {
	return func(c *Client) {
		c.retryDelay = delay
	}
}

// WithUserAgents sets custom user agents
func WithUserAgents(userAgents []string) ClientOption {
	return func(c *Client) {
		if len(userAgents) == 0 {
			return
		}
		c.userAgents = userAgents
		c.currentUserAgent = 0
	}
}

// WithWatchlistType sets the optional "type" query parameter for /quote/watchlist requests (for example: "Rv").
func WithWatchlistType(watchlistType string) ClientOption {
	return func(c *Client) {
		c.watchlistType = strings.TrimSpace(watchlistType)
	}
}

// WithBaseURLs sets custom API base URLs, mainly useful for testing.
func WithBaseURLs(apiURL, wwwURL string) ClientOption {
	return func(c *Client) {
		if apiURL != "" {
			c.baseAPIURL = apiURL
		}
		if wwwURL != "" {
			c.baseWWWURL = wwwURL
		}
	}
}

// Close releases resources associated with the client.
// It is safe to call more than once.
func (c *Client) Close() {
	if c.rateLimiter != nil {
		c.rateLimiter.Stop()
		c.rateLimiter = nil
	}
}

// getUserAgent returns a rotated user agent
func (c *Client) getUserAgent() string {
	c.uaMutex.Lock()
	defer c.uaMutex.Unlock()

	ua := c.userAgents[c.currentUserAgent]
	c.currentUserAgent = (c.currentUserAgent + 1) % len(c.userAgents)
	return ua
}

// doRequest performs an HTTP request with rate limiting and retries
func (c *Client) doRequest(ctx context.Context, method, url string, body []byte, headers map[string]string) ([]byte, error) {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff with jitter
			delay := c.retryDelay * time.Duration(1<<uint(attempt-1))
			jitter := time.Duration(rand.Intn(1000)) * time.Millisecond
			select {
			case <-time.After(delay + jitter):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		// Rate limiting
		if err := c.rateLimiter.Wait(ctx); err != nil {
			return nil, err
		}

		var reqBody io.Reader
		if body != nil {
			reqBody = bytes.NewReader(body)
		}

		req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
		if err != nil {
			lastErr = err
			continue
		}

		// Set headers to mimic browser behavior
		req.Header.Set("User-Agent", c.getUserAgent())
		req.Header.Set("Accept", "application/json, text/plain, */*")
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")
		req.Header.Set("Accept-Encoding", "gzip")
		req.Header.Set("Connection", "keep-alive")
		req.Header.Set("Referer", "https://www.nasdaq.com/")
		req.Header.Set("Origin", "https://www.nasdaq.com")

		// Add custom headers
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		// Read response body and handle gzip decompression
		var respBody []byte
		var bodyReader io.Reader = resp.Body

		// Check if response is gzip compressed
		if resp.Header.Get("Content-Encoding") == "gzip" {
			gzipReader, err := gzip.NewReader(resp.Body)
			if err != nil {
				lastErr = err
				resp.Body.Close()
				continue
			}
			bodyReader = gzipReader
		}

		respBody, err = io.ReadAll(bodyReader)
		if closer, ok := bodyReader.(io.Closer); ok {
			closer.Close()
		} else {
			resp.Body.Close()
		}
		if err != nil {
			lastErr = err
			continue
		}

		// Check for rate limiting (HTTP 429)
		if resp.StatusCode == http.StatusTooManyRequests {
			lastErr = fmt.Errorf("rate limited by server (attempt %d)", attempt+1)
			continue
		}

		// Check for other errors
		if resp.StatusCode >= 400 {
			lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
			continue
		}

		return respBody, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// makeAPIRequest is a helper for making API requests
func (c *Client) makeAPIRequest(ctx context.Context, endpoint string, params url.Values) ([]byte, error) {
	// Build URL with query parameters
	url := fmt.Sprintf("%s%s", c.baseAPIURL, endpoint)
	if len(params) > 0 {
		url += "?" + params.Encode()
	}

	return c.doRequest(ctx, http.MethodGet, url, nil, nil)
}

// makeWWWAPIRequest is a helper for making www.nasdaq.com API requests
func (c *Client) makeWWWAPIRequest(ctx context.Context, endpoint string, params url.Values) ([]byte, error) {
	// Build URL with query parameters
	url := fmt.Sprintf("%s%s", c.baseWWWURL, endpoint)
	if len(params) > 0 {
		url += "?" + params.Encode()
	}

	return c.doRequest(ctx, http.MethodGet, url, nil, nil)
}

// makeAISearchRequest is a helper for making AI search API requests
func (c *Client) makeAISearchRequest(ctx context.Context, endpoint string, params url.Values) ([]byte, error) {
	// AI search uses www.nasdaq.com base URL (use baseWWWURL without /api suffix)
	baseURL := strings.TrimSuffix(c.baseWWWURL, "/api")
	// Build URL with query parameters
	url := fmt.Sprintf("%s%s", baseURL, endpoint)
	if len(params) > 0 {
		url += "?" + params.Encode()
	}

	return c.doRequest(ctx, http.MethodGet, url, nil, nil)
}

// parseJSON is a helper for parsing JSON responses
func parseJSON(data []byte, v interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	return decoder.Decode(v)
}

// rateLimiter implements a simple token bucket rate limiter
type rateLimiter struct {
	ticker   *time.Ticker
	bucket   chan struct{}
	stopChan chan struct{}
	stopOnce sync.Once
}

func newRateLimiter(requestsPerSecond int) *rateLimiter {
	if requestsPerSecond <= 0 {
		requestsPerSecond = 1
	}

	interval := time.Second / time.Duration(requestsPerSecond)
	rl := &rateLimiter{
		ticker:   time.NewTicker(interval),
		bucket:   make(chan struct{}, 1),
		stopChan: make(chan struct{}),
	}

	// Fill bucket initially
	rl.bucket <- struct{}{}

	// Keep bucket filled
	go func() {
		for {
			select {
			case <-rl.ticker.C:
				select {
				case rl.bucket <- struct{}{}:
				default:
				}
			case <-rl.stopChan:
				return
			}
		}
	}()

	return rl
}

func (rl *rateLimiter) Wait(ctx context.Context) error {
	select {
	case <-rl.bucket:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (rl *rateLimiter) Stop() {
	rl.stopOnce.Do(func() {
		close(rl.stopChan)
		rl.ticker.Stop()
	})
}

// Default user agents to rotate through
var defaultUserAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Safari/605.1.15",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36 Edg/119.0.0.0",
}
