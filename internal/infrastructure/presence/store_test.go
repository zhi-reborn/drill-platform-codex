package presence

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestStoreUnavailable(t *testing.T) {
	s := NewStore(nil)
	if err := s.Touch(context.Background(), 1); err == nil {
		t.Fatal("unconfigured store should not report a successful touch")
	}
	if ids, err := s.OnlineIDs(context.Background()); err == nil || ids != nil {
		t.Fatalf("unconfigured store = %v, %v", ids, err)
	}
}

func TestStoreRealRedis(t *testing.T) {
	addr := os.Getenv("PRESENCE_TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("set PRESENCE_TEST_REDIS_ADDR to exercise Lua on Redis")
	}
	ctx := context.Background()
	rc := redis.NewClient(&redis.Options{Addr: addr, ContextTimeoutEnabled: true})
	t.Cleanup(func() { _ = rc.Close() })
	if err := rc.Ping(ctx).Err(); err != nil {
		t.Fatal(err)
	}
	s := NewStore(rc)
	s.key = fmt.Sprintf("drill:test:presence:%d", time.Now().UnixNano())
	t.Cleanup(func() { _ = rc.Del(ctx, s.key).Err() })
	other := NewStore(rc)
	other.key = s.key
	for _, store := range []*Store{s, other, s} {
		if err := store.Touch(ctx, 42); err != nil {
			t.Fatal(err)
		}
	}
	if ids, err := other.OnlineIDs(ctx); err != nil || !reflect.DeepEqual(ids, []uint64{42}) {
		t.Fatalf("cross-node dedup = %v, %v", ids, err)
	}
	if ttl := rc.PTTL(ctx, s.key).Val(); ttl <= 0 || ttl > 5*time.Minute {
		t.Fatalf("unexpected TTL %v", ttl)
	}
	// Seed timestamps with Redis's clock, not the test runner's wall clock.
	now := rc.Time(ctx).Val().UnixMilli()
	if err := rc.ZAdd(ctx, s.key,
		redis.Z{Score: float64(now - 300001), Member: "1"},
		redis.Z{Score: float64(now - 300000), Member: "2"},
		redis.Z{Score: float64(now - 290000), Member: "3"},
	).Err(); err != nil {
		t.Fatal(err)
	}
	if ids, err := s.OnlineIDs(ctx); err != nil || !reflect.DeepEqual(ids, []uint64{3, 42}) {
		t.Fatalf("five-minute filtering = %v, %v", ids, err)
	}
	// A record newer than the server clock must not move backwards on touch.
	future := float64(now + 10000)
	if err := rc.ZAdd(ctx, s.key, redis.Z{Score: future, Member: "42"}).Err(); err != nil {
		t.Fatal(err)
	}
	if err := s.Touch(ctx, 42); err != nil {
		t.Fatal(err)
	}
	if got := rc.ZScore(ctx, s.key, "42").Val(); got != future {
		t.Fatalf("score regressed: %v < %v", got, future)
	}
	if err := s.Touch(ctx, 0); err == nil {
		t.Fatal("anonymous user must not be recorded")
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if err := s.Touch(cancelled, 9); err == nil {
		t.Fatal("cancelled request succeeded")
	}
	if err := rc.Set(ctx, s.key, "wrong type", time.Minute).Err(); err != nil {
		t.Fatal(err)
	}
	if _, err := s.OnlineIDs(ctx); err == nil {
		t.Fatal("Redis errors must not become zero online users")
	}
}
