package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	indexCacheTTL   = 5 * time.Second
	statsCacheTTL   = 60 * time.Second
	candlesCacheTTL = 30 * time.Second
)

// RedisCache wraps a Redis client with helper methods.
type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(redisURL string) *RedisCache {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		// Fallback to default
		opts = &redis.Options{Addr: "localhost:6379"}
	}
	// Connection pool tuned for high concurrency
	opts.PoolSize = 100
	opts.MinIdleConns = 10
	opts.MaxRetries = 3

	return &RedisCache{client: redis.NewClient(opts)}
}

func (c *RedisCache) Close() error {
	return c.client.Close()
}

// ─── Generic Cache Helpers ────────────────────────────────────

func (c *RedisCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	b, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	return c.client.Set(ctx, key, b, ttl).Err()
}

func (c *RedisCache) Get(ctx context.Context, key string, dest any) error {
	b, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dest)
}

func (c *RedisCache) Del(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}

// ─── Domain-specific Keys ────────────────────────────────────

func (c *RedisCache) SetIndex(ctx context.Context, symbol string, v any) error {
	return c.Set(ctx, "idx:"+symbol, v, indexCacheTTL)
}

func (c *RedisCache) GetIndex(ctx context.Context, symbol string, dest any) error {
	return c.Get(ctx, "idx:"+symbol, dest)
}

func (c *RedisCache) SetAllIndices(ctx context.Context, v any) error {
	return c.Set(ctx, "idx:all", v, indexCacheTTL)
}

func (c *RedisCache) GetAllIndices(ctx context.Context, dest any) error {
	return c.Get(ctx, "idx:all", dest)
}

func (c *RedisCache) SetStats(ctx context.Context, symbol string, v any) error {
	return c.Set(ctx, "stats:"+symbol, v, statsCacheTTL)
}

func (c *RedisCache) GetStats(ctx context.Context, symbol string, dest any) error {
	return c.Get(ctx, "stats:"+symbol, dest)
}

func (c *RedisCache) InvalidateIndex(ctx context.Context, symbol string) error {
	return c.Del(ctx, "idx:"+symbol, "idx:all")
}

// ─── Pub/Sub for WebSocket broadcasting ──────────────────────

func (c *RedisCache) Publish(ctx context.Context, channel string, payload any) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	return c.client.Publish(ctx, channel, b).Err()
}

func (c *RedisCache) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return c.client.Subscribe(ctx, channels...)
}

// ─── Rate Limiter (token bucket via Redis) ───────────────────

// AllowRequest returns true if the given key is within rateLimit per window.
// Used to throttle external API calls per IP when needed.
func (c *RedisCache) AllowRequest(ctx context.Context, key string, limit int64, window time.Duration) (bool, error) {
	pipe := c.client.Pipeline()
	incr := pipe.Incr(ctx, "rl:"+key)
	pipe.Expire(ctx, "rl:"+key, window)
	if _, err := pipe.Exec(ctx); err != nil {
		return false, err
	}
	return incr.Val() <= limit, nil
}

// ─── Health Check ────────────────────────────────────────────

func (c *RedisCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}
