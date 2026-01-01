package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user account
type User struct {
	ID               uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email            string     `gorm:"type:varchar(255);unique;not null" json:"email"`
	Name             string     `gorm:"type:varchar(255)" json:"name"`
	BudgetID         *uuid.UUID `gorm:"type:uuid" json:"budget_id"`
	BudgetRole       string     `gorm:"type:varchar(20);default:'read_write'" json:"budget_role"`
	ViewPeriod       string     `gorm:"type:varchar(20);default:'monthly'" json:"view_period"`
	PeriodStartDate  time.Time  `gorm:"type:date;default:CURRENT_DATE" json:"period_start_date"`
	PeriodAnchorDay  *int       `json:"period_anchor_day"`
	IsPremium        bool       `gorm:"default:false" json:"is_premium"`
	PremiumExpiresAt *time.Time `json:"premium_expires_at"`
	StripeCustomerID *string    `gorm:"type:varchar(255)" json:"stripe_customer_id"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// Budget represents a shared budget entity
type Budget struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name       string    `gorm:"type:varchar(255);not null;default:'My Budget'" json:"name"`
	CreatedBy  uuid.UUID `gorm:"type:uuid;not null" json:"created_by"`
	MaxMembers int       `gorm:"default:5" json:"max_members"`
	IsActive   bool      `gorm:"default:true" json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Category represents an expense/income category
type Category struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BudgetID  *uuid.UUID `gorm:"type:uuid" json:"budget_id"`
	Name      string     `gorm:"type:varchar(255);not null" json:"name"`
	Color     string     `gorm:"type:varchar(50);not null" json:"color"`
	Icon      string     `gorm:"type:varchar(50);not null" json:"icon"`
	IsSystem  bool       `gorm:"default:false" json:"is_system"`
	CreatedAt time.Time  `json:"created_at"`
}

// Transaction represents a financial transaction
type Transaction struct {
	ID                 uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID             uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	BudgetID           uuid.UUID  `gorm:"type:uuid;not null" json:"budget_id"`
	Amount             int        `gorm:"not null" json:"amount"` // in cents
	Description        string     `gorm:"type:text" json:"description"`
	MerchantName       string     `gorm:"type:varchar(255)" json:"merchant_name"`
	CategoryID         uuid.UUID  `gorm:"type:uuid;not null" json:"category_id"`
	Date               time.Time  `gorm:"type:date;not null" json:"date"`
	DetectedPatternID  *uuid.UUID `gorm:"type:uuid" json:"detected_pattern_id"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// CategoryBudget represents a monthly budget for a category
type CategoryBudget struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BudgetID       uuid.UUID `gorm:"type:uuid;not null" json:"budget_id"`
	CategoryID     uuid.UUID `gorm:"type:uuid;not null" json:"category_id"`
	Amount         int       `gorm:"not null" json:"amount"` // monthly amount in cents
	AllocationType string    `gorm:"type:varchar(20);default:'pooled'" json:"allocation_type"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// CategoryBudgetSplit represents user-specific allocations for split budgets
type CategoryBudgetSplit struct {
	ID                   uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CategoryBudgetID     uuid.UUID `gorm:"type:uuid;not null" json:"category_budget_id"`
	UserID               uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	AllocationPercentage *float64  `gorm:"type:decimal(5,2)" json:"allocation_percentage"`
	AllocationAmount     *int      `json:"allocation_amount"` // in cents
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// ExpectedIncome represents expected/recurring income
type ExpectedIncome struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BudgetID  uuid.UUID `gorm:"type:uuid;not null" json:"budget_id"`
	Name      string    `gorm:"type:varchar(255);not null" json:"name"`
	Amount    int       `gorm:"not null" json:"amount"` // in cents
	Frequency string    `gorm:"type:varchar(20);not null" json:"frequency"`
	NextDate  time.Time `gorm:"type:date;not null" json:"next_date"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BudgetInvitation represents a pending budget invitation
type BudgetInvitation struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BudgetID     uuid.UUID  `gorm:"type:uuid;not null" json:"budget_id"`
	InviterID    uuid.UUID  `gorm:"type:uuid;not null" json:"inviter_id"`
	InviteeEmail string     `gorm:"type:varchar(255);not null" json:"invitee_email"`
	InvitedRole  string     `gorm:"type:varchar(20);default:'read_write'" json:"invited_role"`
	Token        string     `gorm:"type:varchar(255);unique;not null" json:"token"`
	Status       string     `gorm:"type:varchar(20);default:'pending'" json:"status"`
	ExpiresAt    time.Time  `gorm:"not null" json:"expires_at"`
	CreatedAt    time.Time  `json:"created_at"`
	AcceptedAt   *time.Time `json:"accepted_at"`
}

// BeforeCreate hooks for GORM
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

func (b *Budget) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

func (c *Category) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

func (t *Transaction) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

func (cb *CategoryBudget) BeforeCreate(tx *gorm.DB) error {
	if cb.ID == uuid.Nil {
		cb.ID = uuid.New()
	}
	return nil
}

func (ei *ExpectedIncome) BeforeCreate(tx *gorm.DB) error {
	if ei.ID == uuid.Nil {
		ei.ID = uuid.New()
	}
	return nil
}

func (bi *BudgetInvitation) BeforeCreate(tx *gorm.DB) error {
	if bi.ID == uuid.Nil {
		bi.ID = uuid.New()
	}
	return nil
}
