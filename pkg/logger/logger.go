package logger

import (
	"context"
	"log/slog"
	"os"
)

// Define context keys to avoid string collision
type contextKey string

const (
	RequestIDKey contextKey = "request_id"
	UserIDKey    contextKey = "user_id"
)

// Setup configures the global slog logger to output JSON.
// In production, you call this once in main.go.
func Setup() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo, // Use LevelDebug for local, LevelInfo for prod
	}

	// Output logs as strict JSON to standard out
	handler := slog.NewJSONHandler(os.Stdout, opts)

	// Set the default logger so standard library logs also get formatted
	slog.SetDefault(slog.New(handler))
}

// Info, Error, Debug wrappers
// SDE3 Magic: These functions take a context.Context. They automatically extract
// the RequestID and UserID (if present) and inject them into the JSON log.
func Info(ctx context.Context, msg string, args ...any) {
	slog.Info(msg, appendContextArgs(ctx, args)...)
}

func Error(ctx context.Context, msg string, err error, args ...any) {
	args = append(args, slog.String("error", err.Error()))
	slog.Error(msg, appendContextArgs(ctx, args)...)
}

func Debug(ctx context.Context, msg string, args ...any) {
	slog.Debug(msg, appendContextArgs(ctx, args)...)
}

// appendContextArgs intercepts the log payload and injects tracing metadata
func appendContextArgs(ctx context.Context, args []any) []any {
	if ctx == nil {
		return args
	}

	if reqID, ok := ctx.Value(RequestIDKey).(string); ok {
		args = append(args, slog.String("request_id", reqID))
	}

	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		args = append(args, slog.String("user_id", userID))
	}

	return args
}
