package nasdaq

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const (
	errUnexpectedPath   = "unexpected path: %s"
	errAPIStatusNil     = "expected API status error, got nil"
	errTableOnlyFalse   = "tableonly = %q, want false"
	errTableOnlyTrue    = "tableonly = %q, want true"
	errDownloadTrue     = "download = %q, want true"
	errLimitVal         = "limit = %q, want 10000"
	errRowsWant2        = "len(Rows) = %d, want 2"
	errRowsWant1        = "len(Rows) = %d, want 1"
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
			t.Fatalf(errUnexpectedPath, r.URL.Path)
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
			t.Fatalf(errUnexpectedPath, r.URL.Path)
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
		t.Fatal(errAPIStatusNil)
	}
}

func TestGetQuote(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/quote/watchlist" {
			t.Fatalf(errUnexpectedPath, r.URL.Path)
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

// --- Screener: Stocks ---

func TestGetScreenerStocksTableOnlyFalse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/screener/stocks" {
			t.Fatalf(errUnexpectedPath, r.URL.Path)
		}
		if r.URL.Query().Get("tableonly") != "false" {
			t.Fatalf(errTableOnlyFalse, r.URL.Query().Get("tableonly"))
		}
		if r.URL.Query().Get("download") != "true" {
			t.Fatalf(errDownloadTrue, r.URL.Query().Get("download"))
		}
		w.WriteHeader(http.StatusOK)
		// Real shape returned by /screener/stocks?download=true
		w.Write([]byte(`{
			"status": {"statusCode": 200, "statusDesc": "OK"},
			"data": {
				"asOf": null,
				"headers": {"symbol":"Symbol","name":"Name","lastsale":"Last Sale"},
				"rows": [
					{"symbol":"AAPL","name":"Apple Inc.","lastsale":"$150.00","netchange":"1.00","pctchange":"0.67%","volume":"50000000","marketcap":"2400000000000","country":"United States","ipoyear":"1980","industry":"Computer Hardware","sector":"Technology"},
					{"symbol":"MSFT","name":"Microsoft Corp.","lastsale":"$300.00","netchange":"2.00","pctchange":"0.67%","volume":"30000000","marketcap":"2200000000000","country":"United States","ipoyear":"1986","industry":"Computer Software","sector":"Technology"}
				]
			}
		}`))
	}))
	defer server.Close()

	client := newTestClient(server)
	resp, err := client.GetScreenerStocks(context.Background(), false)
	if err != nil {
		t.Fatalf("GetScreenerStocks() error = %v", err)
	}
	if len(resp.Rows) != 2 {
		t.Fatalf(errRowsWant2, len(resp.Rows))
	}
	if resp.Rows[0].Symbol != "AAPL" {
		t.Errorf("Rows[0].Symbol = %q, want AAPL", resp.Rows[0].Symbol)
	}
	if resp.Rows[0].Name != "Apple Inc." {
		t.Errorf("Rows[0].Name = %q, want Apple Inc.", resp.Rows[0].Name)
	}
	if resp.Rows[0].LastSalePrice != "$150.00" {
		t.Errorf("Rows[0].LastSalePrice = %q, want $150.00", resp.Rows[0].LastSalePrice)
	}
	if resp.Rows[0].Sector != "Technology" {
		t.Errorf("Rows[0].Sector = %q, want Technology", resp.Rows[0].Sector)
	}
}

func TestGetScreenerStocksTableOnlyTrue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("tableonly") != "true" {
			t.Fatalf(errTableOnlyTrue, r.URL.Query().Get("tableonly"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"status": {"statusCode": 200, "statusDesc": "OK"},
			"data": {
				"rows": [{"symbol":"NVDA","name":"NVIDIA","lastsale":"$500.00","netchange":"5.00","pctchange":"1.00%","volume":"10000000","marketcap":"1000000000000","country":"United States","ipoyear":"1999","industry":"Semiconductors","sector":"Technology"}]
			}
		}`))
	}))
	defer server.Close()

	client := newTestClient(server)
	resp, err := client.GetScreenerStocks(context.Background(), true)
	if err != nil {
		t.Fatalf("GetScreenerStocks(tableOnly=true) error = %v", err)
	}
	if len(resp.Rows) != 1 {
		t.Fatalf(errRowsWant1, len(resp.Rows))
	}
	if resp.Rows[0].Symbol != "NVDA" {
		t.Errorf("Rows[0].Symbol = %q, want NVDA", resp.Rows[0].Symbol)
	}
}

func TestGetScreenerStocksAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": {"statusCode": 500, "statusDesc": "Internal Error"}, "data": {"rows": []}}`))
	}))
	defer server.Close()

	client := newTestClient(server)
	_, err := client.GetScreenerStocks(context.Background(), false)
	if err == nil {
		t.Fatal(errAPIStatusNil)
	}
}

