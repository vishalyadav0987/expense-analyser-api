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
	ID           string
	Email        string
	PasswordHash string
	// SDE3 Note: MPINHash is a pointer (*string) because when a user first
	// registers, they haven't set an MPIN yet. It will be NULL in the database.
	MPINHash  *string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// HasSetupMPIN is a domain helper to instantly check if the user needs to be routed
// to the MPIN setup screen or the dashboard.
func (u *User) HasSetupMPIN() bool {
	return u.MPINHash != nil && *u.MPINHash != ""
}
