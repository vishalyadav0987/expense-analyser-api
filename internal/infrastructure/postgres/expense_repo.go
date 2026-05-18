package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/vishalyadav0987/expense-analyser/internal/domain/dashboard"
	"github.com/vishalyadav0987/expense-analyser/internal/domain/expense"
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

func (r *ExpenseRepository) GetCategoryById(
	ctx context.Context,
	categoryId string,
	userId string,
) (*domain.Category, error) {

	query := `
		SELECT id, user_id, name, type
		FROM categories
		WHERE user_id = $1 AND id = $2
	`

	var category domain.Category

	err := r.db.GetContext(ctx, &category, query, userId, categoryId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to get category by id: %w", err)
	}

	return &category, nil
}

func (r *ExpenseRepository) CreateExpense(ctx context.Context, exp *expense.Expense) error {
	query := `
		INSERT INTO expenses (id, user_id, amount, category_id, description, payment_mode, date)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query, exp.ID, exp.UserID, exp.Amount, exp.CategoryID, exp.Description, exp.PaymentMode, exp.Date)
	if err != nil {
		return fmt.Errorf("failed to insert expense: %w", err)
	}

	return nil
}

func (r *ExpenseRepository) GetWeeklySpendByType(
	ctx context.Context,
	userID string,
	categoryType string,
	weekStart time.Time,
) (float64, error) {
	query := `
		SELECT COALESCE(SUM(e.amount), 0) 
		FROM expenses e
		JOIN categories c ON e.category_id = c.id
		WHERE e.user_id = $1 
		AND c.type = $2 
		AND e.date >= $3
	`
	var totalSpent float64
	err := r.db.GetContext(ctx, &totalSpent, query, userID, categoryType, weekStart)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate weekly spend: %w", err)
	}
	return totalSpent, nil
}

func (r *ExpenseRepository) GetAllCategoriesByUserID(ctx context.Context, userID string) ([]*domain.Category, error) {
	query := `
		SELECT id, user_id, name, type 
		FROM categories 
		WHERE user_id = $1
		ORDER BY type, name -- SDE3 Trick: Alphabetical aur Type wise sort karke bhejo
	`

	// sqlx needs an initialized slice
	categories := make([]*domain.Category, 0)

	err := r.db.SelectContext(ctx, &categories, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch categories for user: %w", err)
	}

	return categories, nil
}

func (r *ExpenseRepository) GetMonthlyTransactions(ctx context.Context, userID string, startDate, endDate time.Time) ([]*dashboard.RecentTransaction, error) {
	// SDE3 Trick: JOIN categories to get 'Type' and 'Name' directly in one query
	query := `
		SELECT 
			e.id, e.date, e.description, c.name as category, e.amount, e.payment_mode, c.type
		FROM expenses e
		JOIN categories c ON e.category_id = c.id
		WHERE e.user_id = $1 AND e.date >= $2 AND e.date <= $3
		ORDER BY e.date DESC
	`

	transactions := make([]*dashboard.RecentTransaction, 0)
	err := r.db.SelectContext(ctx, &transactions, query, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch monthly transactions: %w", err)
	}

	return transactions, nil
}
