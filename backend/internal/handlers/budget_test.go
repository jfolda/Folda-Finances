package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/yourusername/folda-finances/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto-migrate all models
	err = db.AutoMigrate(
		&models.User{},
		&models.Budget{},
		&models.Category{},
		&models.CategoryBudget{},
		&models.CategoryBudgetSplit{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

// createTestUser creates a test user with a budget
func createTestUser(t *testing.T, db *gorm.DB, email string) (*models.User, *models.Budget) {
	userID := uuid.New()

	// Create budget first
	budget := &models.Budget{
		ID:        uuid.New(),
		Name:      "Test Budget",
		CreatedBy: userID,
	}
	if err := db.Create(budget).Error; err != nil {
		t.Fatalf("Failed to create test budget: %v", err)
	}

	// Create user
	now := time.Now()
	user := &models.User{
		ID:              userID,
		Email:           email,
		Name:            "Test User",
		BudgetID:        &budget.ID,
		ViewPeriod:      "monthly",
		PeriodStartDate: &now,
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return user, budget
}

// createTestCategory creates a test category
func createTestCategory(t *testing.T, db *gorm.DB, budgetID uuid.UUID, name string) *models.Category {
	category := &models.Category{
		ID:       uuid.New(),
		BudgetID: &budgetID,
		Name:     name,
		Color:    "#FF0000",
		Icon:     "🍔",
	}
	if err := db.Create(category).Error; err != nil {
		t.Fatalf("Failed to create test category: %v", err)
	}
	return category
}

func TestGetBudgetMembers(t *testing.T) {
	db := setupTestDB(t)
	handler := NewBudgetHandler(db)

	user1, budget := createTestUser(t, db, "user1@example.com")
	user2ID := uuid.New()
	user2 := &models.User{
		ID:              user2ID,
		Email:           "user2@example.com",
		Name:            "Test User 2",
		BudgetID:        &budget.ID,
		ViewPeriod:      "monthly",
		PeriodStartDate: func() *time.Time { t := time.Now(); return &t }(),
	}
	db.Create(user2)

	// Create request
	req := httptest.NewRequest("GET", "/budget/members", nil)
	req = req.WithContext(setUserIDContext(req, user1.ID))
	w := httptest.NewRecorder()

	handler.GetBudgetMembers(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	members, ok := response["data"].([]interface{})
	if !ok {
		t.Fatal("Expected data to be an array")
	}

	if len(members) != 2 {
		t.Errorf("Expected 2 members, got %d", len(members))
	}
}

func TestCreateCategoryBudget(t *testing.T) {
	db := setupTestDB(t)
	handler := NewBudgetHandler(db)

	user, budget := createTestUser(t, db, "test@example.com")
	category := createTestCategory(t, db, budget.ID, "Food")

	reqBody := CreateCategoryBudgetRequest{
		CategoryID:     category.ID.String(),
		Amount:         50000, // $500
		AllocationType: stringPtr("pooled"),
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/category-budgets", bytes.NewBuffer(body))
	req = req.WithContext(setUserIDContext(req, user.ID))
	w := httptest.NewRecorder()

	handler.CreateCategoryBudget(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	data, ok := response["data"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected data object in response")
	}

	if data["amount"].(float64) != 50000 {
		t.Errorf("Expected amount 50000, got %v", data["amount"])
	}

	if data["allocation_type"].(string) != "pooled" {
		t.Errorf("Expected allocation_type 'pooled', got %v", data["allocation_type"])
	}
}

func TestUpdateCategoryBudgetSplits(t *testing.T) {
	db := setupTestDB(t)
	handler := NewBudgetHandler(db)

	user1, budget := createTestUser(t, db, "user1@example.com")
	user2ID := uuid.New()
	user2 := &models.User{
		ID:              user2ID,
		Email:           "user2@example.com",
		Name:            "Test User 2",
		BudgetID:        &budget.ID,
		ViewPeriod:      "monthly",
		PeriodStartDate: func() *time.Time { t := time.Now(); return &t }(),
	}
	db.Create(user2)

	category := createTestCategory(t, db, budget.ID, "Food")

	// Create category budget
	categoryBudget := &models.CategoryBudget{
		ID:             uuid.New(),
		BudgetID:       budget.ID,
		CategoryID:     category.ID,
		Amount:         100000, // $1000
		AllocationType: "pooled",
	}
	db.Create(categoryBudget)

	// Create split request
	reqBody := UpdateCategoryBudgetSplitsRequest{
		Splits: []CategoryBudgetSplitInput{
			{
				UserID:           user1.ID.String(),
				AllocationAmount: intPtr(60000), // $600
			},
			{
				UserID:           user2.ID.String(),
				AllocationAmount: intPtr(40000), // $400
			},
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/category-budgets/"+categoryBudget.ID.String()+"/splits", bytes.NewBuffer(body))
	req = req.WithContext(setUserIDContext(req, user1.ID))

	// Add URL parameter
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", categoryBudget.ID.String())
	req = req.WithContext(setRouteContext(req, rctx))

	w := httptest.NewRecorder()

	handler.UpdateCategoryBudgetSplits(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify splits were created
	var splits []models.CategoryBudgetSplit
	db.Where("category_budget_id = ?", categoryBudget.ID).Find(&splits)

	if len(splits) != 2 {
		t.Errorf("Expected 2 splits, got %d", len(splits))
	}

	// Verify category budget allocation type updated to 'split'
	var updatedBudget models.CategoryBudget
	db.First(&updatedBudget, categoryBudget.ID)

	if updatedBudget.AllocationType != "split" {
		t.Errorf("Expected allocation_type 'split', got '%s'", updatedBudget.AllocationType)
	}
}

func TestUpdateCategoryBudgetSplits_ValidationErrors(t *testing.T) {
	db := setupTestDB(t)
	handler := NewBudgetHandler(db)

	user1, budget := createTestUser(t, db, "user1@example.com")

	// Create a second user NOT in the same budget
	otherBudget := &models.Budget{
		ID:        uuid.New(),
		Name:      "Other Budget",
		CreatedBy: uuid.New(),
	}
	db.Create(otherBudget)

	user2ID := uuid.New()
	user2 := &models.User{
		ID:              user2ID,
		Email:           "user2@example.com",
		Name:            "Test User 2",
		BudgetID:        &otherBudget.ID, // Different budget!
		ViewPeriod:      "monthly",
		PeriodStartDate: func() *time.Time { t := time.Now(); return &t }(),
	}
	db.Create(user2)

	category := createTestCategory(t, db, budget.ID, "Food")
	categoryBudget := &models.CategoryBudget{
		ID:             uuid.New(),
		BudgetID:       budget.ID,
		CategoryID:     category.ID,
		Amount:         100000,
		AllocationType: "pooled",
	}
	db.Create(categoryBudget)

	// Try to create split with user from different budget
	reqBody := UpdateCategoryBudgetSplitsRequest{
		Splits: []CategoryBudgetSplitInput{
			{
				UserID:           user1.ID.String(),
				AllocationAmount: intPtr(60000),
			},
			{
				UserID:           user2.ID.String(), // Different budget!
				AllocationAmount: intPtr(40000),
			},
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/category-budgets/"+categoryBudget.ID.String()+"/splits", bytes.NewBuffer(body))
	req = req.WithContext(setUserIDContext(req, user1.ID))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", categoryBudget.ID.String())
	req = req.WithContext(setRouteContext(req, rctx))

	w := httptest.NewRecorder()

	handler.UpdateCategoryBudgetSplits(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for cross-budget split, got %d", w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["error"] != "all users must belong to the same budget" {
		t.Errorf("Expected error message about same budget, got: %v", response["error"])
	}
}

func TestGetCategoryBudgetSplits(t *testing.T) {
	db := setupTestDB(t)
	handler := NewBudgetHandler(db)

	user1, budget := createTestUser(t, db, "user1@example.com")
	user2ID := uuid.New()
	user2 := &models.User{
		ID:              user2ID,
		Email:           "user2@example.com",
		Name:            "Test User 2",
		BudgetID:        &budget.ID,
		ViewPeriod:      "monthly",
		PeriodStartDate: func() *time.Time { t := time.Now(); return &t }(),
	}
	db.Create(user2)

	category := createTestCategory(t, db, budget.ID, "Food")
	categoryBudget := &models.CategoryBudget{
		ID:             uuid.New(),
		BudgetID:       budget.ID,
		CategoryID:     category.ID,
		Amount:         100000,
		AllocationType: "split",
	}
	db.Create(categoryBudget)

	// Create splits
	split1 := &models.CategoryBudgetSplit{
		ID:               uuid.New(),
		CategoryBudgetID: categoryBudget.ID,
		UserID:           user1.ID,
		AllocationAmount: intPtr(60000),
	}
	split2 := &models.CategoryBudgetSplit{
		ID:               uuid.New(),
		CategoryBudgetID: categoryBudget.ID,
		UserID:           user2.ID,
		AllocationAmount: intPtr(40000),
	}
	db.Create(split1)
	db.Create(split2)

	// Get splits
	req := httptest.NewRequest("GET", "/category-budgets/"+categoryBudget.ID.String()+"/splits", nil)
	req = req.WithContext(setUserIDContext(req, user1.ID))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", categoryBudget.ID.String())
	req = req.WithContext(setRouteContext(req, rctx))

	w := httptest.NewRecorder()

	handler.GetCategoryBudgetSplits(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	splits, ok := response["data"].([]interface{})
	if !ok {
		t.Fatal("Expected data to be an array")
	}

	if len(splits) != 2 {
		t.Errorf("Expected 2 splits, got %d", len(splits))
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}
