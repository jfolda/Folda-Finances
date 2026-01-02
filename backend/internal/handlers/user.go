package handlers

import (
	"encoding/json"
	"net/http"
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

	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Auto-create user record on first API call
			user = models.User{
				ID:              userID,
				Email:           "",  // Will be populated from Supabase metadata if available
				Name:            "New User",
				ViewPeriod:      "monthly",
				PeriodStartDate: time.Now(),
			}

			// Create a default budget for the user
			budget := models.Budget{
				Name:      "My Budget",
				CreatedBy: userID,
			}
			if err := h.db.Create(&budget).Error; err != nil {
				respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create budget"})
				return
			}

			// Associate user with the budget
			user.BudgetID = &budget.ID

			if err := h.db.Create(&user).Error; err != nil {
				respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create user"})
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
		updates["period_start_date"] = *req.PeriodStartDate
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
