package expense

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	domain "github.com/vishalyadav0987/expense-analyser/internal/domain/expense"
	"github.com/vishalyadav0987/expense-analyser/internal/domain/setup"
)

type ExpenseService struct {
	repo      domain.ExpenseRepository
	setupRepo setup.SetupRepository
}

func NewExpenseService(repo domain.ExpenseRepository, setupRepo setup.SetupRepository) *ExpenseService {
	return &ExpenseService{repo: repo, setupRepo: setupRepo}
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

func (s *ExpenseService) ProcessNewExpense(ctx context.Context, reqExp *domain.Expense) (*domain.LimitWarning, error) {
	// 1. Fetch the exact Category to know if it's a Need, Want, or Saving
	fmt.Println(reqExp.CategoryID, reqExp.UserID)
	category, err := s.repo.GetCategoryById(ctx, reqExp.CategoryID, reqExp.UserID)
	fmt.Println(category)
	if err != nil {
		return nil, fmt.Errorf("invalid category: %w", err)
	}

	// Populate it so the frontend gets the full object back
	reqExp.Category = category
	reqExp.ID = "txn_" + uuid.NewString()[:8]

	// 2. Fetch User's Profile to get Salary and Smart Rules
	userProfile, err := s.setupRepo.GetInitailSetupDetails(ctx, reqExp.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user profile: %w", err)
	}

	var weeklyLimit float64

	// 🚨 THE SDE3 OVERRIDE PATTERN 🚨
	if userProfile.XXWeeklyLimit != 0 {
		// SCENARIO 1: User ne app ki settings mein jaake hardcode kiya tha ₹25,000
		// Pointer ko dereference (*) karke value nikal lo
		weeklyLimit = userProfile.XXWeeklyLimit

	} else {
		// SCENARIO 2: User ne kuch set nahi kiya hai. Apni Smart Math lagao!
		var percentageLimit int
		switch category.Type {
		case "Need":
			percentageLimit = userProfile.NeedsPercentage
		case "Want":
			percentageLimit = userProfile.WantsPercentage
		case "Saving":
			percentageLimit = userProfile.SavingsPercentage
		}

		monthlyLimit := (userProfile.MonthlySalary * float64(percentageLimit)) / 100
		weeklyLimit = monthlyLimit / 4
	}

	// 4. Calculate how much has ALREADY been spent this week for this type
	now := time.Now()
	// Get Monday of the current week
	offset := int(time.Monday - now.Weekday())
	if offset > 0 {
		offset = -6
	}
	weekStart := time.Date(now.Year(), now.Month(), now.Day()+offset, 0, 0, 0, 0, time.UTC)

	spentThisWeek, _ := s.repo.GetWeeklySpendByType(ctx, reqExp.UserID, string(category.Type), weekStart)

	// Add the CURRENT expense to see if it crosses the limit
	projectedWeeklySpend := spentThisWeek + reqExp.Amount

	var limitWarning *domain.LimitWarning

	// 5. THE RULE ENGINE: Check if they crossed the limit
	if projectedWeeklySpend > weeklyLimit {
		excessAmount := projectedWeeklySpend - weeklyLimit
		limitWarning = &domain.LimitWarning{
			CategoryType:  category.Type,
			Limit:         weeklyLimit,
			SpentThisWeek: projectedWeeklySpend,
			Message:       fmt.Sprintf("Warning: You have exceeded your weekly %s limit by ₹%.2f.", category.Type, excessAmount),
		}
	}

	// 6. Save the expense
	err = s.repo.CreateExpense(ctx, reqExp)
	if err != nil {
		return nil, err
	}

	return limitWarning, nil
}
