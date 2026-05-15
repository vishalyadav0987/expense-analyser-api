package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vishalyadav0987/expense-analyser/interfaces/http/dto"
	app "github.com/vishalyadav0987/expense-analyser/internal/application/setup"
	domain "github.com/vishalyadav0987/expense-analyser/internal/domain/setup"
)

type SetupHandler struct {
	service *app.SetupService
}

func NewSetupHandler(service *app.SetupService) *SetupHandler {
	return &SetupHandler{service: service}
}

func (h *SetupHandler) HandleSetupProfile(c *gin.Context) {
	// 1. Get the UserID from the Gin Context (put there by your AuthMiddleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("unauthorized: missing user context"))
		return
	}

	// 2. Parse the JSON from Flutter
	var req dto.SetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("invalid payload: "+err.Error()))
		return
	}

	// 3. Map DTO to Domain Entity
	profile := &domain.UserInitialSetup{
		UserID: userID.(string),
		Financials: domain.Financials{
			MonthlySalary:        req.Financials.MonthlySalary,
			YearlyHikePercentage: req.Financials.YearlyHikePercentage,
		},
		SmartRules: domain.SmartRules{
			NeedsPercentage:   req.SmartRules.NeedsPercentage,
			WantsPercentage:   req.SmartRules.WantsPercentage,
			SavingsPercentage: req.SmartRules.SavingsPercentage,
		},
	}

	for _, cat := range req.Categories {
		profile.Categories = append(profile.Categories, domain.Category{
			Name: cat.Name,
			Type: domain.CategoryType(cat.Type),
		})
	}

	for _, pm := range req.PaymentMethods {
		profile.PaymentMethods = append(profile.PaymentMethods, domain.PaymentMethod{
			MethodName:  domain.PaymentMethodType(pm.MethodName),
			WeeklyLimit: pm.WeeklyLimit,
			IsActive:    pm.IsActive,
		})
	}

	// 4. Call the Service
	if err := h.service.ProcessInitialSetup(c.Request.Context(), profile); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error()))
		return
	}

	// 5. Return exactly what Flutter requested (the generated IDs)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User profile setup completed successfully.",
		"data": gin.H{
			"user":                gin.H{"id": profile.UserID, "setupCompleted": true},
			"smartRules":          profile.SmartRules,
			"savedCategories":     profile.Categories,
			"savedPaymentMethods": profile.PaymentMethods,
		},
	})
}
