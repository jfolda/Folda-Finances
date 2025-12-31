# Database Schema

## Overview

This document describes the PostgreSQL database schema for Folda Finances.

**Database:** PostgreSQL 15+ (via Supabase)

---

## Tables

### users

Stores user account information. Authentication is handled by Supabase Auth.

```sql
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
    -- View period configuration (for "What Can I Spend?" display only)
    view_period VARCHAR(20) DEFAULT 'monthly' CHECK (view_period IN ('weekly', 'biweekly', 'monthly')),
    period_start_date DATE DEFAULT CURRENT_DATE,
    period_anchor_day INTEGER NULL -- For weekly: 0-6 (Sun-Sat), for biweekly: day of month
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_premium ON users(is_premium, premium_expires_at);
CREATE INDEX idx_users_budget ON users(budget_id);
CREATE INDEX idx_users_budget_role ON users(budget_id, budget_role);
```

**Notes:**
- `id` should match Supabase Auth user ID
- `budget_id` links user to their budget (one budget per user initially)
- `budget_role` defines permissions:
  - `owner`: Budget creator, full admin, cannot be removed
  - `admin`: Can invite/remove members, edit budgets (granted by owner)
  - `read_write`: Can add transactions, view everything (default for invited users)
  - `read_only`: Can only view (set by owner)
- `name` for display in multi-user contexts
- `is_premium` flag for quick premium checks
- `premium_expires_at` for subscription management
- `stripe_customer_id` links to Stripe customer
- `view_period` is ONLY for "What Can I Spend?" display (budgets are always monthly)
- `period_start_date` is the reference date for calculating VIEW period boundaries
- `period_anchor_day` helps calculate recurring VIEW period starts
- **ON DELETE CASCADE**: When budget is deleted (owner deletes account), all users lose budget access

---

### budgets

Stores the budget entity (the shared budget that multiple users belong to).

```sql
CREATE TABLE budgets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL DEFAULT 'My Budget',
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    max_members INTEGER DEFAULT 5, -- configurable for future expansion
    is_active BOOLEAN DEFAULT TRUE
);

CREATE INDEX idx_budgets_created_by ON budgets(created_by);
```

**Notes:**
- This is the parent "budget" entity
- Multiple users can belong to one budget
- `created_by` is the budget creator (has special permissions)
- `max_members` enforced in application logic
- `name` allows users to name their budget (e.g., "Smith Family Budget")

---

### budget_invitations

Stores pending invitations to join a budget.

```sql
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
```

**Notes:**
- `token` is a secure random string for the invitation link
- `invited_role` is the role the user will get when they accept (set by inviter)
- `expires_at` typically 7 days from creation
- When user accepts, their `users.budget_id` and `users.budget_role` are set
- Prevent duplicate invitations for same email+budget
- Only owner and admin can send invitations

---

### categories

Stores expense/income categories (both system and custom).

```sql
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    color VARCHAR(7) NOT NULL, -- hex color (e.g., #10B981)
    icon VARCHAR(50) NOT NULL, -- icon identifier
    is_system BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_categories_user ON categories(user_id);
CREATE INDEX idx_categories_system ON categories(is_system);
```

**Notes:**
- `user_id = NULL` indicates system category (available to all users)
- `user_id != NULL` indicates custom category (premium feature)
- System categories seeded on initial migration
- Icons use a predefined icon set (e.g., Lucide icons)

**Default Categories (to be seeded):**

