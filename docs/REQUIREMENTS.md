# Folda Finances - Requirements Document

## Executive Summary

**Project Name:** Folda Finances
**Type:** Personal Finance SaaS Application
**Target Audience:** Individual consumers managing personal finances
**Business Model:** Freemium (free basic tier, premium subscription)
**Timeline:** 1-2 month MVP, iterative improvements thereafter
**Region:** United States only (initially)

## Vision & Goals

Build a simple, intuitive budgeting application centered around one powerful question: **"What Can I Spend?"**

Unlike traditional budgeting apps that focus on tracking what you *already spent*, Folda Finances helps users make better spending decisions in real-time by showing them exactly how much they have available to spend **right now** in each category, based on their income, bills, goals, and spending patterns.

**Claim to Fame:** The "What Can I Spend?" feature that gives users a real-time, per-category view of their available spending money for the current period (weekly, biweekly, or monthly).

## Tech Stack

### Frontend
- **Framework:** React 18+ with TypeScript
- **Build Tool:** Vite
- **Routing:** React Router v6
- **State Management:** TanStack Query (React Query)
- **Styling:** Tailwind CSS + shadcn/ui components
- **Charts:** Recharts (premium features)
- **Date Handling:** date-fns
- **Auth:** Supabase Auth
- **Hosting:** Vercel

### Backend
- **Language:** Go 1.21+
- **API Style:** REST
- **Router:** Chi or Gin
- **ORM:** GORM
- **Database:** PostgreSQL (Supabase)
- **Auth:** Supabase (JWT validation)
- **Hosting:** Railway, Fly.io, or Render

### Infrastructure
- **Database & Auth:** Supabase
- **Payment Processing:** Stripe (Phase 2)
- **Email:** Supabase Auth emails + SendGrid/Resend (for bill reminders)
- **Monitoring:** TBD (Sentry for errors, optional)

## Feature Requirements

### Phase 1: MVP - Free Tier (Launch Target: 1-2 months)

#### 1. User Authentication & Account Management
**Priority:** P0 (Must Have)

**Features:**
- Email/password registration
- Email verification
- Login/logout
- Password reset via email
- Basic user profile (name, email)
- Account deletion

**Technical Notes:**
- Use Supabase Auth for all authentication flows
- JWT tokens for API authentication
- Secure password requirements (min 8 chars, complexity)

---

#### 2. Budget & User Account Management
**Priority:** P0 (Must Have) - **CORE FEATURE**

**Multi-User Budget System:**
- Each user account belongs to ONE budget (initially)
- Each budget can have up to 5 user accounts associated
- Budget creator invites others via email
- Invitees create their own login/account but join the shared budget

**Budget Structure:**
- All budgets and income are tracked **monthly**
- Budget amounts are always per month (e.g., $300/month for groceries)
- Expected income is monthly (e.g., "$4,000/month salary")

**Category Budget Allocation (Per Budget):**
- Default: Budget pool shared across all users
- Optional: Split budget between users
  - Percentage split (e.g., Alice 60%, Bob 40%)
  - Fixed amount split (e.g., Alice $180, Bob $120)
  - Mixed: Some categories pooled, others split

**UI/UX:**
- Budget creation wizard on first signup
- "Invite to Budget" button generates email invitation
- Pending invitations management
- Budget member list with roles (creator, member)
- Per-category budget allocation settings

**Technical Notes:**
- Store budget_id on user account
- Budget invitation system with token-based links
- Enforce 5-user limit per budget (configurable for future expansion)
- Track which user made each transaction (for split budgets)
- Calculate "available to spend" differently for pooled vs. split categories

---

#### 3. "What Can I Spend?" Dashboard
**Priority:** P0 (Must Have) - **CLAIM TO FAME**

This is the killer feature that sets Folda Finances apart.

