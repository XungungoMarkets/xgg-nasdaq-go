package nasdaq

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

func apiStatusError(code int, desc string) error {
	if code == 200 {
		return nil
	}
	if desc == "" {
		return fmt.Errorf("API error: statusCode=%d", code)
	}
	return fmt.Errorf("API error: statusCode=%d statusDesc=%s", code, desc)
}

// GetWatchlist retrieves quote data for multiple symbols
// Symbols can be stocks, ETFs, or indices
func (c *Client) GetWatchlist(ctx context.Context, symbols []SymbolWithOption) (*WatchlistResponse, error) {
	if len(symbols) == 0 {
		return nil, fmt.Errorf("at least one symbol is required")
	}

	params := url.Values{}
	for _, sym := range symbols {
		params.Add("symbol", sym.String())
	}
	if c.watchlistType != "" {
		params.Set("type", c.watchlistType)
	}

	data, err := c.makeAPIRequest(ctx, "/quote/watchlist", params)
	if err != nil {
		return nil, err
	}

	var response WatchlistResponse
	if err := parseJSON(data, &response); err != nil {
		return nil, err
	}
	if err := apiStatusError(response.Status.StatusCode, response.Status.StatusDesc); err != nil {
		return nil, err
	}

	return &response, nil
}

// GetQuote retrieves quote data for a single symbol
func (c *Client) GetQuote(ctx context.Context, symbol string, symbolType SymbolType) (*QuoteRow, error) {
	sym := SymbolWithOption{
		Symbol: symbol,
		Type:   symbolType,
	}

	resp, err := c.GetWatchlist(ctx, []SymbolWithOption{sym})
	if err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no data found for symbol %s", symbol)
	}

	return &resp.Data[0], nil
}

// GetScreenerStocks retrieves stock screener data.
// With download=true the API returns the full table under data.rows (no pagination).
func (c *Client) GetScreenerStocks(ctx context.Context, tableOnly bool) (*ScreenerResponse, error) {
	params := url.Values{}
	params.Set("tableonly", fmt.Sprintf("%t", tableOnly))
	params.Set("download", "true")

	data, err := c.makeAPIRequest(ctx, "/screener/stocks", params)
	if err != nil {
		return nil, err
	}

	var raw struct {
		Status struct {
			RCode        int    `json:"rCode"`
			BCodeMessage string `json:"bCodeMessage"`
		} `json:"status"`
		Data struct {
			Rows []stocksDownloadRow `json:"rows"`
		} `json:"data"`
	}
	if err := parseJSON(data, &raw); err != nil {
		return nil, err
	}
	if err := apiStatusError(raw.Status.RCode, raw.Status.BCodeMessage); err != nil {
		return nil, err
	}

	resp := &ScreenerResponse{}
	resp.Status.StatusCode = raw.Status.RCode
	resp.Rows = make([]ScreenerRow, len(raw.Data.Rows))
	for i, r := range raw.Data.Rows {
		resp.Rows[i] = r.toScreenerRow()
	}
	return resp, nil
}

// GetScreenerETFs retrieves ETF screener data.
// With download=true the API returns the full table under data.data.rows.
func (c *Client) GetScreenerETFs(ctx context.Context, tableOnly bool) (*ScreenerResponse, error) {
	params := url.Values{}
	params.Set("tableonly", fmt.Sprintf("%t", tableOnly))
	params.Set("download", "true")

	data, err := c.makeAPIRequest(ctx, "/screener/etf", params)
	if err != nil {
		return nil, err
	}

	var raw struct {
		Status struct {
			RCode        int    `json:"rCode"`
			BCodeMessage string `json:"bCodeMessage"`
		} `json:"status"`
		Data struct {
			Data struct {
				Rows []etfDownloadRow `json:"rows"`
			} `json:"data"`
		} `json:"data"`
	}
	if err := parseJSON(data, &raw); err != nil {
		return nil, err
	}
	if err := apiStatusError(raw.Status.RCode, raw.Status.BCodeMessage); err != nil {
		return nil, err
	}

	resp := &ScreenerResponse{}
	resp.Status.StatusCode = raw.Status.RCode
	resp.Rows = make([]ScreenerRow, len(raw.Data.Data.Rows))
	for i, r := range raw.Data.Data.Rows {
		resp.Rows[i] = r.toScreenerRow()
	}
	return resp, nil
}

