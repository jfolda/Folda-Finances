// Shared API type definitions
// These types are used by both frontend and backend (generated)

// ============================================================================
// MULTI-USER BUDGET TYPES
// ============================================================================

export type BudgetRole = 'owner' | 'admin' | 'read_write' | 'read_only';
export type ViewPeriod = 'weekly' | 'biweekly' | 'monthly';
export type AllocationType = 'pooled' | 'split';

// Budget entity (shared by multiple users)
export interface Budget {
  id: string;
  name: string;
  created_by: string;
  max_members: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateBudgetRequest {
  name: string;
}

export interface UpdateBudgetRequest {
  name?: string;
  max_members?: number;
}

// Budget member info
export interface BudgetMember {
  user_id: string;
  name: string;
  email: string;
  budget_role: BudgetRole;
  joined_at: string;
}

// Budget invitation
export interface BudgetInvitation {
  id: string;
  budget_id: string;
  inviter_id: string;
  invitee_email: string;
  invited_role: BudgetRole;
  token: string;
  status: 'pending' | 'accepted' | 'declined' | 'expired';
  expires_at: string;
  created_at: string;
}

export interface CreateBudgetInvitationRequest {
  invitee_email: string;
  invited_role: BudgetRole;
}

export interface AcceptBudgetInvitationRequest {
  token: string;
}

// ============================================================================
// USER TYPES
// ============================================================================

export interface User {
  id: string;
  email: string;
  name: string;
  budget_id: string | null;
  budget_role: BudgetRole | null;
  view_period: ViewPeriod;
  period_start_date: string;
  is_premium: boolean;
  created_at: string;
  updated_at: string;
}

export interface UpdateUserSettingsRequest {
  name?: string;
  view_period?: ViewPeriod;
  period_start_date?: string;
}

// ============================================================================
// CATEGORY BUDGET TYPES (Monthly budgets with optional splits)
// ============================================================================

export interface CategoryBudget {
  id: string;
  budget_id: string;
  category_id: string;
  amount: number; // ALWAYS monthly amount in cents
  allocation_type: AllocationType;
  created_at: string;
  updated_at: string;
}

export interface CategoryBudgetSplit {
  id: string;
  category_budget_id: string;
  user_id: string;
  allocation_percentage: number | null; // e.g., 60.00 for 60%
  allocation_amount: number | null; // in cents
}

export interface CreateCategoryBudgetRequest {
  category_id: string;
  amount: number; // monthly amount in cents
  allocation_type: AllocationType;
  splits?: {
    user_id: string;
    allocation_percentage?: number;
    allocation_amount?: number;
  }[];
}

export interface UpdateCategoryBudgetRequest {
  amount?: number;
  allocation_type?: AllocationType;
  splits?: {
    user_id: string;
    allocation_percentage?: number;
    allocation_amount?: number;
  }[];
}

// ============================================================================
// ACCOUNT TYPES
// ============================================================================

export type AccountType = 'checking' | 'savings' | 'credit_card' | 'cash' | 'investment' | 'other';

export interface Account {
  id: string;
  budget_id: string;
  name: string;
  type: AccountType;
  balance: number; // in cents
  currency: string;
  is_active: boolean;
  notes: string;
  created_at: string;
  updated_at: string;
}

export interface CreateAccountRequest {
  name: string;
  type: AccountType;
  balance?: number; // in cents, defaults to 0
  currency?: string; // defaults to 'USD'
  notes?: string;
}

export interface UpdateAccountRequest {
  name?: string;
  type?: AccountType;
  balance?: number;
  currency?: string;
  is_active?: boolean;
  notes?: string;
}

// ============================================================================
// TRANSACTION TYPES
// ============================================================================

export interface Transaction {
  id: string;
  user_id: string;
  budget_id: string;
  account_id: string | null;
  amount: number;
  description: string;
  category_id: string;
  date: string;
  merchant_name: string; // normalized merchant name for pattern detection
  detected_pattern_id: string | null;
  created_at: string;
  updated_at: string;
}

export interface CreateTransactionRequest {
  amount: number;
  description: string;
  category_id: string;
  date: string;
  account_id?: string | null;
}

export interface UpdateTransactionRequest {
  amount?: number;
  description?: string;
  category_id?: string;
  date?: string;
  account_id?: string | null;
}

// ============================================================================
// CATEGORY TYPES
// ============================================================================

export interface Category {
  id: string;
  budget_id: string | null; // null for system categories
  name: string;
  color: string;
  icon: string;
  is_system: boolean;
  created_at: string;
}

export interface CreateCategoryRequest {
  name: string;
  color: string;
  icon: string;
}

// ============================================================================
// EXPECTED INCOME TYPES
// ============================================================================

export interface ExpectedIncome {
  id: string;
  budget_id: string;
  name: string;
  amount: number;
  frequency: 'weekly' | 'biweekly' | 'monthly' | 'custom';
  next_date: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateExpectedIncomeRequest {
  name: string;
  amount: number;
  frequency: 'weekly' | 'biweekly' | 'monthly' | 'custom';
  next_date: string;
}

export interface UpdateExpectedIncomeRequest {
  name?: string;
  amount?: number;
  frequency?: 'weekly' | 'biweekly' | 'monthly' | 'custom';
  next_date?: string;
  is_active?: boolean;
}

export interface MarkIncomeReceivedRequest {
  received_date: string;
  actual_amount?: number;
}

// ============================================================================
// "WHAT CAN I SPEND?" TYPES (CORE FEATURE)
// ============================================================================

export interface SpendingPeriod {
  type: ViewPeriod;
  start_date: string;
  end_date: string;
  days_remaining: number;
}

export interface CategorySpending {
  category_id: string;
  category_name: string;
  category_icon: string;
  category_color: string;
  budgeted: number; // pro-rated to user's view period
  spent: number; // actual spending in current period
  available: number; // budgeted - spent
  percentage_used: number;
  status: 'on_track' | 'warning' | 'over_budget';
  // For split budgets (premium):
  is_split: boolean;
  my_allocation?: number; // user's allocated portion (if split)
  my_available?: number; // user's available amount (if split)
}

export interface SpendingAvailableResponse {
  period: SpendingPeriod;
  summary: {
    total_available: number;
    total_budgeted: number;
    total_spent: number;
  };
  categories: CategorySpending[];
}

// ============================================================================
// RECURRING PATTERN DETECTION (FREE + PREMIUM)
// ============================================================================

export interface DetectedPattern {
  id: string;
  budget_id: string;
  merchant_name: string;
  average_amount: number;
  frequency: 'weekly' | 'biweekly' | 'monthly' | 'quarterly' | 'yearly';
  confidence_score: number; // 0.00 to 1.00
  first_occurrence: string;
  last_occurrence: string;
  occurrence_count: number;
  user_action: 'pending' | 'accepted' | 'dismissed' | 'ignored';
  created_at: string;
  updated_at: string;
}

export interface AcceptPatternRequest {
  action: 'badge_only' | 'create_subscription'; // badge_only for free, create_subscription for premium
}

// ============================================================================
// SUBSCRIPTIONS (PREMIUM)
// ============================================================================

export interface Subscription {
  id: string;
  budget_id: string;
  created_by_user_id: string;
  name: string;
  amount: number;
  billing_frequency: 'weekly' | 'biweekly' | 'monthly' | 'quarterly' | 'yearly';
  next_billing_date: string;
  category_id: string;
  detected_pattern_id: string | null;
  status: 'active' | 'paused' | 'canceled';
  notes: string;
  created_at: string;
  updated_at: string;
}

export interface CreateSubscriptionRequest {
  name: string;
  amount: number;
  billing_frequency: 'weekly' | 'biweekly' | 'monthly' | 'quarterly' | 'yearly';
  next_billing_date: string;
  category_id: string;
  notes?: string;
}

export interface UpdateSubscriptionRequest {
  name?: string;
  amount?: number;
  billing_frequency?: 'weekly' | 'biweekly' | 'monthly' | 'quarterly' | 'yearly';
  next_billing_date?: string;
  category_id?: string;
  status?: 'active' | 'paused' | 'canceled';
  notes?: string;
}

export interface SubscriptionSummaryResponse {
  total_monthly_cost: number;
  active_count: number;
  paused_count: number;
  upcoming_renewals: {
    subscription_id: string;
    name: string;
    amount: number;
    next_billing_date: string;
    days_until: number;
  }[];
}

// ============================================================================
// BUDGET TRANSFERS (FUTURE PREMIUM)
// ============================================================================

export interface BudgetTransfer {
  id: string;
  user_id: string;
  budget_id: string;
  from_category_id: string;
  to_category_id: string;
  amount: number;
  period_start_date: string;
  notes: string;
  created_at: string;
}

export interface CreateBudgetTransferRequest {
  from_category_id: string;
  to_category_id: string;
  amount: number;
  notes?: string;
}

export interface BudgetTransferHistoryResponse {
  transfers: (BudgetTransfer & {
    from_category_name: string;
    to_category_name: string;
  })[];
  total: number;
}

// ============================================================================
// GOAL TYPES (PREMIUM)
// ============================================================================

export interface Goal {
  id: string;
  budget_id: string;
  name: string;
  target_amount: number;
  current_amount: number;
  target_date: string;
  created_at: string;
  updated_at: string;
}

export interface CreateGoalRequest {
  name: string;
  target_amount: number;
  target_date: string;
}

export interface UpdateGoalRequest {
  name?: string;
  target_amount?: number;
  current_amount?: number;
  target_date?: string;
}

// ============================================================================
// API RESPONSE TYPES
// ============================================================================

export interface ApiResponse<T> {
  data: T;
  message?: string;
}

export interface ApiError {
  error: string;
  message: string;
  status: number;
}

// Pagination
export interface PaginatedResponse<T> {
  data: T[];
  page: number;
  per_page: number;
  total: number;
  total_pages: number;
}
