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

type AccountHandler struct {
	db *gorm.DB
}

func NewAccountHandler(db *gorm.DB) *AccountHandler {
	return &AccountHandler{db: db}
}

type CreateAccountRequest struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Balance  int    `json:"balance"`
	Currency string `json:"currency"`
	Notes    string `json:"notes"`
}

type UpdateAccountRequest struct {
	Name     *string `json:"name"`
	Type     *string `json:"type"`
	Balance  *int    `json:"balance"`
	Currency *string `json:"currency"`
	IsActive *bool   `json:"is_active"`
	Notes    *string `json:"notes"`
}

func (h *AccountHandler) ListAccounts(w http.ResponseWriter, r *http.Request) {
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
		respondJSON(w, http.StatusOK, map[string]interface{}{"data": []models.Account{}})
		return
	}

	var accounts []models.Account
	if err := h.db.Where("budget_id = ?", user.BudgetID).Order("created_at DESC").Find(&accounts).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch accounts"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"data": accounts})
}

func (h *AccountHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req CreateAccountRequest
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

	// Validate account type
	validTypes := map[string]bool{
		"checking":   true,
		"savings":    true,
		"credit_card": true,
		"cash":       true,
		"investment": true,
		"other":      true,
	}
	if !validTypes[req.Type] {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid account type"})
		return
	}

	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}

	account := models.Account{
		BudgetID: *user.BudgetID,
		Name:     req.Name,
		Type:     req.Type,
		Balance:  req.Balance,
		Currency: currency,
		IsActive: true,
		Notes:    req.Notes,
	}

	if err := h.db.Create(&account).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create account"})
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"data":    account,
		"message": "Account created successfully",
	})
}

func (h *AccountHandler) GetAccount(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	accountID := chi.URLParam(r, "id")
	var account models.Account
	if err := h.db.First(&account, "id = ?", accountID).Error; err != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "account not found"})
		return
	}

	// Verify user has access
	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch user"})
		return
	}

	if user.BudgetID == nil || *user.BudgetID != account.BudgetID {
		respondJSON(w, http.StatusForbidden, map[string]string{"error": "access denied"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"data": account})
}

func (h *AccountHandler) UpdateAccount(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	accountID := chi.URLParam(r, "id")
	var account models.Account
	if err := h.db.First(&account, "id = ?", accountID).Error; err != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "account not found"})
		return
	}

	// Verify user has access
	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch user"})
		return
	}

	if user.BudgetID == nil || *user.BudgetID != account.BudgetID {
		respondJSON(w, http.StatusForbidden, map[string]string{"error": "access denied"})
		return
	}

	var req UpdateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Type != nil {
		validTypes := map[string]bool{
			"checking":   true,
			"savings":    true,
			"credit_card": true,
			"cash":       true,
			"investment": true,
			"other":      true,
		}
		if !validTypes[*req.Type] {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid account type"})
			return
		}
		updates["type"] = *req.Type
	}
	if req.Balance != nil {
		updates["balance"] = *req.Balance
	}
	if req.Currency != nil {
		updates["currency"] = *req.Currency
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.Notes != nil {
		updates["notes"] = *req.Notes
	}

	if err := h.db.Model(&account).Updates(updates).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update account"})
		return
	}

	if err := h.db.First(&account, "id = ?", accountID).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch updated account"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"data":    account,
		"message": "Account updated successfully",
	})
}

func (h *AccountHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	accountID := chi.URLParam(r, "id")
	var account models.Account
	if err := h.db.First(&account, "id = ?", accountID).Error; err != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "account not found"})
		return
	}

	// Verify user has access
	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch user"})
		return
	}

	if user.BudgetID == nil || *user.BudgetID != account.BudgetID {
		respondJSON(w, http.StatusForbidden, map[string]string{"error": "access denied"})
		return
	}

	if err := h.db.Delete(&account).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete account"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Account deleted successfully",
	})
}
