# Testing Guide

## Overview

This document explains how to run and understand the tests for the XGG NASDAQ Go Library.

## Running Tests

### Run All Tests
```bash
go test ./...
```

### Run Tests in a Specific Package
```bash
# Run tests for the nasdaq package
go test ./nasdaq

# Run tests for examples
go test ./examples
```

### Run Tests with Verbose Output
```bash
go test -v ./...
```

### Run Specific Test
```bash
# Run a specific test function
go test -v ./nasdaq -run TestNewClient

# Run tests matching a pattern
go test -v ./nasdaq -run TestRate
```

### Run Tests with Coverage
```bash
# Run tests with coverage report
go test -cover ./...

# Run tests with coverage for specific package
go test -cover ./nasdaq

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Run Benchmarks
```bash
# Run all benchmarks
go test -bench=. ./...

# Run specific benchmark
go test -bench=BenchmarkRateLimiter ./nasdaq

# Run benchmarks with memory allocation stats
go test -bench=. -benchmem ./...
```

### Run Tests with Race Detection
```bash
go test -race ./...
```

## Test Structure

### Client Tests (`nasdaq/client_test.go`)

These tests focus on the HTTP client, rate limiting, and retry logic.

| Test | Description |
|------|-------------|
| `TestNewClient` | Tests client initialization |
| `TestNewClientWithOptions` | Tests client with custom options |
| `TestUserAgentRotation` | Verifies user agent rotation |
| `TestSymbolWithOptionString` | Tests symbol formatting |
| `TestRateLimiter` | Tests rate limiter functionality |
| `TestRateLimiterRespectsRate` | Verifies rate limit enforcement |
| `TestRateLimiterContextCancellation` | Tests context cancellation |
| `TestDoRequestWithMockServer` | Tests HTTP requests |
| `TestDoRequestWithRateLimiting` | Tests retry on rate limiting |
| `TestDoRequestContextTimeout` | Tests timeout handling |
| `TestMakeAPIRequest` | Tests API request construction |

### API Tests (`nasdaq/api_test.go`)

These tests focus on the API methods and data structures.

| Test | Description |
|------|-------------|
| `TestNewSymbolWithOption` | Tests symbol creation |
| `TestGetWatchlistWithMockServer` | Tests watchlist API |
| `TestGetQuote` | Tests quote API |
| `TestGetScreenerStocks` | Tests stock screener |
| `TestGetScreenerETFs` | Tests ETF screener |
| `TestGetScreenerIndices` | Tests index screener |
| `TestGetScreenerMutualFunds` | Tests mutual fund screener |
| `TestGetNews` | Tests news API |
| `TestGetTrendingSymbols` | Tests trending symbols API |
| `TestParseJSON` | Tests JSON parsing |
| `TestGetWatchlistEmptySymbols` | Tests error handling |
| `TestAssetClassConstants` | Tests asset class constants |
| `TestSymbolTypeConstants` | Tests symbol type constants |

## Benchmarks

### `BenchmarkUserAgentRotation`
Measures the performance of user agent rotation.

### `BenchmarkRateLimiter`
Measures the performance of the rate limiter at high request rates.

## Writing New Tests

### Test Structure

```go
func TestYourFunction(t *testing.T) {
    // Arrange
    client := NewClient()
    ctx := context.Background()
    
    // Act
    result, err := client.YourFunction(ctx, params)
    
    // Assert
    if err != nil {
        t.Fatalf("Unexpected error: %v", err)
    }
    
    if result.ExpectedField != "expectedValue" {
        t.Errorf("ExpectedField = %v, want %v", result.ExpectedField, "expectedValue")
    }
}
```

### Using Mock Servers

```go
func TestWithMock(t *testing.T) {
    // Create mock server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Verify request
        if r.URL.Path != "/expected/path" {
            t.Errorf("Wrong path: %s", r.URL.Path)
        }
        
        // Return mock response
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status": "ok"}`))
    }))
    defer server.Close()
    
    // Test with mock server
    client := NewClient()
    // ... run tests
}
```

### Table-Driven Tests

```go
func TestMultipleCases(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"Case 1", "input1", "output1"},
        {"Case 2", "input2", "output2"},
        {"Case 3", "input3", "output3"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := YourFunction(tt.input)
            if result != tt.expected {
                t.Errorf("Result = %v, want %v", result, tt.expected)
            }
        })
    }
}
```

## Test Coverage

To view detailed coverage information:

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage in browser
go tool cover -html=coverage.out

# View coverage percentage
go tool cover -func=coverage.out
```

## Continuous Integration

The tests can be run in CI/CD pipelines:

```yaml
# Example GitHub Actions workflow
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: '1.21'
      - run: go test -v -race -cover ./...
```

## Troubleshooting

### Tests Failing with Rate Limiting

If tests fail due to rate limiting from the actual API:

1. Most tests use mock servers to avoid actual API calls
2. Integration tests can be skipped with a build tag:
   ```go
   //go:build integration
   // +build integration
   
   func TestRealAPI(t *testing.T) {
       // Test with real API
   }
   ```

### Slow Tests

If tests are slow:

1. Use `-short` flag to skip slow tests:
   ```bash
   go test -short ./...
   ```

2. Mark slow tests:
   ```go
   func TestSlowFunction(t *testing.T) {
       if testing.Short() {
           t.Skip("Skipping slow test in short mode")
       }
       // ... test code
   }
   ```

### Timeouts

Tests may timeout if the rate limiter is too restrictive. Adjust test timeouts:

```go
func TestWithTimeout(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    // Run test with timeout
}
```

## Best Practices

1. **Use Mock Servers**: Avoid making real API calls in tests
2. **Test Error Cases**: Test both success and failure scenarios
3. **Table-Driven Tests**: Use for testing multiple input cases
4. **Context Usage**: Always use context in tests that involve network calls
5. **Cleanup**: Always defer cleanup operations (closing servers, stopping rate limiters)
6. **Descriptive Names**: Use clear, descriptive test names
7. **Isolation**: Each test should be independent

## Integration Tests

For integration tests that call the real NASDAQ API:

1. Use build tags to separate them:
   ```go
   //go:build integration
   ```

2. Run them separately:
   ```bash
   go test -tags=integration ./nasdaq
   ```

3. Use conservative rate limiting in integration tests:
   ```go
   client := NewClient(
       WithRateLimit(1),  // Conservative for integration tests
   )
   ```

## Contributing Tests

When adding new features:

1. Add unit tests for new functions
2. Add tests for error cases
3. Update test coverage
4. Ensure all tests pass: `go test ./...`
5. Run with race detection: `go test -race ./...`

## Resources

- [Go Testing Package](https://golang.org/pkg/testing/)
- [Table-Driven Tests](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
- [httptest Package](https://golang.org/pkg/net/http/httptest/)
- [Context Package](https://golang.org/pkg/context/)