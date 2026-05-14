package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq" // PostgreSQL driver for specific error code checking
	"github.com/vishalyadav0987/expense-analyser/internal/domain/auth"
)

// UserRepository implements domain.auth.UserRepository
type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser safely inserts a new user, catching duplicate emails via DB constraints.
func (r *UserRepository) CreateUser(ctx context.Context, user *auth.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, mpin_hash, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.MPINHash, // This will naturally insert NULL if it's a nil pointer
		user.IsActive,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		// SDE3 Pattern: Catch Postgres-specific "Unique Violation" error securely
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			// Mask the DB error and return a clean, safe domain error
			return fmt.Errorf("user with email %s already exists", user.Email)
		}
		return fmt.Errorf("failed to insert user: %w", err)
	}

	return nil
}

// GetUserByEmail fetches the user. It maps sql.ErrNoRows to a clean nil return.
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*auth.User, error) {
	query := `
		SELECT id, email, password_hash, mpin_hash, is_active, created_at, updated_at
		FROM users 
		WHERE email = $1
	`

	var user auth.User
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// It's perfectly normal for a user not to exist (e.g., during login).
			// We return nil, nil so the Service layer can decide what to do.
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query user by email: %w", err)
	}

	return &user, nil
}

// UpdateMPIN updates the hash for the quick-login flow.
func (r *UserRepository) UpdateMPIN(ctx context.Context, userID string, mpinHash string) error {
	query := `
		UPDATE users 
		SET mpin_hash = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, mpinHash, userID)
	if err != nil {
		return fmt.Errorf("failed to update MPIN: %w", err)
	}

	// SDE3 Guardrail: Ensure we actually updated a row.
	// If a bad UserID is passed, SQL doesn't throw an error, it just updates 0 rows.
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("user not found or mpin already set to this value")
	}

	return nil
}
