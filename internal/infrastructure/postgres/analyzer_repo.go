package postgres

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

type AnalyzerRepository struct {
	db *sqlx.DB
}

func NewAnalyzerRepository(db *sqlx.DB) *AnalyzerRepository {
	return &AnalyzerRepository{db: db}
}

// Ye struct database columns ko map karega
type RawTransactionData struct {
	ID           string    `db:"id"`
	Amount       float64   `db:"amount"`
	Description  string    `db:"description"`
	PaymentMode  string    `db:"payment_mode"`
	Date         time.Time `db:"date"`
	CategoryName string    `db:"category_name"`
	CategoryType string    `db:"category_type"`
}

func (r *AnalyzerRepository) GetTransactionsForPeriod(
	ctx context.Context,
	userID string,
	startDate,
	endDate time.Time,
) ([]RawTransactionData, error) {
	query := `
		SELECT 
			e.id, e.amount, e.description, e.payment_mode, e.date,
			c.name as category_name, c.type as category_type
		FROM expenses e
		JOIN categories c ON e.category_id = c.id
		WHERE e.user_id = $1 
		AND e.date >= $2 
		AND e.date <= $3
		ORDER BY e.date DESC
	`

	var transactions []RawTransactionData
	err := r.db.SelectContext(ctx, &transactions, query, userID, startDate, endDate)
	return transactions, err
}
