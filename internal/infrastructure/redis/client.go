package redis

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type Config struct {
	Addr           string
	Host           string
	Port           int
	Username       string
	Password       string
	DB             int
	PoolSize       int
	TLS            bool
	SentinelMaster string
	ClusterAddrs   string
}

type Client struct {
	rc redis.UniversalClient
}

func NewClient(cfg *Config) (*Client, error) {
	rdb := newUniversalClient(*cfg)

	client := &Client{rc: rdb}
	if err := client.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	return client, nil
}

func newUniversalClient(cfg Config) redis.UniversalClient {
	if addrs := splitAddrs(cfg.ClusterAddrs); len(addrs) > 0 {
		return redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        addrs,
			Username:     cfg.Username,
			Password:     cfg.Password,
			PoolSize:     cfg.PoolSize,
			DialTimeout:  60 * time.Second,
			ReadTimeout:  60 * time.Second,
			WriteTimeout: 60 * time.Second,
			TLSConfig:    redisTLSConfig(cfg.TLS),
		})
	}

	addr := cfg.Addr
	if addr == "" && cfg.Host != "" {
		addr = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	}

	if cfg.SentinelMaster != "" {
		return redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:    cfg.SentinelMaster,
			SentinelAddrs: splitAddrs(addr),
			Username:      cfg.Username,
			Password:      cfg.Password,
			DB:            cfg.DB,
			PoolSize:      cfg.PoolSize,
			DialTimeout:   60 * time.Second,
			ReadTimeout:   60 * time.Second,
			WriteTimeout:  60 * time.Second,
			TLSConfig:     redisTLSConfig(cfg.TLS),
		})
	}

	return redis.NewClient(&redis.Options{
		Addr:         addr,
		Username:     cfg.Username,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		DialTimeout:  60 * time.Second,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		TLSConfig:    redisTLSConfig(cfg.TLS),
	})
}

func splitAddrs(addrs string) []string {
	parts := strings.Split(addrs, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if addr := strings.TrimSpace(part); addr != "" {
			out = append(out, addr)
		}
	}
	return out
}

func redisTLSConfig(enabled bool) *tls.Config {
	if !enabled {
		return nil
	}
	return &tls.Config{MinVersion: tls.VersionTLS12}
}

func (c *Client) Ping(ctx context.Context) error {
	return c.rc.Ping(ctx).Err()
}

func (c *Client) Raw() redis.UniversalClient {
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
