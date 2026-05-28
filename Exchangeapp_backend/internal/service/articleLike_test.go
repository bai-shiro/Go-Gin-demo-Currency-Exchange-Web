package service

import (
	"os"
	"testing"

	"github.com/go-redis/redis"
)

func newTestRedis(t *testing.T) *redis.Client {
	t.Helper()

	addr := os.Getenv("TEST_REDIS_ADDR")
	if addr == "" {
		addr = "127.0.0.1:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   15,
	})

	if err := client.Ping().Err(); err != nil {
		t.Skipf("skip redis integration test: %v", err)
	}

	if err := client.FlushDB().Err(); err != nil {
		t.Fatalf("FlushDB() error = %v", err)
	}

	t.Cleanup(func() {
		_ = client.FlushDB().Err()
		_ = client.Close()
	})

	return client
}

func TestArticleServiceLikeToggle(t *testing.T) {
	rdb := newTestRedis(t)

	svc := NewArticleService(nil, rdb)

	liked, likes, err := svc.Like("1", 1001)
	if err != nil {
		t.Fatalf("first Like() error = %v", err)
	}
	if !liked || likes != 1 {
		t.Fatalf("first Like() = liked %v, likes %d; want true, 1", liked, likes)
	}

	liked, likes, err = svc.Like("1", 1001)
	if err != nil {
		t.Fatalf("second Like() error = %v", err)
	}
	if liked || likes != 0 {
		t.Fatalf("second Like() = liked %v, likes %d; want false, 0", liked, likes)
	}

	liked, likes, err = svc.Like("1", 1001)
	if err != nil {
		t.Fatalf("third Like() error = %v", err)
	}
	if !liked || likes != 1 {
		t.Fatalf("third Like() = liked %v, likes %d; want true, 1", liked, likes)
	}
}
func TestArticleServiceLikeToggleMultipleUsers(t *testing.T) {
	rdb := newTestRedis(t)
	svc := NewArticleService(nil, rdb)

	liked, likes, err := svc.Like("2", 1001)
	if err != nil {
		t.Fatalf("user 1001 Like() error = %v", err)
	}
	if !liked || likes != 1 {
		t.Fatalf("user 1001 Like() = liked %v, likes %d; want true, 1", liked, likes)
	}

	liked, likes, err = svc.Like("2", 1002)
	if err != nil {
		t.Fatalf("user 1002 Like() error = %v", err)
	}
	if !liked || likes != 2 {
		t.Fatalf("user 1002 Like() = liked %v, likes %d; want true, 2", liked, likes)
	}

	liked, likes, err = svc.Like("2", 1001)
	if err != nil {
		t.Fatalf("user 1001 unlike Like() error = %v", err)
	}
	if liked || likes != 1 {
		t.Fatalf("user 1001 unlike Like() = liked %v, likes %d; want false, 1", liked, likes)
	}

	currentLikes, err := svc.GetLikes("2")
	if err != nil {
		t.Fatalf("GetLikes() error = %v", err)
	}
	if currentLikes != 1 {
		t.Fatalf("GetLikes() = %d, want 1", currentLikes)
	}
}
