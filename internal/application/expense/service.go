package expense

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	domain "github.com/vishalyadav0987/expense-analyser/internal/domain/expense"
	"github.com/vishalyadav0987/expense-analyser/internal/domain/setup"
)

type ExpenseService struct {
	repo domain.ExpenseRepository
}

func NewExpenseService(repo domain.ExpenseRepository) *ExpenseService {
	return &ExpenseService{repo: repo}
}

func (s *ExpenseService) CreateCategory(
	ctx context.Context,
	userID, name string,
	catType setup.CategoryType,
) (*setup.Category, error) {
	// Build the entity
	category := &setup.Category{
		ID:     "cat_" + uuid.NewString()[:8],
		UserID: userID,
		Name:   name,
		Type:   catType,
	}

	// Save it to the database
	if err := s.repo.CreateCategory(ctx, category); err != nil {
		return nil, fmt.Errorf("service failed to create category: %w", err)
	}

	return category, nil
}
