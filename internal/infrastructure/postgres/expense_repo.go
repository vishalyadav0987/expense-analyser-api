package postgres

// import (
// 	"context"
// 	"database/sql"
// 	"fmt"
// 	"time"

// 	"github.com/jmoiron/sqlx"
// 	// Replace "expense-analyzer" with your actual go.mod module name
// 	"expense-analyzer/internal/domain/expense"
// )

// // ExpenseRepository implements the expense.Repository interface defined in the domain.
// type ExpenseRepository struct {
// 	db *sqlx.DB
// }

// // NewExpenseRepository is the constructor used for Dependency Injection.
// func NewExpenseRepository(db *sqlx.DB) *ExpenseRepository {
// 	return &ExpenseRepository{db: db}
// }

// // SaveExpense persists the transaction and updates the daily summary atomically.
// func (r *ExpenseRepository) SaveExpense(ctx context.Context, exp *expense.Expense) error {
// 	// 1. Begin a Database Transaction
// 	tx, err := r.db.BeginTxx(ctx, nil)
// 	if err != nil {
// 		return fmt.Errorf("failed to begin db transaction: %w", err)
// 	}

// 	// SDE3 Pattern: Defers are evaluated LIFO. If tx.Commit() succeeds, Rollback() does nothing.
// 	// If the function panics or returns early with an error, Rollback() ensures no partial data is saved.
// 	defer tx.Rollback()

// 	// 2. Insert into the transactions table
// 	insertTxQuery := `
// 		INSERT INTO transactions (id, user_id, category_id, amount, payment_mode, description, date, type)
// 		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
// 	`
// 	_, err = tx.ExecContext(ctx, insertTxQuery,
// 		exp.ID, exp.UserID, exp.CategoryID, exp.Amount, exp.PaymentMode, exp.Description, exp.Date, exp.Type,
// 	)
// 	if err != nil {
// 		return fmt.Errorf("failed to insert transaction: %w", err)
// 	}

// 	// 3. Upsert into daily_summaries (The SDE3 Performance Trick)
// 	// We use ON CONFLICT to either insert a new row for today, or add the amount to an existing row.
// 	// This happens at the DB level, avoiding race conditions and eliminating the need for a SELECT first.
// 	upsertSummaryQuery := `
// 		INSERT INTO daily_summaries (user_id, date, category_type, total_amount)
// 		VALUES ($1, DATE($2), $3, $4)
// 		ON CONFLICT (user_id, date, category_type)
// 		DO UPDATE SET total_amount = daily_summaries.total_amount + EXCLUDED.total_amount
// 	`
// 	_, err = tx.ExecContext(ctx, upsertSummaryQuery,
// 		exp.UserID, exp.Date, exp.Type, exp.Amount,
// 	)
// 	if err != nil {
// 		return fmt.Errorf("failed to upsert daily summary: %w", err)
// 	}

// 	// 4. Commit the transaction
// 	if err := tx.Commit(); err != nil {
// 		return fmt.Errorf("failed to commit transaction: %w", err)
// 	}

// 	return nil
// }

// // GetWeeklySpendByMode retrieves total spending for limit checks.
// func (r *ExpenseRepository) GetWeeklySpendByMode(ctx context.Context, userID string, mode expense.PaymentMode, start, end time.Time) (float64, error) {
// 	query := `
// 		SELECT SUM(amount)
// 		FROM transactions
// 		WHERE user_id = $1 AND payment_mode = $2 AND date >= $3 AND date <= $4 AND type != 'Saving'
// 	`

// 	var total sql.NullFloat64 // Use NullFloat64 because SUM returns NULL if there are no rows
// 	err := r.db.GetContext(ctx, &total, query, userID, mode, start, end)

// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			return 0, nil
// 		}
// 		return 0, fmt.Errorf("failed to calculate weekly spend: %w", err)
// 	}

// 	return total.Float64, nil
// }

// // GetRecentExpenses fetches data for the Dashboard list view.
// func (r *ExpenseRepository) GetRecentExpenses(ctx context.Context, userID string, limit int) ([]*expense.Expense, error) {
// 	query := `
// 		SELECT id, user_id, category_id, amount, payment_mode, description, date, type
// 		FROM transactions
// 		WHERE user_id = $1
// 		ORDER BY date DESC
// 		LIMIT $2
// 	`

// 	// sqlx handles the slice mapping automatically
// 	var expenses []*expense.Expense
// 	err := r.db.SelectContext(ctx, &expenses, query, userID, limit)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to fetch recent expenses: %w", err)
// 	}

// 	return expenses, nil
// }
