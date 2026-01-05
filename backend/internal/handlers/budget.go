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

type CategoryBudgetSplitInput struct {
	UserID               string   `json:"user_id"`
	AllocationPercentage *float64 `json:"allocation_percentage"`
	AllocationAmount     *int     `json:"allocation_amount"`
}

type UpdateCategoryBudgetSplitsRequest struct {
	Splits []CategoryBudgetSplitInput `json:"splits"`
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

// GetCategoryBudgetSplits retrieves the split allocations for a category budget
func (h *BudgetHandler) GetCategoryBudgetSplits(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	categoryBudgetID := chi.URLParam(r, "id")
	var categoryBudget models.CategoryBudget
	if err := h.db.First(&categoryBudget, "id = ?", categoryBudgetID).Error; err != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "category budget not found"})
		return
	}

	// Verify user has access to this budget
	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch user"})
		return
	}

	if user.BudgetID == nil || *user.BudgetID != categoryBudget.BudgetID {
		respondJSON(w, http.StatusForbidden, map[string]string{"error": "access denied"})
		return
	}

	// Get all splits for this category budget
	var splits []models.CategoryBudgetSplit
	if err := h.db.Where("category_budget_id = ?", categoryBudgetID).Find(&splits).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch splits"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"data": splits})
}

// UpdateCategoryBudgetSplits updates or creates split allocations for a category budget
func (h *BudgetHandler) UpdateCategoryBudgetSplits(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	categoryBudgetID := chi.URLParam(r, "id")
	categoryBudgetUUID, err := uuid.Parse(categoryBudgetID)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid category budget id"})
		return
	}

	var categoryBudget models.CategoryBudget
	if err := h.db.First(&categoryBudget, "id = ?", categoryBudgetID).Error; err != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "category budget not found"})
		return
	}

	// Verify user has access to this budget
	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch user"})
		return
	}

	if user.BudgetID == nil || *user.BudgetID != categoryBudget.BudgetID {
		respondJSON(w, http.StatusForbidden, map[string]string{"error": "access denied"})
		return
	}

	var req UpdateCategoryBudgetSplitsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	// Validate splits
	if len(req.Splits) == 0 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "at least one split required"})
		return
	}

	// Verify all users belong to the same budget
	for _, split := range req.Splits {
		splitUserID, err := uuid.Parse(split.UserID)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user_id in splits"})
			return
		}

		var splitUser models.User
		if err := h.db.First(&splitUser, "id = ?", splitUserID).Error; err != nil {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "user not found in splits"})
			return
		}

		if splitUser.BudgetID == nil || *splitUser.BudgetID != categoryBudget.BudgetID {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "all users must belong to the same budget"})
			return
		}
	}

	// Delete existing splits
	if err := h.db.Where("category_budget_id = ?", categoryBudgetID).Delete(&models.CategoryBudgetSplit{}).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete existing splits"})
		return
	}

	// Create new splits
	var createdSplits []models.CategoryBudgetSplit
	for _, split := range req.Splits {
		splitUserID, _ := uuid.Parse(split.UserID)
		newSplit := models.CategoryBudgetSplit{
			CategoryBudgetID:     categoryBudgetUUID,
			UserID:               splitUserID,
			AllocationPercentage: split.AllocationPercentage,
			AllocationAmount:     split.AllocationAmount,
		}

		if err := h.db.Create(&newSplit).Error; err != nil {
			respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create split"})
			return
		}

		createdSplits = append(createdSplits, newSplit)
	}

	// Update category budget allocation type to 'split'
	if err := h.db.Model(&categoryBudget).Update("allocation_type", "split").Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update allocation type"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"data":    createdSplits,
		"message": "Splits updated successfully",
	})
}

// GetBudgetMembers returns all users who belong to the current user's budget
func (h *BudgetHandler) GetBudgetMembers(w http.ResponseWriter, r *http.Request) {
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
		respondJSON(w, http.StatusOK, map[string]interface{}{"data": []models.User{}})
		return
	}

	// Get all users in the same budget
	var members []models.User
	if err := h.db.Where("budget_id = ?", user.BudgetID).Find(&members).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch budget members"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"data": members})
}
