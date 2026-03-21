package nasdaq

import "time"

// Deprecated: Quote is kept for backward compatibility.
// New integrations should prefer QuoteRow and the response-specific types.
// Quote represents stock quote data
type Quote struct {
	Symbol           string    `json:"symbol"`
	CompanyName      string    `json:"companyName"`
	LastSalePrice    string    `json:"lastSalePrice"`
	NetChange        string    `json:"netChange"`
	PercentageChange string    `json:"percentageChange"`
	MarketCap        string    `json:"marketCap"`
	Volume           string    `json:"volume"`
	OpenPrice        string    `json:"openPrice"`
	HighPrice        string    `json:"highPrice"`
	LowPrice         string    `json:"lowPrice"`
	ClosePrice       string    `json:"closePrice"`
	PreviousClose    string    `json:"previousClose"`
	Week52High       string    `json:"52WeekHigh"`
	Week52Low        string    `json:"52WeekLow"`
	PE               string    `json:"p_e"`
	Dividend         string    `json:"dividend"`
	Yield            string    `json:"yield"`
	Beta             string    `json:"beta"`
	EPS              string    `json:"eps"`
	Shares           string    `json:"shares"`
	Timestamp        time.Time `json:"timestamp"`
}

// WatchlistResponse represents the watchlist API response
type WatchlistResponse struct {
	Status struct {
		StatusCode int    `json:"rCode"`
		StatusDesc string `json:"bCodeMessage"`
	} `json:"status"`
	Data []QuoteRow `json:"data"`
}

// QuoteRow represents a row in the quote table
type QuoteRow struct {
	Symbol           string `json:"symbol"`
	Name             string `json:"name"`
	LastSalePrice    string `json:"lastsale"`
	NetChange        string `json:"netchange"`
	PercentageChange string `json:"pctchange"`
	Volume           string `json:"volume"`
	MarketCap        string `json:"marketcap"`
	Country          string `json:"country"`
	IPOYear          string `json:"ipoyear"`
	Industry         string `json:"industry"`
	Sector           string `json:"sector"`

	// Extended session fields (populated when available)
	MarketStatus          string `json:"marketStatus"`
	DeltaIndicator        string `json:"deltaIndicator"`
	AfterHoursPrice       string `json:"afterHoursLastSalePrice"`
	AfterHoursChange      string `json:"afterHoursLastSaleChange"`
	AfterHoursChangePct   string `json:"afterHoursLastSaleChangePercent"`
	PreMarketPrice        string `json:"preMarketLastSalePrice"`
	PreMarketChange       string `json:"preMarketLastSaleChange"`
	PreMarketChangePct    string `json:"preMarketLastSaleChangePercent"`
}

// ExtendedTradingData holds pre/after-hours price data from the extended-trading endpoint.
type ExtendedTradingData struct {
	Symbol    string `json:"symbol"`
	Price     string `json:"lastSalePrice"`
	Change    string `json:"change"`
	ChangePct string `json:"changePct"`
	Status    string `json:"status"` // e.g. "PRE", "POST", "CLOSED"
}

// ScreenerResponse represents the screener API response (normalized across all endpoints).
type ScreenerResponse struct {
	Status struct {
		StatusCode int    `json:"statusCode"`
		StatusDesc string `json:"statusDesc"`
	} `json:"status"`
	Rows []ScreenerRow
}

// ScreenerRow represents a row in the screener table.
// Fields are normalized across endpoints (stocks, ETFs, indices, mutual funds).
type ScreenerRow struct {
	Symbol           string
	Name             string
	LastSalePrice    string
	NetChange        string
	PercentageChange string
	Volume           string
	MarketCap        string
	Country          string
	IPOYear          string
	Industry         string
	Sector           string
}

// --- internal parse types ---

// stocksDownloadRow matches the JSON shape returned by /screener/stocks?download=true
type stocksDownloadRow struct {
	Symbol           string `json:"symbol"`
	Name             string `json:"name"`
	LastSalePrice    string `json:"lastsale"`
	NetChange        string `json:"netchange"`
	PercentageChange string `json:"pctchange"`
	Volume           string `json:"volume"`
	MarketCap        string `json:"marketCap"`
	Country          string `json:"country"`
	IPOYear          string `json:"ipoyear"`
	Industry         string `json:"industry"`
	Sector           string `json:"sector"`
}

