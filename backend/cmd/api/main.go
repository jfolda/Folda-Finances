package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"github.com/yourusername/folda-finances/internal/database"
	"github.com/yourusername/folda-finances/internal/handlers"
	authmiddleware "github.com/yourusername/folda-finances/internal/middleware"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Database configuration
	dbConfig := database.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", ""),
		DBName:   getEnv("DB_NAME", "folda_finances"),
		SSLMode:  getEnv("DB_SSL_MODE", "disable"),
	}

	// Connect to database
	db, err := database.Connect(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run migrations
	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Seed default categories
	if err := database.SeedDefaultCategories(db); err != nil {
		log.Fatalf("Failed to seed categories: %v", err)
	}

	// Initialize handlers
	userHandler := handlers.NewUserHandler(db)
	categoryHandler := handlers.NewCategoryHandler(db)
	accountHandler := handlers.NewAccountHandler(db)
	transactionHandler := handlers.NewTransactionHandler(db)
	spendingHandler := handlers.NewSpendingHandler(db)
	budgetHandler := handlers.NewBudgetHandler(db)
	incomeHandler := handlers.NewIncomeHandler(db)
	invitationHandler := handlers.NewInvitationHandler(db)

	// Initialize auth middleware
	jwtSecret := getEnv("SUPABASE_JWT_SECRET", "your-secret-key")
	supabaseURL := getEnv("SUPABASE_URL", "")
	authMiddleware := authmiddleware.NewAuthMiddleware(jwtSecret, supabaseURL)

	// Initialize router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:5173", "https://folda-finances.vercel.app"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check (public)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Auth endpoints (protected)
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)

			r.Route("/auth", func(r chi.Router) {
				r.Get("/me", userHandler.GetCurrentUser)
				r.Patch("/me", userHandler.UpdateUser)
			})

			// Spending endpoints (CORE FEATURE)
			r.Route("/spending", func(r chi.Router) {
				r.Get("/available", spendingHandler.GetSpendingAvailable)
			})

			// Category endpoints
			r.Route("/categories", func(r chi.Router) {
				r.Get("/", categoryHandler.GetCategories)
				r.Post("/", categoryHandler.CreateCategory)
			})

			// Account endpoints
			r.Route("/accounts", func(r chi.Router) {
				r.Get("/", accountHandler.ListAccounts)
				r.Post("/", accountHandler.CreateAccount)
				r.Get("/{id}", accountHandler.GetAccount)
				r.Put("/{id}", accountHandler.UpdateAccount)
				r.Delete("/{id}", accountHandler.DeleteAccount)
			})

			// Transaction endpoints
			r.Route("/transactions", func(r chi.Router) {
				r.Get("/", transactionHandler.ListTransactions)
				r.Post("/", transactionHandler.CreateTransaction)
				r.Get("/{id}", transactionHandler.GetTransaction)
				r.Put("/{id}", transactionHandler.UpdateTransaction)
				r.Delete("/{id}", transactionHandler.DeleteTransaction)
			})

			// Category budget endpoints
			r.Route("/category-budgets", func(r chi.Router) {
				r.Get("/", budgetHandler.ListCategoryBudgets)
				r.Post("/", budgetHandler.CreateCategoryBudget)
				r.Put("/{id}", budgetHandler.UpdateCategoryBudget)
				r.Delete("/{id}", budgetHandler.DeleteCategoryBudget)
				r.Get("/{id}/splits", budgetHandler.GetCategoryBudgetSplits)
				r.Put("/{id}/splits", budgetHandler.UpdateCategoryBudgetSplits)
			})

			// Budget member endpoints
			r.Get("/budget/members", budgetHandler.GetBudgetMembers)

			// Expected income endpoints
			r.Route("/expected-income", func(r chi.Router) {
				r.Get("/", incomeHandler.ListExpectedIncome)
				r.Post("/", incomeHandler.CreateExpectedIncome)
				r.Put("/{id}", incomeHandler.UpdateExpectedIncome)
				r.Delete("/{id}", incomeHandler.DeleteExpectedIncome)
			})

			// Budget invitation endpoints
			r.Route("/budgets/{budgetId}/invite", func(r chi.Router) {
				r.Post("/", invitationHandler.InviteToBudget)
			})

			r.Route("/budget-invitations", func(r chi.Router) {
				r.Get("/", invitationHandler.GetBudgetInvitations)
				r.Post("/{token}/accept", invitationHandler.AcceptBudgetInvitation)
				r.Post("/{token}/decline", invitationHandler.DeclineBudgetInvitation)
			})
		})
	})

	// Start server
	port := getEnv("PORT", "8080")
	log.Printf("🚀 Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