// GetScreenerIndices retrieves index screener data.
// download=true does not bypass pagination for this endpoint; rows are under data.records.data.rows.
func (c *Client) GetScreenerIndices(ctx context.Context, tableOnly bool) (*ScreenerResponse, error) {
	params := url.Values{}
	params.Set("tableonly", fmt.Sprintf("%t", tableOnly))
	params.Set("download", "true")
	params.Set("limit", "10000")

	data, err := c.makeAPIRequest(ctx, "/screener/index", params)
	if err != nil {
		return nil, err
	}

	var raw struct {
		Status struct {
			RCode        int    `json:"rCode"`
			BCodeMessage string `json:"bCodeMessage"`
		} `json:"status"`
		Data struct {
			Records struct {
				Data struct {
					Rows []indexMFDownloadRow `json:"rows"`
				} `json:"data"`
			} `json:"records"`
		} `json:"data"`
	}
	if err := parseJSON(data, &raw); err != nil {
		return nil, err
	}
	if err := apiStatusError(raw.Status.RCode, raw.Status.BCodeMessage); err != nil {
		return nil, err
	}

	resp := &ScreenerResponse{}
	resp.Status.StatusCode = raw.Status.RCode
	resp.Rows = make([]ScreenerRow, len(raw.Data.Records.Data.Rows))
	for i, r := range raw.Data.Records.Data.Rows {
		resp.Rows[i] = r.toScreenerRow()
	}
	return resp, nil
}

// GetScreenerMutualFunds retrieves mutual fund screener data.
// download=true does not bypass pagination for this endpoint; rows are under data.records.data.rows.
func (c *Client) GetScreenerMutualFunds(ctx context.Context, tableOnly bool) (*ScreenerResponse, error) {
	params := url.Values{}
	params.Set("tableonly", fmt.Sprintf("%t", tableOnly))
	params.Set("download", "true")
	params.Set("limit", "10000")

	data, err := c.makeAPIRequest(ctx, "/screener/mutualfunds", params)
	if err != nil {
		return nil, err
	}

	var raw struct {
		Status struct {
			RCode        int    `json:"rCode"`
			BCodeMessage string `json:"bCodeMessage"`
		} `json:"status"`
		Data struct {
			Records struct {
				Data struct {
					Rows []indexMFDownloadRow `json:"rows"`
				} `json:"data"`
			} `json:"records"`
		} `json:"data"`
	}
	if err := parseJSON(data, &raw); err != nil {
		return nil, err
	}
	if err := apiStatusError(raw.Status.RCode, raw.Status.BCodeMessage); err != nil {
		return nil, err
	}

	resp := &ScreenerResponse{}
	resp.Status.StatusCode = raw.Status.RCode
	resp.Rows = make([]ScreenerRow, len(raw.Data.Records.Data.Rows))
	for i, r := range raw.Data.Records.Data.Rows {
		resp.Rows[i] = r.toScreenerRow()
	}
	return resp, nil
}

// GetNews retrieves latest news articles
func (c *Client) GetNews(ctx context.Context, offset, limit int, blacklist bool) (*NewsResponse, error) {
	params := url.Values{}
	params.Set("offset", fmt.Sprintf("%d", offset))
	params.Set("limit", fmt.Sprintf("%d", limit))
	params.Set("blacklist", fmt.Sprintf("%t", blacklist))

	data, err := c.makeWWWAPIRequest(ctx, "/news/topic/latestnews", params)
	if err != nil {
		return nil, err
	}

	var response NewsResponse
	if err := parseJSON(data, &response); err != nil {
		return nil, err
	}
	if err := apiStatusError(response.Status.StatusCode, response.Status.StatusDesc); err != nil {
		return nil, err
	}

	return &response, nil
}

