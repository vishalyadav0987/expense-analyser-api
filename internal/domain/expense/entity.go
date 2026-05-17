package expense

import (
	"time"

	setup "github.com/vishalyadav0987/expense-analyser/internal/domain/setup"
)

type Expense struct {
	ID          string          `json:"transactionId" db:"id"`
	UserID      string          `json:"-" db:"user_id"`
	Amount      float64         `json:"amount" db:"amount"`
	CategoryID  string          `json:"categoryId" db:"category_id"`
	Category    *setup.Category `json:"category,omitempty"` // Populated Category
	Description string          `json:"description" db:"description"`
	PaymentMode string          `json:"paymentMode" db:"payment_mode"` // e.g. "Credit Card"
	Date        time.Time       `json:"date" db:"date"`
	CreatedAt   time.Time       `json:"createdAt" db:"created_at"`
}

type LimitWarning struct {
	CategoryType  string  `json:"categoryType"`
	Limit         float64 `json:"limit"`
	SpentThisWeek float64 `json:"spentThisWeek"`
	Message       string  `json:"message"`
}
