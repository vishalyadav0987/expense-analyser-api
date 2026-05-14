package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	// Adjust these import paths to match your actual project structure exactly
	"github.com/vishalyadav0987/expense-analyser/interfaces/http/dto"
	"github.com/vishalyadav0987/expense-analyser/internal/application/auth"
)

// AuthMiddleware creates a Gin middleware that validates JWT tokens.
// It takes the tokenProvider and the specific expectedAudience (e.g., "mpin_setup" or "api_access").
func AuthMiddleware(tokenProvider auth.TokenProvider, expectedAudience string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Extract the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.NewErrorResponse("Authorization header is required"))
			return
		}

		// 2. Validate the "Bearer <token>" format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.NewErrorResponse("Invalid authorization format. Expected 'Bearer <token>'"))
			return
		}

		tokenString := parts[1]

		// 3. Verify the token signature and audience using the provider
		userID, err := tokenProvider.VerifyToken(tokenString, expectedAudience)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.NewErrorResponse("Invalid or expired token: "+err.Error()))
			return
		}

		// 4. Attach the UserID to the Gin Context
		// This is the hand-off! Your HandleSetMPIN handler will extract this exact key later.
		c.Set("userID", userID)

		// 5. Pass control to the next handler
		c.Next()
	}
}
