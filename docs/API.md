# API Documentation

## Base URL
- Development: `http://localhost:8080/api`
- Production: `https://api.foldafinances.com/api` (TBD)

## Authentication

All endpoints except `/health` require authentication via JWT token in the Authorization header:

```
Authorization: Bearer <jwt_token>
```

Tokens are provided by Supabase Auth and validated by the backend.

---

## Endpoints

### Health Check

#### `GET /health`
Check API health status.

**Authentication:** Not required

**Response:**
```json
{
  "status": "healthy"
}
```

---

## Authentication Endpoints

Authentication is handled by Supabase Auth client-side. The backend validates JWT tokens.

#### `GET /api/auth/me`
Get current authenticated user information including spending period settings.

**Authentication:** Required

**Response:**
```json
{
  "data": {
    "id": "uuid",
    "email": "user@example.com",
    "spending_period": "biweekly",
    "period_start_date": "2025-01-01",
    "created_at": "2025-01-15T12:00:00Z",
    "updated_at": "2025-01-15T12:00:00Z"
  }
}
```

#### `PATCH /api/auth/me`
Update user spending period settings.

**Authentication:** Required

**Request Body:**
```json
{
  "spending_period": "biweekly",
  "period_start_date": "2025-01-01"
}
```

**Response:**
```json
{
  "data": {
    "id": "uuid",
    "email": "user@example.com",
    "spending_period": "biweekly",
    "period_start_date": "2025-01-01",
    "updated_at": "2025-01-15T13:00:00Z"
  },
  "message": "Settings updated successfully"
}
```

---

## "What Can I Spend?" Endpoints

### `GET /api/spending/available`
Get "What Can I Spend?" breakdown for the current period. **CORE FEATURE**

**Authentication:** Required

