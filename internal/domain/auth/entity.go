package auth

import (
	"errors"
	"time"
)

var (
	ErrInvalidEmail = errors.New("invalid email format")
	ErrMPINTooShort = errors.New("MPIN must be exactly 4 digits")
)

// User represents the core entity in our authentication domain.
type User struct {
	ID           string `db:"id"`
	Email        string `db:"email"`
	PasswordHash string `db:"password_hash"`

	// SDE3 Critical Detail: This MUST be a pointer (*string)
	// because in our SQL table, mpin_hash can be NULL.
	// If you use a regular string, sqlx will crash when it hits a NULL.
	MPINHash      *string   `db:"mpin_hash"`
	SetupComplete bool      `db:"setup_complete"`
	IsActive      bool      `db:"is_active"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

// HasSetupMPIN is a domain helper to instantly check if the user needs to be routed
// to the MPIN setup screen or the dashboard.
func (u *User) HasSetupMPIN() bool {
	return u.MPINHash != nil && *u.MPINHash != ""
}

func (u *User) MarkSetupCompleted() {
	u.SetupComplete = true
}
