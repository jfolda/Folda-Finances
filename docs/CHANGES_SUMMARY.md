# Major Changes Summary - Multi-User Budgets & Subscriptions

**Date:** 2025-12-30
**Status:** Design Complete - Ready for Implementation

---

## Overview

Folda Finances has been redesigned with four major improvements:

1. **Multi-user budgets** - Up to 5 users share one budget
2. **Monthly budgets only** - "What Can I Spend?" pro-rates to user's view period
3. **Smart recurring detection** - Auto-detect subscriptions and patterns
4. **Subscriptions tab** - Premium feature to manage all subscriptions

---

## Key Changes

### 1. Multi-User Budgets (FREE TIER!)

**What Changed:**
- Users now belong to a **budget** (shared entity)
- Each budget can have up to 5 user accounts
- Budget creator invites others via email
- Each user has their own login but shares the budget

**Why This Matters:**
- Perfect for couples, families, roommates
- See who spent what (accountability)
- Optional: Split category budgets between users (premium)

**Budget Roles & Permissions:**

| Action | Owner | Admin | Read/Write | Read-Only |
|--------|-------|-------|------------|-----------|
| View budget & transactions | ✅ | ✅ | ✅ | ✅ |
| Add transactions | ✅ | ✅ | ✅ | ❌ |
| Edit own transactions | ✅ | ✅ | ✅ | ❌ |
| Delete own transactions | ✅ | ✅ | ✅ | ❌ |
| Edit budgets | ✅ | ✅ | ❌ | ❌ |
| Invite members | ✅ | ✅ | ❌ | ❌ |
| Remove members | ✅ | ✅ | ❌ | ❌ |
| Change member roles | ✅ | ❌ | ❌ | ❌ |
| Delete budget | ✅ | ❌ | ❌ | ❌ |
| Transfer ownership | ✅ | ❌ | ❌ | ❌ |

**Implementation:**
- New `budgets` table (parent entity)
- New `budget_invitations` table (with `invited_role`)
- `users.budget_id` foreign key
- `users.budget_role` (owner/admin/read_write/read_only)
- `transactions.user_id` + `transactions.budget_id`

---

### 2. Monthly Budgets with Pro-Rated View Periods

**What Changed:**
- ALL budgets are **monthly** (stored as monthly amounts)
- Users choose their **view period** (weekly/biweekly/monthly) for "What Can I Spend?"
- App pro-rates monthly budgets to the view period

**Example:**
- Budget: $300/month for groceries
- User's view period: Biweekly
- "What Can I Spend?" shows: $138 available for this 2-week period

**Pro-ration Formula:**
```
Weekly: Monthly × (7 / 30.44) = Monthly × 0.23
Biweekly: Monthly × (14 / 30.44) = Monthly × 0.46
Monthly: Monthly × 1.00
```

**Why This Matters:**
- Simplifies budget management (one monthly number)
- Still provides "What Can I Spend?" flexibility
- Easier for users to think in monthly terms

**Implementation:**
- `users.view_period` instead of `spending_period`
- `category_budgets.amount` ALWAYS monthly
- Backend calculates pro-rated amount on-the-fly

---

### 3. Split Category Budgets (PREMIUM)

**What Changed:**
- Free tier: All category budgets are "pooled" (shared by all users)
- Premium: Can split budgets between users

**Split Types:**
- **Percentage split**: "Alice gets 60%, Bob gets 40%"
- **Fixed amount split**: "Alice gets $180, Bob gets $120"

**Example:**
- Monthly restaurant budget: $300
- Split: Alice 60%, Bob 40%
- Alice's biweekly view: $83 available (60% of $138)
- Bob's biweekly view: $55 available (40% of $138)

**Why This Matters:**
- Perfect for couples with different spending habits
- Accountability without micromanagement
- Flexible: Some categories pooled, others split

**Implementation:**
- `category_budgets.allocation_type` (pooled, split_percentage, split_fixed)
- New `category_budget_splits` table
- Premium feature gate

---

### 4. Recurring Transaction Learning (FREE + PREMIUM)

**What Changed:**
- App auto-detects recurring transactions
- Looks for same merchant + similar amount + regular frequency
- After 2-3 occurrences, suggests adding to subscriptions

**Free Tier:**
- Detects patterns
- Badges recurring transactions
- Shows notification

**Premium Tier:**
- All free features
- One-tap add to Subscriptions tab
- Full subscription management

**Detection Algorithm:**
```
Group transactions by merchant (fuzzy match)
For each merchant with 3+ transactions:
  Calculate average amount and std deviation
  Calculate time intervals between transactions
  If std deviation < 5% AND intervals regular (±3 days):
    Create detected_pattern with confidence score
```

**Why This Matters:**
- Users don't have to manually track subscriptions
- Discover forgotten subscriptions
- Upsell to premium (subscriptions tab)

