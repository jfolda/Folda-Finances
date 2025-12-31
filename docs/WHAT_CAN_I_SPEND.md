# "What Can I Spend?" Feature Specification

## Overview

**"What Can I Spend?"** is the **core differentiator** of Folda Finances. This feature answers the most important question users have when making spending decisions: *"Can I afford this right now?"*

Instead of passively tracking past spending, it actively helps users make better financial decisions in real-time.

---

## User Problem

Traditional budgeting apps show you:
- How much you've spent (past)
- How much you budgeted (static)
- Generic budget progress bars

**But they don't answer:**
- "I'm at a restaurant—can I afford to eat here?"
- "How much can I spend on groceries this week?"
- "Do I have room in my budget for this purchase?"

---

## Our Solution

A **real-time, per-category spending calculator** that shows users exactly how much money they have available to spend in each budget category for the current period.

### Key Features

1. **User-configured spending period** (weekly, biweekly, or monthly)
2. **Per-category available amounts** ("You have $87 left for dining")
3. **Real-time updates** (updates immediately when transactions are added)
4. **Visual clarity** (color-coded status, large numbers, progress bars)
5. **Mobile-first design** (optimized for quick glances while shopping)

---

## How It Works

### Setup Phase

1. **User chooses spending period:**
   - Weekly (every 7 days)
   - Biweekly (every 14 days)
   - Monthly (calendar month)

2. **User sets period start date:**
   - Example: "My paycheck arrives every other Friday starting Jan 5, 2025"
   - App calculates period boundaries automatically

3. **User adds expected income:**
   - Name: "Paycheck"
   - Amount: $2,000
   - Frequency: Biweekly
   - Next date: Jan 19, 2025

4. **User creates budgets per category:**
   - Groceries: $300/period
   - Dining: $150/period
   - Transportation: $100/period
   - etc.

### Daily Usage

**Scenario: User is deciding whether to go to lunch**

1. User opens app → lands on "What Can I Spend?" screen
2. Sees at a glance:
   ```
   Dining & Restaurants
   Available: $87
   (Budget: $150, Spent: $63)
   ```
3. Decides: "Yes, I can afford a $20 lunch"
4. After lunch, adds transaction for $20
5. "Available" immediately updates to $67

**The app makes the decision easy and stress-free.**

---

## Calculation Logic

### Current Period Calculation

```typescript
function getCurrentPeriod(user: User, today: Date): Period {
  const { spending_period, period_start_date } = user;

  if (spending_period === 'weekly') {
    // Find most recent occurrence of start day-of-week
    const startDay = period_start_date.getDay();
    const currentPeriodStart = getMostRecentDayOfWeek(today, startDay);
    const currentPeriodEnd = addDays(currentPeriodStart, 7);
    return { start: currentPeriodStart, end: currentPeriodEnd };
  }

  if (spending_period === 'biweekly') {
    // Find most recent biweekly boundary from anchor date
    const daysSinceAnchor = daysBetween(period_start_date, today);
    const periodNumber = Math.floor(daysSinceAnchor / 14);
    const currentPeriodStart = addDays(period_start_date, periodNumber * 14);
    const currentPeriodEnd = addDays(currentPeriodStart, 14);
    return { start: currentPeriodStart, end: currentPeriodEnd };
  }

  if (spending_period === 'monthly') {
    // Current calendar month
    const currentPeriodStart = startOfMonth(today);
    const currentPeriodEnd = endOfMonth(today);
    return { start: currentPeriodStart, end: currentPeriodEnd };
  }
}
```

### Per-Category Available Calculation

```typescript
function calculateAvailable(
  category: Category,
  budget: Budget | undefined,
  transactions: Transaction[],
  period: Period
): number {
  // If no budget set for this category, return 0
  if (!budget) return 0;

  // Sum all transactions in this category during current period
  const spent = transactions
    .filter(t =>
      t.category_id === category.id &&
      t.date >= period.start &&
      t.date <= period.end
    )
    .reduce((sum, t) => sum + Math.abs(t.amount), 0);

  // Available = Budget - Spent
  const available = budget.amount - spent;

  return available;
}
```

### Status Determination

```typescript
function getStatus(budgeted: number, spent: number): Status {
  const percentageUsed = (spent / budgeted) * 100;

  if (percentageUsed > 100) return 'over_budget';
  if (percentageUsed >= 75) return 'warning';
  return 'on_track';
}
```

---

