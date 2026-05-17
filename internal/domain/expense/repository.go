package expense

import (
	"context"

	"github.com/vishalyadav0987/expense-analyser/internal/domain/setup"
)

type ExpenseRepository interface {
	CreateCategory(ctx context.Context, category *setup.Category) error
	GetCategoryById(ctx context.Context, categoryId string, userId string) (*setup.Category, error)
	CreateExpense(ctx context.Context, exp *Expense) (string, error)
}
