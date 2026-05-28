package exchange

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const DefaultFrankfurterBaseURL = "https://api.frankfurter.dev"

type LatestRate struct {
	Base  string  `json:"base"`
	Quote string  `json:"quote"`
	Rate  float64 `json:"rate"`
	Date  string  `json:"date"`
}

type Client interface {
	FetchLatest(ctx context.Context, base string, quote string) (*LatestRate, error)
}

type FrankfurterClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewFrankfurterClient(baseURL string, timeout time.Duration) *FrankfurterClient {
	if baseURL == "" {
		baseURL = DefaultFrankfurterBaseURL
	}
	if timeout <= 0 {
		timeout = 3 * time.Second
	}

	return &FrankfurterClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *FrankfurterClient) FetchLatest(ctx context.Context, base string, quote string) (*LatestRate, error) {
	base = strings.ToUpper(strings.TrimSpace(base))
	quote = strings.ToUpper(strings.TrimSpace(quote))

	url := fmt.Sprintf("%s/v2/rate/%s/%s", c.baseURL, base, quote)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("frankfurter API returned status %d", resp.StatusCode)
	}

	// 字段检查
	var payload struct {
		Base  string  `json:"base"`
		Quote string  `json:"quote"`
		Rate  float64 `json:"rate"`
		Date  string  `json:"date"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if payload.Rate <= 0 {
		return nil, fmt.Errorf("invalid exchange rate %f", payload.Rate)
	}

	return &LatestRate{
		Base:  strings.ToUpper(payload.Base),
		Quote: strings.ToUpper(payload.Quote),
		Rate:  payload.Rate,
		Date:  payload.Date,
	}, nil
}