```sql
-- Expense categories
INSERT INTO categories (id, user_id, name, color, icon, is_system) VALUES
  (gen_random_uuid(), NULL, 'Housing', '#EF4444', 'home', true),
  (gen_random_uuid(), NULL, 'Utilities', '#F59E0B', 'zap', true),
  (gen_random_uuid(), NULL, 'Groceries', '#10B981', 'shopping-cart', true),
  (gen_random_uuid(), NULL, 'Dining & Restaurants', '#8B5CF6', 'utensils', true),
  (gen_random_uuid(), NULL, 'Transportation', '#3B82F6', 'car', true),
  (gen_random_uuid(), NULL, 'Healthcare', '#EC4899', 'heart', true),
  (gen_random_uuid(), NULL, 'Shopping', '#F97316', 'shopping-bag', true),
  (gen_random_uuid(), NULL, 'Entertainment', '#14B8A6', 'film', true),
  (gen_random_uuid(), NULL, 'Other Expenses', '#6B7280', 'more-horizontal', true);

-- Income categories
INSERT INTO categories (id, user_id, name, color, icon, is_system) VALUES
  (gen_random_uuid(), NULL, 'Salary/Wages', '#22C55E', 'dollar-sign', true),
  (gen_random_uuid(), NULL, 'Freelance', '#84CC16', 'briefcase', true),
  (gen_random_uuid(), NULL, 'Investments', '#06B6D4', 'trending-up', true),
  (gen_random_uuid(), NULL, 'Other Income', '#A3E635', 'plus-circle', true);
```

---

### transactions

Stores all financial transactions (manual and imported).

```sql
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    budget_id UUID NOT NULL REFERENCES budgets(id) ON DELETE CASCADE,
    amount INTEGER NOT NULL, -- stored in cents
    description TEXT,
    merchant_name VARCHAR(255), -- Normalized merchant name for pattern detection
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
```

**Notes:**
- `user_id` tracks WHO made the transaction (for multi-user budgets)
- `budget_id` tracks WHICH budget the transaction belongs to
- `amount` stored in cents (INTEGER) to avoid floating-point precision issues
- Negative amounts = expenses, positive amounts = income
- `date` is the transaction date (not necessarily when it was entered)
- `merchant_name` is normalized/cleaned from description for pattern matching
- `detected_pattern_id` links to recurring pattern (if detected)
- Soft deletes not used; hard delete on user request

---

### expected_income

Stores expected/recurring income sources for "What Can I Spend?" calculations.

```sql
CREATE TABLE expected_income (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL, -- e.g., "Paycheck", "Freelance - Client A"
    amount INTEGER NOT NULL, -- in cents
    frequency VARCHAR(20) NOT NULL CHECK (frequency IN ('weekly', 'biweekly', 'monthly', 'custom')),
    next_date DATE NOT NULL, -- next expected income date
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_expected_income_user ON expected_income(user_id);
CREATE INDEX idx_expected_income_user_date ON expected_income(user_id, next_date);
CREATE INDEX idx_expected_income_active ON expected_income(user_id, is_active);
```

**Notes:**
- This is separate from transactions - tracks EXPECTED income
- When user marks income as "received", it creates a transaction and updates `next_date`
- `frequency` can be custom for irregular income (user manually sets dates)
- `is_active` allows temporarily pausing income sources without deletion
- Used by "What Can I Spend?" to calculate available funds

---

### category_budgets

Stores monthly budget allocations per category per budget (with optional user splits).

```sql
CREATE TABLE category_budgets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    budget_id UUID NOT NULL REFERENCES budgets(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    amount INTEGER NOT NULL, -- ALWAYS monthly amount in cents
    allocation_type VARCHAR(20) DEFAULT 'pooled' CHECK (allocation_type IN ('pooled', 'split_percentage', 'split_fixed')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT unique_budget_category UNIQUE(budget_id, category_id)
);

CREATE INDEX idx_category_budgets_budget ON category_budgets(budget_id);
CREATE INDEX idx_category_budgets_category ON category_budgets(category_id);
```

**Notes:**
- **All budgets are MONTHLY** (not period-specific)
- `amount` is always per month (e.g., $300/month for groceries)
- `allocation_type`:
  - `pooled`: Budget shared across all users (default, free tier)
  - `split_percentage`: Budget split by % per user (premium)
  - `split_fixed`: Budget split by fixed $ amounts per user (premium)
- For split budgets, see `category_budget_splits` table below
- Unique constraint ensures one budget per category per budget group

---

### category_budget_splits (Premium)

Stores user-specific allocations for split budgets.

```sql
CREATE TABLE category_budget_splits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_budget_id UUID NOT NULL REFERENCES category_budgets(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    allocation_percentage DECIMAL(5,2) NULL, -- For percentage splits (must total 100)
    allocation_amount INTEGER NULL, -- For fixed amount splits (must total category_budget.amount)
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
```

