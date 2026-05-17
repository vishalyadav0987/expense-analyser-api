package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vishalyadav0987/expense-analyser/interfaces/http/dto"
	app "github.com/vishalyadav0987/expense-analyser/internal/application/expense"
	"github.com/vishalyadav0987/expense-analyser/internal/domain/expense"
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

func (h *ExpenseHandler) HandleAddExpense(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req dto.AddExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("invalid payload"))
		return
	}

	// Map to Domain Entity
	newExp := &expense.Expense{
		UserID:      userID.(string),
		Amount:      req.Amount,
		CategoryID:  req.CategoryID,
		Description: req.Description,
		PaymentMode: setup.PaymentMethodType(req.PaymentMode),
		Date:        req.Date,
	}

	// Call the Smart Rule Engine
	warning, err := h.service.ProcessNewExpense(c.Request.Context(), newExp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err.Error()))
		return
	}

	// Return the perfect JSON format requested by Frontend
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Expense added successfully.",
		"data": gin.H{
			"transactionId": newExp.ID,
			"limitWarning":  warning,
			"transaction":   newExp, // Since Category is a pointer inside newExp, it will render as a nested JSON object perfectly!
		},
	})
}
