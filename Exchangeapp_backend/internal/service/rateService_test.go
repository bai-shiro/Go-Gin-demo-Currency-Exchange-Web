package service

import (
	"context"
	"errors"
	"exchangeapp/internal/apperrors"
	"exchangeapp/internal/client/exchange"
	"exchangeapp/internal/models"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

type fakeExchangeClient struct {
	latest *exchange.LatestRate
	err    error
	calls  int
}

func (f *fakeExchangeClient) FetchLatest(ctx context.Context, base string, quote string) (*exchange.LatestRate, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	return f.latest, nil
}

type fakeRateStore struct {
	history          []models.ExchangeRate
	historyErr       error
	historyCalls     int
	historyFrom      string
	historyTo        string
	historyStartDate time.Time
	historyEndDate   time.Time
}

func (f *fakeRateStore) Create(exchangeRate *models.ExchangeRate) error {
	return nil
}

func (f *fakeRateStore) Upsert(exchangeRate *models.ExchangeRate) error {
	return nil
}

func (f *fakeRateStore) Latest() ([]models.ExchangeRate, error) {
	return nil, nil
}

func (f *fakeRateStore) FindHistory(fromCurrency string, toCurrency string, startDate time.Time, endDate time.Time) ([]models.ExchangeRate, error) {
	f.historyCalls++
	f.historyFrom = fromCurrency
	f.historyTo = toCurrency
	f.historyStartDate = startDate
	f.historyEndDate = endDate
	return f.history, f.historyErr
}

func TestRateServiceLatestPairCacheMissThenHit(t *testing.T) {
	rdb := newTestRedis(t)
	client := &fakeExchangeClient{latest: &exchange.LatestRate{Base: "USD", Quote: "CNY", Rate: decimal.RequireFromString("7.25"), Date: "2026-05-28"}}
	svc := NewRateService(nil, rdb, client)

	latest, err := svc.LatestPair(context.Background(), "usd", "cny")
	if err != nil {
		t.Fatalf("LatestPair() error = %v", err)
	}
	if latest.Base != "USD" || latest.Quote != "CNY" || !latest.Rate.Equal(decimal.RequireFromString("7.25")) {
		t.Fatalf("LatestPair() = %+v, want USD/CNY 7.25", latest)
	}
	if client.calls != 1 {
		t.Fatalf("client calls = %d, want 1", client.calls)
	}

	latest, err = svc.LatestPair(context.Background(), "USD", "CNY")
	if err != nil {
		t.Fatalf("LatestPair() second call error = %v", err)
	}
	if !latest.Rate.Equal(decimal.RequireFromString("7.25")) {
		t.Fatalf("LatestPair() cached rate = %v, want 7.25", latest.Rate)
	}
	if client.calls != 1 {
		t.Fatalf("client calls after cache hit = %d, want 1", client.calls)
	}
}

func TestRateServiceLatestPairBadCacheFallsBackToClient(t *testing.T) {
	rdb := newTestRedis(t)
	client := &fakeExchangeClient{latest: &exchange.LatestRate{Base: "EUR", Quote: "USD", Rate: decimal.RequireFromString("1.08"), Date: "2026-05-28"}}
	svc := NewRateService(nil, rdb, client)

	if err := rdb.Set(latestRateCacheKey("EUR", "USD"), "bad-json", 0).Err(); err != nil {
		t.Fatalf("seed bad cache error = %v", err)
	}

	latest, err := svc.LatestPair(context.Background(), "EUR", "USD")
	if err != nil {
		t.Fatalf("LatestPair() error = %v", err)
	}
	if !latest.Rate.Equal(decimal.RequireFromString("1.08")) {
		t.Fatalf("LatestPair() rate = %v, want 1.08", latest.Rate)
	}
	if client.calls != 1 {
		t.Fatalf("client calls = %d, want 1", client.calls)
	}
}

func TestRateServiceConvert(t *testing.T) {
	rdb := newTestRedis(t)
	client := &fakeExchangeClient{latest: &exchange.LatestRate{Base: "USD", Quote: "JPY", Rate: decimal.RequireFromString("150"), Date: "2026-05-28"}}
	svc := NewRateService(nil, rdb, client)

	converted, err := svc.Convert(context.Background(), "USD", "JPY", decimal.RequireFromString("2"))
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}
	if !converted.ConvertedAmount.Equal(decimal.RequireFromString("300")) {
		t.Fatalf("ConvertedAmount = %v, want 300", converted.ConvertedAmount)
	}
}

