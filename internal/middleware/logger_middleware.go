package middleware

import (
	"context"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid" // Run: go get github.com/google/uuid

	"github.com/vishalyadav0987/expense-analyser/pkg/logger"
)

// RequestLogger is an SDE3 grade middleware that traces HTTP traffic
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 1. Generate a unique Request-ID
		requestID := uuid.New().String()

		// 2. Inject it into the Gin context headers (so the Flutter app gets it)
		c.Writer.Header().Set("X-Request-ID", requestID)

		// 3. Inject it into the Go Context (so our database/service layers get it)
		// We use standard context.WithValue so it works outside of Gin
		ctx := context.WithValue(c.Request.Context(), logger.RequestIDKey, requestID)
		c.Request = c.Request.WithContext(ctx)

		// 4. Process the request
		c.Next()

		// 5. Calculate latency and get status
		latency := time.Since(start)
		statusCode := c.Writer.Status()

		// 6. Log the HTTP Request as strict JSON
		logArgs := []any{
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", statusCode),
			slog.String("ip", c.ClientIP()),
			slog.Duration("latency", latency),
		}

		// SDE3 Rule: 500s are Errors, 400s are Warnings, 200s are Info
		if statusCode >= 500 {
			logger.Error(ctx, "HTTP Request Failed", nil, logArgs...)
		} else if statusCode >= 400 {
			// e.g., Bad Request, Unauthorized
			logger.Info(ctx, "HTTP Client Error", logArgs...)
		} else {
			logger.Info(ctx, "HTTP Request Completed", logArgs...)
		}
	}
}
