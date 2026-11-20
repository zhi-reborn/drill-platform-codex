// Package nodereg maintains a Redis-backed registry of backend node health
// snapshots. Each instance reports its JSON snapshot every few seconds; the
// registry pairs snapshots with Redis-clock heartbeats so offline detection is
// immune to node clock skew (same approach as the presence store).
package nodereg

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// key is the hash storing instanceID -> snapshot JSON plus per-node heartbeats.
const key = "drill:health:nodes"

const (
	// OfflineAfter is the heartbeat age after which a node counts as offline.
	OfflineAfter = 30 * time.Second

	// PruneAfter is the heartbeat age after which a crashed node's entry is
	// removed from the registry entirely.
	PruneAfter = time.Hour

	keyTTL = 24 * time.Hour
)

var ErrUnavailable = errors.New("node registry unavailable")

// reportScript stores the snapshot plus a Redis-clock heartbeat atomically.
// Redis TIME supplies one clock for all nodes, so offline detection stays
// correct even when node clocks drift.
const reportScript = `
local clock = redis.call('TIME')
local now = tonumber(clock[1])
redis.call('HSET', KEYS[1], ARGV[1], ARGV[2], ARGV[1] .. ':hb', now)
redis.call('EXPIRE', KEYS[1], ARGV[3])
return now
`

// Reporter publishes one instance's health snapshot into the registry.
type Reporter struct {
	client     redis.UniversalClient
	instanceID string
}

func NewReporter(client redis.UniversalClient, instanceID string) *Reporter {
	return &Reporter{client: client, instanceID: instanceID}
}

// Report stores the snapshot JSON with a fresh Redis-clock heartbeat.
func (r *Reporter) Report(ctx context.Context, snapshot []byte) error {
	if r == nil || r.client == nil || r.instanceID == "" {
		return ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	return r.client.Eval(ctx, reportScript, []string{key},
		r.instanceID, string(snapshot), strconv.Itoa(int(keyTTL/time.Second))).Err()
}

// Deregister removes the instance from the registry during graceful shutdown.
func (r *Reporter) Deregister(ctx context.Context) error {
	if r == nil || r.client == nil || r.instanceID == "" {
		return ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	return r.client.HDel(ctx, key, r.instanceID, r.instanceID+":hb").Err()
}

// Record is one registry entry: the raw snapshot JSON plus the Redis-clock
// heartbeat used for offline detection.
type Record struct {
	InstanceID  string
	Data        []byte
	HeartbeatAt int64 // unix seconds on the Redis clock
}

// Registry reads and prunes the node registry.
type Registry struct {
	client redis.UniversalClient
}

func NewRegistry(client redis.UniversalClient) *Registry {
	return &Registry{client: client}
}

// List returns all registry records plus the Redis clock reading they should
// be compared against. Entries whose heartbeat is older than PruneAfter (left
// behind by crashed nodes) are removed on the way out.
func (g *Registry) List(ctx context.Context) ([]Record, int64, error) {
	if g == nil || g.client == nil {
		return nil, 0, ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	now, err := g.client.Time(ctx).Result()
	if err != nil {
		return nil, 0, err
	}
	all, err := g.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, 0, err
	}

	snapshots := make(map[string]string)
	heartbeats := make(map[string]int64)
	for field, value := range all {
		if id, ok := strings.CutSuffix(field, ":hb"); ok {
			hb, _ := strconv.ParseInt(value, 10, 64)
			heartbeats[id] = hb
		} else {
			snapshots[field] = value
		}
	}

	records := make([]Record, 0, len(snapshots))
	var stale []string
	for id, data := range snapshots {
		hb, ok := heartbeats[id]
		if !ok || now.Unix()-hb > int64(PruneAfter/time.Second) {
			stale = append(stale, id, id+":hb")
			continue
		}
		records = append(records, Record{InstanceID: id, Data: []byte(data), HeartbeatAt: hb})
	}
	if len(stale) > 0 {
		// 独立超时：调用方 ctx 可能已接近 deadline
		pctx, pcancel := context.WithTimeout(context.Background(), time.Second)
		g.client.HDel(pctx, key, stale...)
		pcancel()
	}
	return records, now.Unix(), nil
}
