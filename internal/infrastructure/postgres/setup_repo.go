package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/vishalyadav0987/expense-analyser/internal/domain/setup"
)

type SetupRepository struct {
	db *sqlx.DB
}

func NewSetupRepository(db *sqlx.DB) *SetupRepository {
	return &SetupRepository{db: db}
}

// SaveCompleteSetup executes the entire setup process atomically.
func (r *SetupRepository) SaveCompleteSetup(ctx context.Context, p *setup.UserInitialSetup) error {
	// 1. Begin the Transaction
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// SDE3 Safety Net: defer a rollback.
	// If the function returns an error or panics, the transaction rolls back safely.
	// If tx.Commit() is successful later, this rollback becomes a no-op.
	defer tx.Rollback()

	// 2. Insert User Profile (Financials & Smart Rules)
	profileQuery := `
		INSERT INTO user_profiles (
			user_id, monthly_salary, yearly_hike_percentage, 
			needs_percentage, wants_percentage, savings_percentage, setup_completed, xx_weekly_limit
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err = tx.ExecContext(ctx, profileQuery,
		p.UserID,
		p.Financials.MonthlySalary,
		p.Financials.YearlyHikePercentage,
		p.SmartRules.NeedsPercentage,
		p.SmartRules.WantsPercentage,
		p.SmartRules.SavingsPercentage,
		p.SetupCompleted,
		p.Financials.XXWeeklyLimit,
	)
	if err != nil {
		return fmt.Errorf("failed to insert user profile: %w", err)
	}

	// 3. Insert Categories
	categoryQuery := `
		INSERT INTO categories (id, user_id, name, type) 
		VALUES ($1, $2, $3, $4)
	`
	for _, cat := range p.Categories {
		_, err = tx.ExecContext(ctx, categoryQuery, cat.ID, cat.UserID, cat.Name, cat.Type)
		if err != nil {
			return fmt.Errorf("failed to insert category '%s': %w", cat.Name, err)
		}
	}

	// 4. Insert Payment Methods
	paymentQuery := `
		INSERT INTO payment_methods (id, user_id, method_name, weekly_limit, is_active)
		VALUES ($1, $2, $3, $4, $5)
	`
	for _, pm := range p.PaymentMethods {
		_, err = tx.ExecContext(ctx, paymentQuery, pm.ID, pm.UserID, pm.MethodName, pm.WeeklyLimit, pm.IsActive)
		if err != nil {
			return fmt.Errorf("failed to insert payment method '%s': %w", pm.MethodName, err)
		}
	}

	// 5. Commit the Transaction
	// If we reach this line, EVERYTHING succeeded. We lock it into the database permanently.
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit setup transaction: %w", err)
	}

	return nil
}

func (r *SetupRepository) GetInitailSetupDetails(
	ctx context.Context,
	userId string,
) (*setup.UserInitialSetupDTO, error) {

	query := `
		SELECT
			user_id,
			monthly_salary,
			yearly_hike_percentage,
			needs_percentage,
			wants_percentage,
			savings_percentage,
			setup_completed,
			created_at,
			updated_at
		FROM user_profiles
		WHERE user_id = $1
	`

	var userSetupData setup.UserInitialSetupDTO

	err := r.db.GetContext(ctx, &userSetupData, query, userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to get initial setup details: %w", err)
	}

	return &userSetupData, nil
}