// --- Screener: ETFs ---

func TestGetScreenerETFsTableOnlyFalse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/screener/etf" {
			t.Fatalf(errUnexpectedPath, r.URL.Path)
		}
		if r.URL.Query().Get("tableonly") != "false" {
			t.Fatalf(errTableOnlyFalse, r.URL.Query().Get("tableonly"))
		}
		if r.URL.Query().Get("download") != "true" {
			t.Fatalf(errDownloadTrue, r.URL.Query().Get("download"))
		}
		w.WriteHeader(http.StatusOK)
		// Real shape returned by /screener/etf?download=true
		w.Write([]byte(`{
			"status": {"statusCode": 200, "statusDesc": "OK"},
			"data": {
				"dataAsOf": "3/16/2026 8:00:00 PM",
				"data": {
					"asOf": null,
					"headers": {"symbol":"SYMBOL","companyName":"NAME","lastSalePrice":"LAST PRICE"},
					"rows": [
						{"symbol":"QQQ","companyName":"Invesco QQQ Trust","lastSalePrice":"$450.00","netChange":"2.00","percentageChange":"0.44%","oneYearPercentage":"15.00%","deltaIndicator":"up"},
						{"symbol":"SPY","companyName":"SPDR S&P 500 ETF Trust","lastSalePrice":"$500.00","netChange":"1.00","percentageChange":"0.20%","oneYearPercentage":"12.00%","deltaIndicator":"up"}
					]
				}
			}
		}`))
	}))
	defer server.Close()

	client := newTestClient(server)
	resp, err := client.GetScreenerETFs(context.Background(), false)
	if err != nil {
		t.Fatalf("GetScreenerETFs() error = %v", err)
	}
	if len(resp.Rows) != 2 {
		t.Fatalf(errRowsWant2, len(resp.Rows))
	}
	if resp.Rows[0].Symbol != "QQQ" {
		t.Errorf("Rows[0].Symbol = %q, want QQQ", resp.Rows[0].Symbol)
	}
	// companyName must be mapped to Name
	if resp.Rows[0].Name != "Invesco QQQ Trust" {
		t.Errorf("Rows[0].Name = %q, want Invesco QQQ Trust", resp.Rows[0].Name)
	}
	// lastSalePrice must be mapped to LastSalePrice
	if resp.Rows[0].LastSalePrice != "$450.00" {
		t.Errorf("Rows[0].LastSalePrice = %q, want $450.00", resp.Rows[0].LastSalePrice)
	}
	if resp.Rows[0].NetChange != "2.00" {
		t.Errorf("Rows[0].NetChange = %q, want 2.00", resp.Rows[0].NetChange)
	}
}

func TestGetScreenerETFsTableOnlyTrue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("tableonly") != "true" {
			t.Fatalf(errTableOnlyTrue, r.URL.Query().Get("tableonly"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"status": {"statusCode": 200, "statusDesc": "OK"},
			"data": {"data": {"rows": [{"symbol":"IVV","companyName":"iShares Core S&P 500 ETF","lastSalePrice":"$480.00","netChange":"1.50","percentageChange":"0.31%","oneYearPercentage":"11.00%","deltaIndicator":"up"}]}}
		}`))
	}))
	defer server.Close()

	client := newTestClient(server)
	resp, err := client.GetScreenerETFs(context.Background(), true)
	if err != nil {
		t.Fatalf("GetScreenerETFs(tableOnly=true) error = %v", err)
	}
	if len(resp.Rows) != 1 {
		t.Fatalf(errRowsWant1, len(resp.Rows))
	}
	if resp.Rows[0].Symbol != "IVV" {
		t.Errorf("Rows[0].Symbol = %q, want IVV", resp.Rows[0].Symbol)
	}
}

func TestGetScreenerETFsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": {"statusCode": 500, "statusDesc": "Error"}, "data": {"data": {"rows": []}}}`))
	}))
	defer server.Close()

	client := newTestClient(server)
	_, err := client.GetScreenerETFs(context.Background(), false)
	if err == nil {
		t.Fatal(errAPIStatusNil)
	}
}

// --- Screener: Indices ---