**Implementation:**
- New `detected_patterns` table
- `transactions.merchant_name` (normalized)
- `transactions.detected_pattern_id` (link)
- Background job runs weekly
- Badge component in UI

---

### 5. Subscriptions Tab (PREMIUM)

**What Changed:**
- Dedicated tab to manage all subscriptions
- Auto-populated from detected patterns OR manually added
- Shows total monthly subscription cost
- Renewal reminders

**Features:**
- List all active/paused/canceled subscriptions
- Service name, amount, billing frequency, next billing date
- Visual calendar of upcoming bills
- Monthly cost summary
- Cancel reminders (e.g., "Annual renewal in 7 days")
- Export to CSV/PDF

**Why This Matters:**
- High-value premium feature
- Solves real pain point (subscription management)
- Sticky feature (users check it regularly)
- Natural upsell from free tier pattern detection

**Implementation:**
- New `subscriptions` table
- Link to `detected_patterns` (if auto-created)
- Background job for renewal reminders
- Premium feature gate
- Calendar UI component

---

## Database Schema Changes

### New Tables

1. **budgets** - Parent budget entity (multiple users belong to one)
2. **budget_invitations** - Pending budget invites
3. **category_budgets** - Budget amounts per category per budget (MONTHLY)
4. **category_budget_splits** - User splits for split budgets (premium)
5. **detected_patterns** - Auto-detected recurring patterns
6. **subscriptions** - User-managed subscriptions (premium)

### Modified Tables

1. **users**
   - Added: `budget_id`, `name`
   - Changed: `spending_period` → `view_period`

2. **transactions**
   - Added: `budget_id`, `merchant_name`, `detected_pattern_id`

3. **expected_income**
   - No changes (still exists, separate feature)

### Removed Concepts

- ~~Period-based budgets~~ → All budgets are monthly now

---

## API Changes Needed

### New Endpoints

**Budgets:**
- `POST /api/budgets` - Create budget
- `GET /api/budgets/:id` - Get budget details
- `GET /api/budgets/:id/members` - List budget members
- `POST /api/budgets/:id/invite` - Invite user to budget
- `DELETE /api/budgets/:id/members/:user_id` - Remove member

**Budget Invitations:**
- `GET /api/budget-invitations` - List my invitations
- `POST /api/budget-invitations/:token/accept` - Accept invitation
- `POST /api/budget-invitations/:token/decline` - Decline invitation

**Category Budgets:**
- `POST /api/category-budgets` - Create/update category budget
- `PUT /api/category-budgets/:id/allocation` - Set split allocation (premium)
- `GET /api/category-budgets` - List all category budgets for my budget

**Detected Patterns:**
- `GET /api/detected-patterns` - List detected patterns
- `POST /api/detected-patterns/:id/accept` - Accept pattern (free: badge only, premium: add to subscriptions)
- `POST /api/detected-patterns/:id/dismiss` - Dismiss pattern
- `POST /api/detected-patterns/:id/ignore` - Ignore suggestion

**Subscriptions (Premium):**
- `GET /api/subscriptions` - List all subscriptions
- `POST /api/subscriptions` - Create subscription
- `PUT /api/subscriptions/:id` - Update subscription
- `DELETE /api/subscriptions/:id` - Cancel subscription
- `GET /api/subscriptions/summary` - Get monthly cost summary

### Modified Endpoints

**"What Can I Spend?"**
- `GET /api/spending/available`
  - Now calculates pro-rated budgets based on `view_period`
  - For split budgets: Only show user's allocation
  - For pooled budgets: Show total for all users

**Transactions:**
- `POST /api/transactions` - Now requires `budget_id`, auto-extracts `merchant_name`
- `GET /api/transactions` - Can filter by `user_id` (who made it)

**User Settings:**
- `PATCH /api/auth/me` - Update `view_period` (not `spending_period`)

---

## Frontend Changes Needed

### New Pages

1. **Budget Settings** - Manage budget members, send invitations
2. **Subscriptions** (Premium) - Full subscription management tab
3. **Budget Invitations** - Accept/decline invitations

### Modified Pages

1. **"What Can I Spend?"**
   - Show "Shared" vs "Your share: $X" for split budgets
   - Period toggle: Weekly/Biweekly/Monthly (view only)
   - Pro-rated amounts in UI

2. **Budget Creation**
   - Set monthly amounts (not period-specific)
   - Split allocation UI (premium)
   - Visual split pie chart

3. **Transactions List**
   - Show user avatar/badge (who made transaction)
   - Filter by user
   - Badge for detected recurring transactions

4. **Settings**
   - View period selector (weekly/biweekly/monthly)
   - Budget member management
   - Invite to budget button

### New Components

1. **UserAvatar** - Show user initial/avatar on transactions
2. **RecurringBadge** - Badge for detected recurring transactions
3. **PatternSuggestionBanner** - "We detected Netflix $15.99/month. Add to subscriptions?"
4. **SubscriptionCard** - Card component for subscription list
5. **SplitAllocationPicker** - UI for setting budget splits
6. **MemberList** - List of budget members with roles