func TestRateServiceConvertRejectsInvalidAmount(t *testing.T) {
	rdb := newTestRedis(t)
	client := &fakeExchangeClient{latest: &exchange.LatestRate{Base: "USD", Quote: "JPY", Rate: decimal.RequireFromString("150"), Date: "2026-05-28"}}
	svc := NewRateService(nil, rdb, client)

	_, err := svc.Convert(context.Background(), "USD", "JPY", decimal.Zero)
	if err == nil {
		t.Fatal("Convert() expected error for zero amount, got nil")
	}
}

func TestRateServiceLatestPairReturnsClientError(t *testing.T) {
	rdb := newTestRedis(t)
	client := &fakeExchangeClient{err: errors.New("upstream unavailable")}
	svc := NewRateService(nil, rdb, client)

	_, err := svc.LatestPair(context.Background(), "USD", "CNY")
	if err == nil {
		t.Fatal("LatestPair() expected client error, got nil")
	}
}

func TestRateServiceHistory(t *testing.T) {
	repo := &fakeRateStore{
		history: []models.ExchangeRate{
			{
				FromCurrency: "USD",
				ToCurrency:   "CNY",
				Rate:         decimal.RequireFromString("7.2500000000"),
				RateDate:     mustParseDate(t, "2026-05-01"),
			},
			{
				FromCurrency: "USD",
				ToCurrency:   "CNY",
				Rate:         decimal.RequireFromString("7.2600000000"),
				RateDate:     mustParseDate(t, "2026-05-02"),
			},
		},
	}
	svc := NewRateService(repo, nil, nil)

	rates, err := svc.History("usd", "cny", "2026-05-01", "2026-05-31")
	if err != nil {
		t.Fatalf("History() error = %v", err)
	}
	if repo.historyCalls != 1 {
		t.Fatalf("FindHistory calls = %d, want 1", repo.historyCalls)
	}
	if repo.historyFrom != "USD" || repo.historyTo != "CNY" {
		t.Fatalf("FindHistory pair = %s/%s, want USD/CNY", repo.historyFrom, repo.historyTo)
	}
	if !sameDate(repo.historyStartDate, mustParseDate(t, "2026-05-01")) {
		t.Fatalf("FindHistory startDate = %v, want 2026-05-01", repo.historyStartDate)
	}
	if !sameDate(repo.historyEndDate, mustParseDate(t, "2026-05-31")) {
		t.Fatalf("FindHistory endDate = %v, want 2026-05-31", repo.historyEndDate)
	}
	if len(rates) != 2 {
		t.Fatalf("History() len = %d, want 2", len(rates))
	}
	if !rates[0].Rate.Equal(decimal.RequireFromString("7.2500000000")) {
		t.Fatalf("History()[0].Rate = %v, want 7.2500000000", rates[0].Rate)
	}
}

func TestRateServiceHistoryRejectsInvalidParams(t *testing.T) {
	tests := []struct {
		name  string
		base  string
		quote string
		start string
		end   string
	}{
		{
			name:  "invalid currency pair",
			base:  "US",
			quote: "CNY",
			start: "2026-05-01",
			end:   "2026-05-31",
		},
		{
			name:  "invalid start date",
			base:  "USD",
			quote: "CNY",
			start: "2026/05/01",
			end:   "2026-05-31",
		},
		{
			name:  "invalid end date",
			base:  "USD",
			quote: "CNY",
			start: "2026-05-01",
			end:   "2026/05/31",
		},
		{
			name:  "start after end",
			base:  "USD",
			quote: "CNY",
			start: "2026-06-01",
			end:   "2026-05-31",
		},
		{
			name:  "range over one year",
			base:  "USD",
			quote: "CNY",
			start: "2025-01-01",
			end:   "2026-01-02",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRateStore{}
			svc := NewRateService(repo, nil, nil)

			_, err := svc.History(tt.base, tt.quote, tt.start, tt.end)
			if err != apperrors.ErrInvalidParams {
				t.Fatalf("History() error = %v, want ErrInvalidParams", err)
			}
			if repo.historyCalls != 0 {
				t.Fatalf("FindHistory calls = %d, want 0", repo.historyCalls)
			}
		})
	}
}

func mustParseDate(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		t.Fatalf("parse date %q error = %v", value, err)
	}
	return parsed
}

func sameDate(a time.Time, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}
