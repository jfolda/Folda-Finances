# Folda Finances Backend

Go REST API for Folda Finances.

## Setup

1. Copy `.env.example` to `.env` and fill in your values
2. Install dependencies: `go mod download`
3. Run the server: `go run cmd/api/main.go`

## Development

The API will be available at `http://localhost:8080`

## Project Structure

```
backend/
├── cmd/
│   └── api/           # Application entry point
├── internal/
│   ├── handlers/      # HTTP request handlers
│   ├── models/        # Data models
│   ├── database/      # Database connection and migrations
│   ├── middleware/    # HTTP middleware (auth, logging, etc.)
│   └── services/      # Business logic
└── go.mod
```

## API Endpoints

See [../docs/API.md](../docs/API.md) for API documentation.
