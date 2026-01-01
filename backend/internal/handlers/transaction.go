package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/yourusername/folda-finances/internal/middleware"
	"github.com/yourusername/folda-finances/internal/models"
	"gorm.io/gorm"
)

type TransactionHandler struct {
	db *gorm.DB
}

func NewTransactionHandler(db *gorm.DB) *TransactionHandler {
	return &TransactionHandler{db: db}
}

type CreateTransactionRequest struct {
	Amount      int    `json:"amount"`
	Description string `json:"description"`
	CategoryID  string `json:"category_id"`
	Date        string `json:"date"`
}

type UpdateTransactionRequest struct {
	Amount      *int    `json:"amount"`
	Description *string `json:"description"`
	CategoryID  *string `json:"category_id"`
	Date        *string `json:"date"`
}

func (h *TransactionHandler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	// Get user's budget_id
	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch user"})
		return
	}

	if user.BudgetID == nil {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"data": map[string]interface{}{
				"data":        []models.Transaction{},
				"page":        1,
				"per_page":    50,
				"total":       0,
				"total_pages": 0,
			},
		})
		return
	}

	query := h.db.Where("budget_id = ?", user.BudgetID)

	// Filters
	if categoryID := r.URL.Query().Get("category_id"); categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}
	if filterUserID := r.URL.Query().Get("user_id"); filterUserID != "" {
		query = query.Where("user_id = ?", filterUserID)
	}
	if startDate := r.URL.Query().Get("start_date"); startDate != "" {
		query = query.Where("date >= ?", startDate)
	}
	if endDate := r.URL.Query().Get("end_date"); endDate != "" {
		query = query.Where("date <= ?", endDate)
	}

	var transactions []models.Transaction
	if err := query.Order("date DESC, created_at DESC").Find(&transactions).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch transactions"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"data":        transactions,
			"page":        1,
			"per_page":    50,
			"total":       len(transactions),
			"total_pages": 1,
		},
	})
}

func (h *TransactionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	// Get user's budget_id
	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch user"})
		return
	}

	if user.BudgetID == nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "user does not belong to a budget"})
		return
	}

	// Parse category ID
	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid category_id"})
		return
	}

	// Parse date
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid date format"})
		return
	}

	// Extract merchant name from description (simple version)
	merchantName := extractMerchantName(req.Description)

	transaction := models.Transaction{
		UserID:       userID,
		BudgetID:     *user.BudgetID,
		Amount:       req.Amount,
		Description:  req.Description,
		MerchantName: merchantName,
		CategoryID:   categoryID,
		Date:         date,
	}

	if err := h.db.Create(&transaction).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create transaction"})
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"data":    transaction,
		"message": "Transaction created successfully",
	})
}

func (h *TransactionHandler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	transactionID := chi.URLParam(r, "id")
	if transactionID == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "transaction ID required"})
		return
	}

	var transaction models.Transaction
	if err := h.db.First(&transaction, "id = ?", transactionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondJSON(w, http.StatusNotFound, map[string]string{"error": "transaction not found"})
			return
		}
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch transaction"})
		return
	}

	// Verify user has access to this transaction (same budget)
	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch user"})
		return
	}

	if user.BudgetID == nil || transaction.BudgetID != *user.BudgetID {
		respondJSON(w, http.StatusForbidden, map[string]string{"error": "access denied"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"data": transaction})
}

func (h *TransactionHandler) UpdateTransaction(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	transactionID := chi.URLParam(r, "id")
	var transaction models.Transaction
	if err := h.db.First(&transaction, "id = ?", transactionID).Error; err != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "transaction not found"})
		return
	}

	// Verify ownership or budget access
	if transaction.UserID != userID {
		respondJSON(w, http.StatusForbidden, map[string]string{"error": "can only update your own transactions"})
		return
	}

	var req UpdateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	updates := map[string]interface{}{}
	if req.Amount != nil {
		updates["amount"] = *req.Amount
	}
	if req.Description != nil {
		updates["description"] = *req.Description
		updates["merchant_name"] = extractMerchantName(*req.Description)
	}
	if req.CategoryID != nil {
		categoryID, err := uuid.Parse(*req.CategoryID)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid category_id"})
			return
		}
		updates["category_id"] = categoryID
	}
	if req.Date != nil {
		date, err := time.Parse("2006-01-02", *req.Date)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid date format"})
			return
		}
		updates["date"] = date
	}

	if err := h.db.Model(&transaction).Updates(updates).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update transaction"})
		return
	}

	// Fetch updated transaction
	if err := h.db.First(&transaction, "id = ?", transactionID).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch updated transaction"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"data":    transaction,
		"message": "Transaction updated successfully",
	})
}

func (h *TransactionHandler) DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	transactionID := chi.URLParam(r, "id")
	var transaction models.Transaction
	if err := h.db.First(&transaction, "id = ?", transactionID).Error; err != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "transaction not found"})
		return
	}

	// Verify ownership
	if transaction.UserID != userID {
		respondJSON(w, http.StatusForbidden, map[string]string{"error": "can only delete your own transactions"})
		return
	}

	if err := h.db.Delete(&transaction).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete transaction"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Transaction deleted successfully",
	})
}

// extractMerchantName extracts a normalized merchant name from description
func extractMerchantName(description string) string {
	// Simple implementation: take first word, uppercase
	words := strings.Fields(description)
	if len(words) > 0 {
		return strings.ToUpper(words[0])
	}
	return ""
}
