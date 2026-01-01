package handlers

import (
	"math"
	"net/http"
	"time"

	"github.com/yourusername/folda-finances/internal/middleware"
	"github.com/yourusername/folda-finances/internal/models"
	"gorm.io/gorm"
)

type SpendingHandler struct {
	db *gorm.DB
}

func NewSpendingHandler(db *gorm.DB) *SpendingHandler {
	return &SpendingHandler{db: db}
}

type SpendingPeriod struct {
	Type          string `json:"type"`
	StartDate     string `json:"start_date"`
	EndDate       string `json:"end_date"`
	DaysRemaining int    `json:"days_remaining"`
}

type CategorySpending struct {
	CategoryID     string  `json:"category_id"`
	CategoryName   string  `json:"category_name"`
	CategoryIcon   string  `json:"category_icon"`
	CategoryColor  string  `json:"category_color"`
	Budgeted       int     `json:"budgeted"`
	Spent          int     `json:"spent"`
	Available      int     `json:"available"`
	PercentageUsed float64 `json:"percentage_used"`
	Status         string  `json:"status"`
	IsSplit        bool    `json:"is_split"`
	MyAllocation   *int    `json:"my_allocation,omitempty"`
	MyAvailable    *int    `json:"my_available,omitempty"`
}

type SpendingAvailableResponse struct {
	Period   SpendingPeriod     `json:"period"`
	Summary  SpendingSummary    `json:"summary"`
	Categories []CategorySpending `json:"categories"`
}

type SpendingSummary struct {
	TotalAvailable int `json:"total_available"`
	TotalBudgeted  int `json:"total_budgeted"`
	TotalSpent     int `json:"total_spent"`
}

func (h *SpendingHandler) GetSpendingAvailable(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	// Get user with budget
	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch user"})
		return
	}

	if user.BudgetID == nil {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"data": SpendingAvailableResponse{
				Period: SpendingPeriod{
					Type:          user.ViewPeriod,
					StartDate:     time.Now().Format("2006-01-02"),
					EndDate:       time.Now().AddDate(0, 1, 0).Format("2006-01-02"),
					DaysRemaining: 30,
				},
				Summary: SpendingSummary{
					TotalAvailable: 0,
					TotalBudgeted:  0,
					TotalSpent:     0,
				},
				Categories: []CategorySpending{},
			},
		})
		return
	}

	// Calculate current period
	period := calculatePeriod(user.ViewPeriod, user.PeriodStartDate)

	// Get category budgets for this budget
	var categoryBudgets []models.CategoryBudget
	if err := h.db.Where("budget_id = ?", user.BudgetID).Find(&categoryBudgets).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch budgets"})
		return
	}

	// Get categories
	var categories []models.Category
	categoryIDs := make([]string, len(categoryBudgets))
	for i, cb := range categoryBudgets {
		categoryIDs[i] = cb.CategoryID.String()
	}
	if err := h.db.Where("id IN ?", categoryIDs).Find(&categories).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch categories"})
		return
	}

	// Create category map
	categoryMap := make(map[string]models.Category)
	for _, cat := range categories {
		categoryMap[cat.ID.String()] = cat
	}

	// Get transactions for current period
	var transactions []models.Transaction
	if err := h.db.Where("budget_id = ? AND date >= ? AND date <= ?",
		user.BudgetID, period.StartDate, period.EndDate).Find(&transactions).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch transactions"})
		return
	}

	// Calculate spending per category
	categorySpendingList := []CategorySpending{}
	totalBudgeted := 0
	totalSpent := 0

	for _, categoryBudget := range categoryBudgets {
		category, ok := categoryMap[categoryBudget.CategoryID.String()]
		if !ok {
			continue
		}

		// Pro-rate monthly budget to view period
		proratedBudget := prorateBudget(categoryBudget.Amount, user.ViewPeriod)

		// Calculate spent in this period for this category
		spent := 0
		for _, tx := range transactions {
			if tx.CategoryID == categoryBudget.CategoryID && tx.Amount < 0 {
				spent += int(math.Abs(float64(tx.Amount)))
			}
		}

		available := proratedBudget - spent
		percentageUsed := 0.0
		if proratedBudget > 0 {
			percentageUsed = (float64(spent) / float64(proratedBudget)) * 100
		}

		status := getStatus(percentageUsed)

		categorySpendingList = append(categorySpendingList, CategorySpending{
			CategoryID:     category.ID.String(),
			CategoryName:   category.Name,
			CategoryIcon:   category.Icon,
			CategoryColor:  category.Color,
			Budgeted:       proratedBudget,
			Spent:          spent,
			Available:      available,
			PercentageUsed: percentageUsed,
			Status:         status,
			IsSplit:        categoryBudget.AllocationType != "pooled",
		})

		totalBudgeted += proratedBudget
		totalSpent += spent
	}

	totalAvailable := totalBudgeted - totalSpent

	response := SpendingAvailableResponse{
		Period: period,
		Summary: SpendingSummary{
			TotalAvailable: totalAvailable,
			TotalBudgeted:  totalBudgeted,
			TotalSpent:     totalSpent,
		},
		Categories: categorySpendingList,
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"data": response})
}

func calculatePeriod(viewPeriod string, startDate time.Time) SpendingPeriod {
	now := time.Now()
	var periodStart, periodEnd time.Time

	switch viewPeriod {
	case "weekly":
		// Find most recent start day
		daysSinceStart := int(now.Sub(startDate).Hours() / 24)
		weeksPassed := daysSinceStart / 7
		periodStart = startDate.AddDate(0, 0, weeksPassed*7)
		periodEnd = periodStart.AddDate(0, 0, 7).Add(-time.Second)

	case "biweekly":
		// Find most recent biweekly boundary
		daysSinceStart := int(now.Sub(startDate).Hours() / 24)
		periodsPassed := daysSinceStart / 14
		periodStart = startDate.AddDate(0, 0, periodsPassed*14)
		periodEnd = periodStart.AddDate(0, 0, 14).Add(-time.Second)

	case "monthly":
		// Current calendar month
		periodStart = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		periodEnd = periodStart.AddDate(0, 1, 0).Add(-time.Second)

	default:
		// Default to monthly
		periodStart = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		periodEnd = periodStart.AddDate(0, 1, 0).Add(-time.Second)
	}

	daysRemaining := int(periodEnd.Sub(now).Hours() / 24)
	if daysRemaining < 0 {
		daysRemaining = 0
	}

	return SpendingPeriod{
		Type:          viewPeriod,
		StartDate:     periodStart.Format("2006-01-02"),
		EndDate:       periodEnd.Format("2006-01-02"),
		DaysRemaining: daysRemaining,
	}
}

func prorateBudget(monthlyAmount int, viewPeriod string) int {
	const daysPerMonth = 30.44

	switch viewPeriod {
	case "weekly":
		return int(math.Round(float64(monthlyAmount) * (7 / daysPerMonth)))
	case "biweekly":
		return int(math.Round(float64(monthlyAmount) * (14 / daysPerMonth)))
	case "monthly":
		return monthlyAmount
	default:
		return monthlyAmount
	}
}

func getStatus(percentageUsed float64) string {
	if percentageUsed > 100 {
		return "over_budget"
	}
	if percentageUsed >= 75 {
		return "warning"
	}
	return "on_track"
}
