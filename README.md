# XGG NASDAQ Go Library

A comprehensive Go library for accessing NASDAQ market data with built-in rate limiting to prevent IP bans.

## Features

- ✅ Rate limiting (configurable requests per second)
- ✅ Automatic retry with exponential backoff
- ✅ User agent rotation to mimic browser behavior
- ✅ Context support for cancellation
- ✅ Comprehensive market data access
- ✅ Anti-ban measures

## Installation

```bash
go get github.com/XungungoMarkets/xgg-nasdaq-go
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/XungungoMarkets/xgg-nasdaq-go/nasdaq"
)

func main() {
    // Create client with rate limiting
    client := nasdaq.NewClient(
        nasdaq.WithRateLimit(2),  // 2 requests per second
        nasdaq.WithWatchlistType("Rv"), // Optional parity with some NASDAQ watchlist calls
    )
    defer client.Close()
    
    ctx := context.Background()
    
    // Get stock quote
    quote, err := client.GetQuote(ctx, "AAPL", nasdaq.SymbolTypeStock)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("%s: %s\n", quote.Symbol, quote.LastSalePrice)
}
```

## Anti-Ban Measures

This library implements several measures to avoid IP bans:

### 1. Rate Limiting
Default: 2 requests per second (configurable)

```go
client := nasdaq.NewClient(
    nasdaq.WithRateLimit(1),  // Conservative: 1 request per second
)
```

### 2. Automatic Retries with Exponential Backoff
- Default: 3 retries
- Initial delay: 2 seconds
- Exponential backoff with jitter

```go
client := nasdaq.NewClient(
    nasdaq.WithMaxRetries(5),              // More retries
    nasdaq.WithRetryDelay(3 * time.Second), // Longer initial delay
)
```

### 3. User Agent Rotation
Automatically rotates through 5 different browser user agents to mimic real browser traffic.

### 4. Browser Headers
All requests include proper browser headers:
- User-Agent (rotated)
- Accept, Accept-Language, Accept-Encoding
- Referer, Origin
- Connection: keep-alive

### 5. HTTP 429 Detection
Automatically detects rate limiting responses (HTTP 429) and retries with backoff.

## API Methods

### Get Quote
Get real-time quote data for a single symbol.

```go
quote, err := client.GetQuote(ctx, "AAPL", nasdaq.SymbolTypeStock)
```

**Parameters:**
- `ctx`: Context for cancellation
- `symbol`: Stock symbol (e.g., "AAPL", "GOOGL")
- `symbolType`: Type of symbol (Stock, ETF, Index, Future)

**Returns:** `QuoteRow` with price, change, volume, market cap, etc.

### Get Watchlist
Get quotes for multiple symbols in a single request.

```go
symbols := []nasdaq.SymbolWithOption{
    nasdaq.NewSymbolWithOption("NVDA", nasdaq.SymbolTypeStock),
    nasdaq.NewSymbolWithOption("MSFT", nasdaq.SymbolTypeStock),
    nasdaq.NewSymbolWithOption("NDX", nasdaq.SymbolTypeIndex),
}

watchlist, err := client.GetWatchlist(ctx, symbols)
```

### Get Screener Data
Get comprehensive lists of stocks, ETFs, indices, or mutual funds.

```go
// Stocks
stocks, err := client.GetScreenerStocks(ctx, false)

// ETFs
etfs, err := client.GetScreenerETFs(ctx, false)

// Indices
indices, err := client.GetScreenerIndices(ctx, false)

// Mutual Funds
funds, err := client.GetScreenerMutualFunds(ctx, false)
```

**Parameters:**
- `tableOnly`: If true, returns only table data (faster, less data)

All screener methods send `download=true` to the API, which instructs it to return the **complete table** without pagination limits. This is the equivalent of the "Download" option on the NASDAQ website.

### Get News
Get latest news articles.

```go
news, err := client.GetNews(ctx, 0, 10, true)
```

**Parameters:**
- `offset`: Starting position for pagination
- `limit`: Number of articles to retrieve
- `blacklist`: Whether to apply blacklist filters

### Get Trending Symbols
Get currently trending symbols by asset class.