## UI/UX Design Principles

### 1. **Make it the default view**
   - App opens directly to "What Can I Spend?"
   - No navigation required
   - Instant access to the answer

### 2. **Big, clear numbers**
   - Available amount is the **largest element** on screen
   - Easy to read at a glance
   - No cognitive load

### 3. **Color-coded status**
   - 🟢 Green: Plenty of budget left (<75% used)
   - 🟡 Yellow: Running low (75-100% used)
   - 🔴 Red: Over budget (>100% used)

### 4. **Minimal friction**
   - No scrolling to see top categories
   - Pull-to-refresh to recalculate
   - One tap to add transaction
   - Works offline (caches last calculation)

### 5. **Contextual information**
   - Show period dates ("Jan 15 - Jan 28")
   - Days remaining in period
   - Visual timeline of period progress

---

## Example UI Mockup (Text)

```
┌─────────────────────────────────────┐
│  What Can I Spend?                  │
│                                     │
│  Current Period                     │
│  Jan 15 - Jan 28 (10 days left)    │
│  ▓▓▓▓▓▓▓░░░░░░                     │
│                                     │
├─────────────────────────────────────┤
│                                     │
│  🍽️  Dining & Restaurants           │
│                                     │
│       $87                           │
│       available                     │
│                                     │
│  ▓▓▓▓▓▓░░░░ 42% used               │
│  Budget: $150 | Spent: $63         │
│                                     │
├─────────────────────────────────────┤
│                                     │
│  🛒  Groceries                       │
│                                     │
│       $180                          │
│       available                     │
│                                     │
│  ▓▓▓▓▓▓▓▓░░ 60% used               │
│  Budget: $300 | Spent: $120        │
│                                     │
├─────────────────────────────────────┤
│                                     │
│  🚗  Transportation                  │
│                                     │
│       $100                          │
│       available                     │
│                                     │
│  ░░░░░░░░░░ 0% used                │
│  Budget: $100 | Spent: $0          │
│                                     │
└─────────────────────────────────────┘
        ┌─────────────┐
        │ + Add Expense│
        └─────────────┘
```

---

## Backend Implementation

### Database Tables

1. **users** - Store period settings
   - `spending_period` (weekly/biweekly/monthly)
   - `period_start_date`
   - `period_anchor_day`

2. **expected_income** - Track expected income
   - `name`, `amount`, `frequency`, `next_date`

