package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/vishalyadav0987/expense-analyser/interfaces/http/handlers"
	"github.com/vishalyadav0987/expense-analyser/internal/application/auth"
	"github.com/vishalyadav0987/expense-analyser/internal/middleware"
)

// SetupRouter organizes all endpoints and middleware for the application.
// Passing the Gin engine allows us to inject dependencies cleanly.
func SetupRouter(
	router *gin.Engine,
	authHandler *handlers.AuthHandler,
	setupHandler *handlers.SetupHandler,
	expenseHandler *handlers.ExpenseHandler,
	tokenProvider auth.TokenProvider,
) {

	// API Versioning Group
	apiAuthMiddleware := middleware.AuthMiddleware(tokenProvider, "api_access")
	v1 := router.Group("/api/v1")
	{
		// ------------------------------------------------------------------
		// Public Routes (No Auth Required)
		// ------------------------------------------------------------------
		auth := v1.Group("/auth")
		{
			auth.POST("/request-otp", authHandler.HandleRequestOTP)
			auth.POST("/verify-otp", authHandler.HandleVerifyOTP)
			auth.POST("/mpin-login", authHandler.HandleLoginMPIN)
			auth.POST("/biometric-login", authHandler.HandleBiometricLogin)
		}

		// ------------------------------------------------------------------
		// Protected Routes (Requires JWT Middleware)
		// ------------------------------------------------------------------
		secureAuth := auth.Group("/")
		{
			// This route requires the user to pass the OTPAccessToken
			secureAuth.POST("/set-mpin", authHandler.HandleSetMPIN)
		}

		userRoutes := v1.Group("/user")
		userRoutes.Use(apiAuthMiddleware)
		{
			// POST /api/v1/user/setup
			userRoutes.POST("/setup", setupHandler.HandleSetupProfile)
		}

		expenseRoutes := v1.Group("/expense")
		expenseRoutes.Use(apiAuthMiddleware)
		{
			expenseRoutes.POST("/create-category", expenseHandler.HandleCreateCategory)
		}

	}
}
