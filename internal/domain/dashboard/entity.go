package dashboard

import "time"

type DashboardResponse struct {
	Success bool          `json:"success"`
	Message string        `json:"message"`
	Data    DashboardData `json:"data"`
}

type DashboardData struct {
	TopCard            TopCard              `json:"topCard"`
	FinancialScore     int                  `json:"financialScore"`
	RuleProgress       RuleProgress         `json:"ruleProgress"`
	RecentTransactions []*RecentTransaction `json:"recentTransactions"`
}

type TopCard struct {
	TotalSalary       float64 `json:"totalSalary"`
	BudgetPending     float64 `json:"budgetPending"`
	SavingTarget      float64 `json:"savingTarget"`
	CurrentSavingRate float64 `json:"currentSavingRate"`
}

type RuleProgress struct {
	Needs   CategoryProgress `json:"needs"`
	Wants   CategoryProgress `json:"wants"`
	Savings SavingsProgress  `json:"savings"`
}

type CategoryProgress struct {
	Spent              float64 `json:"spent"`
	Limit              float64 `json:"limit"`
	PercentageConsumed float64 `json:"percentageConsumed"`
}

type SavingsProgress struct {
	Invested           float64 `json:"invested"`
	Target             float64 `json:"target"`
	PercentageAchieved float64 `json:"percentageAchieved"`
}

type RecentTransaction struct {
	ID          string    `json:"id"`
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Amount      float64   `json:"amount"`
	PaymentMode string    `json:"paymentMode" db:"payment_mode"`
	Type        string    `json:"type"` // "Need", "Want", "Saving"
}