3. **budgets** - Budget per category
   - `category_id`, `amount` (per user's period)

4. **transactions** - Actual spending
   - `amount`, `category_id`, `date`

### API Endpoint

**`GET /api/spending/available`**

Returns:
- Current period info
- Per-category breakdown
- Summary totals

**Backend Logic:**
1. Get user's period settings
2. Calculate current period boundaries
3. Fetch all budgets for user
4. Fetch all transactions in current period
5. Group transactions by category
6. Calculate available for each category
7. Determine status (on_track/warning/over_budget)
8. Return formatted response

**Performance Considerations:**
- Cache period boundaries for current user
- Index on `(user_id, date)` for transactions
- Aggregate queries should be <100ms
- Consider materialized view for frequent access

---

## Frontend Implementation

### Component Structure

```
WhatCanISpendPage/
  ├── PeriodHeader
  │   ├── Period dates display
  │   └── Days remaining badge
  ├── SummaryCard (optional)
  │   └── Total available across all categories
  ├── CategorySpendingList
  │   └── CategorySpendingCard (per category)
  │       ├── Category icon + name
  │       ├── Available amount (large)
  │       ├── Progress bar
  │       └── Budget/Spent breakdown
  └── QuickAddButton (FAB)
```

### State Management

```typescript
// React Query hook
function useSpendingAvailable() {
  return useQuery({
    queryKey: ['spending', 'available'],
    queryFn: () => api.getSpendingAvailable(),
    staleTime: 30000, // 30 seconds
    refetchOnWindowFocus: true,
  });
}

// Auto-refresh on transaction add
const queryClient = useQueryClient();
queryClient.invalidateQueries(['spending', 'available']);
```

---

## User Onboarding Flow

### First-Time Setup

1. **Welcome screen**
   - "Welcome to Folda Finances!"
   - "Let's set up your spending period"

2. **Period selection**
   - "How often do you get paid?"
   - [ ] Weekly
   - [ ] Biweekly (most common)
   - [ ] Monthly

3. **Start date**
   - "When is your next payday?"
   - [Date picker]

4. **Expected income**
   - "How much do you expect to earn per period?"
   - [$___]

5. **Budget setup**
   - "Let's create your first budgets"
   - Show top 3-5 categories
   - Suggest amounts based on income (e.g., 30% groceries, 15% dining, etc.)

6. **Done!**
   - "You're all set!"
   - Show "What Can I Spend?" screen

**Goal:** Get user to value in <2 minutes

---

## Future Enhancements

### Phase 2: Premium Features

1. **Savings Goals Integration**
   - Subtract goal contributions from available spending
   - "After saving $250 for emergency fund, you have $X to spend"

2. **Bill Integration**
   - Subtract upcoming bills from available
   - "Rent is due in 5 days ($800), so you have $X for everything else"

3. **Smart Recommendations**
   - "You're spending faster than usual this period—slow down!"
   - "You have $50 left for 7 days = $7/day average"

4. **Multi-Period View**
   - Toggle between current period, next period, month view
   - "If you save $X this period, you'll have $Y next period"

### Phase 3: Advanced Features

5. **Predictive Spending**
   - AI learns your spending patterns
   - "Based on your history, you'll likely spend $X more on groceries"

6. **Shared Budgets**
   - Family/household budgets
   - "You have $X left, and your partner has $Y left = $Z total"

7. **Voice Interface**
   - "Hey Siri, how much can I spend on dining?"
   - "You have $87 available for dining this week"

---

## Success Metrics

### MVP Success Criteria

1. **Engagement**
   - 70%+ of users check "What Can I Spend?" at least 3x per week
   - Average session duration on this screen: 30-60 seconds (quick glance)

2. **Value Perception**
   - 80%+ of surveyed users say this feature helps them make better decisions
   - NPS score 40+

3. **Behavior Change**
   - Users stay under budget 60%+ of the time (vs. 30% industry average)
   - Reduced "surprise" overspending

### Long-Term Goals

- **Claim to fame:** Users say "Folda Finances shows me what I can spend"
- **Word of mouth:** "You should try Folda—it tells you how much you can spend"
- **Market position:** #1 app for "real-time spending guidance"

---

## Competitive Analysis

### What Competitors Do

| App | Budget Tracking | Real-Time Guidance |
|-----|-----------------|-------------------|
| Mint | ✅ Yes | ❌ No (only shows spent vs budget) |
| YNAB | ✅ Yes | 🟡 Partial (category balances, but complex) |
| Monarch | ✅ Yes | ❌ No (analytics-focused) |
| Simplifi | ✅ Yes | 🟡 Partial ("spending plan" but not per-category) |
| **Folda** | ✅ Yes | ✅ **YES - Core feature** |

### Our Advantage

- **Simpler** than YNAB (no envelope complexity)
- **More actionable** than Mint (not just tracking)
- **More real-time** than any competitor
- **Built for mobile-first** decision-making

---

## Marketing Angle

### Tagline Options

- "Know what you can spend, before you spend it"
- "Stop guessing. Start spending smart."
- "The budgeting app that answers: Can I afford this?"
- "Budget in real-time, not in hindsight"

### Key Messages

1. **No more surprises** - Always know where you stand
2. **Decision confidence** - Make purchases stress-free
3. **Simple setup** - Set your period, budget, and go
4. **Always accurate** - Updates in real-time with every transaction

### Target Users

- **Primary:** People who struggle to stay on budget
- **Secondary:** People who want spending confidence
- **Pain point:** "I never know if I can afford something until it's too late"

---

## Implementation Priority

### Week 1-2: Foundation
- [ ] Database schema (users table with period settings)
- [ ] User settings API (update period)
- [ ] Period calculation logic (backend utility)

### Week 3-4: Core Feature
- [ ] Expected income CRUD (backend + frontend)
- [ ] Budget CRUD with period awareness
- [ ] `GET /api/spending/available` endpoint
- [ ] Frontend "What Can I Spend?" page

### Week 5-6: Polish
- [ ] Onboarding flow
- [ ] Pull-to-refresh
- [ ] Offline caching
- [ ] Color coding and visual polish

### Week 7-8: Testing & Launch
- [ ] User testing
- [ ] Performance optimization
- [ ] Mobile responsiveness
- [ ] Beta launch

---

**Document Version:** 1.0
**Last Updated:** 2025-12-30
**Status:** Ready for Implementation

This feature is the heart of Folda Finances. Let's build it right. 🚀
