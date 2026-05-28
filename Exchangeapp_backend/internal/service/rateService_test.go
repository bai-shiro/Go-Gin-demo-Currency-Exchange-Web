package service

import (
	"context"
	"errors"
	"exchangeapp/internal/client/exchange"
	"testing"
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

func TestRateServiceLatestPairCacheMissThenHit(t *testing.T) {
	rdb := newTestRedis(t)
	client := &fakeExchangeClient{latest: &exchange.LatestRate{Base: "USD", Quote: "CNY", Rate: 7.25, Date: "2026-05-28"}}
	svc := NewRateService(nil, rdb, client)

	latest, err := svc.LatestPair(context.Background(), "usd", "cny")
	if err != nil {
		t.Fatalf("LatestPair() error = %v", err)
	}
	if latest.Base != "USD" || latest.Quote != "CNY" || latest.Rate != 7.25 {
		t.Fatalf("LatestPair() = %+v, want USD/CNY 7.25", latest)
	}
	if client.calls != 1 {
		t.Fatalf("client calls = %d, want 1", client.calls)
	}

	latest, err = svc.LatestPair(context.Background(), "USD", "CNY")
	if err != nil {
		t.Fatalf("LatestPair() second call error = %v", err)
	}
	if latest.Rate != 7.25 {
		t.Fatalf("LatestPair() cached rate = %f, want 7.25", latest.Rate)
	}
	if client.calls != 1 {
		t.Fatalf("client calls after cache hit = %d, want 1", client.calls)
	}
}

func TestRateServiceLatestPairBadCacheFallsBackToClient(t *testing.T) {
	rdb := newTestRedis(t)
	client := &fakeExchangeClient{latest: &exchange.LatestRate{Base: "EUR", Quote: "USD", Rate: 1.08, Date: "2026-05-28"}}
	svc := NewRateService(nil, rdb, client)

	if err := rdb.Set(latestRateCacheKey("EUR", "USD"), "bad-json", 0).Err(); err != nil {
		t.Fatalf("seed bad cache error = %v", err)
	}

	latest, err := svc.LatestPair(context.Background(), "EUR", "USD")
	if err != nil {
		t.Fatalf("LatestPair() error = %v", err)
	}
	if latest.Rate != 1.08 {
		t.Fatalf("LatestPair() rate = %f, want 1.08", latest.Rate)
	}
	if client.calls != 1 {
		t.Fatalf("client calls = %d, want 1", client.calls)
	}
}

func TestRateServiceConvert(t *testing.T) {
	rdb := newTestRedis(t)
	client := &fakeExchangeClient{latest: &exchange.LatestRate{Base: "USD", Quote: "JPY", Rate: 150, Date: "2026-05-28"}}
	svc := NewRateService(nil, rdb, client)

	converted, err := svc.Convert(context.Background(), "USD", "JPY", 2)
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}
	if converted.ConvertedAmount != 300 {
		t.Fatalf("ConvertedAmount = %f, want 300", converted.ConvertedAmount)
	}
}

func TestRateServiceConvertRejectsInvalidAmount(t *testing.T) {
	rdb := newTestRedis(t)
	client := &fakeExchangeClient{latest: &exchange.LatestRate{Base: "USD", Quote: "JPY", Rate: 150, Date: "2026-05-28"}}
	svc := NewRateService(nil, rdb, client)

	_, err := svc.Convert(context.Background(), "USD", "JPY", 0)
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
