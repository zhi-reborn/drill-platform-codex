package redis

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

var errEscapedLeaseScript = errors.New("lease script contains escaped quotes")

type fakeLeaseStore struct {
	mu     sync.Mutex
	values map[string]interface{}
}

var _ LeaseStore = (*fakeLeaseStore)(nil)

func newTestLease(store LeaseStore, key, workerID string, ttl time.Duration) *Lease {
	return NewLease(store, key, workerID, ttl)
}

func newFakeLeaseStore() *fakeLeaseStore {
	return &fakeLeaseStore{values: make(map[string]interface{})}
}

func (f *fakeLeaseStore) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, ok := f.values[key]; ok {
		return false, nil
	}
	f.values[key] = value
	return true, nil
}

func (f *fakeLeaseStore) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	if strings.Contains(script, `\"`) {
		return nil, errEscapedLeaseScript
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	current, ok := f.values[keys[0]]
	if !ok || current != args[0] {
		return int64(0), nil
	}

	if len(args) == 1 {
		delete(f.values, keys[0])
	}
	return int64(1), nil
}

func TestLeaseUsesTokenFenceForRenewAndRelease(t *testing.T) {
	ctx := context.Background()
	store := newFakeLeaseStore()
	first := newTestLease(store, "lease:worker", "worker-1", time.Minute)
	second := newTestLease(store, "lease:worker", "worker-2", time.Minute)

	acquired, err := first.Acquire(ctx)
	if err != nil {
		t.Fatalf("first acquire returned error: %v", err)
	}
	if !acquired {
		t.Fatal("first acquire = false, want true")
	}

	renewed, err := second.Renew(ctx)
	if err != nil {
		t.Fatalf("second renew returned error: %v", err)
	}
	if renewed {
		t.Fatal("second renew = true, want false")
	}

	released, err := second.Release(ctx)
	if err != nil {
		t.Fatalf("second release returned error: %v", err)
	}
	if released {
		t.Fatal("second release = true, want false")
	}

	renewed, err = first.Renew(ctx)
	if err != nil {
		t.Fatalf("first renew returned error: %v", err)
	}
	if !renewed {
		t.Fatal("first renew = false, want true")
	}

	released, err = first.Release(ctx)
	if err != nil {
		t.Fatalf("first release returned error: %v", err)
	}
	if !released {
		t.Fatal("first release = false, want true")
	}
}