```go
trending, err := client.GetTrendingSymbols(ctx, nasdaq.AssetClassStock)
```

**Asset Classes:**
- `AssetClassStock` - Stocks
- `AssetClassETF` - ETFs
- `AssetClassIndex` - Indices
- `AssetClassFutures` - Futures

### Get Bell Notifications
Get current bell events including IPOs, stock upgrades, and other market events.

```go
bellEvents, err := client.GetBellNotifications(ctx)
```

**Returns:** Array of flexible map structures containing event data such as:
- IPO announcements
- Stock upgrades/downgrades
- Market opening/closing bells
- Other corporate events

### Get Market Info
Get general market information including market status and summary data.

```go
marketInfo, err := client.GetMarketInfo(ctx)
```

**Returns:** Map containing market information such as:
- Market status (open/closed)
- Market indices
- Trading hours
- Other market summary data

### Get Symbol Info
Get detailed information about a specific symbol.

```go
symbolInfo, err := client.GetSymbolInfo(ctx, "AAPL", nasdaq.AssetClassStock)
```

**Parameters:**
- `symbol`: Stock symbol (e.g., "AAPL", "NDX")
- `assetClass`: Asset class (Stock, ETF, Index, Futures)

**Returns:** Map containing detailed symbol information

### Get Basic Quotes
Get basic quote data for multiple symbols (alternative to watchlist).

```go
symbols := []nasdaq.SymbolWithOption{
    nasdaq.NewSymbolWithOption("AAPL", nasdaq.SymbolTypeStock),
    nasdaq.NewSymbolWithOption("MSFT", nasdaq.SymbolTypeStock),
}

basicQuotes, err := client.GetBasicQuotes(ctx, symbols)
```

**Returns:** Array of map structures with basic quote data

### Get Symbol Chart
Get chart data for a specific symbol.

```go
chartData, err := client.GetSymbolChart(ctx, "AAPL", nasdaq.AssetClassStock)
```

**Parameters:**
- `symbol`: Stock symbol (e.g., "AAPL", "NDX")
- `assetClass`: Asset class (Stock, ETF, Index, Futures)

**Returns:** Map containing chart data with price history

### Search
Search for symbols using the autosuggest feature to find stocks, ETFs, indices, and more.

```go
searchResults, err := client.Search(ctx, "NVDA", 50, false)
```

**Parameters:**
- `ctx`: Context for cancellation
- `query`: Search query string (e.g., "NVDA", "Apple", "Technology")
- `limit`: Maximum number of results to return (default: 50)
- `includeMarketData`: Whether to include market data in results

**Returns:** `SearchResponse` with array of `SearchSuggestion` objects

**Example:**
```go
results, err := client.Search(ctx, "NVDA", 50, false)
if err != nil {
    log.Fatal(err)
}

for _, result := range results.Data {
    fmt.Printf("Symbol: %s\n", result.Symbol)
    fmt.Printf("Name: %s\n", result.Name)
    fmt.Printf("Type: %s\n", result.Type)
    fmt.Printf("Description: %s\n\n", result.Description)
}
```

## Configuration Options

```go
client := nasdaq.NewClient(
    nasdaq.WithHTTPClient(customHTTPClient),
    nasdaq.WithRateLimit(2),
    nasdaq.WithMaxRetries(5),
    nasdaq.WithRetryDelay(3 * time.Second),
    nasdaq.WithWatchlistType("Rv"),
    nasdaq.WithUserAgents([]string{
        "Custom User Agent 1",
        "Custom User Agent 2",
    }),
)
```

## Best Practices to Avoid Bans

### 1. Use Conservative Rate Limiting
```go
// For production use
client := nasdaq.NewClient(
    nasdaq.WithRateLimit(1),  // 1 request per second
)
```

### 2. Add Delays Between Batches
```go
for _, batch := range symbolBatches {
    data, err := client.GetWatchlist(ctx, batch)
    if err != nil {
        log.Printf("Error: %v", err)
    }
    
    // Wait 5 seconds between batches
    time.Sleep(5 * time.Second)
}
```

