package events

import (
	"context"
	"encoding/json"
	"errors"
	"sync/atomic"
	"time"

	infraredis "drill-platform/internal/infrastructure/redis"

	goredis "github.com/redis/go-redis/v9"
)

const EventsChannel = "drill:events"

// dedupCapacity bounds how many recently seen event IDs the subscriber tracks
// for duplicate suppression. When the table fills it is reset wholesale — a
// simple, lock-free tradeoff that bounds memory while tolerating bursts.
const dedupCapacity = 4096

type redisPublisher interface {
	Publish(context.Context, string, interface{}) error
}

type redisSubscriber interface {
	Subscribe(context.Context, ...string) pubSubSession
}

type pubSubSession interface {
	Receive(context.Context) (interface{}, error)
	Channel() <-chan *goredis.Message
	Close() error
}

type redisClient interface {
	PublishContext(context.Context, string, interface{}) error
	SubscribeContext(context.Context, ...string) *goredis.PubSub
}

type RedisBus struct {
	publisher  redisPublisher
	subscriber redisSubscriber
	healthy    atomic.Bool
	backoff    time.Duration
	maxBackoff time.Duration

	// dedup state — touched only by the Subscribe goroutine, so no mutex.
	seen    map[string]struct{}
	seenCnt int
}

func NewRedisBus(client interface{}) *RedisBus {
	bus := &RedisBus{
		backoff:    100 * time.Millisecond,
		maxBackoff: 5 * time.Second,
		seen:       map[string]struct{}{},
	}

	switch c := client.(type) {
	case *infraredis.Client:
		wrapped := redisClientAdapter{client: c}
		bus.publisher = wrapped
		bus.subscriber = wrapped
	case redisClient:
		wrapped := redisClientAdapter{client: c}
		bus.publisher = wrapped
		bus.subscriber = wrapped
	case *goredis.Client:
		raw := rawRedisClient{client: c}
		bus.publisher = raw
		bus.subscriber = raw
	case redisPublisher:
		bus.publisher = c
		if subscriber, ok := client.(redisSubscriber); ok {
			bus.subscriber = subscriber
		}
	}

	return bus
}

func (b *RedisBus) Publish(ctx context.Context, msg WSMessage) error {
	if b.publisher == nil {
		return errors.New("redis publisher is not configured")
	}

	if msg.Version == 0 {
		msg.Version = CurrentVersion
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return b.publisher.Publish(ctx, EventsChannel, string(data))
}

func (b *RedisBus) Subscribe(ctx context.Context, handler func(WSMessage)) error {
	if b.subscriber == nil {
		return errors.New("redis subscriber is not configured")
	}

	backoff := b.backoff
	if backoff <= 0 {
		backoff = 100 * time.Millisecond
	}
	maxBackoff := b.maxBackoff
	if maxBackoff <= 0 {
		maxBackoff = 5 * time.Second
	}

	for {
		if err := ctx.Err(); err != nil {
			b.healthy.Store(false)
			return err
		}

		pubsub := b.subscriber.Subscribe(ctx, EventsChannel)
		if _, err := pubsub.Receive(ctx); err != nil {
			b.healthy.Store(false)
			pubsub.Close()
			if err := sleepWithContext(ctx, backoff); err != nil {
				return err
			}
			backoff = growBackoff(backoff, maxBackoff)
			continue
		}

		b.healthy.Store(true)
		backoff = b.backoff // reset backoff after a successful connect

		ch := pubsub.Channel()
		for {
			select {
			case <-ctx.Done():
				b.healthy.Store(false)
				pubsub.Close()
				return ctx.Err()
			case msg, ok := <-ch:
				if !ok {
					// Subscription channel closed — mark unready immediately so
					// /ready drains traffic while we reconnect.
					b.healthy.Store(false)
					pubsub.Close()
					if err := sleepWithContext(ctx, backoff); err != nil {
						return err
					}
					backoff = growBackoff(backoff, maxBackoff)
					goto reconnect
				}
				b.handleMessage(msg.Payload, handler)
			}
		}

	reconnect:
	}
}

func (b *RedisBus) Healthy() bool {
	return b.healthy.Load()
}

func (b *RedisBus) handleMessage(payload string, handler func(WSMessage)) {
	var msg WSMessage
	if err := json.Unmarshal([]byte(payload), &msg); err != nil {
		return
	}
	if msg.Version != 0 && msg.Version != CurrentVersion {
		return
	}
	if !b.markSeen(msg.ID) {
		return
	}
	handler(msg)
}

// markSeen records the event ID and returns true if this is the first time
// we've seen it. Duplicate IDs return false so the handler is not re-invoked.
// When the table reaches dedupCapacity it resets, bounding memory.
func (b *RedisBus) markSeen(id string) bool {
	if id == "" {
		return true
	}
	if b.seen == nil {
		b.seen = map[string]struct{}{}
	}
	if _, exists := b.seen[id]; exists {
		return false
	}
	if b.seenCnt >= dedupCapacity {
		b.seen = map[string]struct{}{}
		b.seenCnt = 0
	}
	b.seen[id] = struct{}{}
	b.seenCnt++
	return true
}

func growBackoff(current, max time.Duration) time.Duration {
	next := current * 2
	if next > max {
		next = max
	}
	return next
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

type redisClientAdapter struct {
	client redisClient
}

func (r redisClientAdapter) Publish(ctx context.Context, channel string, message interface{}) error {
	return r.client.PublishContext(ctx, channel, message)
}

func (r redisClientAdapter) Subscribe(ctx context.Context, channels ...string) pubSubSession {
	return redisPubSubSession{pubsub: r.client.SubscribeContext(ctx, channels...)}
}

type redisPubSubSession struct {
	pubsub *goredis.PubSub
}

func (s redisPubSubSession) Receive(ctx context.Context) (interface{}, error) {
	return s.pubsub.Receive(ctx)
}

func (s redisPubSubSession) Channel() <-chan *goredis.Message {
	return s.pubsub.Channel()
}

func (s redisPubSubSession) Close() error {
	return s.pubsub.Close()
}

type rawRedisClient struct {
	client *goredis.Client
}

func (r rawRedisClient) Publish(ctx context.Context, channel string, message interface{}) error {
	return r.client.Publish(ctx, channel, message).Err()
}

func (r rawRedisClient) Subscribe(ctx context.Context, channels ...string) pubSubSession {
	return redisPubSubSession{pubsub: r.client.Subscribe(ctx, channels...)}
}
