package setup

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/vishalyadav0987/expense-analyser/internal/domain/auth"
	domain "github.com/vishalyadav0987/expense-analyser/internal/domain/setup"
)

type SetupService struct {
	repo     domain.SetupRepository
	authRepo auth.UserRepository
}

func NewSetupService(repo domain.SetupRepository, authRepo auth.UserRepository) *SetupService {
	return &SetupService{repo: repo, authRepo: authRepo}
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
	err := s.repo.SaveCompleteSetup(ctx, p)
	if err != nil {
		return fmt.Errorf("failed to save setup data: %w", err)
	}

	err = s.authRepo.MarkSetupComplete(ctx, p.UserID)
	if err != nil {
		// Even if this fails, the profile data was saved.
		// You can either return the error or just log it, depending on your strictness.
		return fmt.Errorf("setup saved, but failed to update user status: %w", err)
	}

	return nil
}
