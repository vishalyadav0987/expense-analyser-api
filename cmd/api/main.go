package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/vishalyadav0987/expense-analyser/db/connect"
	routes "github.com/vishalyadav0987/expense-analyser/interfaces/http"
	"github.com/vishalyadav0987/expense-analyser/interfaces/http/handlers"
	"github.com/vishalyadav0987/expense-analyser/internal/application/analyzer"
	"github.com/vishalyadav0987/expense-analyser/internal/application/auth"
	"github.com/vishalyadav0987/expense-analyser/internal/application/dashboard"
	expense "github.com/vishalyadav0987/expense-analyser/internal/application/expense"
	"github.com/vishalyadav0987/expense-analyser/internal/application/setup"
	"github.com/vishalyadav0987/expense-analyser/internal/config"
	"github.com/vishalyadav0987/expense-analyser/internal/infrastructure/email"
	"github.com/vishalyadav0987/expense-analyser/internal/infrastructure/jwt"
	"github.com/vishalyadav0987/expense-analyser/internal/infrastructure/postgres"
	redisInfra "github.com/vishalyadav0987/expense-analyser/internal/infrastructure/redis"
	"github.com/vishalyadav0987/expense-analyser/pkg/logger"
)

func main() {

	// ------------------------------------------------------------------
	// 1. Configuration & Secrets (In production, load from .env)
	// ------------------------------------------------------------------

	cfg := config.MustLoad() // Loads the .env file

	// ------------------------------------------------------------------
	// 2. Initialize Database Connections (The "Duffer" Check)
	// ------------------------------------------------------------------

	// Connect to PostgreSQL
	db, err := connect.NewConnection(cfg.DBConnectionString)
	if err != nil {
		log.Fatalf("❌ Fatal DB Error: %v", err)
	}
	defer db.Close()

	// Connect to Redis
	rdb, err := connect.NewClient(cfg.RedisAddr)
	if err != nil {
		log.Fatalf("❌ Fatal Redis DB Error: %v", err)
	}
	defer rdb.Close()

	logger.Setup()

	// ------------------------------------------------------------------
	// 3. Dependency Injection (Wiring the Layers)
	// ------------------------------------------------------------------

	// A. Infrastructure Layer (Adapters)
	userRepo := postgres.NewUserRepository(db)
	otpRepo := redisInfra.NewOTPRepository(rdb)
	tokenProvider := jwt.NewTokenProvider(cfg.JWTSecret, "expense_app")
	emailProvider := email.NewSMTPEmailService()
	securityService := redisInfra.NewSecurityRepository(rdb)

	setupRepo := postgres.NewSetupRepository(db)
	setupService := setup.NewSetupService(setupRepo, userRepo)

	epxenseRepo := postgres.NewExpenseRepository(db)
	expenseService := expense.NewExpenseService(epxenseRepo, setupRepo)

	analyzerRepo := postgres.NewAnalyzerRepository(db)
	analyzerService := analyzer.NewAnalyzerService(analyzerRepo)

	dashboardService := dashboard.NewDashboardService(epxenseRepo, setupRepo)

	tokenRepo := redisInfra.NewTokenRepository(rdb)

	// B. Application Layer (The Brain)
	authService := auth.NewService(userRepo, otpRepo, tokenProvider, emailProvider, securityService, tokenRepo)

	// C. Delivery Layer (HTTP Handlers)
	authHandler := handlers.NewAuthHandler(authService, tokenProvider)
	setupHandler := handlers.NewSetupHandler(setupService)
	expenseHandler := handlers.NewExpenseHandler(expenseService)
	analyzerHandler := handlers.NewAnalyzerHandler(analyzerService)
	dashboardHandler := handlers.NewDashboardHandler(*dashboardService)

	// ------------------------------------------------------------------
	// 3. Setup Gin Router
	// ------------------------------------------------------------------

	// Set Gin to release mode in production: gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// Call our new Routing Layer
	routes.SetupRouter(router, authHandler, setupHandler, expenseHandler, analyzerHandler, dashboardHandler, tokenProvider)

	// ------------------------------------------------------------------
	// 5. Start Server with Graceful Shutdown
	// ------------------------------------------------------------------
	srv := &http.Server{
		Addr:    ":" + cfg.AppPort,
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
