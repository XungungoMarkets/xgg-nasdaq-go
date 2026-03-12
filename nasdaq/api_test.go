package nasdaq

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestClient(server *httptest.Server) *Client {
	return NewClient(
		WithHTTPClient(server.Client()),
		WithBaseURLs(server.URL+"/api", server.URL+"/api"),
		WithRateLimit(1000),
		WithMaxRetries(0),
	)
}

func TestNewSymbolWithOption(t *testing.T) {
	sym := NewSymbolWithOption("AAPL", SymbolTypeStock)
	if sym.Symbol != "AAPL" {
		t.Errorf("Symbol = %v, want AAPL", sym.Symbol)
	}
	if sym.Type != SymbolTypeStock {
		t.Errorf("Type = %v, want SymbolTypeStock", sym.Type)
	}
}

func TestGetWatchlistWithMultipleSymbols(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/quote/watchlist" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		symbols := r.URL.Query()["symbol"]
		if len(symbols) != 2 {
			t.Fatalf("expected 2 symbol params, got %d (%v)", len(symbols), symbols)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"status": {"statusCode": 200, "statusDesc": "OK"},
			"data": [
				{"symbol":"AAPL","name":"Apple","lastsale":"$150","netchange":"+1","pctchange":"+0.6%","volume":"100","marketcap":"1"},
				{"symbol":"MSFT","name":"Microsoft","lastsale":"$300","netchange":"+2","pctchange":"+0.7%","volume":"200","marketcap":"2"}
			]
		}`))
	}))
	defer server.Close()

	client := newTestClient(server)
	ctx := context.Background()
	resp, err := client.GetWatchlist(ctx, []SymbolWithOption{
		NewSymbolWithOption("AAPL", SymbolTypeStock),
		NewSymbolWithOption("MSFT", SymbolTypeStock),
	})
	if err != nil {
		t.Fatalf("GetWatchlist() error = %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("len(resp.Data) = %d, want 2", len(resp.Data))
	}
}

func TestGetWatchlistIncludesTypeWhenConfigured(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/quote/watchlist" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("type") != "Rv" {
			t.Fatalf("expected type=Rv, got type=%q", r.URL.Query().Get("type"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"status": {"statusCode": 200, "statusDesc": "OK"},
			"data": [
				{"symbol":"AAPL","name":"Apple","lastsale":"$150","netchange":"+1","pctchange":"+0.6%","volume":"100","marketcap":"1"}
			]
		}`))
	}))
	defer server.Close()

	client := NewClient(
		WithHTTPClient(server.Client()),
		WithBaseURLs(server.URL+"/api", server.URL+"/api"),
		WithRateLimit(1000),
		WithMaxRetries(0),
		WithWatchlistType("Rv"),
	)

	_, err := client.GetWatchlist(context.Background(), []SymbolWithOption{
		NewSymbolWithOption("AAPL", SymbolTypeStock),
	})
	if err != nil {
		t.Fatalf("GetWatchlist() error = %v", err)
	}
}

func TestGetWatchlistReturnsErrorOnAPINon200Status(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"status": {"statusCode": 500, "statusDesc": "Internal Error"},
			"data": []
		}`))
	}))
	defer server.Close()

	client := newTestClient(server)
	_, err := client.GetWatchlist(context.Background(), []SymbolWithOption{
		NewSymbolWithOption("AAPL", SymbolTypeStock),
	})
	if err == nil {
		t.Fatal("expected API status error, got nil")
	}
}

func TestGetQuote(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/quote/watchlist" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"status": {"statusCode": 200, "statusDesc": "OK"},
			"data": [
				{"symbol":"GOOGL","name":"Alphabet","lastsale":"$120","netchange":"+2","pctchange":"+1.6%","volume":"300","marketcap":"3"}
			]
		}`))
	}))
	defer server.Close()

	client := newTestClient(server)
	ctx := context.Background()
	quote, err := client.GetQuote(ctx, "GOOGL", SymbolTypeStock)
	if err != nil {
		t.Fatalf("GetQuote() error = %v", err)
	}
	if quote.Symbol != "GOOGL" {
		t.Fatalf("quote.Symbol = %s, want GOOGL", quote.Symbol)
	}
}

func TestGetScreenerStocks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/screener/stocks" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("tableonly") != "false" {
			t.Fatalf("unexpected tableonly: %s", r.URL.Query().Get("tableonly"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"status": {"statusCode": 200, "statusDesc": "OK"},
			"data": {"table": {"rows":[{"symbol":"AAPL","name":"Apple","lastsale":"$150"}], "headings":[]}}
		}`))
	}))
	defer server.Close()

	client := newTestClient(server)
	ctx := context.Background()
	resp, err := client.GetScreenerStocks(ctx, false)
	if err != nil {
		t.Fatalf("GetScreenerStocks() error = %v", err)
	}
	if len(resp.Data.Table.Rows) != 1 {
		t.Fatalf("len(rows) = %d, want 1", len(resp.Data.Table.Rows))
	}
}

func TestGetNews(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/news/topic/latestnews" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("limit") != "10" {
			t.Fatalf("unexpected limit: %s", r.URL.Query().Get("limit"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"status": {"statusCode": 200, "statusDesc": "OK"},
			"data": {"rows":[{"title":"Test","url":"https://example.com","publishTime":"2025-01-01T00:00:00Z"}]}
		}`))
	}))
	defer server.Close()

	client := newTestClient(server)
	ctx := context.Background()
	resp, err := client.GetNews(ctx, 0, 10, true)
	if err != nil {
		t.Fatalf("GetNews() error = %v", err)
	}
	if len(resp.Data.NewsArticles) != 1 {
		t.Fatalf("len(news) = %d, want 1", len(resp.Data.NewsArticles))
	}
}

