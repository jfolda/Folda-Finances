# Folda Finances

A personal finance SaaS application for budget tracking, savings goals, and financial insights.

## Project Structure

```
folda-finances/
├── backend/          # Go REST API
├── frontend/         # React + TypeScript web app
├── shared/           # Shared type definitions
└── docs/            # Documentation
```

## Tech Stack

### Frontend
- React 18+ with TypeScript
- Vite
- TanStack Query (React Query)
- Tailwind CSS + shadcn/ui
- Supabase Auth

### Backend
- Go 1.21+
- PostgreSQL (Supabase)
- Chi/Gin router
- GORM

### Infrastructure
- Frontend: Vercel
- Backend: Railway/Fly.io/Render
- Database & Auth: Supabase

## Getting Started

### Prerequisites
- Go 1.21+
- Node.js 18+
- PostgreSQL (or Supabase account)

### Backend Setup
```bash
cd backend
go mod download
go run cmd/api/main.go
```

### Frontend Setup
```bash
cd frontend
npm install
npm run dev
```

## Development

See [docs/REQUIREMENTS.md](docs/REQUIREMENTS.md) for detailed requirements and features.

## License

Proprietary
