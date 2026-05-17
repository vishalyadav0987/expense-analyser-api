package setup

import (
	"errors"
	"time"
)

// 1. Enums for strict type safety
type CategoryType string

const (
	CategoryTypeNeed   CategoryType = "Need"
	CategoryTypeWant   CategoryType = "Want"
	CategoryTypeSaving CategoryType = "Saving"
)

type PaymentMethodType string

const (
	PaymentMethodCash   PaymentMethodType = "Cash"
	PaymentMethodDebit  PaymentMethodType = "Debit Card"
	PaymentMethodCredit PaymentMethodType = "Credit Card"
	PaymentMethodUpi    PaymentMethodType = "Upi"
	PaymentMethodBank   PaymentMethodType = "Bank"
)

// 2. Financial & Rule Value Objects
type Financials struct {
	MonthlySalary        float64
	YearlyHikePercentage float64
	XXWeeklyLimit        float64
}

type SmartRules struct {
	NeedsPercentage   int
	WantsPercentage   int
	SavingsPercentage int
}

// 3. Main Entities
type Category struct {
	ID     string       `db:"id"`
	UserID string       `db:"user_id"`
	Name   string       `db:"name"`
	Type   CategoryType `db:"type"`
}

type PaymentMethod struct {
	ID          string
	UserID      string
	MethodName  PaymentMethodType
	WeeklyLimit float64
	IsActive    bool
}

// 4. The Aggregate Root (Groups everything for the Setup Transaction)
type UserInitialSetup struct {
	UserID         string
	SetupCompleted bool
	Financials     Financials
	SmartRules     SmartRules
	Categories     []Category
	PaymentMethods []PaymentMethod
}

func (u *UserInitialSetup) Validate() error {
	if u.Financials.MonthlySalary <= 0 {
		return errors.New("monthly salary must be greater than 0")
	}

	totalPercentage := u.SmartRules.NeedsPercentage + u.SmartRules.WantsPercentage + u.SmartRules.SavingsPercentage
	if totalPercentage != 100 {
		return errors.New("smart rules percentages must exactly equal 100")
	}

	return nil
}

func (u *UserInitialSetup) ValidateWeeklyLimit() error {
	if u.Financials.XXWeeklyLimit <= 0 {
		return errors.New("weekly limit must be greater than 0")
	}

	if u.Financials.XXWeeklyLimit > u.Financials.MonthlySalary {
		return errors.New("weekly limit should be less the monthly salary")
	}

	return nil
}

type UserInitialSetupDTO struct {
	UserID               string    `db:"user_id"`
	SetupCompleted       bool      `db:"setup_completed"`
	MonthlySalary        float64   `db:"monthly_salary"`
	YearlyHikePercentage float64   `db:"yearly_hike_percentage"`
	XXWeeklyLimit        float64   `db:"xx_weekly_limit"`
	NeedsPercentage      int       `db:"needs_percentage"`
	WantsPercentage      int       `db:"wants_percentage"`
	SavingsPercentage    int       `db:"savings_percentage"`
	CreatedAt            time.Time `db:"created_at"`
	UpdatedAt            time.Time `db:"updated_at"`
}
