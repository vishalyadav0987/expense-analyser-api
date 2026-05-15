package connect

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

// NewClient initializes the Redis client and pings the server to ensure connectivity.
func NewClient(addr string) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	// Use a brief timeout context just for the ping, so it doesn't hang forever if Redis is down
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return rdb, nil
}
