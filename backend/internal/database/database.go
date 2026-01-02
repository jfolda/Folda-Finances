package database

import (
	"fmt"
	"log"

	"github.com/yourusername/folda-finances/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func Connect(config Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.DBName,
		config.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("✓ Database connection established")
	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	log.Println("Running database migrations...")

	// Enable UUID extension
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		return fmt.Errorf("failed to create uuid extension: %w", err)
	}

	// Run auto migrations
	err := db.AutoMigrate(
		&models.User{},
		&models.Budget{},
		&models.Category{},
		&models.Account{},
		&models.Transaction{},
		&models.CategoryBudget{},
		&models.CategoryBudgetSplit{},
		&models.ExpectedIncome{},
		&models.BudgetInvitation{},
	)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("✓ Database migrations completed")
	return nil
}

func SeedDefaultCategories(db *gorm.DB) error {
	log.Println("Seeding default categories...")

	categories := []models.Category{
		// Expenses
		{Name: "Housing", Color: "#8B5CF6", Icon: "🏠", IsSystem: true},
		{Name: "Utilities", Color: "#3B82F6", Icon: "⚡", IsSystem: true},
		{Name: "Groceries", Color: "#10B981", Icon: "🛒", IsSystem: true},
		{Name: "Dining & Restaurants", Color: "#F59E0B", Icon: "🍽️", IsSystem: true},
		{Name: "Transportation", Color: "#EF4444", Icon: "🚗", IsSystem: true},
		{Name: "Healthcare", Color: "#EC4899", Icon: "🏥", IsSystem: true},
		{Name: "Entertainment", Color: "#6366F1", Icon: "🎬", IsSystem: true},
		{Name: "Shopping", Color: "#8B5CF6", Icon: "🛍️", IsSystem: true},
		{Name: "Personal Care", Color: "#14B8A6", Icon: "💆", IsSystem: true},
		{Name: "Education", Color: "#F97316", Icon: "📚", IsSystem: true},
		{Name: "Subscriptions", Color: "#A855F7", Icon: "📱", IsSystem: true},
		{Name: "Insurance", Color: "#06B6D4", Icon: "🛡️", IsSystem: true},
		{Name: "Savings", Color: "#22C55E", Icon: "💰", IsSystem: true},
		{Name: "Debt Payments", Color: "#DC2626", Icon: "💳", IsSystem: true},
		{Name: "Gifts & Donations", Color: "#F472B6", Icon: "🎁", IsSystem: true},
		{Name: "Miscellaneous", Color: "#6B7280", Icon: "📦", IsSystem: true},

		// Income
		{Name: "Salary", Color: "#059669", Icon: "💵", IsSystem: true},
		{Name: "Freelance", Color: "#0891B2", Icon: "💼", IsSystem: true},
		{Name: "Investments", Color: "#7C3AED", Icon: "📈", IsSystem: true},
		{Name: "Other Income", Color: "#84CC16", Icon: "💸", IsSystem: true},
	}

	for _, category := range categories {
		var existing models.Category
		result := db.Where("name = ? AND is_system = true AND budget_id IS NULL", category.Name).First(&existing)

		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&category).Error; err != nil {
				return fmt.Errorf("failed to seed category %s: %w", category.Name, err)
			}
		}
	}

	log.Println("✓ Default categories seeded")
	return nil
}
