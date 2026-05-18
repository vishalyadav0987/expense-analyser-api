package analyser

import (
	"context"
	"time"
)

type RawTransactionData struct {
	ID           string    `db:"id"`
	Amount       float64   `db:"amount"`
	Description  string    `db:"description"`
	PaymentMode  string    `db:"payment_mode"`
	Date         time.Time `db:"date"`
	CategoryName string    `db:"category_name"`
	CategoryType string    `db:"category_type"`
}

type AnalyzerRepository interface {
	GetTransactionsForPeriod(
		ctx context.Context,
		userID string,
		startDate,
		endDate time.Time,
	) ([]RawTransactionData, error)
}
