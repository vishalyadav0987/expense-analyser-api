package dto

type WeeklyAnalyzerResponse struct {
	Summary            AnalyzerSummary      `json:"summary"`
	BarGraph           []DailyBar           `json:"barGraph"`
	PieChartRules      []PieChartRule       `json:"pieChartRules"`
	PieChartCategories []PieChartCategory   `json:"pieChartCategories"`
	DailySummaries     []DailySummaryDetail `json:"dailySummaries"`
}

type AnalyzerSummary struct {
	TotalSpent float64 `json:"totalSpent"`
	TotalSaved float64 `json:"totalSaved"`
}

type DailyBar struct {
	Day    string  `json:"day"`
	Date   string  `json:"date"`
	Amount float64 `json:"amount"`
}

type PieChartRule struct {
	Type       string  `json:"type"`
	Amount     float64 `json:"amount"`
	Percentage float64 `json:"percentage"`
}

type PieChartCategory struct {
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
}

type DailySummaryDetail struct {
	Date              string            `json:"date"`
	DayName           string            `json:"dayName"`
	TotalDailyExpense float64           `json:"totalDailyExpense"`
	Transactions      []TransactionItem `json:"transactions"`
}

type TransactionItem struct {
	ID          string  `json:"id"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Amount      float64 `json:"amount"`
	Type        string  `json:"type"`
	PaymentMode string  `json:"paymentMode"`
}
