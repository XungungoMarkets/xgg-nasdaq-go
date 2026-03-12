package nasdaq

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	if client.httpClient == nil {
		t.Error("httpClient is nil")
	}

	if client.rateLimiter == nil {
		t.Error("rateLimiter is nil")
	}
}

func TestNewClientWithOptions(t *testing.T) {
	customClient := &http.Client{Timeout: 60 * time.Second}
	client := NewClient(
		WithHTTPClient(customClient),
		WithRateLimit(5),
		WithMaxRetries(5),
		WithRetryDelay(5*time.Second),
		WithBaseURLs("https://example.com/api", "https://example.com/www-api"),
	)

	if client.httpClient != customClient {
		t.Error("HTTP client not set correctly")
	}
	if client.baseAPIURL != "https://example.com/api" {
		t.Errorf("baseAPIURL = %s, want https://example.com/api", client.baseAPIURL)
	}
	if client.baseWWWURL != "https://example.com/www-api" {
		t.Errorf("baseWWWURL = %s, want https://example.com/www-api", client.baseWWWURL)
	}
}

func TestUserAgentRotation(t *testing.T) {
	client := NewClient()

	// Get user agents multiple times
	uas := make([]string, 10)
	for i := 0; i < 10; i++ {
		uas[i] = client.getUserAgent()
	}

	// Check that they rotate
	if uas[0] == uas[1] && uas[1] == uas[2] {
		t.Error("User agents are not rotating")
	}
}

func TestSymbolWithOptionString(t *testing.T) {
	tests := []struct {
		symbol     string
		symbolType SymbolType
		expected   string
	}{
		{"AAPL", SymbolTypeStock, "aapl|stocks"},
		{"NVDA", SymbolTypeETF, "nvda|etf"},
		{"NDX", SymbolTypeIndex, "ndx|index"},
		{"GOOGL", SymbolTypeStock, "googl|stocks"},
	}

	for _, tt := range tests {
		sym := NewSymbolWithOption(tt.symbol, tt.symbolType)
		result := sym.String()
		if result != tt.expected {
			t.Errorf("SymbolWithOption.String() = %v, want %v", result, tt.expected)
		}
	}
}

func TestRateLimiter(t *testing.T) {
	rl := newRateLimiter(10)
	defer rl.Stop()

	ctx := context.Background()

	// Should allow immediate first request
	start := time.Now()
	err := rl.Wait(ctx)
	if err != nil {
		t.Fatalf("Wait() error = %v", err)
	}
	duration := time.Since(start)

	// Should be almost instant
	if duration > 100*time.Millisecond {
		t.Errorf("First request took too long: %v", duration)
	}
}

func TestRateLimiterRespectsRate(t *testing.T) {
	rl := newRateLimiter(2) // 2 requests per second
	defer rl.Stop()

	ctx := context.Background()

	start := time.Now()

	// Make 5 requests
	for i := 0; i < 5; i++ {
		err := rl.Wait(ctx)
		if err != nil {
			t.Fatalf("Wait() error = %v", err)
		}
	}

	duration := time.Since(start)

	// Should take at least 2 seconds for 5 requests at 2/sec
	// First request is instant, then 4 more at 2/sec = ~2 seconds
	if duration < 1*time.Second {
		t.Errorf("Rate limiter not working: completed 5 requests in %v", duration)
	}

	// Should not take more than 3 seconds
	if duration > 3*time.Second {
		t.Errorf("Rate limiter too slow: completed 5 requests in %v", duration)
	}
}

func TestRateLimiterContextCancellation(t *testing.T) {
	rl := newRateLimiter(1)
	defer rl.Stop()

	// Use context with timeout instead
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Drain the bucket first to ensure we need to wait
	select {
	case <-rl.bucket:
		// Token consumed
	default:
		// No token available
	}

	// Now try to wait with a timeout context
	err := rl.Wait(ctx)
	if err != nil {
		// Should get timeout error
		if err != context.DeadlineExceeded {
			t.Errorf("Expected DeadlineExceeded error, got %v", err)
		}
	} else {
		t.Error("Expected timeout error, got nil")
	}
}

func TestDoRequestWithMockServer(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check headers
		if r.Header.Get("User-Agent") == "" {
			t.Error("User-Agent header not set")
		}
		if r.Header.Get("Referer") == "" {
			t.Error("Referer header not set")
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"test": "data"}`))
	}))
	defer server.Close()

	client := NewClient()
	ctx := context.Background()

	// Make request
	data, err := client.doRequest(ctx, http.MethodGet, server.URL, nil, nil)
	if err != nil {
		t.Fatalf("doRequest() error = %v", err)
	}

	if string(data) != `{"test": "data"}` {
		t.Errorf("Response = %v, want %v", string(data), `{"test": "data"}`)
	}
}

func TestDoRequestWithRateLimiting(t *testing.T) {
	callCount := 0

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		// Return rate limit on second request
		if callCount == 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	client := NewClient(
		WithMaxRetries(2),
		WithRetryDelay(100*time.Millisecond),
	)
	ctx := context.Background()

	// First request should succeed
	_, err := client.doRequest(ctx, http.MethodGet, server.URL, nil, nil)
	if err != nil {
		t.Fatalf("First request failed: %v", err)
	}

	// Second request should retry and eventually succeed
	start := time.Now()
	_, err = client.doRequest(ctx, http.MethodGet, server.URL, nil, nil)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Second request failed: %v", err)
	}

	// Should have taken some time due to retry
	if duration < 100*time.Millisecond {
		t.Error("Retry delay not applied")
	}

	// Should have been called at least 3 times (initial + 2 retries)
	if callCount < 3 {
		t.Errorf("Expected at least 3 calls, got %d", callCount)
	}
}

func TestDoRequestContextTimeout(t *testing.T) {
	// Create slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Should fail due to timeout
	_, err := client.doRequest(ctx, http.MethodGet, server.URL, nil, nil)
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
}

func TestMakeAPIRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/quote/watchlist" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("test") != "value" {
			t.Fatalf("unexpected query test=%s", r.URL.Query().Get("test"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"test":"data"}`))
	}))
	defer server.Close()

	client := NewClient(WithBaseURLs(server.URL+"/api", server.URL+"/www-api"))
	ctx := context.Background()
	params := url.Values{}
	params.Set("test", "value")

	data, err := client.makeAPIRequest(ctx, "/quote/watchlist", params)
	if err != nil {
		t.Fatalf("makeAPIRequest() error = %v", err)
	}
	if string(data) != `{"test":"data"}` {
		t.Fatalf("unexpected response: %s", string(data))
	}
}

func TestWithUserAgentsEmptyDoesNotPanic(t *testing.T) {
	client := NewClient(WithUserAgents([]string{}))
	ua := client.getUserAgent()
	if ua == "" {
		t.Fatal("expected non-empty user agent")
	}
}

func TestClientCloseIsIdempotent(t *testing.T) {
	client := NewClient()
	client.Close()
	client.Close()
}

func BenchmarkUserAgentRotation(b *testing.B) {
	client := NewClient()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.getUserAgent()
	}
}

func BenchmarkRateLimiter(b *testing.B) {
	rl := newRateLimiter(100)
	defer rl.Stop()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl.Wait(ctx)
	}
}
