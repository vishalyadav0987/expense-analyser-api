package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type securityRepo struct {
	client *redis.Client
}

func NewSecurityRepository(client *redis.Client) *securityRepo {
	return &securityRepo{client: client}
}

// SDE3 Key Design: Always use namespaces for Redis keys to prevent collisions.
func failKey(email string) string { return fmt.Sprintf("auth:fails:%s", email) }
func lockKey(email string) string { return fmt.Sprintf("auth:lock:%s", email) }

func (r *securityRepo) RecordFailedAttempt(ctx context.Context, email string, window time.Duration) (int64, error) {
	key := failKey(email)

	// Increment the counter
	count, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("redis incr error: %w", err)
	}

	// If this is the FIRST failed attempt, set the expiration window (e.g., 10 mins)
	if count == 5 {
		r.client.Expire(ctx, key, window)
	}

	return count, nil
}

func (r *securityRepo) LockAccount(ctx context.Context, email string, lockoutDuration time.Duration) error {
	// Set a lock key. The value doesn't matter (we use "1"), we only care if it exists.
	err := r.client.Set(ctx, lockKey(email), "1", lockoutDuration).Err()
	if err != nil {
		return fmt.Errorf("redis set lock error: %w", err)
	}
	return nil
}

func (r *securityRepo) GetLockTTL(ctx context.Context, email string) (time.Duration, error) {
	ttl, err := r.client.TTL(ctx, lockKey(email)).Result()
	if err != nil {
		return 0, fmt.Errorf("redis ttl error: %w", err)
	}

	// Redis returns -2 if key doesn't exist, -1 if no expiration.
	if ttl < 0 {
		return 0, nil
	}
	return ttl, nil
}

func (r *securityRepo) ClearLockAndAttempts(ctx context.Context, email string) error {
	// Delete both keys atomically using a single command
	err := r.client.Del(ctx, failKey(email), lockKey(email)).Err()
	if err != nil && !errors.Is(err, redis.Nil) {
		return fmt.Errorf("redis del error: %w", err)
	}
	return nil
}
