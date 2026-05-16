package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	domain "github.com/vishalyadav0987/expense-analyser/internal/domain/setup"
)

type ExpenseRepository struct {
	db *sqlx.DB
}

func NewExpenseRepository(db *sqlx.DB) *ExpenseRepository {
	return &ExpenseRepository{db: db}
}

func (r *ExpenseRepository) CreateCategory(ctx context.Context, cat *domain.Category) error {
	query := `
		INSERT INTO categories (id, user_id, name, type) 
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.ExecContext(ctx, query, cat.ID, cat.UserID, cat.Name, cat.Type)
	if err != nil {
		return fmt.Errorf("failed to create category: %w", err)
	}
	return nil
}
