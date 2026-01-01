-- ============================================================================
-- FOLDA FINANCES - SUPABASE DATABASE INITIALIZATION SCRIPT
-- ============================================================================
-- Version: 1.0
-- Date: 2025-12-30
-- Description: Complete database schema for Folda Finances budgeting app
-- ============================================================================

-- Enable UUID extension (if not already enabled)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================================
-- DROP TABLES (for clean reinstall - use with caution!)
-- ============================================================================
-- Uncomment the following lines if you need to drop all tables and start fresh:
/*
DROP TABLE IF EXISTS budget_transfers CASCADE;
DROP TABLE IF EXISTS subscriptions CASCADE;
DROP TABLE IF EXISTS detected_patterns CASCADE;
DROP TABLE IF EXISTS recurring_bills CASCADE;
DROP TABLE IF EXISTS goals CASCADE;
DROP TABLE IF EXISTS category_budget_splits CASCADE;
DROP TABLE IF EXISTS category_budgets CASCADE;
DROP TABLE IF EXISTS expected_income CASCADE;
DROP TABLE IF EXISTS transactions CASCADE;
DROP TABLE IF EXISTS budget_invitations CASCADE;
DROP TABLE IF EXISTS categories CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS budgets CASCADE;
*/

-- ============================================================================
-- TABLE: budgets
-- ============================================================================
-- Parent budget entity (must be created before users due to foreign key)
CREATE TABLE budgets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL DEFAULT 'My Budget',
    created_by UUID NOT NULL, -- Will reference users(id) after users table is created
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    max_members INTEGER DEFAULT 5,
    is_active BOOLEAN DEFAULT TRUE
);

CREATE INDEX idx_budgets_created_by ON budgets(created_by);

-- ============================================================================
-- TABLE: users
-- ============================================================================
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    is_premium BOOLEAN DEFAULT FALSE,
    premium_expires_at TIMESTAMP WITH TIME ZONE NULL,
    stripe_customer_id VARCHAR(255) NULL,
    -- Multi-user budget support
    budget_id UUID NULL REFERENCES budgets(id) ON DELETE CASCADE,
    budget_role VARCHAR(20) DEFAULT 'read_write' CHECK (budget_role IN ('owner', 'admin', 'read_write', 'read_only')),
    -- View period configuration
    view_period VARCHAR(20) DEFAULT 'monthly' CHECK (view_period IN ('weekly', 'biweekly', 'monthly')),
    period_start_date DATE DEFAULT CURRENT_DATE,
    period_anchor_day INTEGER NULL
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_premium ON users(is_premium, premium_expires_at);
CREATE INDEX idx_users_budget ON users(budget_id);
CREATE INDEX idx_users_budget_role ON users(budget_id, budget_role);

-- Add foreign key constraint to budgets.created_by now that users table exists
ALTER TABLE budgets ADD CONSTRAINT fk_budgets_created_by FOREIGN KEY (created_by) REFERENCES users(id);

-- ============================================================================
-- TABLE: budget_invitations
-- ============================================================================
CREATE TABLE budget_invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    budget_id UUID NOT NULL REFERENCES budgets(id) ON DELETE CASCADE,
    inviter_id UUID NOT NULL REFERENCES users(id),
    invitee_email VARCHAR(255) NOT NULL,
    invited_role VARCHAR(20) DEFAULT 'read_write' CHECK (invited_role IN ('admin', 'read_write', 'read_only')),
    token VARCHAR(255) UNIQUE NOT NULL,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'declined', 'expired')),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    accepted_at TIMESTAMP WITH TIME ZONE NULL
);

CREATE INDEX idx_budget_invitations_budget ON budget_invitations(budget_id);
CREATE INDEX idx_budget_invitations_email ON budget_invitations(invitee_email);
CREATE INDEX idx_budget_invitations_token ON budget_invitations(token);

