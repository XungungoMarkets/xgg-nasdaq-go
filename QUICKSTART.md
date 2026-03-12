# Quick Start Guide

## Prerequisites

1. Install Go 1.21 or later from https://golang.org/dl/
2. Verify installation: `go version`

## Setup

1. Clone or download this project
2. Navigate to project directory: `cd xgg-nasdaq-go`
3. Initialize module (if needed): `go mod init github.com/XungungoMarkets/xgg-nasdaq-go`

## Run Examples

```bash
cd examples
go run main.go
```

## Create Your First Program

Create `main.go`:

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/XungungoMarkets/xgg-nasdaq-go/nasdaq"
)

func main() {
    // Create client with rate limiting to avoid bans
    client := nasdaq.NewClient(
        nasdaq.WithRateLimit(2),  // 2 requests per second
        nasdaq.WithWatchlistType("Rv"), // Optional watchlist query type from HAR
    )
    defer client.Close()
    
    ctx := context.Background()
    
    // Get Apple stock quote
    quote, err := client.GetQuote(ctx, "AAPL", nasdaq.SymbolTypeStock)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Symbol: %s\n", quote.Symbol)
    fmt.Printf("Name: %s\n", quote.Name)
    fmt.Printf("Price: %s\n", quote.LastSalePrice)
    fmt.Printf("Change: %s\n", quote.NetChange)
}
```

Run it:
```bash
go run main.go
```

## Common Use Cases

### 1. Get Multiple Stocks

```go
symbols := []nasdaq.SymbolWithOption{
    nasdaq.NewSymbolWithOption("AAPL", nasdaq.SymbolTypeStock),
    nasdaq.NewSymbolWithOption("GOOGL", nasdaq.SymbolTypeStock),
    nasdaq.NewSymbolWithOption("MSFT", nasdaq.SymbolTypeStock),
}

watchlist, err := client.GetWatchlist(ctx, symbols)
for _, row := range watchlist.Data {
    fmt.Printf("%s: %s\n", row.Symbol, row.LastSalePrice)
}
```

### 2. Get All Stocks

```go
stocks, err := client.GetScreenerStocks(ctx, false)
fmt.Printf("Total stocks: %d\n", len(stocks.Data.Table.Rows))
```

### 3. Get Latest News

```go
news, err := client.GetNews(ctx, 0, 10, true)
for _, article := range news.Data.NewsArticles {
    fmt.Printf("%s\n", article.Title)
    fmt.Printf("URL: %s\n\n", article.URL)
}
```

### 4. Get Trending Stocks

```go
trending, err := client.GetTrendingSymbols(ctx, nasdaq.AssetClassStock)
for _, sym := range trending.Data {
    fmt.Printf("%s: %s\n", sym.Symbol, sym.LastSalePrice)
}
```

## Avoiding IP Bans

### Conservative Settings

```go
client := nasdaq.NewClient(
    nasdaq.WithRateLimit(1),              // 1 request per second
    nasdaq.WithMaxRetries(5),             // More retries
    nasdaq.WithRetryDelay(3 * time.Second), // Longer delay
)
```

### Add Delays Between Requests

```go
for _, symbol := range symbols {
    quote, err := client.GetQuote(ctx, symbol, nasdaq.SymbolTypeStock)
    // Process quote...
    
    // Wait 2 seconds between requests
    time.Sleep(2 * time.Second)
}
```

## Module Setup for Your Project

If you want to use this in your own project:

1. Copy the `nasdaq/` folder to your project
2. Update your `go.mod`:

```go
module your-project-name

go 1.21
```

3. Import in your code:

```go
import "your-project-name/nasdaq"
```

Or if you publish it:

```bash
go get github.com/XungungoMarkets/xgg-nasdaq-go
```

```go
import "github.com/XungungoMarkets/xgg-nasdaq-go/nasdaq"
```

## Troubleshooting

### "package not found"
- Make sure you're in the correct directory
- Run `go mod tidy` to update dependencies
- Check your `go.mod` file

### Rate Limited Errors
- Reduce `WithRateLimit` value (e.g., 1 instead of 2)
- Increase `WithRetryDelay` value
- Add `time.Sleep()` between batches of requests

### Connection Errors
- Check your internet connection
- Verify firewall settings
- Try with a custom HTTP client if needed

## Next Steps

- Read the full [README.md](README.md) for detailed documentation
- Explore [examples/main.go](examples/main.go) for more examples
- Check the source code in `nasdaq/` directory

## Support

For issues or questions:
1. Check the README.md for detailed API documentation
2. Review the example code
3. Examine error messages for specific issues