### 3. Handle Errors Gracefully
```go
for i := 0; i < len(symbols); i++ {
    quote, err := client.GetQuote(ctx, symbols[i], nasdaq.SymbolTypeStock)
    if err != nil {
        log.Printf("Error getting %s: %v", symbols[i], err)
        // Continue with next symbol
        continue
    }
    
    // Process quote...
}
```

### 4. Use Context for Timeouts
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

quote, err := client.GetQuote(ctx, "AAPL", nasdaq.SymbolTypeStock)
```

## Data Structures

### QuoteRow
```go
type QuoteRow struct {
    Symbol           string // Ticker symbol
    Name             string // Company name
    LastSalePrice    string // Current price
    NetChange        string // Price change
    PercentageChange string // Percent change
    Volume           string // Trading volume
    MarketCap        string // Market capitalization
    Country          string // Country
    IPOYear          string // Year of IPO
    Industry         string // Industry sector
    Sector           string // Market sector
}
```

### ScreenerRow
Similar to QuoteRow with additional screener-specific fields.

### NewsArticle
```go
type NewsArticle struct {
    Title       string    // Article title
    URL         string    // Article URL
    ImageURL    string    // Article image URL
    Source      string    // News source
    PublishTime time.Time // Publication time
    Summary     string    // Article summary
}
```

### SearchSuggestion
```go
type SearchSuggestion struct {
    Symbol      string // Ticker symbol
    Name        string // Company/instrument name
    Type        string // Type (stocks, etf, index, futures, options)
    Description string // Description or sector
}
```

### SearchResponse
```go
type SearchResponse struct {
    Status struct {
        StatusCode int    // HTTP status code
        StatusDesc string // Status description
    }
    Data []SearchSuggestion // Array of search results
}
```

## Error Handling

The library returns errors for various scenarios:

```go
quote, err := client.GetQuote(ctx, "INVALID", nasdaq.SymbolTypeStock)
if err != nil {
    // Handle different error types
    if strings.Contains(err.Error(), "no data found") {
        // Symbol doesn't exist
    } else if strings.Contains(err.Error(), "rate limited") {
        // Too many requests, wait and retry
        time.Sleep(10 * time.Second)
    } else {
        // Other errors
        log.Fatal(err)
    }
}
```

## Examples

See the `examples/` directory for complete working examples:

```bash
cd examples
go run main.go
```

## API Endpoints Used

Based on HAR analysis, this library uses the following NASDAQ API endpoints:

- `https://api.nasdaq.com/api/quote/watchlist` - Stock quotes
- `https://api.nasdaq.com/api/quote/basic` - Basic quotes for multiple symbols
- `https://api.nasdaq.com/api/quote/{symbol}/info` - Detailed symbol information
- `https://api.nasdaq.com/api/quote/{symbol}/chart` - Chart data for symbol
- `https://api.nasdaq.com/api/market-info` - General market information
- `https://api.nasdaq.com/api/screener/stocks` - Stock screener
- `https://api.nasdaq.com/api/screener/etf` - ETF screener
- `https://api.nasdaq.com/api/screener/index` - Index screener
- `https://api.nasdaq.com/api/screener/mutualfunds` - Mutual fund screener
- `https://www.nasdaq.com/api/news/topic/latestnews` - Latest news
- `https://www.nasdaq.com/api/ga/trending-symbols` - Trending symbols
- `https://www.nasdaq.com/api/nasdaq-bell-notifications/current-events` - Bell events (IPOs, upgrades, etc.)
- `https://www.nasdaq.com/ai-search/external/content-search-bff/v1/autosuggest` - Symbol search/autosuggest

## Important Notes

⚠️ **Rate Limiting**: The NASDAQ API may rate limit your IP if you make too many requests. Always use rate limiting and retry logic.

⚠️ **Data Format**: All price and numeric values are returned as strings to preserve formatting (commas, decimals, etc.). You'll need to parse them if you need numeric calculations.

⚠️ **Terms of Service**: Ensure your use complies with NASDAQ's Terms of Service. This library is for educational purposes.

⚠️ **No Authentication Required**: The public API endpoints used by this library do not require authentication keys.

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License

MIT License - See LICENSE file for details

## Disclaimer

This library is for educational purposes only. The authors are not responsible for any misuse or violations of NASDAQ's Terms of Service. Always use rate limiting and respect the API's usage policies.
