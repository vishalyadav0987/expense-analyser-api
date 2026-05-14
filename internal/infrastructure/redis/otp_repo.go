package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8" // Standard Redis client for Go
)

// otpRepository is the concrete implementation of auth.OTPRepository
type otpRepository struct {
	client *redis.Client
}

// NewOTPRepository creates a new Redis-backed OTP store
func NewOTPRepository(client *redis.Client) *otpRepository {
	return &otpRepository{client: client}
}

// SaveOTP stores the OTP with an automatic expiration (TTL)
func (r *otpRepository) SaveOTP(ctx context.Context, email, otp string, ttl time.Duration) error {
	key := fmt.Sprintf("otp:%s", email)

	// Redis handles the 5-minute expiration automatically.
	err := r.client.Set(ctx, key, otp, time.Minute*5).Err()
	if err != nil {
		return fmt.Errorf("redis failed to save otp: %w", err)
	}
	return nil
}

// GetOTP retrieves the OTP. It returns an error if the OTP doesn't exist or expired.
func (r *otpRepository) GetOTP(ctx context.Context, email string) (string, error) {
	key := fmt.Sprintf("otp:%s", email)

	otp, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", errors.New("otp expired or not found")
	} else if err != nil {
		return "", fmt.Errorf("redis failed to get otp: %w", err)
	}

	return otp, nil
}

// DeleteOTP ensures an OTP cannot be reused after successful verification (Replay Attack prevention)
func (r *otpRepository) DeleteOTP(ctx context.Context, email string) error {
	key := fmt.Sprintf("otp:%s", email)
	return r.client.Del(ctx, key).Err()
}