// GetTrendingSymbols retrieves trending symbols by asset class
func (c *Client) GetTrendingSymbols(ctx context.Context, assetClass AssetClass) (*TrendingSymbolsResponse, error) {
	params := url.Values{}
	params.Set("assetclass", string(assetClass))

	data, err := c.makeWWWAPIRequest(ctx, "/ga/trending-symbols", params)
	if err != nil {
		return nil, err
	}

	var response TrendingSymbolsResponse
	if err := parseJSON(data, &response); err != nil {
		return nil, err
	}
	if err := apiStatusError(response.Status.StatusCode, response.Status.StatusDesc); err != nil {
		return nil, err
	}

	return &response, nil
}

// GetBellNotifications retrieves current bell events (IPOs, stock upgrades, etc.)
func (c *Client) GetBellNotifications(ctx context.Context) ([]map[string]interface{}, error) {
	data, err := c.makeWWWAPIRequest(ctx, "/nasdaq-bell-notifications/current-events", nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Status struct {
			StatusCode int    `json:"rCode"`
			StatusDesc string `json:"bCodeMessage"`
		} `json:"status"`
		Data []map[string]interface{} `json:"data"`
	}

	if err := parseJSON(data, &response); err != nil {
		return nil, err
	}

	if err := apiStatusError(response.Status.StatusCode, response.Status.StatusDesc); err != nil {
		return nil, err
	}

	return response.Data, nil
}

// GetMarketInfo retrieves general market information
func (c *Client) GetMarketInfo(ctx context.Context) (map[string]interface{}, error) {
	data, err := c.makeAPIRequest(ctx, "/market-info", nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Status struct {
			StatusCode int    `json:"rCode"`
			StatusDesc string `json:"bCodeMessage"`
		} `json:"status"`
		Data map[string]interface{} `json:"data"`
	}

	if err := parseJSON(data, &response); err != nil {
		return nil, err
	}

	if err := apiStatusError(response.Status.StatusCode, response.Status.StatusDesc); err != nil {
		return nil, err
	}

	return response.Data, nil
}

// GetSymbolInfo retrieves detailed information about a specific symbol
func (c *Client) GetSymbolInfo(ctx context.Context, symbol string, assetClass AssetClass) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("assetclass", string(assetClass))

	endpoint := fmt.Sprintf("/quote/%s/info", symbol)
	data, err := c.makeAPIRequest(ctx, endpoint, params)
	if err != nil {
		return nil, err
	}

	var response struct {
		Status struct {
			StatusCode int    `json:"rCode"`
			StatusDesc string `json:"bCodeMessage"`
		} `json:"status"`
		Data map[string]interface{} `json:"data"`
	}

	if err := parseJSON(data, &response); err != nil {
		return nil, err
	}

	if err := apiStatusError(response.Status.StatusCode, response.Status.StatusDesc); err != nil {
		return nil, err
	}

	return response.Data, nil
}

// GetBasicQuotes retrieves basic quote data for multiple symbols
func (c *Client) GetBasicQuotes(ctx context.Context, symbols []SymbolWithOption) ([]map[string]interface{}, error) {
	if len(symbols) == 0 {
		return nil, fmt.Errorf("at least one symbol is required")
	}

	params := url.Values{}
	for _, sym := range symbols {
		params.Add("symbol", sym.String())
	}

	data, err := c.makeAPIRequest(ctx, "/quote/basic", params)
	if err != nil {
		return nil, err
	}

	var response struct {
		Status struct {
			StatusCode int    `json:"rCode"`
			StatusDesc string `json:"bCodeMessage"`
		} `json:"status"`
		Data []map[string]interface{} `json:"data"`
	}

	if err := parseJSON(data, &response); err != nil {
		return nil, err
	}

	if err := apiStatusError(response.Status.StatusCode, response.Status.StatusDesc); err != nil {
		return nil, err
	}

	return response.Data, nil
}

