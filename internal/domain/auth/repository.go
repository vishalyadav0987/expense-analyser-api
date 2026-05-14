package auth

import (
	"context"
	"time"
)

// UserRepository defines how we persist permanent user data.
// Notice there is no SQL here. It's just a contract.
type UserRepository interface {
	CreateUser(ctx context.Context, user *User) error
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, userID string) (*User, error)
	UpdateMPIN(ctx context.Context, userID string, mpinHash string) error
}

// OTPRepository defines how we handle short-lived, transient auth data.
// In production, this is ALWAYS implemented using Redis or Memcached, never Postgres.
type OTPRepository interface {
	SaveOTP(ctx context.Context, email string, otp string, ttl time.Duration) error
	GetOTP(ctx context.Context, email string) (string, error)
	DeleteOTP(ctx context.Context, email string) error
}