**Query Parameters:**
- None (uses current period based on user's settings)

**Response:**
```json
{
  "data": {
    "period": {
      "type": "biweekly",
      "start_date": "2025-01-15",
      "end_date": "2025-01-28",
      "days_remaining": 10
    },
    "summary": {
      "total_available": 43500,
      "total_budgeted": 100000,
      "total_spent": 56500
    },
    "categories": [
      {
        "category_id": "uuid",
        "category_name": "Dining & Restaurants",
        "category_icon": "utensils",
        "category_color": "#8B5CF6",
        "budgeted": 15000,
        "spent": 6300,
        "available": 8700,
        "percentage_used": 42,
        "status": "on_track"
      },
      {
        "category_id": "uuid",
        "category_name": "Groceries",
        "category_icon": "shopping-cart",
        "category_color": "#10B981",
        "budgeted": 30000,
        "spent": 18000,
        "available": 12000,
        "percentage_used": 60,
        "status": "on_track"
      }
    ]
  }
}
```

**Status Values:**
- `on_track` - Under 75% of budget used
- `warning` - 75-100% of budget used
- `over_budget` - Over 100% of budget used

---

## Expected Income Endpoints

### `GET /api/expected-income`
List all expected income sources.

**Authentication:** Required

**Response:**
```json
{
  "data": [
    {
      "id": "uuid",
      "user_id": "uuid",
      "name": "Paycheck",
      "amount": 200000,
      "frequency": "biweekly",
      "next_date": "2025-01-22",
      "is_active": true,
      "created_at": "2025-01-01T12:00:00Z",
      "updated_at": "2025-01-01T12:00:00Z"
    }
  ]
}
```

### `POST /api/expected-income`
Create a new expected income source.

**Authentication:** Required

**Request Body:**
```json
{
  "name": "Paycheck",
  "amount": 200000,
  "frequency": "biweekly",
  "next_date": "2025-01-22"
}
```

**Response:**
```json
{
  "data": {
    "id": "uuid",
    "user_id": "uuid",
    "name": "Paycheck",
    "amount": 200000,
    "frequency": "biweekly",
    "next_date": "2025-01-22",
    "is_active": true,
    "created_at": "2025-01-15T12:00:00Z",
    "updated_at": "2025-01-15T12:00:00Z"
  },
  "message": "Expected income created successfully"
}
```

### `PUT /api/expected-income/:id`
Update an expected income source.

**Authentication:** Required

**Request Body:**
```json
{
  "name": "Updated Paycheck",
  "amount": 210000,
  "frequency": "biweekly",
  "next_date": "2025-02-05"
}
```

**Response:**
```json
{
  "data": {
    "id": "uuid",
    "user_id": "uuid",
    "name": "Updated Paycheck",
    "amount": 210000,
    "frequency": "biweekly",
    "next_date": "2025-02-05",
    "is_active": true,
    "updated_at": "2025-01-15T13:00:00Z"
  },
  "message": "Expected income updated successfully"
}
```

### `DELETE /api/expected-income/:id`
Delete an expected income source.

**Authentication:** Required

**Response:**
```json
{
  "message": "Expected income deleted successfully"
}
```

### `POST /api/expected-income/:id/mark-received`
Mark expected income as received (creates transaction and updates next_date).

**Authentication:** Required

**Request Body:**
```json
{
  "received_date": "2025-01-22",
  "actual_amount": 200000
}
```

**Response:**
```json
{
  "data": {
    "expected_income": {
      "id": "uuid",
      "next_date": "2025-02-05"
    },
    "transaction": {
      "id": "uuid",
      "amount": 200000,
      "description": "Paycheck",
      "date": "2025-01-22"
    }
  },
  "message": "Income marked as received"
}
```

---

## Transaction Endpoints

### `GET /api/transactions`
List transactions with optional filtering and pagination.

**Authentication:** Required

**Query Parameters:**
- `page` (int, default: 1) - Page number
- `per_page` (int, default: 50, max: 100) - Items per page
- `start_date` (string, YYYY-MM-DD) - Filter by start date
- `end_date` (string, YYYY-MM-DD) - Filter by end date
- `category_id` (uuid) - Filter by category
- `search` (string) - Search in description

**Response:**
```json
{
  "data": [
    {
      "id": "uuid",
      "user_id": "uuid",
      "amount": -4599, // in cents, negative = expense
      "description": "Grocery shopping",
      "category_id": "uuid",
      "date": "2025-01-15",
      "created_at": "2025-01-15T12:00:00Z",
      "updated_at": "2025-01-15T12:00:00Z"
    }
  ],
  "page": 1,
  "per_page": 50,
  "total": 150,
  "total_pages": 3
}
```

### `POST /api/transactions`
Create a new transaction.

**Authentication:** Required

**Request Body:**
```json
{
  "amount": -4599,
  "description": "Grocery shopping",
  "category_id": "uuid",
  "date": "2025-01-15"
}
```

**Response:**
```json
{
  "data": {
    "id": "uuid",
    "user_id": "uuid",
    "amount": -4599,
    "description": "Grocery shopping",
    "category_id": "uuid",
    "date": "2025-01-15",
    "created_at": "2025-01-15T12:00:00Z",
    "updated_at": "2025-01-15T12:00:00Z"
  },
  "message": "Transaction created successfully"
}
```

### `GET /api/transactions/:id`
Get a single transaction by ID.

**Authentication:** Required

**Response:**
```json
{
  "data": {
    "id": "uuid",
    "user_id": "uuid",
    "amount": -4599,
    "description": "Grocery shopping",
    "category_id": "uuid",
    "date": "2025-01-15",
    "created_at": "2025-01-15T12:00:00Z",
    "updated_at": "2025-01-15T12:00:00Z"
  }
}
```

### `PUT /api/transactions/:id`
Update a transaction.

**Authentication:** Required

**Request Body:**
```json
{
  "amount": -5000,
  "description": "Updated description",
  "category_id": "uuid",
  "date": "2025-01-16"
}
```

**Response:**
```json
{
  "data": {
    "id": "uuid",
    "user_id": "uuid",
    "amount": -5000,
    "description": "Updated description",
    "category_id": "uuid",
    "date": "2025-01-16",
    "created_at": "2025-01-15T12:00:00Z",
    "updated_at": "2025-01-15T13:00:00Z"
  },
  "message": "Transaction updated successfully"
}
```

### `DELETE /api/transactions/:id`
Delete a transaction.

**Authentication:** Required

**Response:**
```json
{
  "message": "Transaction deleted successfully"
}
```

---

## Category Endpoints

### `GET /api/categories`
List all categories (system + user's custom).

**Authentication:** Required

**Response:**
```json
{
  "data": [
    {
      "id": "uuid",
      "user_id": null,
      "name": "Groceries",
      "color": "#10B981",
      "icon": "shopping-cart",
      "is_system": true,
      "created_at": "2025-01-01T00:00:00Z"
    },
    {
      "id": "uuid",
      "user_id": "uuid",
      "name": "Coffee",
      "color": "#8B4513",
      "icon": "coffee",
      "is_system": false,
      "created_at": "2025-01-15T12:00:00Z"
    }
  ]
}
```

### `POST /api/categories` (Premium)
Create a custom category.

**Authentication:** Required (Premium)

**Request Body:**
```json
{
  "name": "Coffee",
  "color": "#8B4513",
  "icon": "coffee"
}
```

**Response:**
```json
{
  "data": {
    "id": "uuid",
    "user_id": "uuid",
    "name": "Coffee",
    "color": "#8B4513",
    "icon": "coffee",
    "is_system": false,
    "created_at": "2025-01-15T12:00:00Z"
  },
  "message": "Category created successfully"
}
```

### `PUT /api/categories/:id` (Premium)
Update a custom category.

**Authentication:** Required (Premium)

**Request Body:**
```json
{
  "name": "Coffee & Tea",
  "color": "#654321",
  "icon": "coffee"
}
```

**Response:**
```json
{
  "data": {
    "id": "uuid",
    "user_id": "uuid",
    "name": "Coffee & Tea",
    "color": "#654321",
    "icon": "coffee",
    "is_system": false,
    "created_at": "2025-01-15T12:00:00Z"
  },
  "message": "Category updated successfully"
}
```

### `DELETE /api/categories/:id` (Premium)
Delete a custom category (only if no transactions use it).

**Authentication:** Required (Premium)

**Response:**
```json
{
  "message": "Category deleted successfully"
}
```

---

## Budget Endpoints

### `GET /api/budgets`
List budgets for the current or specified month.

**Authentication:** Required

**Query Parameters:**
- `month` (string, YYYY-MM, optional) - Specific month (defaults to current month)

**Response:**
```json
{
  "data": [
    {
      "id": "uuid",
      "user_id": "uuid",
      "category_id": "uuid",
      "amount": 50000,
      "period": "monthly",
      "start_date": "2025-01-01",
      "created_at": "2025-01-01T12:00:00Z",
      "updated_at": "2025-01-01T12:00:00Z"
    }
  ]
}
```

### `POST /api/budgets`
Create a budget.

**Authentication:** Required

**Request Body:**
```json
{
  "category_id": "uuid",
  "amount": 50000,
  "period": "monthly",
  "start_date": "2025-01-01"
}
```

**Response:**
```json
{
  "data": {
    "id": "uuid",
    "user_id": "uuid",
    "category_id": "uuid",
    "amount": 50000,
    "period": "monthly",
    "start_date": "2025-01-01",
    "created_at": "2025-01-01T12:00:00Z",
    "updated_at": "2025-01-01T12:00:00Z"
  },
  "message": "Budget created successfully"
}
```

### `PUT /api/budgets/:id`
Update a budget.

**Authentication:** Required

**Request Body:**
```json
{
  "amount": 60000,
  "period": "monthly"
}
```

**Response:**
```json
{
  "data": {
    "id": "uuid",
    "user_id": "uuid",
    "category_id": "uuid",
    "amount": 60000,
    "period": "monthly",
    "start_date": "2025-01-01",
    "created_at": "2025-01-01T12:00:00Z",
    "updated_at": "2025-01-15T12:00:00Z"
  },
  "message": "Budget updated successfully"
}
```

### `DELETE /api/budgets/:id`
Delete a budget.

**Authentication:** Required

**Response:**
```json
{
  "message": "Budget deleted successfully"
}
```

### `GET /api/budgets/summary`
Get budget summary with actual spending.

**Authentication:** Required

**Query Parameters:**
- `month` (string, YYYY-MM, optional) - Defaults to current month

**Response:**
```json
{
  "data": {
    "month": "2025-01",
    "total_budgeted": 200000,
    "total_spent": 150000,
    "categories": [
      {
        "category_id": "uuid",
        "category_name": "Groceries",
        "budgeted": 50000,
        "spent": 35000,
        "remaining": 15000,
        "percentage": 70
      }
    ],
    "on_track_count": 5,
    "over_budget_count": 1
  }
}
```

---

## Dashboard Endpoints

### `GET /api/dashboard/summary`
Get dashboard summary data.

**Authentication:** Required

**Query Parameters:**
- `month` (string, YYYY-MM, optional) - Defaults to current month

**Response:**
```json
{
  "data": {
    "month": "2025-01",
    "total_income": 500000,
    "total_expenses": 350000,
    "net": 150000,
    "budget_health": {
      "on_track": 8,
      "over_budget": 2,
      "no_budget": 3
    }
  }
}
```

### `GET /api/dashboard/recent-transactions`
Get recent transactions for dashboard.

**Authentication:** Required

**Query Parameters:**
- `limit` (int, default: 10, max: 20) - Number of transactions

**Response:**
```json
{
  "data": [
    {
      "id": "uuid",
      "amount": -4599,
      "description": "Grocery shopping",
      "category_id": "uuid",
      "date": "2025-01-15",
      "created_at": "2025-01-15T12:00:00Z"
    }
  ]
}
```

---

## Error Responses

All error responses follow this format:

```json
{
  "error": "ErrorType",
  "message": "Human-readable error message",
  "status": 400
}
```

### Common Error Codes:

- `400` - Bad Request (validation errors)
- `401` - Unauthorized (missing or invalid token)
- `403` - Forbidden (premium feature, insufficient permissions)
- `404` - Not Found
- `409` - Conflict (duplicate resource)
- `422` - Unprocessable Entity (business logic error)
- `500` - Internal Server Error

### Example Error Response:

```json
{
  "error": "ValidationError",
  "message": "Amount is required",
  "status": 400
}
```

---

## Rate Limiting

- **Free Tier:** 100 requests per minute
- **Premium Tier:** 500 requests per minute

Rate limit headers included in all responses:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1640000000
```

---

## Pagination

List endpoints support pagination with these query parameters:
- `page` (int, default: 1)
- `per_page` (int, default: 50, max: 100)

Paginated responses include:
```json
{
  "data": [...],
  "page": 1,
  "per_page": 50,
  "total": 150,
  "total_pages": 3
}
```

---

## Notes

- All amounts are in **cents** (integer) to avoid floating-point issues
- Negative amounts represent expenses, positive represent income
- All dates are in **YYYY-MM-DD** format
- All timestamps are in **ISO 8601** format with UTC timezone
- UUIDs are used for all IDs