// GetSymbolChart retrieves chart data for a specific symbol
func (c *Client) GetSymbolChart(ctx context.Context, symbol string, assetClass AssetClass) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("assetclass", string(assetClass))

	endpoint := fmt.Sprintf("/quote/%s/chart", symbol)
	data, err := c.makeAPIRequest(ctx, endpoint, params)
	if err != nil {
		return nil, err
	}

	var response struct {
		Status struct {
			StatusCode int    `json:"rCode"`
			StatusDesc string `json:"bCodeMessage"`
		} `json:"status"`
		Data map[string]interface{} `json:"data"`
	}

	if err := parseJSON(data, &response); err != nil {
		return nil, err
	}

	if response.Status.StatusCode != 200 {
		return nil, fmt.Errorf("API error: %s", response.Status.StatusDesc)
	}

	return response.Data, nil
}

// GetExtendedTrading retrieves pre/after-hours trading data for a single stock symbol.
// It calls /api/quote/{symbol}/extended-trading?assetclass=stocks.
func (c *Client) GetExtendedTrading(ctx context.Context, symbol string) (*ExtendedTradingData, error) {
	params := url.Values{}
	params.Set("assetclass", "stocks")

	endpoint := fmt.Sprintf("/quote/%s/extended-trading", strings.ToUpper(symbol))
	data, err := c.makeAPIRequest(ctx, endpoint, params)
	if err != nil {
		return nil, err
	}

	var raw struct {
		Status struct {
			RCode        int    `json:"rCode"`
			BCodeMessage string `json:"bCodeMessage"`
		} `json:"status"`
		Data struct {
			Summary struct {
				Symbol    string `json:"symbol"`
				Price     string `json:"lastSalePrice"`
				Change    string `json:"change"`
				ChangePct string `json:"changePct"`
				Status    string `json:"status"`
			} `json:"summary"`
		} `json:"data"`
	}
	if err := parseJSON(data, &raw); err != nil {
		return nil, err
	}
	if err := apiStatusError(raw.Status.RCode, raw.Status.BCodeMessage); err != nil {
		return nil, err
	}

	s := raw.Data.Summary
	return &ExtendedTradingData{
		Symbol:    s.Symbol,
		Price:     s.Price,
		Change:    s.Change,
		ChangePct: s.ChangePct,
		Status:    s.Status,
	}, nil
}

// Search performs an autosuggest search for symbols
func (c *Client) Search(ctx context.Context, query string, limit int, includeMarketData bool) (*SearchResponse, error) {
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	params := url.Values{}
	params.Set("query", query)
	params.Set("limit", fmt.Sprintf("%d", limit))
	params.Set("use_cache", "true")
	params.Set("include_market_data", fmt.Sprintf("%t", includeMarketData))

	data, err := c.makeAISearchRequest(ctx, "/ai-search/external/content-search-bff/v1/autosuggest", params)
	if err != nil {
		return nil, err
	}

	var response SearchResponse
	if err := parseJSON(data, &response); err != nil {
		return nil, err
	}
	if err := apiStatusError(response.Status.StatusCode, response.Status.StatusDesc); err != nil {
		return nil, err
	}

	return &response, nil
}

// SymbolWithOption represents a symbol with its type option
type SymbolWithOption struct {
	Symbol string
	Type   SymbolType
}

// String returns to formatted symbol string for API requests
func (s SymbolWithOption) String() string {
	return fmt.Sprintf("%s|%s", strings.ToLower(s.Symbol), s.Type)
}

// NewSymbolWithOption creates a new SymbolWithOption
func NewSymbolWithOption(symbol string, symbolType SymbolType) SymbolWithOption {
	return SymbolWithOption{
		Symbol: symbol,
		Type:   symbolType,
	}
}
