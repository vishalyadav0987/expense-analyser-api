package auth

import (
	"context"
	"time"
)

// SecurityRepository handles brute-force protection and rate limiting.
type SecurityRepository interface {
	// RecordFailedAttempt increments the fail count and returns the new total.
	RecordFailedAttempt(ctx context.Context, email string, window time.Duration) (int64, error)

	// LockAccount sets a temporary lock on the account.
	LockAccount(ctx context.Context, email string, lockoutDuration time.Duration) error

	// GetLockTTL returns how much time is left on the lock. Returns 0 if not locked.
	GetLockTTL(ctx context.Context, email string) (time.Duration, error)

	// ClearLockAndAttempts resets everything upon a successful login.
	ClearLockAndAttempts(ctx context.Context, email string) error
}
