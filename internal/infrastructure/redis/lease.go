package redis

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"
)

type LeaseStore interface {
	SetNX(context.Context, string, interface{}, time.Duration) (bool, error)
	Eval(context.Context, string, []string, ...interface{}) (interface{}, error)
}

type Lease struct {
	store LeaseStore
	key   string
	value string
	ttl   time.Duration
}

func NewLease(store LeaseStore, key, workerID string, ttl time.Duration) *Lease {
	return &Lease{
		store: store,
		key:   key,
		value: fmt.Sprintf("%s/%s", workerID, newLeaseToken()),
		ttl:   ttl,
	}
}

func (l *Lease) Acquire(ctx context.Context) (bool, error) {
	return l.store.SetNX(ctx, l.key, l.value, l.ttl)
}

func (l *Lease) Renew(ctx context.Context) (bool, error) {
	result, err := l.store.Eval(ctx, `if redis.call("GET", KEYS[1]) == ARGV[1] then return redis.call("PEXPIRE", KEYS[1], ARGV[2]) end return 0`, []string{l.key}, l.value, l.ttl.Milliseconds())
	if err != nil {
		return false, err
	}
	return evalBool(result), nil
}

func (l *Lease) Release(ctx context.Context) (bool, error) {
	result, err := l.store.Eval(ctx, `if redis.call("GET", KEYS[1]) == ARGV[1] then return redis.call("DEL", KEYS[1]) end return 0`, []string{l.key}, l.value)
	if err != nil {
		return false, err
	}
	return evalBool(result), nil
}

func (l *Lease) Value() string {
	return l.value
}

func evalBool(result interface{}) bool {
	switch v := result.(type) {
	case int64:
		return v == 1
	case int:
		return v == 1
	case bool:
		return v
	default:
		return false
	}
}

func newLeaseToken() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(fmt.Errorf("generate lease token: %w", err))
	}

	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