func (r stocksDownloadRow) toScreenerRow() ScreenerRow {
	return ScreenerRow{
		Symbol: r.Symbol, Name: r.Name, LastSalePrice: r.LastSalePrice,
		NetChange: r.NetChange, PercentageChange: r.PercentageChange,
		Volume: r.Volume, MarketCap: r.MarketCap, Country: r.Country,
		IPOYear: r.IPOYear, Industry: r.Industry, Sector: r.Sector,
	}
}

// etfDownloadRow matches the JSON shape returned by /screener/etf?download=true
type etfDownloadRow struct {
	Symbol            string `json:"symbol"`
	Name              string `json:"companyName"`
	LastSalePrice     string `json:"lastSalePrice"`
	NetChange         string `json:"netChange"`
	PercentageChange  string `json:"percentageChange"`
	OneYearPercentage string `json:"oneYearPercentage"`
	DeltaIndicator    string `json:"deltaIndicator"`
}

func (r etfDownloadRow) toScreenerRow() ScreenerRow {
	return ScreenerRow{
		Symbol: r.Symbol, Name: r.Name, LastSalePrice: r.LastSalePrice,
		NetChange: r.NetChange, PercentageChange: r.PercentageChange,
	}
}

// indexMFDownloadRow matches the JSON shape returned by /screener/index and /screener/mutualfunds
type indexMFDownloadRow struct {
	Symbol           string `json:"symbol"`
	Name             string `json:"companyName"`
	LastSalePrice    string `json:"lastSalePrice"`
	NetChange        string `json:"netChange"`
	PercentageChange string `json:"percentageChange"`
}

func (r indexMFDownloadRow) toScreenerRow() ScreenerRow {
	return ScreenerRow{
		Symbol: r.Symbol, Name: r.Name, LastSalePrice: r.LastSalePrice,
		NetChange: r.NetChange, PercentageChange: r.PercentageChange,
	}
}

// NewsResponse represents the news API response
type NewsResponse struct {
	Status struct {
		StatusCode int    `json:"rCode"`
		StatusDesc string `json:"bCodeMessage"`
	} `json:"status"`
	Data struct {
		NewsArticles []NewsArticle `json:"rows"`
	} `json:"data"`
}

// NewsArticle represents a news article
type NewsArticle struct {
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	ImageURL    string    `json:"imageUrl"`
	Source      string    `json:"source"`
	PublishTime time.Time `json:"publishTime"`
	Summary     string    `json:"summary"`
}

// TrendingSymbolsResponse represents the trending symbols API response
type TrendingSymbolsResponse struct {
	Status struct {
		StatusCode int    `json:"rCode"`
		StatusDesc string `json:"bCodeMessage"`
	} `json:"status"`
	Data []TrendingSymbol `json:"data"`
}

// TrendingSymbol represents a trending symbol
type TrendingSymbol struct {
	Symbol           string `json:"symbol"`
	Name             string `json:"name"`
	LastSalePrice    string `json:"lastsale"`
	NetChange        string `json:"netchange"`
	PercentageChange string `json:"pctchange"`
	Volume           string `json:"volume"`
}

// SymbolType represents the type of symbol
type SymbolType string

const (
	SymbolTypeStock  SymbolType = "stocks"
	SymbolTypeETF    SymbolType = "etf"
	SymbolTypeIndex  SymbolType = "index"
	SymbolTypeFuture SymbolType = "futures"
)

// AssetClass represents the asset class
type AssetClass string

const (
	AssetClassStock   AssetClass = "stocks"
	AssetClassETF     AssetClass = "etf"
	AssetClassIndex   AssetClass = "index"
	AssetClassFutures AssetClass = "futures"
	AssetClassOptions AssetClass = "options"
)

// SearchSuggestion represents a single search result from autosuggest
type SearchSuggestion struct {
	Symbol      string `json:"symbol"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

// SearchResponse represents the autosuggest API response
type SearchResponse struct {
	Status struct {
		StatusCode int    `json:"rCode"`
		StatusDesc string `json:"bCodeMessage"`
	} `json:"status"`
	Data []SearchSuggestion `json:"data"`
}
