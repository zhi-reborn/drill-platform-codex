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
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	return &Client{rc: rdb}, nil
}

func (c *Client) Get(key string) (string, error) {
	return c.rc.Get(ctx, key).Result()
}

func (c *Client) Set(key string, value interface{}, expiration time.Duration) error {
	return c.rc.Set(ctx, key, value, expiration).Err()
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
	return c.rc.Publish(ctx, channel, message).Err()
}

func (c *Client) Subscribe(channels ...string) *redis.PubSub {
	return c.rc.Subscribe(ctx, channels...)
}

func (c *Client) Close() error {
	return c.rc.Close()
}
