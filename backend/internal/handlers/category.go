package handlers

import (
	"net/http"

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