func TestGetScreenerIndicesTableOnlyFalse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/screener/index" {
			t.Fatalf(errUnexpectedPath, r.URL.Path)
		}
		if r.URL.Query().Get("tableonly") != "false" {
			t.Fatalf(errTableOnlyFalse, r.URL.Query().Get("tableonly"))
		}
		if r.URL.Query().Get("download") != "true" {
			t.Fatalf(errDownloadTrue, r.URL.Query().Get("download"))
		}
		if r.URL.Query().Get("limit") != "10000" {
			t.Fatalf(errLimitVal, r.URL.Query().Get("limit"))
		}
		w.WriteHeader(http.StatusOK)
		// Real shape returned by /screener/index?download=true
		w.Write([]byte(`{
			"status": {"statusCode": 200, "statusDesc": "OK"},
			"data": {
				"filters": [],
				"records": {
					"totalrecords": 2,
					"limit": 10000,
					"offset": 0,
					"data": {
						"asOf": null,
						"headers": {"symbol":"SYMBOL","companyName":"NAME","lastSalePrice":"LAST"},
						"rows": [
							{"symbol":"NDX","companyName":"NASDAQ-100 Index","lastSalePrice":"$19000.00","netChange":"50.00","percentageChange":"0.26%"},
							{"symbol":"COMP","companyName":"NASDAQ Composite Index","lastSalePrice":"$17000.00","netChange":"30.00","percentageChange":"0.18%"}
						]
					}
				}
			}
		}`))
	}))
	defer server.Close()

	client := newTestClient(server)
	resp, err := client.GetScreenerIndices(context.Background(), false)
	if err != nil {
		t.Fatalf("GetScreenerIndices() error = %v", err)
	}
	if len(resp.Rows) != 2 {
		t.Fatalf(errRowsWant2, len(resp.Rows))
	}
	if resp.Rows[0].Symbol != "NDX" {
		t.Errorf("Rows[0].Symbol = %q, want NDX", resp.Rows[0].Symbol)
	}
	if resp.Rows[0].Name != "NASDAQ-100 Index" {
		t.Errorf("Rows[0].Name = %q, want NASDAQ-100 Index", resp.Rows[0].Name)
	}
	if resp.Rows[0].LastSalePrice != "$19000.00" {
		t.Errorf("Rows[0].LastSalePrice = %q, want $19000.00", resp.Rows[0].LastSalePrice)
	}
}

func TestGetScreenerIndicesTableOnlyTrue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("tableonly") != "true" {
			t.Fatalf(errTableOnlyTrue, r.URL.Query().Get("tableonly"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"status": {"statusCode": 200, "statusDesc": "OK"},
			"data": {"records": {"data": {"rows": [{"symbol":"SPX","companyName":"S&P 500 Index","lastSalePrice":"$5000.00","netChange":"10.00","percentageChange":"0.20%"}]}}}
		}`))
	}))
	defer server.Close()

	client := newTestClient(server)
	resp, err := client.GetScreenerIndices(context.Background(), true)
	if err != nil {
		t.Fatalf("GetScreenerIndices(tableOnly=true) error = %v", err)
	}
	if len(resp.Rows) != 1 {
		t.Fatalf(errRowsWant1, len(resp.Rows))
	}
	if resp.Rows[0].Symbol != "SPX" {
		t.Errorf("Rows[0].Symbol = %q, want SPX", resp.Rows[0].Symbol)
	}
}

func TestGetScreenerIndicesAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": {"statusCode": 500, "statusDesc": "Error"}, "data": {"records": {"data": {"rows": []}}}}`))
	}))
	defer server.Close()

	client := newTestClient(server)
	_, err := client.GetScreenerIndices(context.Background(), false)
	if err == nil {
		t.Fatal(errAPIStatusNil)
	}
}

// --- Screener: Mutual Funds ---

func TestGetScreenerMutualFundsTableOnlyFalse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/screener/mutualfunds" {
			t.Fatalf(errUnexpectedPath, r.URL.Path)
		}
		if r.URL.Query().Get("tableonly") != "false" {
			t.Fatalf(errTableOnlyFalse, r.URL.Query().Get("tableonly"))
		}
		if r.URL.Query().Get("download") != "true" {
			t.Fatalf(errDownloadTrue, r.URL.Query().Get("download"))
		}
		if r.URL.Query().Get("limit") != "10000" {
			t.Fatalf(errLimitVal, r.URL.Query().Get("limit"))
		}
		w.WriteHeader(http.StatusOK)
		// Same shape as /screener/index?download=true
		w.Write([]byte(`{
			"status": {"statusCode": 200, "statusDesc": "OK"},
			"data": {
				"filters": [],
				"records": {
					"totalrecords": 2,
					"limit": 10000,
					"offset": 0,
					"data": {
						"asOf": null,
						"headers": {"symbol":"SYMBOL","companyName":"NAME","lastSalePrice":"LAST"},
						"rows": [
							{"symbol":"FXAIX","companyName":"Fidelity 500 Index Fund","lastSalePrice":"$180.00","netChange":"0.50","percentageChange":"0.28%"},
							{"symbol":"VFIAX","companyName":"Vanguard 500 Index Fund Admiral","lastSalePrice":"$420.00","netChange":"1.20","percentageChange":"0.29%"}
						]
					}
				}
			}
		}`))
	}))
	defer server.Close()

	client := newTestClient(server)
	resp, err := client.GetScreenerMutualFunds(context.Background(), false)
	if err != nil {
		t.Fatalf("GetScreenerMutualFunds() error = %v", err)
	}
	if len(resp.Rows) != 2 {
		t.Fatalf(errRowsWant2, len(resp.Rows))
	}
	if resp.Rows[0].Symbol != "FXAIX" {
		t.Errorf("Rows[0].Symbol = %q, want FXAIX", resp.Rows[0].Symbol)
	}
	if resp.Rows[0].Name != "Fidelity 500 Index Fund" {
		t.Errorf("Rows[0].Name = %q, want Fidelity 500 Index Fund", resp.Rows[0].Name)
	}
	if resp.Rows[0].LastSalePrice != "$180.00" {
		t.Errorf("Rows[0].LastSalePrice = %q, want $180.00", resp.Rows[0].LastSalePrice)
	}
}

