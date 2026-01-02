package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/yourusername/folda-finances/internal/middleware"
	"github.com/yourusername/folda-finances/internal/models"
	"gorm.io/gorm"
)

type InvitationHandler struct {
	db *gorm.DB
}

func NewInvitationHandler(db *gorm.DB) *InvitationHandler {
	return &InvitationHandler{db: db}
}

type CreateBudgetInvitationRequest struct {
	Email string `json:"invitee_email"`
	Role  string `json:"invited_role"`
}

// InviteToBudget creates a new budget invitation
func (h *InvitationHandler) InviteToBudget(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	budgetIDParam := chi.URLParam(r, "budgetId")
	budgetID, err := uuid.Parse(budgetIDParam)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid budget_id"})
		return
	}

	var req CreateBudgetInvitationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	// Validate email
	if req.Email == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "email is required"})
		return
	}

	// Validate role
	role := req.Role
	if role == "" {
		role = "read_write"
	}
	if role != "read_only" && role != "read_write" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid role"})
		return
	}

	// Check if user has permission to invite to this budget
	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch user"})
		return
	}

	if user.BudgetID == nil || *user.BudgetID != budgetID {
		respondJSON(w, http.StatusForbidden, map[string]string{"error": "you do not have permission to invite to this budget"})
		return
	}

	// Check if invitation already exists for this email
	var existingInvitation models.BudgetInvitation
	err = h.db.Where("budget_id = ? AND invitee_email = ? AND status = ?", budgetID, req.Email, "pending").First(&existingInvitation).Error
	if err == nil {
		respondJSON(w, http.StatusConflict, map[string]string{"error": "invitation already exists for this email"})
		return
	}

	// Check if user already has access to this budget
	var existingUser models.User
	err = h.db.Where("email = ? AND budget_id = ?", req.Email, budgetID).First(&existingUser).Error
	if err == nil {
		respondJSON(w, http.StatusConflict, map[string]string{"error": "user already has access to this budget"})
		return
	}

	// Generate invitation token
	token, err := generateToken()
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
		return
	}

	// Create invitation
	invitation := models.BudgetInvitation{
		ID:           uuid.New(),
		BudgetID:     budgetID,
		InviterID:    userID,
		InviteeEmail: req.Email,
		InvitedRole:  role,
		Token:        token,
		Status:       "pending",
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour), // 7 days
	}

	if err := h.db.Create(&invitation).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create invitation"})
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"data":    invitation,
		"message": "Invitation sent successfully",
	})
}

// GetBudgetInvitations lists all invitations for the current user
func (h *InvitationHandler) GetBudgetInvitations(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	// Get user's email
	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch user"})
		return
	}

	// Get all pending invitations for this user's email
	var invitations []models.BudgetInvitation
	if err := h.db.Where("invitee_email = ? AND status = ?", user.Email, "pending").Find(&invitations).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch invitations"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"data": invitations})
}

// AcceptBudgetInvitation accepts an invitation
func (h *InvitationHandler) AcceptBudgetInvitation(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	token := chi.URLParam(r, "token")
	if token == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "token is required"})
		return
	}

	// Get the invitation
	var invitation models.BudgetInvitation
	if err := h.db.Where("token = ?", token).First(&invitation).Error; err != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "invitation not found"})
		return
	}

	// Validate invitation
	if invitation.Status != "pending" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invitation has already been processed"})
		return
	}

	if time.Now().After(invitation.ExpiresAt) {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invitation has expired"})
		return
	}

	// Get user and verify email matches
	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch user"})
		return
	}

	if user.Email != invitation.InviteeEmail {
		respondJSON(w, http.StatusForbidden, map[string]string{"error": "this invitation is for a different email address"})
		return
	}

	// Update user's budget_id
	user.BudgetID = &invitation.BudgetID
	if err := h.db.Save(&user).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update user"})
		return
	}

	// Update invitation status
	now := time.Now()
	invitation.Status = "accepted"
	invitation.AcceptedAt = &now
	if err := h.db.Save(&invitation).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update invitation"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Invitation accepted successfully",
	})
}

// DeclineBudgetInvitation declines an invitation
func (h *InvitationHandler) DeclineBudgetInvitation(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	token := chi.URLParam(r, "token")
	if token == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "token is required"})
		return
	}

	// Get the invitation
	var invitation models.BudgetInvitation
	if err := h.db.Where("token = ?", token).First(&invitation).Error; err != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "invitation not found"})
		return
	}

	// Validate invitation
	if invitation.Status != "pending" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invitation has already been processed"})
		return
	}

	// Get user and verify email matches
	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch user"})
		return
	}

	if user.Email != invitation.InviteeEmail {
		respondJSON(w, http.StatusForbidden, map[string]string{"error": "this invitation is for a different email address"})
		return
	}

	// Update invitation status
	invitation.Status = "declined"
	if err := h.db.Save(&invitation).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update invitation"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Invitation declined",
	})
}

// Helper function to generate a random token
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
