package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vishalyadav0987/expense-analyser/interfaces/http/dto"
	app "github.com/vishalyadav0987/expense-analyser/internal/application/expense"
	"github.com/vishalyadav0987/expense-analyser/internal/domain/setup"
)

type ExpenseHandler struct {
	service *app.ExpenseService
}

func NewExpenseHandler(service *app.ExpenseService) *ExpenseHandler {
	return &ExpenseHandler{service: service}
}

func (h *ExpenseHandler) HandleCreateCategory(c *gin.Context) {
	// 1. Get UserID from Middleware
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("unauthorized"))
		return
	}

	// 2. Bind and Validate JSON
	var req dto.CreateExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("invalid payload: "+err.Error()))
		return
	}

	// 3. Call Service
	Expense, err := h.service.CreateCategory(
		c.Request.Context(),
		userID.(string),
		req.Name,
		setup.CategoryType(req.Type),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err.Error()))
		return
	}

	// 4. Return formatted response
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Expense created successfully.",
		"data": gin.H{
			"id":   Expense.ID,
			"name": Expense.Name,
			"type": Expense.Type,
		},
	})
}
