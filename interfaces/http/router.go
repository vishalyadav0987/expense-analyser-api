package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/vishalyadav0987/expense-analyser/interfaces/http/handlers"
)

// SetupRouter organizes all endpoints and middleware for the application.
// Passing the Gin engine allows us to inject dependencies cleanly.
func SetupRouter(router *gin.Engine, authHandler *handlers.AuthHandler) {

	// API Versioning Group
	v1 := router.Group("/api/v1")
	{
		// ------------------------------------------------------------------
		// Public Routes (No Auth Required)
		// ------------------------------------------------------------------
		auth := v1.Group("/auth")
		{
			auth.POST("/request-otp", authHandler.HandleRequestOTP)
			auth.POST("/verify-otp", authHandler.HandleVerifyOTP)
		}

		// ------------------------------------------------------------------
		// Protected Routes (Requires JWT Middleware)
		// ------------------------------------------------------------------
		// SDE3 Setup: Later we will add: secureAuth := auth.Group("/").Use(middleware.RequireJWT())
		secureAuth := auth.Group("/")
		{
			// This route requires the user to pass the OTPAccessToken
			secureAuth.POST("/submit-mpin", authHandler.HandleSubmitMPIN)
		}

		// Future implementation:
		// expenses := v1.Group("/expenses").Use(middleware.RequireJWT())
		// {
		// 	expenses.POST("/", expenseHandler.CreateExpense)
		// 	expenses.GET("/", expenseHandler.GetRecentExpenses)
		// }
	}
}
