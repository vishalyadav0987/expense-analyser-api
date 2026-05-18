package analyzer

import (
	"context"
	"math"
	"sort"
	"time"

	"github.com/vishalyadav0987/expense-analyser/interfaces/http/dto"
	"github.com/vishalyadav0987/expense-analyser/internal/infrastructure/postgres"
)

type AnalyzerService struct {
	repo *postgres.AnalyzerRepository
}

func NewAnalyzerService(repo *postgres.AnalyzerRepository) *AnalyzerService {
	return &AnalyzerService{repo: repo}
}

func (s *AnalyzerService) GenerateWeeklyReport(ctx context.Context, userID string, startDate, endDate time.Time) (*dto.WeeklyAnalyzerResponse, error) {
	// 1. Fetch ALL transactions for the week in one single DB call
	rawTxns, err := s.repo.GetTransactionsForPeriod(ctx, userID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	report := &dto.WeeklyAnalyzerResponse{
		BarGraph:           make([]dto.DailyBar, 0),
		PieChartRules:      make([]dto.PieChartRule, 0),
		PieChartCategories: make([]dto.PieChartCategory, 0),
		DailySummaries:     make([]dto.DailySummaryDetail, 0),
	}

	// Internal trackers for aggregation
	var totalSpent, totalSaved float64
	typeTotals := make(map[string]float64)
	categoryTotals := make(map[string]float64)
	dailyMap := make(map[string]*dto.DailySummaryDetail)

	// 2. Aggregate the raw data
	for _, t := range rawTxns {
		dateStr := t.Date.Format("2006-01-02")

		// Track Totals
		if t.CategoryType == "Saving" {
			totalSaved += t.Amount
		} else {
			totalSpent += t.Amount
		}

		typeTotals[t.CategoryType] += t.Amount
		categoryTotals[t.CategoryName] += t.Amount

		// Group for Daily Summaries
		if _, exists := dailyMap[dateStr]; !exists {
			dailyMap[dateStr] = &dto.DailySummaryDetail{
				Date:              dateStr,
				DayName:           t.Date.Format("Mon"),
				TotalDailyExpense: 0,
				Transactions:      []dto.TransactionItem{},
			}
		}

		item := dto.TransactionItem{
			ID:          t.ID,
			Description: t.Description,
			Category:    t.CategoryName,
			Amount:      t.Amount,
			Type:        t.CategoryType,
			PaymentMode: t.PaymentMode,
		}

		dailyMap[dateStr].Transactions = append(dailyMap[dateStr].Transactions, item)
		if t.CategoryType != "Saving" { // Don't count savings in daily expense graph
			dailyMap[dateStr].TotalDailyExpense += t.Amount
		}
	}

	report.Summary.TotalSpent = totalSpent
	report.Summary.TotalSaved = totalSaved

	// 3. GENERATE BAR GRAPH (Fill missing days with 0)
	// We loop through the 7 days from StartDate to EndDate
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		amount := 0.0
		if summary, exists := dailyMap[dateStr]; exists {
			amount = summary.TotalDailyExpense
		}
		report.BarGraph = append(report.BarGraph, dto.DailyBar{
			Day:    d.Format("Mon"),
			Date:   dateStr,
			Amount: amount,
		})
	}

	// 4. GENERATE PIE CHARTS
	for typeName, amt := range typeTotals {
		// Avoid division by zero
		percentage := 0.0
		if typeName == "Saving" && amt > 0 {
			percentage = 100 // Savings is its own isolated pool usually
		} else if totalSpent > 0 {
			percentage = math.Round((amt/totalSpent)*1000) / 10 // Round to 1 decimal
		}

		report.PieChartRules = append(report.PieChartRules, dto.PieChartRule{
			Type:       typeName,
			Amount:     amt,
			Percentage: percentage,
		})
	}

	for catName, amt := range categoryTotals {
		report.PieChartCategories = append(report.PieChartCategories, dto.PieChartCategory{
			Category: catName,
			Amount:   amt,
		})
	}

	// 5. GENERATE DAILY SUMMARIES (Only for days with transactions, sorted newest first)
	for _, summary := range dailyMap {
		report.DailySummaries = append(report.DailySummaries, *summary)
	}

	// Sort descending by date (Frontend usually wants most recent first)
	sort.Slice(report.DailySummaries, func(i, j int) bool {
		return report.DailySummaries[i].Date > report.DailySummaries[j].Date
	})

	return report, nil
}
