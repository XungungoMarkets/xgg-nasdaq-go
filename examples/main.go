package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

func main() {
	httpClient := &http.Client{Timeout: 6 * time.Second}
	endpoints := []struct {
		name string
		url  string
	}{
		{"Watchlist", buildWatchlistURL()},
		{"Screener Stocks", "https://api.nasdaq.com/api/screener/stocks?tableonly=false"},
		{"Screener ETF", "https://api.nasdaq.com/api/screener/etf?tableonly=false"},
		{"News", "https://www.nasdaq.com/api/news/topic/latestnews?offset=0&limit=5&blacklist=true"},
		{"Trending", "https://www.nasdaq.com/api/ga/trending-symbols?assetclass=stocks"},
		{"Bell Notifications", "https://www.nasdaq.com/api/nasdaq-bell-notifications/current-events"},
		{"Market Info", "https://api.nasdaq.com/api/market-info"},
		{"Symbol Info (NDX)", "https://api.nasdaq.com/api/quote/NDX/info?assetclass=index"},
		{"Basic Quotes", buildBasicURL()},
		{"Chart (AAPL)", "https://api.nasdaq.com/api/quote/AAPL/chart?assetclass=stocks"},
	}

	for _, ep := range endpoints {
		ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
		body, err := fetchJSON(ctx, httpClient, ep.url)
		cancel()
		if err != nil {
			log.Printf("[%s] ERROR: %v", ep.name, err)
			continue
		}
		log.Printf("[%s] OK status=%s keys=%v sample=%s", ep.name, statusSummary(body), topKeys(body), sampleResponse(body))
	}
}

func buildWatchlistURL() string {
	q := url.Values{}
	q.Add("symbol", "aapl|stocks")
	q.Add("symbol", "msft|stocks")
	q.Add("type", "Rv")
	return "https://api.nasdaq.com/api/quote/watchlist?" + q.Encode()
}

func buildBasicURL() string {
	q := url.Values{}
	q.Add("symbol", "aapl|stocks")
	q.Add("symbol", "msft|stocks")
	return "https://api.nasdaq.com/api/quote/basic?" + q.Encode()
}

func fetchJSON(ctx context.Context, httpClient *http.Client, endpoint string) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", "https://www.nasdaq.com/")
	req.Header.Set("Origin", "https://www.nasdaq.com")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var out map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

func statusSummary(m map[string]interface{}) string {
	raw, ok := m["status"].(map[string]interface{})
	if !ok {
		return "<no-status>"
	}
	if v, ok := raw["rCode"]; ok {
		return fmt.Sprintf("rCode=%v", v)
	}
	if v, ok := raw["code"]; ok {
		return fmt.Sprintf("code=%v", v)
	}
	return "<unknown-status-shape>"
}

func topKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func sampleResponse(m map[string]interface{}) string {
	if events, ok := m["events"].([]interface{}); ok {
		if len(events) == 0 {
			return "events=[]"
		}
		return compact(events[0], 220)
	}

	data, ok := m["data"]
	if !ok {
		return "<no-data>"
	}

	if rows, ok := nestedArray(data, "rows"); ok && len(rows) > 0 {
		return compact(rows[0], 220)
	}
	if records, ok := nestedArray(data, "records"); ok && len(records) > 0 {
		return compact(records[0], 220)
	}
	if arr, ok := data.([]interface{}); ok {
		if len(arr) == 0 {
			return "data=[]"
		}
		return compact(arr[0], 220)
	}
	return compact(data, 220)
}

func nestedArray(v interface{}, key string) ([]interface{}, bool) {
	obj, ok := v.(map[string]interface{})
	if !ok {
		return nil, false
	}
	arr, ok := obj[key].([]interface{})
	return arr, ok
}

func compact(v interface{}, max int) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "<marshal-error>"
	}
	s := string(b)
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
