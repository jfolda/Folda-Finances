package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/yourusername/folda-finances/internal/middleware"
	"github.com/yourusername/folda-finances/internal/models"
	"gorm.io/gorm"
)

type IncomeHandler struct {
	db *gorm.DB
}

func NewIncomeHandler(db *gorm.DB) *IncomeHandler {
	return &IncomeHandler{db: db}
}

type CreateExpectedIncomeRequest struct {
	Name      string `json:"name"`
	Amount    int    `json:"amount"`
	Frequency string `json:"frequency"`
	NextDate  string `json:"next_date"`
}

type UpdateExpectedIncomeRequest struct {
	Name      *string `json:"name"`
	Amount    *int    `json:"amount"`
	Frequency *string `json:"frequency"`
	NextDate  *string `json:"next_date"`
	IsActive  *bool   `json:"is_active"`
}

func (h *IncomeHandler) ListExpectedIncome(w http.ResponseWriter, r *http.Request) {
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
		respondJSON(w, http.StatusOK, map[string]interface{}{"data": []models.ExpectedIncome{}})
		return
	}

	var incomes []models.ExpectedIncome
	if err := h.db.Where("budget_id = ?", user.BudgetID).Order("next_date ASC").Find(&incomes).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch expected income"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"data": incomes})
}

func (h *IncomeHandler) CreateExpectedIncome(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req CreateExpectedIncomeRequest
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

	nextDate, err := time.Parse("2006-01-02", req.NextDate)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid date format"})
		return
	}

	income := models.ExpectedIncome{
		BudgetID:  *user.BudgetID,
		Name:      req.Name,
		Amount:    req.Amount,
		Frequency: req.Frequency,
		NextDate:  nextDate,
		IsActive:  true,
	}

	if err := h.db.Create(&income).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create expected income"})
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"data":    income,
		"message": "Expected income created successfully",
	})
}

func (h *IncomeHandler) UpdateExpectedIncome(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	incomeID := chi.URLParam(r, "id")
	var income models.ExpectedIncome
	if err := h.db.First(&income, "id = ?", incomeID).Error; err != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "expected income not found"})
		return
	}

	// Verify user has access
	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch user"})
		return
	}

	if user.BudgetID == nil || *user.BudgetID != income.BudgetID {
		respondJSON(w, http.StatusForbidden, map[string]string{"error": "access denied"})
		return
	}

	var req UpdateExpectedIncomeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Amount != nil {
		updates["amount"] = *req.Amount
	}
	if req.Frequency != nil {
		updates["frequency"] = *req.Frequency
	}
	if req.NextDate != nil {
		nextDate, err := time.Parse("2006-01-02", *req.NextDate)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid date format"})
			return
		}
		updates["next_date"] = nextDate
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if err := h.db.Model(&income).Updates(updates).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update expected income"})
		return
	}

	if err := h.db.First(&income, "id = ?", incomeID).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch updated income"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"data":    income,
		"message": "Expected income updated successfully",
	})
}

func (h *IncomeHandler) DeleteExpectedIncome(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	incomeID := chi.URLParam(r, "id")
	var income models.ExpectedIncome
	if err := h.db.First(&income, "id = ?", incomeID).Error; err != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "expected income not found"})
		return
	}

	// Verify user has access
	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch user"})
		return
	}

	if user.BudgetID == nil || *user.BudgetID != income.BudgetID {
		respondJSON(w, http.StatusForbidden, map[string]string{"error": "access denied"})
		return
	}

	if err := h.db.Delete(&income).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete expected income"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Expected income deleted successfully",
	})
}
