package setup

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	domain "github.com/vishalyadav0987/expense-analyser/internal/domain/setup"
)

type SetupService struct {
	repo domain.SetupRepository
}

func NewSetupService(repo domain.SetupRepository) *SetupService {
	return &SetupService{repo: repo}
}

// ProcessInitialSetup builds the entity, validates it, and saves it.
func (s *SetupService) ProcessInitialSetup(ctx context.Context, p *domain.UserInitialSetup) error {
	// 1. Generate secure UUIDs for all categories
	for i := range p.Categories {
		p.Categories[i].ID = "cat_" + uuid.NewString()[:8] // Short, unique IDs for Flutter
		p.Categories[i].UserID = p.UserID
	}

	// 2. Generate secure UUIDs for all payment methods
	for i := range p.PaymentMethods {
		p.PaymentMethods[i].ID = "pay_" + uuid.NewString()[:8]
		p.PaymentMethods[i].UserID = p.UserID
	}

	// 3. Domain Validation (Checks the 100% rule and > 0 salary)
	if err := p.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// 4. Save to Database via the Transaction Repository
	p.SetupCompleted = true
	return s.repo.SaveCompleteSetup(ctx, p)
}
