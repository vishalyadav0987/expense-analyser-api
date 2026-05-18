package expense

import (
	"context"
	"time"

	"github.com/vishalyadav0987/expense-analyser/internal/domain/setup"
)

type ExpenseRepository interface {
	CreateCategory(ctx context.Context, category *setup.Category) error
	GetCategoryById(ctx context.Context, categoryId string, userId string) (*setup.Category, error)
	CreateExpense(ctx context.Context, exp *Expense) error
	GetWeeklySpendByType(
		ctx context.Context,
		userID string,
		categoryType string,
		weekStart time.Time,
	) (float64, error)
	GetAllCategoriesByUserID(ctx context.Context, userID string) ([]*setup.Category, error)
}
