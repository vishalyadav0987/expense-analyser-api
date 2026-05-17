package dto

import (
	"time"
)

type CreateExpenseRequest struct {
	Name string `json:"name" binding:"required"`
	Type string `json:"type" binding:"required,oneof=Need Want Saving"`
}

type AddExpenseRequest struct {
	Amount      float64   `json:"amount" binding:"required,gt=0"`
	CategoryID  string    `json:"categoryId" binding:"required"`
	Description string    `json:"description" binding:"required"`
	PaymentMode string    `json:"paymentMode" binding:"required"`
	Date        time.Time `json:"date" binding:"required"` // Gin automatically parses ISO8601 dates!
}
