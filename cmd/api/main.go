package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // Postgres driver

	// Adjust these imports to match your actual module name
	routes "github.com/vishalyadav0987/expense-analyser/interfaces/http"
	"github.com/vishalyadav0987/expense-analyser/interfaces/http/handlers"
	"github.com/vishalyadav0987/expense-analyser/internal/application/auth"
	"github.com/vishalyadav0987/expense-analyser/internal/infrastructure/email"
	"github.com/vishalyadav0987/expense-analyser/internal/infrastructure/jwt"
	"github.com/vishalyadav0987/expense-analyser/internal/infrastructure/postgres"
	redisInfra "github.com/vishalyadav0987/expense-analyser/internal/infrastructure/redis"
	"github.com/vishalyadav0987/expense-analyser/pkg/logger"
)

func main() {

	_ = godotenv.Load() // Loads the .env file

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)
	// ------------------------------------------------------------------
	// 1. Configuration & Secrets (In production, load from .env)
	// ------------------------------------------------------------------
	redisAddr := "localhost:6379"
	jwtSecret := "super_secret_key_change_in_production"

	// ------------------------------------------------------------------
	// 2. Initialize Database Connections (The "Duffer" Check)
	// ------------------------------------------------------------------

	// Connect to PostgreSQL
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("❌ Failed to connect to PostgreSQL: %v", err)
	}
	// SDE3 Connection Pool Tuning
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)
	log.Println("✅ Connected to PostgreSQL")

	// Connect to Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("❌ Failed to connect to Redis: %v", err)
	}
	log.Println("✅ Connected to Redis")

	logger.Setup()

	// ------------------------------------------------------------------
	// 3. Dependency Injection (Wiring the Layers)
	// ------------------------------------------------------------------

	// A. Infrastructure Layer (Adapters)
	userRepo := postgres.NewUserRepository(db)
	otpRepo := redisInfra.NewOTPRepository(rdb)
	tokenProvider := jwt.NewTokenProvider(jwtSecret, "expense_app")
	emailProvider := email.NewSMTPEmailService()

	// B. Application Layer (The Brain)
	authService := auth.NewService(userRepo, otpRepo, tokenProvider, emailProvider)

	// C. Delivery Layer (HTTP Handlers)
	authHandler := handlers.NewAuthHandler(authService, tokenProvider)

	// ------------------------------------------------------------------
	// 3. Setup Gin Router
	// ------------------------------------------------------------------

	// Set Gin to release mode in production: gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// Call our new Routing Layer
	routes.SetupRouter(router, authHandler)

	// ------------------------------------------------------------------
	// 5. Start Server with Graceful Shutdown
	// ------------------------------------------------------------------
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
		// SDE3 Security: Prevent slow-loris attacks by enforcing timeouts
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start the server in a goroutine so it doesn't block
	go func() {
		log.Println("🚀 Server starting on port 8080...")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Server failed: %v", err)
		}
	}()

	// ------------------------------------------------------------------
	// Wait for Interrupt Signal (Ctrl+C or Docker stop)
	// ------------------------------------------------------------------
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 Shutting down server...")

	// Give active connections 5 seconds to finish before killing them
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("❌ Server forced to shutdown: %v", err)
	}

	// Clean up database connections
	db.Close()
	rdb.Close()
	log.Println("✅ Server exited gracefully")
}
