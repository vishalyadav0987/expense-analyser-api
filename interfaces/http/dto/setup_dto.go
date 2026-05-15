package dto

// REQUEST DTOs
type SetupRequest struct {
	Financials FinancialsDTO `json:"financials" binding:"required"`
	SmartRules SmartRulesDTO `json:"smartRules" binding:"required"`
	// SDE3 Tip: 'dive' tells Gin to go inside the array and validate every single item!
	Categories     []CategoryDTO      `json:"categories" binding:"required,min=1,dive"`
	PaymentMethods []PaymentMethodDTO `json:"paymentMethods" binding:"required,min=1,dive"`
}

type FinancialsDTO struct {
	MonthlySalary        float64 `json:"monthlySalary" binding:"required,gt=0"` // Must be Greater Than (gt) 0
	YearlyHikePercentage float64 `json:"yearlyHikePercentage" binding:"gte=0"`  // Greater Than or Equal (gte) 0
	XXWeeklyLimit        float64 `json:"xxWeeklyLimit" binding:"gte=0"`
}

type SmartRulesDTO struct {
	NeedsPercentage   int `json:"needsPercentage" binding:"required"`
	WantsPercentage   int `json:"wantsPercentage" binding:"required"`
	SavingsPercentage int `json:"savingsPercentage" binding:"required"`
}

type CategoryDTO struct {
	Name string `json:"name" binding:"required"`
	Type string `json:"type" binding:"required,oneof=Need Want Saving"` // Strictly enforces spelling!
}

type PaymentMethodDTO struct {
	MethodName  string  `json:"methodName" binding:"required"`
	WeeklyLimit float64 `json:"weeklyLimit" binding:"gte=0"`
	IsActive    bool    `json:"isActive"`
}
