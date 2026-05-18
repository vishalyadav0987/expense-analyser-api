package dashboard

import (
	"context"
	"time"

	"github.com/vishalyadav0987/expense-analyser/internal/domain/dashboard"
	"github.com/vishalyadav0987/expense-analyser/internal/domain/expense"
	"github.com/vishalyadav0987/expense-analyser/internal/domain/setup"
	// Import your repositories...
)

type DashboardService struct {
	expenseRepo expense.ExpenseRepository
	setupRepo   setup.SetupRepository
}

func NewDashboardService(repo expense.ExpenseRepository, setupRepo setup.SetupRepository) *DashboardService {
	return &DashboardService{expenseRepo: repo, setupRepo: setupRepo}
}

func (s *DashboardService) GetDashboardSummary(ctx context.Context, userID string, month, year int) (*dashboard.DashboardData, error) {
	// 1. Fetch Profile (Salary & Smart Rules)
	profile, err := s.setupRepo.GetInitailSetupDetails(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 2. Define Time Range for the month
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1).Add(23 * time.Hour).Add(59 * time.Minute)

	// 3. Fetch Transactions (O(N) processing)
	txns, err := s.expenseRepo.GetMonthlyTransactions(ctx, userID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// --- THE O(N) AGGREGATION LOOP ---
	var totalNeeds, totalWants, totalSavings, totalCCSpend float64
	savingTxnCount := 0

	for _, txn := range txns {
		switch txn.Type {
		case "Need":
			totalNeeds += txn.Amount
		case "Want":
			totalWants += txn.Amount
		case "Saving":
			totalSavings += txn.Amount
			savingTxnCount++
		}

		if txn.PaymentMode == "Credit Card" || txn.PaymentMode == "Credit" {
			totalCCSpend += txn.Amount
		}
	}

	totalSpent := totalNeeds + totalWants

	// --- 🚨 CALCULATE FINANCIAL SCORE (Using your exact logic) ---
	score := s.calculateScore(profile.MonthlySalary, totalSavings, totalSpent, totalWants, savingTxnCount, totalCCSpend)

	// --- CALCULATE WIDGET DATA ---
	needLimit := profile.MonthlySalary * float64(profile.NeedsPercentage) / 100
	wantLimit := profile.MonthlySalary * float64(profile.WantsPercentage) / 100
	savingTarget := profile.MonthlySalary * float64(profile.SavingsPercentage) / 100

	budgetPending := profile.MonthlySalary - (totalSpent + totalSavings)
	currentSavingRate := 0.0
	if profile.MonthlySalary > 0 {
		currentSavingRate = (totalSavings / profile.MonthlySalary) * 100
	}

	// Return top 5 recent transactions for the list view
	recentTxns := txns
	if len(txns) > 5 {
		recentTxns = txns[:5]
	}

	// --- BUILD FINAL JSON ---
	return &dashboard.DashboardData{
		TopCard: dashboard.TopCard{
			TotalSalary:       profile.MonthlySalary,
			BudgetPending:     budgetPending,
			SavingTarget:      savingTarget,
			CurrentSavingRate: currentSavingRate,
		},
		FinancialScore: score,
		RuleProgress: dashboard.RuleProgress{
			Needs: dashboard.CategoryProgress{
				Spent:              totalNeeds,
				Limit:              needLimit,
				PercentageConsumed: safePercentage(totalNeeds, needLimit),
			},
			Wants: dashboard.CategoryProgress{
				Spent:              totalWants,
				Limit:              wantLimit,
				PercentageConsumed: safePercentage(totalWants, wantLimit),
			},
			Savings: dashboard.SavingsProgress{
				Invested:           totalSavings,
				Target:             savingTarget,
				PercentageAchieved: safePercentage(totalSavings, savingTarget),
			},
		},
		RecentTransactions: recentTxns,
	}, nil
}

// safePercentage prevents divide by zero
func safePercentage(val, limit float64) float64 {
	if limit == 0 {
		return 0
	}
	return (val / limit) * 100
}

// 🧠 THE SCORING ENGINE (Exact matching of your logic)
func (s *DashboardService) calculateScore(salary, savings, spent, wants float64, saveCount int, ccSpend float64) int {
	score := 0

	// 1. Savings Health (Max 30)
	saveRate := (savings / salary) * 100
	if saveRate >= 20 {
		score += 30
	} else if saveRate >= 15 {
		score += 22
	} else if saveRate >= 10 {
		score += 15
	} else if saveRate > 0 {
		score += 7
	}

	// 2. Spending Discipline (Max 25)
	maxAllowed := salary * 0.80
	spendRatio := (spent / maxAllowed) * 100
	if spendRatio <= 90 {
		score += 25
	} else if spendRatio <= 100 {
		score += 20
	} else if spendRatio <= 110 {
		score += 10
	}

	// 3. Needs vs Wants Balance (Max 15)
	wantRatio := (wants / salary) * 100
	if wantRatio <= 30 {
		score += 15
	} else if wantRatio <= 40 {
		score += 10
	} else if wantRatio <= 50 {
		score += 5
	}

	// 4. Investment Consistency (Max 15)
	if saveCount >= 2 {
		score += 15
	} else if saveCount == 1 {
		score += 10
	}

	// 5. Cash Flow Stability (Max 10)
	pending := salary - (spent + savings)
	if pending > (salary * 0.05) {
		score += 10
	} else if pending > 0 {
		score += 5
	}

	// 6. Credit Reliance (Max 5)
	if spent > 0 {
		ccRatio := (ccSpend / spent) * 100
		if ccRatio <= 20 {
			score += 5
		} else if ccRatio <= 40 {
			score += 3
		}
	} else {
		score += 5 // If 0 spent, credit reliance is inherently 0 (perfect)
	}

	return score
}
