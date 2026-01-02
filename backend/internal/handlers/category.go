package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/yourusername/folda-finances/internal/middleware"
	"github.com/yourusername/folda-finances/internal/models"
	"gorm.io/gorm"
)

type CategoryHandler struct {
	db *gorm.DB
}

func NewCategoryHandler(db *gorm.DB) *CategoryHandler {
	return &CategoryHandler{db: db}
}

func (h *CategoryHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	// Get user to check their budget_id
	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch user"})
		return
	}

	var categories []models.Category

	// Get system categories (budget_id IS NULL) and user's budget categories
	query := h.db.Where("budget_id IS NULL AND is_system = true")
	if user.BudgetID != nil {
		query = query.Or("budget_id = ?", user.BudgetID)
	}

	if err := query.Order("name ASC").Find(&categories).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch categories"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"data": categories})
}

type CreateCategoryRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
	Icon  string `json:"icon"`
}

func (h *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	// Validate required fields
	if req.Name == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "category name is required"})
		return
	}
	if req.Icon == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "category icon is required"})
		return
	}
	if req.Color == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "category color is required"})
		return
	}

	// Get user to access their budget_id
	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch user"})
		return
	}

	if user.BudgetID == nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "user does not belong to a budget"})
		return
	}

	// Create the category
	category := models.Category{
		ID:       uuid.New(),
		BudgetID: user.BudgetID,
		Name:     req.Name,
		Color:    req.Color,
		Icon:     req.Icon,
		IsSystem: false,
	}

	if err := h.db.Create(&category).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create category"})
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"data":    category,
		"message": "Category created successfully",
	})
}
