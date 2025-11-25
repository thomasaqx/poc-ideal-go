package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	YahooFinanceAPIURL = "https://yfapi.net"
	DefaultTimeout     = 30 * time.Second
)

type YahooClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewYahooClient create a new instance of Yahoo Finance client
func NewYahooClient(apiKey string) *YahooClient {
	return &YahooClient{
		apiKey:  apiKey,
		baseURL: YahooFinanceAPIURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
}

// doRequest executes an HTTP GET request to the Yahoo Finance API
func (c *YahooClient) doRequest(endpoint string, params url.Values) ([]byte, error) {
	// Build the full URL
	fullURL := fmt.Sprintf("%s%s", c.baseURL, endpoint)

	// Add query parameters
	if len(params) > 0 {
		fullURL = fmt.Sprintf("%s?%s", fullURL, params.Encode())
	}

	// Create the request
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Add authentication headers
	req.Header.Add("X-API-KEY", c.apiKey)
	req.Header.Add("Accept", "application/json")

	// Execute the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// GetQuote fetches information for a specific stock
func (c *YahooClient) GetQuote(symbol string) (*QuoteResponse, error) {
	params := url.Values{}
	params.Add("symbols", symbol)

	body, err := c.doRequest("/v6/finance/quote", params)
	if err != nil {
		return nil, err
	}

	var response QuoteResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	return &response, nil
}

// GetQuotes fetches information for multiple stocks
func (c *YahooClient) GetQuotes(symbols []string) (*QuoteResponse, error) {
	params := url.Values{}

	symbolsStr := ""
	for i, symbol := range symbols {
		if i > 0 {
			symbolsStr += ","
		}
		symbolsStr += symbol
	}
	params.Add("symbols", symbolsStr)

	body, err := c.doRequest("/v6/finance/quote", params)
	if err != nil {
		return nil, err
	}

	var response QuoteResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	return &response, nil
}

// GetHistoricalData fetches historical data for a stock
func (c *YahooClient) GetHistoricalData(symbol string, period1, period2 int64, interval string) (*HistoricalDataResponse, error) {
	params := url.Values{}
	params.Add("period1", fmt.Sprintf("%d", period1))
	params.Add("period2", fmt.Sprintf("%d", period2))
	params.Add("interval", interval)

	endpoint := fmt.Sprintf("/v8/finance/chart/%s", symbol)
	body, err := c.doRequest(endpoint, params)
	if err != nil {
		return nil, err
	}

	var response HistoricalDataResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	return &response, nil
}
