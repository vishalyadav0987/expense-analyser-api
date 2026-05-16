package expense

import (
	"context"

	domain "github.com/vishalyadav0987/expense-analyser/internal/domain/setup"
)

type ExpenseRepository interface {
	CreateCategory(ctx context.Context, category *domain.Category) error
}