**Notes:**
- Premium feature only
- Either percentage OR fixed amount, never both
- Application validates:
  - For percentage: sum of all allocations = 100%
  - For fixed: sum of all allocations = category_budget.amount
- Only exists for categories with `allocation_type` = 'split_*'

---

### goals (Premium)

Stores savings goals.

```sql
CREATE TABLE goals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    target_amount INTEGER NOT NULL, -- in cents
    current_amount INTEGER DEFAULT 0, -- in cents
    target_date DATE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_goals_user ON goals(user_id);
CREATE INDEX idx_goals_user_date ON goals(user_id, target_date);
```

**Notes:**
- Premium feature only
- Users manually update `current_amount`
- Future: auto-calculate from linked transactions

---

### recurring_bills (Premium)

Stores recurring bill information.

```sql
CREATE TABLE recurring_bills (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    amount INTEGER NOT NULL, -- in cents
    category_id UUID NOT NULL REFERENCES categories(id),
    frequency VARCHAR(20) NOT NULL CHECK (frequency IN ('weekly', 'biweekly', 'monthly', 'yearly')),
    next_due_date DATE NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_recurring_bills_user ON recurring_bills(user_id);
CREATE INDEX idx_recurring_bills_user_date ON recurring_bills(user_id, next_due_date);
CREATE INDEX idx_recurring_bills_active ON recurring_bills(is_active, next_due_date);
```

**Notes:**
- Premium feature only
- `next_due_date` automatically updated when bill is marked paid
- `is_active` allows pausing bills without deletion
- Background job checks for upcoming bills and sends reminders

---

### detected_patterns

Stores auto-detected recurring transaction patterns for smart suggestions.

```sql
CREATE TABLE detected_patterns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    budget_id UUID NOT NULL REFERENCES budgets(id) ON DELETE CASCADE,
    merchant_name VARCHAR(255) NOT NULL,
    average_amount INTEGER NOT NULL, -- in cents
    frequency VARCHAR(20) NOT NULL CHECK (frequency IN ('weekly', 'biweekly', 'monthly', 'yearly')),
    confidence_score DECIMAL(3,2) NOT NULL, -- 0.00 to 1.00
    transaction_count INTEGER NOT NULL, -- Number of transactions in pattern
    last_transaction_date DATE NOT NULL,
    user_action VARCHAR(20) DEFAULT 'pending' CHECK (user_action IN ('pending', 'accepted', 'dismissed', 'ignored')),
    suggested_as_subscription BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_detected_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_detected_patterns_budget ON detected_patterns(budget_id);
CREATE INDEX idx_detected_patterns_merchant ON detected_patterns(merchant_name);
CREATE INDEX idx_detected_patterns_action ON detected_patterns(user_action);
```

**Notes:**
- Auto-populated by background job (weekly)
- `confidence_score` based on amount variance and interval regularity
- `user_action`:
  - `pending`: Not yet shown to user
  - `accepted`: User confirmed and added to subscriptions
  - `dismissed`: User said "not recurring"
  - `ignored`: User dismissed the suggestion banner
- Don't re-suggest patterns user has dismissed
- Links to transactions via `transactions.detected_pattern_id`

---

### subscriptions (Premium)

Stores user-managed subscriptions and recurring expenses.

```sql
CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    budget_id UUID NOT NULL REFERENCES budgets(id) ON DELETE CASCADE,
    created_by_user_id UUID NOT NULL REFERENCES users(id),
    name VARCHAR(255) NOT NULL,
    amount INTEGER NOT NULL, -- in cents
    billing_frequency VARCHAR(20) NOT NULL CHECK (billing_frequency IN ('weekly', 'monthly', 'yearly')),
    next_billing_date DATE NOT NULL,
    category_id UUID NOT NULL REFERENCES categories(id),
    detected_pattern_id UUID NULL REFERENCES detected_patterns(id), -- If auto-detected
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'paused', 'canceled')),
    cancel_reminder_days INTEGER NULL, -- Days before renewal to send reminder (e.g., 7)
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    canceled_at TIMESTAMP WITH TIME ZONE NULL
);

CREATE INDEX idx_subscriptions_budget ON subscriptions(budget_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);
CREATE INDEX idx_subscriptions_next_billing ON subscriptions(next_billing_date);
CREATE INDEX idx_subscriptions_pattern ON subscriptions(detected_pattern_id);
```