-- ============================================================================
-- TABLE: categories
-- ============================================================================
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    budget_id UUID NULL REFERENCES budgets(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    color VARCHAR(7) NOT NULL,
    icon VARCHAR(50) NOT NULL,
    is_system BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_categories_budget ON categories(budget_id);
CREATE INDEX idx_categories_system ON categories(is_system);

-- ============================================================================
-- TABLE: detected_patterns
-- ============================================================================
CREATE TABLE detected_patterns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    budget_id UUID NOT NULL REFERENCES budgets(id) ON DELETE CASCADE,
    merchant_name VARCHAR(255) NOT NULL,
    average_amount INTEGER NOT NULL,
    frequency VARCHAR(20) NOT NULL CHECK (frequency IN ('weekly', 'biweekly', 'monthly', 'quarterly', 'yearly')),
    confidence_score DECIMAL(3,2) NOT NULL,
    first_occurrence DATE NOT NULL,
    last_occurrence DATE NOT NULL,
    occurrence_count INTEGER NOT NULL DEFAULT 0,
    user_action VARCHAR(20) DEFAULT 'pending' CHECK (user_action IN ('pending', 'accepted', 'dismissed', 'ignored')),
    suggested_as_subscription BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_detected_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_detected_patterns_budget ON detected_patterns(budget_id);
CREATE INDEX idx_detected_patterns_merchant ON detected_patterns(merchant_name);
CREATE INDEX idx_detected_patterns_action ON detected_patterns(user_action);

-- ============================================================================
-- TABLE: transactions
-- ============================================================================
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    budget_id UUID NOT NULL REFERENCES budgets(id) ON DELETE CASCADE,
    amount INTEGER NOT NULL,
    description TEXT,
    merchant_name VARCHAR(255),
    category_id UUID NOT NULL REFERENCES categories(id),
    date DATE NOT NULL,
    detected_pattern_id UUID NULL REFERENCES detected_patterns(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_transactions_user_date ON transactions(user_id, date DESC);
CREATE INDEX idx_transactions_budget_date ON transactions(budget_id, date DESC);
CREATE INDEX idx_transactions_category ON transactions(category_id);
CREATE INDEX idx_transactions_merchant ON transactions(merchant_name);
CREATE INDEX idx_transactions_pattern ON transactions(detected_pattern_id);

-- ============================================================================
-- TABLE: expected_income
-- ============================================================================
CREATE TABLE expected_income (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    budget_id UUID NOT NULL REFERENCES budgets(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    amount INTEGER NOT NULL,
    frequency VARCHAR(20) NOT NULL CHECK (frequency IN ('weekly', 'biweekly', 'monthly', 'custom')),
    next_date DATE NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_expected_income_budget ON expected_income(budget_id);
CREATE INDEX idx_expected_income_budget_date ON expected_income(budget_id, next_date);
CREATE INDEX idx_expected_income_active ON expected_income(budget_id, is_active);

-- ============================================================================
-- TABLE: category_budgets
-- ============================================================================
CREATE TABLE category_budgets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    budget_id UUID NOT NULL REFERENCES budgets(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    amount INTEGER NOT NULL,
    allocation_type VARCHAR(20) DEFAULT 'pooled' CHECK (allocation_type IN ('pooled', 'split_percentage', 'split_fixed')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT unique_budget_category UNIQUE(budget_id, category_id)
);

CREATE INDEX idx_category_budgets_budget ON category_budgets(budget_id);
CREATE INDEX idx_category_budgets_category ON category_budgets(category_id);

-- ============================================================================
-- TABLE: category_budget_splits (Premium)
-- ============================================================================
CREATE TABLE category_budget_splits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_budget_id UUID NOT NULL REFERENCES category_budgets(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    allocation_percentage DECIMAL(5,2) NULL,
    allocation_amount INTEGER NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT unique_category_budget_user UNIQUE(category_budget_id, user_id),
    CONSTRAINT check_allocation CHECK (
        (allocation_percentage IS NOT NULL AND allocation_amount IS NULL) OR
        (allocation_percentage IS NULL AND allocation_amount IS NOT NULL)
    )
);

CREATE INDEX idx_category_budget_splits_category_budget ON category_budget_splits(category_budget_id);
CREATE INDEX idx_category_budget_splits_user ON category_budget_splits(user_id);

-- ============================================================================
-- TABLE: goals (Premium)
-- ============================================================================
CREATE TABLE goals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    budget_id UUID NOT NULL REFERENCES budgets(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    target_amount INTEGER NOT NULL,
    current_amount INTEGER DEFAULT 0,
    target_date DATE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_goals_budget ON goals(budget_id);
CREATE INDEX idx_goals_budget_date ON goals(budget_id, target_date);

-- ============================================================================
-- TABLE: recurring_bills (Premium)
-- ============================================================================
CREATE TABLE recurring_bills (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    budget_id UUID NOT NULL REFERENCES budgets(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    amount INTEGER NOT NULL,
    category_id UUID NOT NULL REFERENCES categories(id),
    frequency VARCHAR(20) NOT NULL CHECK (frequency IN ('weekly', 'biweekly', 'monthly', 'yearly')),
    next_due_date DATE NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_recurring_bills_budget ON recurring_bills(budget_id);
CREATE INDEX idx_recurring_bills_budget_date ON recurring_bills(budget_id, next_due_date);
CREATE INDEX idx_recurring_bills_active ON recurring_bills(is_active, next_due_date);

-- ============================================================================
-- TABLE: subscriptions (Premium)
-- ============================================================================
CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    budget_id UUID NOT NULL REFERENCES budgets(id) ON DELETE CASCADE,
    created_by_user_id UUID NOT NULL REFERENCES users(id),
    name VARCHAR(255) NOT NULL,
    amount INTEGER NOT NULL,
    billing_frequency VARCHAR(20) NOT NULL CHECK (billing_frequency IN ('weekly', 'biweekly', 'monthly', 'quarterly', 'yearly')),
    next_billing_date DATE NOT NULL,
    category_id UUID NOT NULL REFERENCES categories(id),
    detected_pattern_id UUID NULL REFERENCES detected_patterns(id),
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'paused', 'canceled')),
    cancel_reminder_days INTEGER NULL,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    canceled_at TIMESTAMP WITH TIME ZONE NULL
);

CREATE INDEX idx_subscriptions_budget ON subscriptions(budget_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);
CREATE INDEX idx_subscriptions_next_billing ON subscriptions(next_billing_date);
CREATE INDEX idx_subscriptions_pattern ON subscriptions(detected_pattern_id);

-- ============================================================================
-- TABLE: budget_transfers (Future Feature)
-- ============================================================================
CREATE TABLE budget_transfers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    budget_id UUID NOT NULL REFERENCES budgets(id),
    from_category_id UUID NOT NULL REFERENCES categories(id),
    to_category_id UUID NOT NULL REFERENCES categories(id),
    amount INTEGER NOT NULL,
    period_start_date DATE NOT NULL,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_budget_transfers_user ON budget_transfers(user_id);
CREATE INDEX idx_budget_transfers_budget ON budget_transfers(budget_id);
CREATE INDEX idx_budget_transfers_period ON budget_transfers(period_start_date);

-- ============================================================================
-- SEED DEFAULT CATEGORIES
-- ============================================================================

-- Expense Categories
INSERT INTO categories (budget_id, name, color, icon, is_system) VALUES
  (NULL, 'Housing', '#8B5CF6', '🏠', true),
  (NULL, 'Utilities', '#3B82F6', '⚡', true),
  (NULL, 'Groceries', '#10B981', '🛒', true),
  (NULL, 'Dining & Restaurants', '#F59E0B', '🍽️', true),
  (NULL, 'Transportation', '#EF4444', '🚗', true),
  (NULL, 'Healthcare', '#EC4899', '🏥', true),
  (NULL, 'Entertainment', '#6366F1', '🎬', true),
  (NULL, 'Shopping', '#8B5CF6', '🛍️', true),
  (NULL, 'Personal Care', '#14B8A6', '💆', true),
  (NULL, 'Education', '#F97316', '📚', true),
  (NULL, 'Subscriptions', '#A855F7', '📱', true),
  (NULL, 'Insurance', '#06B6D4', '🛡️', true),
  (NULL, 'Savings', '#22C55E', '💰', true),
  (NULL, 'Debt Payments', '#DC2626', '💳', true),
  (NULL, 'Gifts & Donations', '#F472B6', '🎁', true),
  (NULL, 'Miscellaneous', '#6B7280', '📦', true);

-- Income Categories
INSERT INTO categories (budget_id, name, color, icon, is_system) VALUES
  (NULL, 'Salary', '#059669', '💵', true),
  (NULL, 'Freelance', '#0891B2', '💼', true),
  (NULL, 'Investments', '#7C3AED', '📈', true),
  (NULL, 'Other Income', '#84CC16', '💸', true);

-- ============================================================================
-- ROW LEVEL SECURITY (RLS) POLICIES
-- ============================================================================
-- Note: Adjust these based on your Supabase auth setup

-- Enable RLS on all tables
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE budgets ENABLE ROW LEVEL SECURITY;
ALTER TABLE budget_invitations ENABLE ROW LEVEL SECURITY;
ALTER TABLE categories ENABLE ROW LEVEL SECURITY;
ALTER TABLE transactions ENABLE ROW LEVEL SECURITY;
ALTER TABLE expected_income ENABLE ROW LEVEL SECURITY;
ALTER TABLE category_budgets ENABLE ROW LEVEL SECURITY;
ALTER TABLE category_budget_splits ENABLE ROW LEVEL SECURITY;
ALTER TABLE goals ENABLE ROW LEVEL SECURITY;
ALTER TABLE recurring_bills ENABLE ROW LEVEL SECURITY;
ALTER TABLE subscriptions ENABLE ROW LEVEL SECURITY;
ALTER TABLE detected_patterns ENABLE ROW LEVEL SECURITY;
ALTER TABLE budget_transfers ENABLE ROW LEVEL SECURITY;

-- Users can only see their own user record
CREATE POLICY users_own_record ON users
  FOR ALL
  USING (id = auth.uid());

-- Users can see budgets they belong to
CREATE POLICY budgets_member_access ON budgets
  FOR ALL
  USING (
    id IN (SELECT budget_id FROM users WHERE id = auth.uid())
    OR created_by = auth.uid()
  );

-- Users can see transactions for their budget
CREATE POLICY transactions_budget_access ON transactions
  FOR ALL
  USING (
    budget_id IN (SELECT budget_id FROM users WHERE id = auth.uid())
  );

-- Users can see categories (system categories + their budget's categories)
CREATE POLICY categories_access ON categories
  FOR ALL
  USING (
    is_system = true OR
    budget_id IN (SELECT budget_id FROM users WHERE id = auth.uid())
  );

-- Users can see expected income for their budget
CREATE POLICY expected_income_budget_access ON expected_income
  FOR ALL
  USING (
    budget_id IN (SELECT budget_id FROM users WHERE id = auth.uid())
  );

-- Users can see category budgets for their budget
CREATE POLICY category_budgets_budget_access ON category_budgets
  FOR ALL
  USING (
    budget_id IN (SELECT budget_id FROM users WHERE id = auth.uid())
  );

-- Users can see budget splits for their budget
CREATE POLICY category_budget_splits_budget_access ON category_budget_splits
  FOR ALL
  USING (
    category_budget_id IN (
      SELECT id FROM category_budgets
      WHERE budget_id IN (SELECT budget_id FROM users WHERE id = auth.uid())
    )
  );

-- Users can see goals for their budget
CREATE POLICY goals_budget_access ON goals
  FOR ALL
  USING (
    budget_id IN (SELECT budget_id FROM users WHERE id = auth.uid())
  );

-- Users can see subscriptions for their budget
CREATE POLICY subscriptions_budget_access ON subscriptions
  FOR ALL
  USING (
    budget_id IN (SELECT budget_id FROM users WHERE id = auth.uid())
  );

-- Users can see detected patterns for their budget
CREATE POLICY detected_patterns_budget_access ON detected_patterns
  FOR ALL
  USING (
    budget_id IN (SELECT budget_id FROM users WHERE id = auth.uid())
  );

-- Users can see budget transfers for their budget
CREATE POLICY budget_transfers_budget_access ON budget_transfers
  FOR ALL
  USING (
    budget_id IN (SELECT budget_id FROM users WHERE id = auth.uid())
  );

-- ============================================================================
-- COMPLETION MESSAGE
-- ============================================================================

DO $$
BEGIN
  RAISE NOTICE '============================================================================';
  RAISE NOTICE 'Folda Finances Database Initialization Complete!';
  RAISE NOTICE '============================================================================';
  RAISE NOTICE 'Tables created: 13';
  RAISE NOTICE 'Default categories seeded: 20';
  RAISE NOTICE 'Row Level Security: Enabled on all tables';
  RAISE NOTICE '';
  RAISE NOTICE 'Next steps:';
  RAISE NOTICE '1. Configure your backend .env with database credentials';
  RAISE NOTICE '2. Set SUPABASE_JWT_SECRET in your backend .env';
  RAISE NOTICE '3. Start your Go backend: go run cmd/api/main.go';
  RAISE NOTICE '4. Start your React frontend: npm run dev';
  RAISE NOTICE '============================================================================';
END $$;