func TestGetTrendingSymbols(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/ga/trending-symbols" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("assetclass") != "stocks" {
			t.Fatalf("unexpected assetclass: %s", r.URL.Query().Get("assetclass"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"status": {"statusCode": 200, "statusDesc": "OK"},
			"data": [{"symbol":"AAPL","name":"Apple","lastsale":"$150","netchange":"+5","pctchange":"+3%","volume":"100"}]
		}`))
	}))
	defer server.Close()

	client := newTestClient(server)
	ctx := context.Background()
	resp, err := client.GetTrendingSymbols(ctx, AssetClassStock)
	if err != nil {
		t.Fatalf("GetTrendingSymbols() error = %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("len(trending) = %d, want 1", len(resp.Data))
	}
}

func TestGetBasicQuotesWithMultipleSymbols(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/quote/basic" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		symbols := r.URL.Query()["symbol"]
		if len(symbols) != 2 {
			t.Fatalf("expected 2 symbol params, got %d (%v)", len(symbols), symbols)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"status": {"statusCode": 200, "statusDesc": "OK"},
			"data": [{"symbol":"AAPL"},{"symbol":"MSFT"}]
		}`))
	}))
	defer server.Close()

	client := newTestClient(server)
	ctx := context.Background()
	resp, err := client.GetBasicQuotes(ctx, []SymbolWithOption{
		NewSymbolWithOption("AAPL", SymbolTypeStock),
		NewSymbolWithOption("MSFT", SymbolTypeStock),
	})
	if err != nil {
		t.Fatalf("GetBasicQuotes() error = %v", err)
	}
	if len(resp) != 2 {
		t.Fatalf("len(resp) = %d, want 2", len(resp))
	}
}

func TestParseJSON(t *testing.T) {
	validJSON := []byte(`{"test": "value", "number": 123}`)
	var result struct {
		Test   string `json:"test"`
		Number int    `json:"number"`
	}
	if err := parseJSON(validJSON, &result); err != nil {
		t.Fatalf("parseJSON() error = %v", err)
	}
	if result.Test != "value" || result.Number != 123 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestParseJSONInvalid(t *testing.T) {
	var result struct{}
	if err := parseJSON([]byte(`{invalid json}`), &result); err == nil {
		t.Fatal("expected parseJSON error for invalid JSON")
	}
}

func TestGetWatchlistEmptySymbols(t *testing.T) {
	client := NewClient()
	ctx := context.Background()
	if _, err := client.GetWatchlist(ctx, []SymbolWithOption{}); err == nil {
		t.Fatal("expected error for empty symbols slice")
	}
}

func TestAssetClassConstants(t *testing.T) {
	tests := []struct {
		name       string
		assetClass AssetClass
		expected   string
	}{
		{"Stock", AssetClassStock, "stocks"},
		{"ETF", AssetClassETF, "etf"},
		{"Index", AssetClassIndex, "index"},
		{"Futures", AssetClassFutures, "futures"},
		{"Options", AssetClassOptions, "options"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.assetClass) != tt.expected {
				t.Errorf("AssetClass = %v, want %v", tt.assetClass, tt.expected)
			}
		})
	}
}

func TestSymbolTypeConstants(t *testing.T) {
	tests := []struct {
		name       string
		symbolType SymbolType
		expected   string
	}{
		{"Stock", SymbolTypeStock, "stocks"},
		{"ETF", SymbolTypeETF, "etf"},
		{"Index", SymbolTypeIndex, "index"},
		{"Future", SymbolTypeFuture, "futures"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.symbolType) != tt.expected {
				t.Errorf("SymbolType = %v, want %v", tt.symbolType, tt.expected)
			}
		})
	}
}

func TestPublishTimeParsing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"status": {"statusCode": 200, "statusDesc": "OK"},
			"data": {"rows":[{"title":"Test","url":"https://example.com","publishTime":"2026-01-01T12:30:00Z"}]}
		}`))
	}))
	defer server.Close()

	client := newTestClient(server)
	resp, err := client.GetNews(context.Background(), 0, 1, true)
	if err != nil {
		t.Fatalf("GetNews() error = %v", err)
	}

	expected := time.Date(2026, 1, 1, 12, 30, 0, 0, time.UTC)
	if !resp.Data.NewsArticles[0].PublishTime.Equal(expected) {
		t.Fatalf("publish time = %v, want %v", resp.Data.NewsArticles[0].PublishTime, expected)
	}
}
