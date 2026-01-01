package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/yourusername/folda-finances/internal/middleware"
	"github.com/yourusername/folda-finances/internal/models"
	"gorm.io/gorm"
)

type BudgetHandler struct {
	db *gorm.DB
}

func NewBudgetHandler(db *gorm.DB) *BudgetHandler {
	return &BudgetHandler{db: db}
}

type CreateCategoryBudgetRequest struct {
	CategoryID     string  `json:"category_id"`
	Amount         int     `json:"amount"` // monthly amount in cents
	AllocationType *string `json:"allocation_type"`
}

type UpdateCategoryBudgetRequest struct {
	Amount         *int    `json:"amount"`
	AllocationType *string `json:"allocation_type"`
}

func (h *BudgetHandler) ListCategoryBudgets(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch user"})
		return
	}

	if user.BudgetID == nil {
		respondJSON(w, http.StatusOK, map[string]interface{}{"data": []models.CategoryBudget{}})
		return
	}

	var budgets []models.CategoryBudget
	if err := h.db.Where("budget_id = ?", user.BudgetID).Find(&budgets).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch budgets"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"data": budgets})
}

func (h *BudgetHandler) CreateCategoryBudget(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req CreateCategoryBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch user"})
		return
	}

	if user.BudgetID == nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "user does not belong to a budget"})
		return
	}

	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid category_id"})
		return
	}

	allocationType := "pooled"
	if req.AllocationType != nil {
		allocationType = *req.AllocationType
	}

	budget := models.CategoryBudget{
		BudgetID:       *user.BudgetID,
		CategoryID:     categoryID,
		Amount:         req.Amount,
		AllocationType: allocationType,
	}

	if err := h.db.Create(&budget).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create budget"})
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"data":    budget,
		"message": "Budget created successfully",
	})
}

func (h *BudgetHandler) UpdateCategoryBudget(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	budgetID := chi.URLParam(r, "id")
	var budget models.CategoryBudget
	if err := h.db.First(&budget, "id = ?", budgetID).Error; err != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "budget not found"})
		return
	}

	// Verify user has access
	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch user"})
		return
	}

	if user.BudgetID == nil || *user.BudgetID != budget.BudgetID {
		respondJSON(w, http.StatusForbidden, map[string]string{"error": "access denied"})
		return
	}

	var req UpdateCategoryBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	updates := map[string]interface{}{}
	if req.Amount != nil {
		updates["amount"] = *req.Amount
	}
	if req.AllocationType != nil {
		updates["allocation_type"] = *req.AllocationType
	}

	if err := h.db.Model(&budget).Updates(updates).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update budget"})
		return
	}

	if err := h.db.First(&budget, "id = ?", budgetID).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch updated budget"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"data":    budget,
		"message": "Budget updated successfully",
	})
}

func (h *BudgetHandler) DeleteCategoryBudget(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	budgetID := chi.URLParam(r, "id")
	var budget models.CategoryBudget
	if err := h.db.First(&budget, "id = ?", budgetID).Error; err != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "budget not found"})
		return
	}

	// Verify user has access
	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch user"})
		return
	}

	if user.BudgetID == nil || *user.BudgetID != budget.BudgetID {
		respondJSON(w, http.StatusForbidden, map[string]string{"error": "access denied"})
		return
	}

	if err := h.db.Delete(&budget).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete budget"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Budget deleted successfully",
	})
}
