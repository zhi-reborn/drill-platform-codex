package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type Config struct {
	Host     string
	Port     int
	Password string
	DB       int
	PoolSize int
}

type Client struct {
	rc *redis.Client
}

func NewClient(cfg *Config) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		DialTimeout:  60 * time.Second,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
	})

	client := &Client{rc: rdb}
	if err := client.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	return client, nil
}

func (c *Client) Ping(ctx context.Context) error {
	return c.rc.Ping(ctx).Err()
}

func (c *Client) Raw() *redis.Client {
	return c.rc
}

func (c *Client) Get(key string) (string, error) {
	return c.rc.Get(ctx, key).Result()
}

func (c *Client) Set(key string, value interface{}, expiration time.Duration) error {
	return c.rc.Set(ctx, key, value, expiration).Err()
}

func (c *Client) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return c.rc.SetNX(ctx, key, value, expiration).Result()
}

func (c *Client) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	return c.rc.Eval(ctx, script, keys, args...).Result()
}

func (c *Client) Delete(keys ...string) error {
	return c.rc.Del(ctx, keys...).Err()
}

func (c *Client) Exists(keys ...string) (int64, error) {
	return c.rc.Exists(ctx, keys...).Result()
}

func (c *Client) Expire(key string, expiration time.Duration) error {
	return c.rc.Expire(ctx, key, expiration).Err()
}

func (c *Client) Publish(channel string, message interface{}) error {
	return c.PublishContext(ctx, channel, message)
}

func (c *Client) PublishContext(ctx context.Context, channel string, message interface{}) error {
	return c.rc.Publish(ctx, channel, message).Err()
}

func (c *Client) Subscribe(channels ...string) *redis.PubSub {
	return c.SubscribeContext(ctx, channels...)
}

func (c *Client) SubscribeContext(ctx context.Context, channels ...string) *redis.PubSub {
	return c.rc.Subscribe(ctx, channels...)
}

func (c *Client) Close() error {
	return c.rc.Close()
}