**Notes:**
- Premium feature only
- Can be manually added OR auto-created from detected_patterns
- `status`:
  - `active`: Currently subscribed
  - `paused`: Temporarily disabled (no reminders)
  - `canceled`: Canceled but kept for history
- `cancel_reminder_days`: Send reminder N days before renewal (useful for annual subscriptions)
- Background job sends reminders before `next_billing_date`
- Calculate monthly cost for budgeting: yearly ÷ 12, weekly × 4.33

---

## Migrations

We'll use a migration tool for database versioning:

**Options:**
- **golang-migrate** - Simple, CLI-based
- **goose** - Popular, supports both SQL and Go migrations
- **GORM AutoMigrate** - Built into GORM (less control)

**Recommendation:** `golang-migrate` for explicit control

### Migration Files Structure

```
backend/migrations/
├── 000001_create_users_table.up.sql
├── 000001_create_users_table.down.sql
├── 000002_create_categories_table.up.sql
├── 000002_create_categories_table.down.sql
├── 000003_seed_categories.up.sql
├── 000003_seed_categories.down.sql
├── 000004_create_transactions_table.up.sql
├── 000004_create_transactions_table.down.sql
└── ...
```

---

## Indexes Strategy

**Primary Indexes (already included):**
- All tables have primary key on `id`
- Foreign keys have indexes for joins
- User-scoped queries indexed on `(user_id, ...)`

**Query Patterns:**
1. "Get all transactions for user X in date range" → `idx_transactions_user_date`
2. "Get all budgets for user X in month Y" → `idx_budgets_user_period`
3. "Get upcoming bills for all users" → `idx_recurring_bills_active`

**Future Optimizations:**
- Partial indexes for premium users if needed
- Covering indexes for common SELECT queries
- Consider materialized views for complex reports (premium)

---

## Data Retention Policy

**Free Tier:**
- Keep transactions for 1 year
- Budgets: current + 12 months
- Goals: N/A (premium only)

**Premium Tier:**
- Unlimited transaction history
- Unlimited budget history
- Unlimited goals

**Deleted Users:**
- Hard delete all user data within 30 days
- Anonymize transaction data for analytics (optional)

---

## Backup Strategy

**Supabase Automatic Backups:**
- Daily backups (retained for 7 days on free, 30 days on Pro)
- Point-in-time recovery (PITR) available on Pro plan

**Additional Backups (Production):**
- Weekly exports to S3 (via pg_dump)
- Monthly archival backups

---

## Security Considerations

1. **Row Level Security (RLS):**
   Enable RLS policies to ensure users can only access their own data:
   ```sql
   ALTER TABLE transactions ENABLE ROW LEVEL SECURITY;

   CREATE POLICY transactions_user_policy ON transactions
   FOR ALL
   USING (user_id = auth.uid());
   ```

2. **Sensitive Data:**
   - No credit card info stored (handled by Stripe)
   - No SSN or sensitive PII
   - Email is only PII (encrypted at rest by Supabase)

3. **Audit Logging (Future):**
   - Track sensitive operations (account deletion, data exports)
   - Store in separate `audit_logs` table

---

## Performance Monitoring

**Queries to Monitor:**
1. Transaction list with filters (most common)
2. Budget summary calculations
3. Dashboard summary aggregations
4. Report generation queries (premium)

**Tools:**
- Supabase Dashboard (query performance)
- pg_stat_statements extension
- Slow query log

**Optimization Targets:**
- All queries < 100ms (p95)
- Dashboard load < 500ms total
- Transaction list < 200ms

---

## Future Schema Additions

**Phase 3+:**
- `accounts` - Bank accounts (for Plaid sync)
- `plaid_items` - Plaid connection metadata
- `subscriptions` - Stripe subscription tracking
- `notifications` - In-app notifications
- `audit_logs` - Security audit trail
- `app_sessions` - Mobile app sessions

---

**Document Version:** 1.0
**Last Updated:** 2025-12-30
