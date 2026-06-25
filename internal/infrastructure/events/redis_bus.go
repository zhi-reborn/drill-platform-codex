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
}

func NewRedisBus(client interface{}) *RedisBus {
	bus := &RedisBus{backoff: 100 * time.Millisecond}

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

func (b *RedisBus) Publish(ctx context.Context, event Event) error {
	if b.publisher == nil {
		return errors.New("redis publisher is not configured")
	}

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return b.publisher.Publish(ctx, EventsChannel, string(data))
}

func (b *RedisBus) Subscribe(ctx context.Context, handler func(Event)) error {
	if b.subscriber == nil {
		return errors.New("redis subscriber is not configured")
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
			if err := sleepWithContext(ctx, b.backoff); err != nil {
				return err
			}
			continue
		}

		b.healthy.Store(true)
		ch := pubsub.Channel()
		for {
			select {
			case <-ctx.Done():
				b.healthy.Store(false)
				pubsub.Close()
				return ctx.Err()
			case msg, ok := <-ch:
				if !ok {
					b.healthy.Store(false)
					pubsub.Close()
					if err := sleepWithContext(ctx, b.backoff); err != nil {
						return err
					}
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

func (b *RedisBus) handleMessage(payload string, handler func(Event)) {
	var event Event
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		return
	}
	handler(event)
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