func TestGetScreenerMutualFundsTableOnlyTrue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("tableonly") != "true" {
			t.Fatalf(errTableOnlyTrue, r.URL.Query().Get("tableonly"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"status": {"statusCode": 200, "statusDesc": "OK"},
			"data": {"records": {"data": {"rows": [{"symbol":"VTSAX","companyName":"Vanguard Total Stock Market Index Fund Admiral","lastSalePrice":"$110.00","netChange":"0.30","percentageChange":"0.27%"}]}}}
		}`))
	}))
	defer server.Close()

	client := newTestClient(server)
	resp, err := client.GetScreenerMutualFunds(context.Background(), true)
	if err != nil {
		t.Fatalf("GetScreenerMutualFunds(tableOnly=true) error = %v", err)
	}
	if len(resp.Rows) != 1 {
		t.Fatalf(errRowsWant1, len(resp.Rows))
	}
	if resp.Rows[0].Symbol != "VTSAX" {
		t.Errorf("Rows[0].Symbol = %q, want VTSAX", resp.Rows[0].Symbol)
	}
}

func TestGetScreenerMutualFundsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": {"statusCode": 500, "statusDesc": "Error"}, "data": {"records": {"data": {"rows": []}}}}`))
	}))
	defer server.Close()

	client := newTestClient(server)
	_, err := client.GetScreenerMutualFunds(context.Background(), false)
	if err == nil {
		t.Fatal(errAPIStatusNil)
	}
}

func TestGetNews(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/news/topic/latestnews" {
			t.Fatalf(errUnexpectedPath, r.URL.Path)
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
			t.Fatalf(errUnexpectedPath, r.URL.Path)
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
			t.Fatalf(errUnexpectedPath, r.URL.Path)
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

func TestSearch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"status": {"statusCode": 200, "statusDesc": "OK"},
			"data": [
				{"symbol":"NVDA","name":"NVIDIA Corporation","type":"stocks","description":"Technology sector"},
				{"symbol":"NVDAW","name":"NVIDIA Corp","type":"stocks","description":"Technology"}
			]
		}`))
	}))
	defer server.Close()

	client := NewClient(
		WithHTTPClient(server.Client()),
		WithBaseURLs(server.URL+"/api", server.URL),
		WithRateLimit(1000),
		WithMaxRetries(0),
	)
	ctx := context.Background()
	resp, err := client.Search(ctx, "NVDA", 10, false)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("len(resp.Data) = %d, want 2", len(resp.Data))
	}
	if resp.Data[0].Symbol != "NVDA" {
		t.Fatalf("first symbol = %s, want NVDA", resp.Data[0].Symbol)
	}
}

func TestSearchEmptyQuery(t *testing.T) {
	client := NewClient()
	ctx := context.Background()
	_, err := client.Search(ctx, "", 10, false)
	if err == nil {
		t.Fatal("expected error for empty query")
	}
}

func TestSearchWithMarketData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"status": {"statusCode": 200, "statusDesc": "OK"},
			"data": [{"symbol":"AAPL","name":"Apple Inc","type":"stocks","description":"Consumer Electronics"}]
		}`))
	}))
	defer server.Close()

	client := NewClient(
		WithHTTPClient(server.Client()),
		WithBaseURLs(server.URL+"/api", server.URL),
		WithRateLimit(1000),
		WithMaxRetries(0),
	)
	ctx := context.Background()
	_, err := client.Search(ctx, "AAPL", 5, true)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
}