---

## Implementation Priority

### Phase 1: Multi-User Budgets (Week 1-3)
1. Database migrations (all new tables)
2. Budget creation + invitation system
3. Transaction tracking with user_id
4. "What Can I Spend?" with pooled budgets
5. User filtering in transactions list

### Phase 2: Pro-Rated View Periods (Week 4-5)
1. Update budget storage to monthly only
2. Pro-ration logic in "What Can I Spend?" API
3. View period selector UI
4. Update onboarding flow

### Phase 3: Recurring Detection (Week 6-7)
1. Background job for pattern detection
2. Merchant name normalization
3. Pattern suggestion UI
4. Badge recurring transactions
5. "Detected Patterns" settings page

### Phase 4: Subscriptions Tab (Week 8-9) - PREMIUM
1. Subscriptions CRUD API
2. Subscriptions page UI
3. Calendar view
4. Monthly cost summary
5. Renewal reminders (background job)
6. Link pattern detection → subscriptions

### Phase 5: Split Budgets (Week 10-11) - PREMIUM
1. Category budget splits API
2. Split allocation UI
3. "What Can I Spend?" split calculations
4. Per-user spending insights

---

## Migration Strategy

### For Existing Users (If Any)

1. **Create default budget** for each user
2. **Migrate** `user.spending_period` → `user.view_period`
3. **Convert** existing budgets to monthly (if period-based)
4. **Set** all `category_budgets.allocation_type` = 'pooled'

### For New Users

1. **Onboarding** creates budget automatically
2. **Ask** for view period preference
3. **Default** to monthly budgets, pooled allocation
4. **Offer** invitation flow during onboarding

---

## Testing Checklist

### Multi-User Budgets
- [ ] Create budget
- [ ] Send invitation email
- [ ] Accept invitation (join budget)
- [ ] Decline invitation
- [ ] View budget members
- [ ] Remove budget member
- [ ] Enforce 5-member limit
- [ ] Transactions show correct user
- [ ] Filter transactions by user

### Pro-Rated View Periods
- [ ] Set view period to weekly
- [ ] "What Can I Spend?" shows correct pro-rated amount
- [ ] Set view period to biweekly
- [ ] Verify calculation accuracy
- [ ] Change budget amount, verify pro-ration updates

### Recurring Detection
- [ ] Add 3 Netflix transactions
- [ ] Pattern detected after weekly job
- [ ] Suggestion shown to user
- [ ] Accept → adds badge (free tier)
- [ ] Accept → adds to subscriptions (premium)
- [ ] Dismiss → not shown again

### Subscriptions
- [ ] Add subscription manually
- [ ] Edit subscription
- [ ] Cancel subscription
- [ ] View monthly cost summary
- [ ] Receive renewal reminder
- [ ] Export to CSV

### Split Budgets
- [ ] Set category to split_percentage
- [ ] Assign % to each user (must total 100%)
- [ ] "What Can I Spend?" shows correct user allocation
- [ ] Set category to split_fixed
- [ ] Assign $ to each user (must total budget)
- [ ] User can only spend their allocation

---

## Metrics to Track

### Engagement
- % of users in multi-user budgets
- Average budget member count
- Invitation acceptance rate

### Premium Conversion
- % users who upgrade for split budgets
- % users who upgrade for subscriptions tab
- Subscription management frequency

### Feature Usage
- Pattern detection accuracy
- % patterns accepted vs dismissed
- Subscription count per premium user
- Average monthly subscription cost tracked

---

## Open Questions - RESOLVED

1. **Budget Permissions System** ✅
   - **Owner** (budget creator): Full admin permissions
   - **Admin**: Can invite/remove members, edit budgets (owner can grant)
   - **Read/Write** (default for invited users): Can add transactions, view everything
   - **Read-Only**: Can only view (owner can set)
   - Owner controls permissions when inviting users

2. **What happens when owner deletes their account?** ✅
   - Budget is deleted
   - All budget data is deleted
   - All members lose access to that budget
   - (Future: Prompt owner to transfer ownership before deleting account)

3. **Can users switch budgets?** ✅
   - Phase 1: No (one budget per user)
   - Phase 4: Yes (multi-budget support - premium)

4. **Pattern detection scope** ✅
   - Detect all patterns across all categories
   - Don't limit to specific categories

5. **Pattern detection frequency** ✅
   - Weekly background job

---

## Documentation Updates Needed

- [x] REQUIREMENTS.md - Updated with all new features
- [x] DATABASE.md - All new tables and schema changes
- [ ] API.md - New endpoints and modified responses
- [ ] WHAT_CAN_I_SPEND.md - Update with pro-ration logic
- [x] This summary document

---

**Document Version:** 1.0
**Last Updated:** 2025-12-30
**Status:** Ready for review and implementation
