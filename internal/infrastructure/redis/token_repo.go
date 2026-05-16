package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

type TokenRepository struct {
	client *redis.Client
}

func NewTokenRepository(client *redis.Client) *TokenRepository {
	return &TokenRepository{client: client}
}

func (r *TokenRepository) MarkTokenUsed(ctx context.Context, jti string, expiration time.Duration) error {
	// Save the token ID to Redis. It automatically deletes itself when it naturally expires.
	key := fmt.Sprintf("used_rt:%s", jti)
	return r.client.Set(ctx, key, "used", expiration).Err()
}

func (r *TokenRepository) IsTokenUsed(ctx context.Context, jti string) (bool, error) {
	key := fmt.Sprintf("used_rt:%s", jti)
	res, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return res > 0, nil
}

func (r *TokenRepository) ActivateKillSwitch(ctx context.Context, userID string) error {
	// If a token is stolen, we save the exact second we detected the theft.
	// Keep this kill switch active for 30 days (the max life of any token).
	key := fmt.Sprintf("kill_switch:%s", userID)
	now := time.Now().Unix()
	return r.client.Set(ctx, key, now, 30*24*time.Hour).Err()
}

func (r *TokenRepository) IsKillSwitchActive(ctx context.Context, userID string, tokenIssuedAt time.Time) (bool, error) {
	key := fmt.Sprintf("kill_switch:%s", userID)
	val, err := r.client.Get(ctx, key).Result()

	if err == redis.Nil {
		return false, nil // No kill switch active
	} else if err != nil {
		return false, err
	}

	killTimeUnix, _ := strconv.ParseInt(val, 10, 64)

	// If the token was issued BEFORE the kill switch was activated, it is dead.
	if tokenIssuedAt.Unix() <= killTimeUnix {
		return true, nil
	}
	return false, nil
}