**View Period Selection:**
- User chooses their view period:
  - Weekly (see what's left this week)
  - Biweekly (see what's left this 2-week period)
  - Monthly (see what's left this month)
- User sets period start date (e.g., "payday is every other Friday")
- **Important:** Budgets are ALWAYS monthly, but app pro-rates to show for the chosen period

**Features:**
- Prominently displayed "What Can I Spend?" view
- Per-category breakdown showing:
  - Category name and icon
  - Budget allocated for this VIEW period (pro-rated from monthly)
  - Amount spent so far this period
  - **Available to spend** (large, prominent number)
  - Visual progress bar
  - Color coding (green = plenty left, yellow = running low, red = over budget)
- For split categories: Show "Your share: $X available"
- For pooled categories: Show "Shared: $X available"
- Quick tap to see calculation breakdown:
  - Monthly budget: $X
  - Period budget: $Y (pro-rated)
  - Minus spent: -$Z
  - Equals available: $W
- Overall summary card:
  - Total available across all categories
  - Days remaining in period
  - Daily average you can spend

**UI/UX:**
- Make this the DEFAULT landing page (not dashboard)
- Large, easy-to-read numbers
- Period selector at top (Weekly/Biweekly/Monthly toggle)
- One-tap access from anywhere in app
- Optimized for quick glance while shopping
- Offline-capable (cache last calculation)
- Pull-to-refresh to recalculate

**Calculation Logic (per category):**
```
Monthly Budget = $300 (example: Groceries)

Period Budget (pro-rated) =
  If Weekly: Monthly Budget × (7 / 30.44) = $300 × 0.23 = $69
  If Biweekly: Monthly Budget × (14 / 30.44) = $300 × 0.46 = $138
  If Monthly: Monthly Budget = $300

Spent This Period = Sum of transactions in current period

For Pooled Categories:
  Available = Period Budget - Spent by All Users

For Split Categories:
  Your Allocation = Period Budget × Your Split %
  Your Spent = Sum of YOUR transactions
  Your Available = Your Allocation - Your Spent

If negative: Show "Over budget by $X" in red
If positive: Show "Available: $X" in green/yellow
```

**Example Use Case:**
Alice has her period set to "biweekly" (every other Friday):
- Monthly restaurant budget: $300
- Split: Alice 60%, Bob 40%
- Current biweekly period: Jan 5 - Jan 18

Alice's view:
1. Opens app (lands on "What Can I Spend?")
2. Looks at "Dining & Restaurants" category
3. Sees:
   - Period budget: $138 (pro-rated from $300/month)
   - Your share: $83 (60% of $138)
   - You've spent: $45
   - **Available: $38**
4. Decides she can afford a $25 lunch
5. After lunch, adds transaction
6. Available updates to $13

---

#### 4. Manual Transaction Entry
**Priority:** P0 (Must Have)

**Features:**
- Create transaction with:
  - Amount (positive for income, negative for expense)
  - Description/memo
  - Category (from predefined list)
  - Date (defaults to today)
- Edit existing transactions
- Delete transactions
- View transaction list (sorted by date, newest first)
- Filter transactions by:
  - Date range
  - Category
  - Type (income/expense)

**UI/UX:**
- Quick-add button prominently displayed
- Form validation (amount required, category required)
- Mobile-friendly transaction entry
- Confirmation before deletion

**Technical Notes:**
- Store amounts as integers (cents) to avoid floating-point issues
- Support negative amounts for expenses OR use transaction_type field
- Pagination for transaction list (50 per page)

---

#### 5. Predefined Categories
**Priority:** P0 (Must Have)

**Default Categories (System-Provided):**

**Expenses:**
- Housing (rent, mortgage)
- Utilities (electric, water, gas, internet)
- Groceries
- Dining & Restaurants
- Transportation (gas, public transit, parking)
- Healthcare
- Shopping (clothing, personal items)
- Entertainment
- Other Expenses

**Income:**
- Salary/Wages
- Freelance
- Investments
- Other Income

**Features:**
- Categories are read-only in free tier
- Each category has a color and icon
- Categories displayed in transaction entry form

**Technical Notes:**
- Seed categories in database on initial migration
- Associate categories with user_id = NULL for system categories
- Custom categories reserved for premium tier

---

#### 6. Monthly Budget Creation & Allocation
**Priority:** P0 (Must Have)

**Features:**
- Create **monthly** budget for each category
- Set budget amount per month (e.g., $300/month groceries)
- One active budget per category per budget group
- Edit budget amounts
- Delete budgets
- Configure category allocation:
  - **Pooled** (default): Shared by all budget members
  - **Split by percentage**: Assign % to each member (must total 100%)
  - **Split by fixed amount**: Assign dollar amounts (must total monthly budget)

**Budget Display:**
- List all budgets with monthly amounts
- Show current month spending vs. budget
- For split budgets: Show each member's allocation
- Visual progress bar per category
- Color coding:
  - Green: under budget (< 75%)
  - Yellow: approaching limit (75-100%)
  - Red: over budget (> 100%)

**UI/UX:**
- Simple budget setup wizard for first-time users
- "Smart suggest" based on monthly income (suggest budgets totaling 70% of income)
- Category allocation modal:
  - Toggle: Pooled / Split
  - If split: Choose percentage or fixed amount
  - Visual split pie chart
- Copy previous month's budgets option
- Budget overview page showing all categories

**Technical Notes:**
- All budgets stored as monthly amounts
- Split configuration stored per category_budget
- When calculating "What Can I Spend?", pro-rate monthly to user's view period
- Track transaction creator (user_id) for split budget calculations
- Validate: split percentages = 100%, split amounts = monthly budget

---

#### 7. Summary Dashboard (Secondary View)
**Priority:** P1 (Should Have)

**Note:** The "What Can I Spend?" view is the PRIMARY/DEFAULT view. This is a secondary dashboard for historical overview.

**Dashboard Components:**
- Current period summary:
  - Total income (expected + received)
  - Total expenses
  - Net (income - expenses)
  - Days remaining in period
- Period progress:
  - Visual timeline showing where in period you are
  - Spending pace (on track, ahead, behind)
- Budget health overview:
  - Number of categories on track
  - Number of categories over budget
- Recent transactions (last 5-10)
- Quick actions:
  - Add transaction
  - Mark income received
  - Adjust budget

**UI/UX:**
- Accessible via tab/nav from "What Can I Spend?"
- Clean, scannable layout
- Mobile-responsive
- Loading states for all data
- Empty states with helpful prompts and onboarding nudges

---

#### 8. Transaction List & Filtering
**Priority:** P0 (Must Have)

**Features:**
- Paginated transaction list
- Sort by date (newest first, oldest first)
- Filter by:
  - Date range (preset: this month, last month, last 3 months, custom)
  - Category (multi-select)
  - Type (income, expense, all)
  - User (for multi-user budgets - filter by who made the transaction)
- Search by description
- Bulk select (future: bulk delete, bulk categorize)
- Show transaction creator badge (for multi-user budgets)

**UI/UX:**
- Sticky filters bar
- Clear/reset filters button
- Transaction count displayed
- Loading skeletons during fetch
- User avatar/initial on each transaction (multi-user budgets)
- Badge for detected recurring transactions

---

#### 9. Recurring Transaction Learning (Smart Detection)
**Priority:** P1 (Should Have - MVP)

**Features:**
- Automatically detect recurring transaction patterns:
  - Same merchant name (fuzzy match, e.g., "NETFLIX.COM" = "Netflix")
  - Similar amount (within 5% variance)
  - Regular frequency (weekly, biweekly, monthly, yearly)
- When pattern detected (after 2-3 occurrences):
  - Show smart notification: "We detected Netflix $15.99 monthly. Add to subscriptions?"
  - One-tap to add to Subscriptions tab (premium feature)
  - Option to dismiss or "don't ask about this again"
- Badge recurring transactions in transaction list
- Suggest categorization for future similar transactions

**Detection Patterns:**
- Subscriptions (Netflix, Spotify, Disney+, gym membership)
- Bills (utilities, phone, internet, insurance)
- Recurring expenses (weekly grocery trips, monthly haircuts)

**UI/UX:**
- Non-intrusive notification banner on transaction list or dashboard
- Badge on recurring transactions: "Recurring" with frequency
- "Detected Patterns" section in settings to review/manage all detected patterns
- User can confirm/deny/ignore each pattern

**Technical Notes:**
- Run pattern detection weekly via background job
- Algorithm:
  ```
  Group transactions by merchant name (fuzzy match using Levenshtein distance)
  For each merchant group with 3+ transactions:
    Calculate average amount and standard deviation
    Calculate time intervals between transactions
    If std deviation < 5% AND intervals are regular (±3 days):
      Create "detected_pattern" record with confidence score
  ```
- Store detection metadata (confidence score, frequency, last_detected)
- Don't re-suggest patterns user has dismissed
- For free tier: Detect and badge only
- For premium: Detect, badge, and add to Subscriptions tab

---

### Phase 2: Premium Features (Post-MVP)

#### 10. Subscriptions Tab
**Priority:** P1 (Premium) - **HIGH VALUE FEATURE**

Track and manage all recurring subscriptions in one place.

**Features:**
- List all subscriptions with:
  - Service name (e.g., "Netflix")
  - Amount
  - Billing frequency (monthly, yearly)
  - Next billing date
  - Category
  - Auto-detected or manually added
- Add subscription manually:
  - Name, amount, frequency, next billing date, category
- Edit/cancel subscriptions
- "Upcoming" view: See all subscriptions due in next 30 days
- Total monthly subscription cost summary
- "Cancel reminder" - Set reminder before renewal (e.g., 7 days before annual renewal)
- Link to detected recurring transactions

**UI/UX:**
- Dedicated "Subscriptions" tab/page
- Card-based layout with service logos (if available)
- Sort by: next billing date, cost, alphabetical
- Filter by: active, canceled, category
- Visual calendar showing when each subscription bills
- Monthly cost breakdown chart
- "You're spending $X/month on subscriptions" summary card
- Integration with transaction learning: "Add this recurring charge to subscriptions?"

**Subscription Actions:**
- Mark as canceled (keeps history but removes from active list)
- "Pause subscription" (temporarily disable renewal reminders)
- Export list to CSV/PDF
- Share subscription split with budget members

**Technical Notes:**
- Store in `subscriptions` table (separate from `recurring_bills`)
- Link to merchant from transaction learning
- Send email/push notifications before billing date
- Calculate total monthly cost accounting for different frequencies:
  ```
  Monthly Cost =
    (Monthly subscriptions sum) +
    (Yearly subscriptions sum / 12) +
    (Weekly subscriptions sum × 4.33)
  ```
- Premium feature gate

---

#### 11. Savings Goals
**Priority:** P1 (Premium)

**Features:**
- Create savings goal with:
  - Name (e.g., "Emergency Fund")
  - Target amount
  - Target date
  - Current amount (optional)
- Track progress toward goal
- Update current amount (manual contributions)
- Calculate monthly savings needed
- Visual progress indicators

**UI/UX:**
- Goals dashboard with progress cards
- Celebrate milestone achievements (25%, 50%, 75%, 100%)
- Suggested monthly contribution calculator

---

#### 12. Advanced Reports & Analytics
**Priority:** P1 (Premium)

**Features:**
- Spending trends over time (line/bar charts)
- Category breakdown (pie/donut charts)
- Income vs. expenses over time
- Month-over-month comparison
- Year-over-year comparison
- Top spending categories
- Average monthly spending
- Custom date range analysis

**Export Options:**
- Download reports as PDF
- Export transaction data as CSV
- Monthly spending summary email

**UI/UX:**
- Interactive charts
- Drill-down capabilities
- Date range selector
- Save favorite report views

---

#### 13. Recurring Bill Tracking & Reminders
**Priority:** P1 (Premium)

**Features:**
- Add recurring bills with:
  - Name (e.g., "Netflix")
  - Amount
  - Category
  - Frequency (weekly, biweekly, monthly, yearly)
  - Next due date
- Automatic next-date calculation
- Email reminders (3 days before, day of)
- Mark bill as paid (creates transaction)
- Bill calendar view

**UI/UX:**
- Upcoming bills widget on dashboard
- Calendar view showing all due dates
- One-click "mark as paid" action

**Technical Notes:**
- Background job to send reminder emails
- Option to auto-create transactions from recurring bills

---

#### 14. Custom Categories
**Priority:** P1 (Premium)

**Features:**
- Create custom categories with:
  - Name
  - Color (color picker)
  - Icon (icon selector)
- Edit custom categories
- Delete custom categories (only if no transactions)
- Merge categories (move all transactions to another category)

**Limits:**
- Max 50 custom categories per user

---

### Phase 3: Future Features (Post-Launch)

#### 15. Bank Account Sync
**Priority:** P2 (Future Premium)

**Features:**
- Connect bank accounts via Plaid
- Automatic transaction import
- Reconciliation view (match imported to manual)
- Multi-account support
- Account balance tracking
- Net worth calculation

**Technical Notes:**
- Plaid integration for US banks
- Webhook handling for transaction updates
- Secure credential storage

---

#### 16. Mobile Applications
**Priority:** P2 (Future Premium)

**Features:**
- iOS and Android native apps
- All web features available
- Mobile-optimized UI
- Push notifications for bill reminders
- Quick transaction entry widget

**Technical Notes:**
- React Native or Flutter
- Share API with web app
- App Store and Google Play distribution

---

#### 17. Multi-Budget Support (Phase 4)
**Priority:** P3 (Future)

Allow users to belong to multiple budgets simultaneously.

**Features:**
- User can be member of multiple budgets
- Switch between budgets via dropdown/selector
- Different roles per budget (creator, admin, member)
- Budget-specific settings and permissions
- Cross-budget reporting (premium)

**Use Cases:**
- Personal budget + household budget
- Multiple household budgets (parents' house + own apartment)
- Business budget + personal budget

---

#### 18. Budget Reallocation (Quick Transfer)
**Priority:** P2 (Future Premium)

Allow users to transfer budgeted funds between categories on-the-fly.

**Features:**
- **Quick transfer** from "What Can I Spend?" view:
  - User sees "Groceries: $5 available" but needs $50
  - Tap small square "+" button next to category
  - See list of other categories with available funds
  - Select "Restaurants: $87 available"
  - Transfer $45 from Restaurants to Groceries
  - New balances: Groceries $50, Restaurants $42
- Transfer history/audit trail
- Undo last transfer (within current period)
- Smart suggestions: "You have $X available in [category], transfer some?"

**Use Cases:**
- At grocery store, total is higher than budgeted
- Unexpected expense in one category
- Month-end budget juggling
- Emergency flexibility

**For Split Budgets:**
- Can only transfer YOUR allocated portion
- Cannot transfer from pooled categories
- Cannot transfer from other users' allocations
- Premium feature: Transfer between users (with approval)

**UI/UX:**
- Small square "+" button on category card
- Transfer modal with slider or amount input
- Visual before/after comparison
- Confirmation: "Transfer $X from [Cat A] to [Cat B]?"
- Toast notification: "Transferred $X. Groceries now has $Y available"

**Technical Notes:**
- Create `budget_transfers` table to log transfers
- Transfers are period-specific (don't affect monthly budget amounts)
- Recalculate "What Can I Spend?" after transfer
- Validate: can't transfer more than available in source category
- Premium feature gate

**Database:**
```sql
CREATE TABLE budget_transfers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    budget_id UUID NOT NULL REFERENCES budgets(id),
    from_category_id UUID NOT NULL REFERENCES categories(id),
    to_category_id UUID NOT NULL REFERENCES categories(id),
    amount INTEGER NOT NULL, -- in cents
    period_start_date DATE NOT NULL,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

---

#### 19. Additional Future Features
- Investment tracking
- Bill negotiation suggestions
- Spending insights & AI recommendations
- Receipt scanning & OCR
- Multi-currency support
- Budget templates
- Financial advisor integrations
- Shared expense splitting (Venmo-style within budget)

---

## User Tiers & Monetization

### Free Tier

**Included Features:**
- ✅ **"What Can I Spend?" real-time view** (CORE FEATURE)
- ✅ View period configuration (weekly/biweekly/monthly display)
- ✅ **Multi-user budgets** (up to 5 users per budget)
- ✅ Budget member invitations
- ✅ Pooled category budgets
- ✅ Manual transaction entry (unlimited)
- ✅ Transaction filtering by user
- ✅ Predefined categories
- ✅ Monthly budgets with pro-rated period view
- ✅ Summary dashboard
- ✅ Transaction filtering & search
- ✅ **Recurring transaction detection** (badge only)
- ✅ Up to 1 year of transaction history

**Limitations:**
- ❌ No split category budgets (percentage/fixed splits)
- ❌ No custom categories
- ❌ No savings goals
- ❌ **No Subscriptions tab** (can see detected patterns but can't manage them)
- ❌ No advanced reports
- ❌ No recurring bill reminders
- ❌ No data export
- ❌ No bank sync
- ❌ No mobile apps
- ❌ Limited to 1 budget per user (multi-budget support is future premium)

### Premium Tier ($9.99/month or $99/year)

**All Free Features Plus:**
- ✅ **Split category budgets** (percentage & fixed amount splits)
- ✅ **Subscriptions tab** - Full subscription management & tracking
- ✅ Subscription renewal reminders
- ✅ Custom categories (up to 50)
- ✅ Savings goals (unlimited)
- ✅ Advanced reports & analytics
- ✅ Spending insights by user (for multi-user budgets)
- ✅ Recurring bill tracking & reminders
- ✅ PDF reports & CSV export
- ✅ Subscription cost breakdown charts
- ✅ Unlimited transaction history
- ✅ Priority email support
- 🔮 Bank account sync (future)
- 🔮 Mobile apps (future)
- 🔮 Multi-budget support (future - belong to multiple budgets)

**Payment Processing:**
- Stripe integration
- Monthly and annual billing options
- 14-day free trial for premium features
- Cancel anytime

---

## Database Schema

### Users Table
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    is_premium BOOLEAN DEFAULT FALSE,
    premium_expires_at TIMESTAMP NULL
);
```

### Categories Table
```sql
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    color VARCHAR(7) NOT NULL, -- hex color
    icon VARCHAR(50) NOT NULL,
    is_system BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW()
);
```

### Transactions Table
```sql
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount INTEGER NOT NULL, -- stored in cents
    description TEXT,
    category_id UUID NOT NULL REFERENCES categories(id),
    date DATE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_transactions_user_date ON transactions(user_id, date DESC);
CREATE INDEX idx_transactions_category ON transactions(category_id);
```

### Budgets Table
```sql
CREATE TABLE budgets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories(id),
    amount INTEGER NOT NULL, -- stored in cents
    period VARCHAR(20) DEFAULT 'monthly',
    start_date DATE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, category_id, start_date)
);

CREATE INDEX idx_budgets_user ON budgets(user_id);
```

### Goals Table (Premium)
```sql
CREATE TABLE goals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    target_amount INTEGER NOT NULL,
    current_amount INTEGER DEFAULT 0,
    target_date DATE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_goals_user ON goals(user_id);
```

### Recurring Bills Table (Premium)
```sql
CREATE TABLE recurring_bills (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    amount INTEGER NOT NULL,
    category_id UUID NOT NULL REFERENCES categories(id),
    frequency VARCHAR(20) NOT NULL, -- weekly, biweekly, monthly, yearly
    next_due_date DATE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_recurring_bills_user_date ON recurring_bills(user_id, next_due_date);
```

---

## API Endpoints

### Authentication (Supabase handles most of this)
- `POST /api/auth/signup` - Register new user
- `POST /api/auth/login` - Login
- `POST /api/auth/logout` - Logout
- `POST /api/auth/reset-password` - Request password reset
- `GET /api/auth/me` - Get current user

### Transactions
- `GET /api/transactions` - List transactions (with filters, pagination)
- `POST /api/transactions` - Create transaction
- `GET /api/transactions/:id` - Get single transaction
- `PUT /api/transactions/:id` - Update transaction
- `DELETE /api/transactions/:id` - Delete transaction

### Categories
- `GET /api/categories` - List all categories (system + user's custom)
- `POST /api/categories` - Create custom category (premium)
- `PUT /api/categories/:id` - Update custom category (premium)
- `DELETE /api/categories/:id` - Delete custom category (premium)

### Budgets
- `GET /api/budgets` - List budgets for current month
- `GET /api/budgets?month=YYYY-MM` - List budgets for specific month
- `POST /api/budgets` - Create budget
- `PUT /api/budgets/:id` - Update budget
- `DELETE /api/budgets/:id` - Delete budget
- `GET /api/budgets/summary` - Get budget summary (totals, progress)

### Dashboard
- `GET /api/dashboard/summary` - Get dashboard summary data
- `GET /api/dashboard/recent-transactions` - Get recent transactions

### Goals (Premium)
- `GET /api/goals` - List all goals
- `POST /api/goals` - Create goal
- `GET /api/goals/:id` - Get single goal
- `PUT /api/goals/:id` - Update goal
- `DELETE /api/goals/:id` - Delete goal

### Recurring Bills (Premium)
- `GET /api/recurring-bills` - List all recurring bills
- `POST /api/recurring-bills` - Create recurring bill
- `PUT /api/recurring-bills/:id` - Update recurring bill
- `DELETE /api/recurring-bills/:id` - Delete recurring bill
- `POST /api/recurring-bills/:id/pay` - Mark bill as paid (create transaction)

### Reports (Premium)
- `GET /api/reports/spending-trends` - Get spending trends data
- `GET /api/reports/category-breakdown` - Get category breakdown
- `GET /api/reports/income-vs-expenses` - Get income vs expenses over time
- `POST /api/reports/export/csv` - Export transactions as CSV
- `POST /api/reports/export/pdf` - Generate PDF report

---

## Non-Functional Requirements

### Performance
- Page load time < 2 seconds
- API response time < 500ms (p95)
- Support 1000+ transactions per user without performance degradation

### Security
- HTTPS only (enforce)
- JWT authentication with secure token storage
- SQL injection prevention (use parameterized queries)
- XSS prevention (sanitize inputs)
- CSRF protection
- Rate limiting on API endpoints
- Secure password hashing (bcrypt, handled by Supabase)

### Accessibility
- WCAG 2.1 AA compliance
- Keyboard navigation support
- Screen reader friendly
- Proper ARIA labels
- Color contrast ratios met

### Browser Support
- Chrome (latest 2 versions)
- Firefox (latest 2 versions)
- Safari (latest 2 versions)
- Edge (latest 2 versions)
- Mobile browsers (iOS Safari, Chrome Mobile)

### Data Privacy
- GDPR compliance (for future EU expansion)
- User data export capability
- Account deletion (hard delete all user data)
- Privacy policy and terms of service

---

## Success Metrics

### MVP Success Criteria
- 100 beta users signed up
- 80% of users create at least one budget
- 70% of users add at least 10 transactions
- Average session duration > 5 minutes
- < 5% error rate on critical flows

### Premium Conversion Goals
- 5% free-to-premium conversion rate
- $10 average revenue per user (ARPU)
- < 5% monthly churn rate
- 90% payment success rate

---

## Development Phases & Timeline

### Phase 1: MVP (Weeks 1-8)

**Week 1-2: Foundation**
- Set up monorepo structure ✅
- Initialize frontend (React + Vite) ✅
- Initialize backend (Go + Chi) ✅
- Set up Supabase project
- Configure authentication
- Database schema & migrations
- Seed default categories

**Week 3-4: Core Features**
- User registration & login
- Transaction CRUD operations
- Category display
- Basic transaction list

**Week 5-6: Budgeting**
- Budget CRUD operations
- Budget vs. actual calculations
- Budget dashboard
- Progress indicators

**Week 7-8: Polish & Launch**
- Dashboard summary
- Transaction filtering
- UI polish & responsiveness
- Testing & bug fixes
- Deploy to production
- Beta user onboarding

### Phase 2: Premium Features (Weeks 9-16)

**Week 9-10: Goals**
- Goal CRUD operations
- Progress tracking
- Goal dashboard

**Week 11-12: Reports**
- Chart components (Recharts)
- Spending trends
- Category breakdown
- Export functionality

**Week 13-14: Recurring Bills**
- Bill CRUD operations
- Bill calendar
- Email reminder system
- Mark as paid functionality

**Week 15-16: Premium Launch**
- Stripe integration
- Subscription management
- Premium feature gating
- Marketing site updates

### Phase 3: Growth Features (Month 4+)

- Bank sync (Plaid integration)
- Mobile apps (React Native)
- Advanced features based on user feedback

---

## Open Questions & Decisions Needed

1. **Backend Router:** Chi or Gin?
   - Chi: More minimalist, standard library oriented
   - Gin: More features out of the box, faster performance
   - **Recommendation:** Chi (simpler, easier to maintain)

2. **Deployment Platform for Backend:**
   - Railway: Easy setup, good for PostgreSQL + Go
   - Fly.io: More control, global distribution
   - Render: Simple, good free tier
   - **Recommendation:** Railway (easiest for MVP)

3. **Email Service for Bill Reminders:**
   - SendGrid: Popular, good free tier
   - Resend: Modern, developer-friendly
   - Supabase Auth emails: Limited to auth flows
   - **Recommendation:** Resend (better DX, modern)

4. **Component Library:**
   - shadcn/ui: Copy/paste components, full control
   - Material UI: Comprehensive, heavier
   - Chakra UI: Good middle ground
   - **Recommendation:** shadcn/ui (best with Tailwind)

5. **Testing Strategy:**
   - Backend: Go testing package + testify
   - Frontend: Vitest + React Testing Library
   - E2E: Playwright or Cypress
   - **Decision needed:** How much test coverage for MVP?

---

## Risk & Mitigation

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Supabase outage | High | Implement retry logic, monitor status |
| Stripe payment failures | Medium | Clear error messages, support email |
| Data loss | High | Regular backups, point-in-time recovery |
| Slow database queries | Medium | Proper indexing, query optimization |
| Security breach | High | Regular security audits, penetration testing |
| Scope creep | Medium | Strict MVP feature set, defer non-critical features |
| Solo developer burnout | Medium | Set realistic timelines, take breaks, MVP first |

---

## Next Steps

1. **Immediate (This Week):**
   - [ ] Create Supabase project
   - [ ] Set up database schema & migrations
   - [ ] Implement authentication flow (frontend + backend)
   - [ ] Create first API endpoint (health check) ✅

2. **Next Week:**
   - [ ] Build transaction CRUD (backend)
   - [ ] Build transaction UI (frontend)
   - [ ] Set up React Query for API calls
   - [ ] Implement transaction list & filtering

3. **Ongoing:**
   - [ ] Set up CI/CD pipelines
   - [ ] Write tests for critical flows
   - [ ] Deploy to staging environment
   - [ ] User testing & feedback collection

---

## Appendix

### Useful Resources
- [Supabase Documentation](https://supabase.com/docs)
- [Go Best Practices](https://go.dev/doc/effective_go)
- [React Query Docs](https://tanstack.com/query/latest)
- [Tailwind CSS](https://tailwindcss.com/docs)
- [Stripe API](https://stripe.com/docs/api)

### Design Inspiration
- Mint (legacy)
- YNAB (You Need A Budget)
- Monarch Money
- Lunch Money
- Actual Budget

---

**Document Version:** 1.0
**Last Updated:** 2025-12-30
**Owner:** Solo Developer (with Claude Code assistance)
