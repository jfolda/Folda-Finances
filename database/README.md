# Database Setup Guide

This directory contains the SQL script to initialize your Supabase database for Folda Finances.

## Quick Start

### Option 1: Run in Supabase SQL Editor (Recommended)

1. Go to your Supabase project dashboard
2. Navigate to **SQL Editor** in the left sidebar
3. Click **New Query**
4. Copy the entire contents of `init_supabase.sql`
5. Paste into the SQL editor
6. Click **Run** (or press Ctrl+Enter)

### Option 2: Run via psql CLI

```bash
# Connect to your Supabase database
psql "postgresql://postgres:[YOUR-PASSWORD]@db.[YOUR-PROJECT-REF].supabase.co:5432/postgres"

# Run the script
\i database/init_supabase.sql
```

## What the Script Does

### 1. Creates 13 Tables
- ✅ `budgets` - Shared budget entities
- ✅ `users` - User accounts with budget association
- ✅ `budget_invitations` - Budget sharing invitations
- ✅ `categories` - Expense/income categories
- ✅ `transactions` - Financial transactions
- ✅ `expected_income` - Expected income sources
- ✅ `category_budgets` - Monthly budget allocations
- ✅ `category_budget_splits` - Split budget allocations (premium)
- ✅ `goals` - Savings goals (premium)
- ✅ `recurring_bills` - Recurring bills (premium)
- ✅ `subscriptions` - Subscription tracking (premium)
- ✅ `detected_patterns` - Auto-detected recurring patterns
- ✅ `budget_transfers` - Budget reallocation history

### 2. Seeds Default Categories
- **16 Expense Categories:**
  - 🏠 Housing, ⚡ Utilities, 🛒 Groceries, 🍽️ Dining & Restaurants
  - 🚗 Transportation, 🏥 Healthcare, 🎬 Entertainment, 🛍️ Shopping
  - 💆 Personal Care, 📚 Education, 📱 Subscriptions, 🛡️ Insurance
  - 💰 Savings, 💳 Debt Payments, 🎁 Gifts & Donations, 📦 Miscellaneous

- **4 Income Categories:**
  - 💵 Salary, 💼 Freelance, 📈 Investments, 💸 Other Income

### 3. Enables Row Level Security (RLS)
All tables have RLS policies that ensure:
- Users can only see their own data
- Users can only see budgets they belong to
- System categories are visible to everyone
- Budget members can see shared budget data

### 4. Creates Indexes
Optimized indexes for common queries:
- User lookups by email
- Transaction filtering by date
- Budget queries by user
- Category filtering

## Verification

After running the script, verify the setup:

```sql
-- Check that all tables were created
SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'public'
ORDER BY table_name;

-- Count default categories
SELECT COUNT(*) FROM categories WHERE is_system = true;
-- Should return: 20

-- Check RLS is enabled
SELECT tablename, rowsecurity
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY tablename;
-- All should show 't' (true) for rowsecurity
```

## Clean Reinstall

If you need to drop all tables and start fresh, uncomment the DROP TABLE statements at the top of `init_supabase.sql` (lines 15-28).

**⚠️ WARNING:** This will delete ALL data! Only use in development.

## Troubleshooting

### Error: "relation already exists"
- Tables already exist. Either drop them first or use a fresh database.

### Error: "extension uuid-ossp does not exist"
- Run: `CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`
- This should be handled automatically by the script.

### Error: "foreign key constraint violation"
- The script creates tables in dependency order.
- If you see this error, ensure you're running the entire script, not parts of it.

### RLS Policies Not Working
- Ensure your backend is passing the correct JWT token
- Verify `auth.uid()` matches your Supabase user ID
- Check that RLS is enabled: `ALTER TABLE users ENABLE ROW LEVEL SECURITY;`

## Next Steps

After running this script:

1. **Configure Backend:**
   ```bash
   cd backend
   cp .env.example .env
   # Edit .env with your Supabase credentials
   ```

2. **Get Supabase JWT Secret:**
   - Go to Supabase Dashboard → Settings → API
   - Copy the "JWT Secret"
   - Add to `backend/.env` as `SUPABASE_JWT_SECRET`

3. **Start Backend:**
   ```bash
   go run cmd/api/main.go
   ```

4. **Start Frontend:**
   ```bash
   cd frontend
   npm run dev
   ```

## Database Connection Details

Find your connection details in Supabase:
- Dashboard → Settings → Database
- Connection String (Direct)
- Or use the Supabase client library

## Schema Documentation

For detailed schema documentation, see:
- `docs/DATABASE.md` - Complete schema reference
- `docs/REQUIREMENTS.md` - Feature requirements
- `docs/API.md` - API endpoint documentation

## Support

If you encounter issues:
1. Check the Supabase Dashboard logs
2. Verify your database connection string
3. Ensure you're using PostgreSQL 14+
4. Check that all environment variables are set correctly
