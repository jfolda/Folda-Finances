package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/yourusername/folda-finances/internal/middleware"
	"github.com/yourusername/folda-finances/internal/models"
	"gorm.io/gorm"
)

type UserHandler struct {
	db *gorm.DB
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{db: db}
}

type UpdateUserRequest struct {
	Name            *string `json:"name"`
	ViewPeriod      *string `json:"view_period"`
	PeriodStartDate *string `json:"period_start_date"`
}

func (h *UserHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	email, err := middleware.GetUserEmail(r)
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Auto-create user record on first API call
			// Step 1: Create user without budget first (to satisfy foreign key)
			user = models.User{
				ID:              userID,
				Email:           email, // Will be populated from Supabase metadata if available
				Name:            "New User",
				ViewPeriod:      "monthly",
				PeriodStartDate: time.Now(),
				BudgetID:        nil, // No budget yet
			}

			if err := h.db.Create(&user).Error; err != nil {
				respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create user"})
				return
			}

			// Step 2: Create a default budget (now user exists for foreign key)
			budget := models.Budget{
				Name:      "My Budget",
				CreatedBy: userID,
			}
			if err := h.db.Create(&budget).Error; err != nil {
				respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create budget"})
				return
			}

			// Step 3: Update user with budget ID
			user.BudgetID = &budget.ID
			if err := h.db.Save(&user).Error; err != nil {
				respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update user budget"})
				return
			}
		} else {
			respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch user"})
			return
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"data": user})
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.ViewPeriod != nil {
		if *req.ViewPeriod != "weekly" && *req.ViewPeriod != "biweekly" && *req.ViewPeriod != "monthly" {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid view_period"})
			return
		}
		updates["view_period"] = *req.ViewPeriod
	}
	if req.PeriodStartDate != nil {
		// Convert string to period_anchor_day integer
		anchorDay, err := strconv.Atoi(*req.PeriodStartDate)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid period_start_date"})
			return
		}
		updates["period_anchor_day"] = anchorDay
	}

	if err := h.db.Model(&user).Updates(updates).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update user"})
		return
	}

	// Fetch updated user
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch updated user"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"data":    user,
		"message": "Settings updated successfully",
	})
}
