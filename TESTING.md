# Testing Guide

This document describes how to run tests for the Folda Finances application.

## Quick Start

### Run All Tests (Pre-Deployment)

**Windows:**
```bash
test.bat
```

**Linux/Mac:**
```bash
./test.sh
```

This will run all backend and frontend tests and provide a summary.

## Backend Tests (Go)

### Prerequisites
- Go 1.21 or later
- SQLite (for in-memory test database)

### Running Tests

```bash
cd backend

# Run all tests
make test

# Run with verbose output
make test-verbose

# Run with coverage report
make test-coverage
```

### Test Coverage

The coverage report will be generated at `backend/coverage.html`. Open this file in a browser to see detailed coverage information.

### What's Being Tested

- **Budget Handlers** (`budget_test.go`)
  - Creating category budgets
  - Updating category budgets
  - Budget splitting functionality
  - Get budget members
  - Get/update category budget splits
  - Validation (e.g., users must be in same budget)

## Frontend Tests (TypeScript/Vitest)

### Prerequisites
- Node.js 18 or later
- npm

### Running Tests

```bash
cd frontend

# Install dependencies (first time only)
npm install

# Run all tests
npm test

# Run tests with UI
npm run test:ui

# Run with coverage
npm run test:coverage
```

### What's Being Tested

- **API Client** (`api.test.ts`)
  - Budget member fetching
  - Category budget split operations
  - Budget creation with different allocation types
  - Token management
  - Error handling

### Test Coverage

Coverage reports are generated in `frontend/coverage/` directory.

## Writing New Tests

### Backend (Go)

Tests should be placed in `*_test.go` files next to the code they're testing.

```go
func TestMyFeature(t *testing.T) {
    db := setupTestDB(t)
    handler := NewMyHandler(db)

    // Create test data
    user, budget := createTestUser(t, db, "test@example.com")

    // Make request
    req := httptest.NewRequest("GET", "/my-endpoint", nil)
    req = req.WithContext(setUserIDContext(req, user.ID))
    w := httptest.NewRecorder()

    handler.MyEndpoint(w, req)

    // Assert
    if w.Code != http.StatusOK {
        t.Errorf("Expected 200, got %d", w.Code)
    }
}
```

### Frontend (TypeScript)

Tests should be placed in `*.test.ts` or `*.test.tsx` files.

```typescript
import { describe, it, expect, vi } from 'vitest';

describe('MyFeature', () => {
  it('should do something', async () => {
    // Arrange
    const mockData = { id: '1', name: 'Test' };
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ data: mockData }),
    });

    // Act
    const result = await apiClient.getSomething();

    // Assert
    expect(result.data).toEqual(mockData);
  });
});
```

## Continuous Integration

These tests should be run automatically before:
1. Merging pull requests
2. Deploying to staging
3. Deploying to production

### GitHub Actions (Example)

```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run tests
        run: ./test.sh
```

## Test Data

- Backend tests use an in-memory SQLite database
- Each test gets a fresh database
- Frontend tests mock all API calls
- No real data or external services are used

## Troubleshooting

### Backend Tests Fail with CGO Error (Windows)

**Error**: `Binary was compiled with 'CGO_ENABLED=0', go-sqlite3 requires cgo to work`

**Solution**: Backend tests use SQLite for in-memory testing, which requires CGO (C compiler integration). On Windows, you need:

1. **Option 1: Install GCC** (Recommended for local development)
   - Install TDM-GCC from https://jmeubank.github.io/tdm-gcc/ or MinGW-w64
   - Add to PATH and restart your terminal
   - Set `CGO_ENABLED=1` before running tests:
     ```bash
     set CGO_ENABLED=1
     go test ./...
     ```

2. **Option 2: Skip Backend Tests** (If you don't have a C compiler)
   - Run only frontend tests: `cd frontend && npm test`
   - Backend tests will run automatically in CI/CD on Linux where CGO is available

3. **Option 3: Use WSL or Linux**
   - Backend tests work out of the box on Linux/Mac

### Backend Tests Fail (General)

```bash
cd backend
go mod tidy  # Ensure dependencies are up to date
CGO_ENABLED=1 go test ./... -v  # Run with verbose output
```

### Frontend Tests Fail

```bash
cd frontend
rm -rf node_modules package-lock.json
npm install  # Reinstall dependencies
npm test  # Run again
```

### Coverage Not Generating

Make sure you have the coverage tools installed:

```bash
# Backend
go install golang.org/x/tools/cmd/cover@latest

# Frontend
cd frontend
npm install --save-dev @vitest/coverage-v8
```

## Best Practices

1. **Run tests before committing**
   ```bash
   ./test.bat  # or ./test.sh
   ```

2. **Write tests for new features**
   - Add tests when implementing new endpoints
   - Test both success and error cases

3. **Keep tests fast**
   - Use in-memory databases
   - Mock external services
   - Avoid unnecessary delays

4. **Test realistic scenarios**
   - Test with multiple users
   - Test edge cases (empty lists, invalid IDs, etc.)
   - Test validation errors

5. **Maintain test coverage**
   - Aim for >80% coverage on critical paths
   - Business logic should be 100% covered
