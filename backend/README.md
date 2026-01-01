# Folda Finances - Backend API

Go backend API for Folda Finances budgeting application.

## Features Implemented

### ✅ Core API Endpoints
- **Authentication** - JWT middleware for Supabase tokens
- **User Management** - Get/update user settings
- **"What Can I Spend?"** - Real-time spending calculations (CORE FEATURE)
- **Transactions** - Full CRUD for financial transactions
- **Categories** - List system and custom categories
- **Category Budgets** - Manage monthly budget allocations
- **Expected Income** - Track expected income sources

### Tech Stack
- **Go 1.21+** - Backend language
- **Chi Router** - HTTP routing
- **GORM** - ORM for PostgreSQL
- **PostgreSQL** - Database
- **Supabase Auth** - JWT authentication

## Project Structure

```
backend/
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/
│   ├── database/
│   │   └── database.go          # DB connection & migrations
│   ├── handlers/
│   │   ├── user.go              # User endpoints
│   │   ├── category.go          # Category endpoints
│   │   ├── transaction.go       # Transaction endpoints
│   │   ├── spending.go          # "What Can I Spend?" (CORE)
│   │   ├── budget.go            # Budget endpoints
│   │   ├── income.go            # Expected income endpoints
│   │   └── helpers.go           # Utility functions
│   ├── middleware/
│   │   └── auth.go              # JWT authentication
│   └── models/
│       └── models.go            # Database models
├── go.mod
├── go.sum
└── .env.example
```

## Getting Started

### 1. Prerequisites
- Go 1.21 or higher
- PostgreSQL 14 or higher
- Supabase project (for authentication)

### 2. Install Dependencies

```bash
cd backend
go mod download
```

### 3. Database Setup

Create a PostgreSQL database:

```sql
CREATE DATABASE folda_finances;
```

### 4. Environment Configuration

Copy `.env.example` to `.env`:

```bash
cp .env.example .env
```

Edit `.env` with your configuration:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=folda_finances
DB_SSL_MODE=disable

SUPABASE_JWT_SECRET=your_supabase_jwt_secret
PORT=8080
```

### 5. Run the Server

```bash
go run cmd/api/main.go
```

The API will be available at `http://localhost:8080`

### 6. Database Migrations

Migrations run automatically on startup. The server will:
- Create all required tables
- Seed default categories (20 system categories)

## API Endpoints

### Health Check
```
GET /health
```

### Authentication
```
GET    /api/auth/me        # Get current user
PATCH  /api/auth/me        # Update user settings
```

### "What Can I Spend?" (CORE FEATURE)
```
GET    /api/spending/available   # Get real-time spending data
```

### Categories
```
GET    /api/categories     # List all categories
```

### Transactions
```
GET    /api/transactions           # List transactions
POST   /api/transactions           # Create transaction
GET    /api/transactions/:id       # Get transaction
PUT    /api/transactions/:id       # Update transaction
DELETE /api/transactions/:id       # Delete transaction
```

### Category Budgets
```
GET    /api/category-budgets       # List budgets
POST   /api/category-budgets       # Create budget
PUT    /api/category-budgets/:id   # Update budget
DELETE /api/category-budgets/:id   # Delete budget
```

### Expected Income
```
GET    /api/expected-income        # List expected income
POST   /api/expected-income        # Create expected income
PUT    /api/expected-income/:id    # Update expected income
DELETE /api/expected-income/:id    # Delete expected income
```

## Authentication

All API endpoints (except `/health`) require JWT authentication.

Include the Supabase JWT token in the `Authorization` header:

```
Authorization: Bearer <jwt_token>
```

## Development

### Run with Auto-Reload

Install `air` for hot reloading:

```bash
go install github.com/cosmtrek/air@latest
air
```

### Run Tests

```bash
go test ./...
```

## Deployment

### Build for Production

```bash
go build -o folda-api cmd/api/main.go
```

### Run Production Binary

```bash
./folda-api
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | Database user | `postgres` |
| `DB_PASSWORD` | Database password | - |
| `DB_NAME` | Database name | `folda_finances` |
| `DB_SSL_MODE` | SSL mode | `disable` |
| `SUPABASE_JWT_SECRET` | Supabase JWT secret | - |
| `PORT` | Server port | `8080` |

## Default Categories

The system seeds 20 default categories on startup:

**Expenses:**
- Housing, Utilities, Groceries, Dining & Restaurants
- Transportation, Healthcare, Entertainment, Shopping
- Personal Care, Education, Subscriptions, Insurance
- Savings, Debt Payments, Gifts & Donations, Miscellaneous

**Income:**
- Salary, Freelance, Investments, Other Income

## Scripts

```bash
# Run server
go run cmd/api/main.go

# Build
go build -o folda-api cmd/api/main.go

# Format code
go fmt ./...

# Vet code
go vet ./...

# Install dependencies
go mod download
```

## Next Steps

- [ ] Add unit tests
- [ ] Add integration tests
- [ ] Implement multi-user budget endpoints
- [ ] Implement subscription detection
- [ ] Add logging middleware
- [ ] Add rate limiting
- [ ] Deploy to production

## Notes

- All monetary amounts are stored as **integers in cents** to avoid floating-point issues
- Budgets are stored as **monthly amounts** and pro-rated for display based on user's view period
- JWT tokens are validated using Supabase's JWT secret
- CORS is configured for `localhost:3000` and `localhost:5173`
