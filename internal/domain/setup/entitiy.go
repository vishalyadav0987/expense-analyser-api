package setup

import "errors"

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
}

type SmartRules struct {
	NeedsPercentage   int
	WantsPercentage   int
	SavingsPercentage int
}

// 3. Main Entities
type Category struct {
	ID     string
	UserID string
	Name   string
	Type   CategoryType
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
