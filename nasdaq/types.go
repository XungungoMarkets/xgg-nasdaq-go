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
		StatusCode int    `json:"statusCode"`
		StatusDesc string `json:"statusDesc"`
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
}

// ScreenerResponse represents the screener API response
type ScreenerResponse struct {
	Status struct {
		StatusCode int    `json:"statusCode"`
		StatusDesc string `json:"statusDesc"`
	} `json:"status"`
	Data struct {
		Table struct {
			Rows     []ScreenerRow     `json:"rows"`
			Headings []ScreenerHeading `json:"headings"`
		} `json:"table"`
	} `json:"data"`
}

// ScreenerRow represents a row in the screener table
type ScreenerRow struct {
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
}

// ScreenerHeading represents a column heading
type ScreenerHeading struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Type  string `json:"type"`
}

// NewsResponse represents the news API response
type NewsResponse struct {
	Status struct {
		StatusCode int    `json:"statusCode"`
		StatusDesc string `json:"statusDesc"`
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
		StatusCode int    `json:"statusCode"`
		StatusDesc string `json:"statusDesc"`
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
