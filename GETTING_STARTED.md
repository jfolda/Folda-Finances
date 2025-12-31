# Getting Started with Folda Finances

Welcome! This guide will help you get the project set up and running locally.

## Prerequisites

Make sure you have the following installed:
- **Go 1.21+** - [Download](https://go.dev/dl/)
- **Node.js 18+** - [Download](https://nodejs.org/)
- **Git** - [Download](https://git-scm.com/)
- **Supabase Account** - [Sign up](https://supabase.com/)

## Project Setup

### 1. Clone & Install

```bash
# Navigate to your project directory
cd "Folda Finances"

# Install frontend dependencies
cd frontend
npm install
cd ..

# Backend dependencies are managed by Go modules
cd backend
go mod download
cd ..
```

### 2. Set Up Supabase

1. Go to [Supabase Dashboard](https://app.supabase.com/)
2. Create a new project
3. Note your project URL and anon key
4. Go to SQL Editor and run the migration scripts from `docs/DATABASE.md`

### 3. Configure Environment Variables

**Backend:**
```bash
cd backend
cp .env.example .env
```

Edit `backend/.env`:
```
PORT=8080
DATABASE_URL=postgresql://postgres:[YOUR-PASSWORD]@db.[YOUR-PROJECT-REF].supabase.co:5432/postgres
SUPABASE_URL=https://[YOUR-PROJECT-REF].supabase.co
SUPABASE_ANON_KEY=your-anon-key
SUPABASE_SERVICE_KEY=your-service-key
JWT_SECRET=your-jwt-secret
ENVIRONMENT=development
```

**Frontend:**
```bash
cd frontend
cp .env.example .env.local
```

Edit `frontend/.env.local`:
```
VITE_API_URL=http://localhost:8080
VITE_SUPABASE_URL=https://[YOUR-PROJECT-REF].supabase.co
VITE_SUPABASE_ANON_KEY=your-anon-key
```

### 4. Run the Application

**Terminal 1 - Backend:**
```bash
cd backend
go run cmd/api/main.go
```

Backend will be available at `http://localhost:8080`

**Terminal 2 - Frontend:**
```bash
cd frontend
npm run dev
```

Frontend will be available at `http://localhost:3000`

### 5. Verify Setup

1. Open `http://localhost:3000` in your browser
2. You should see the Folda Finances welcome page
3. Check backend health: `http://localhost:8080/health` should return `{"status":"healthy"}`

## Next Steps

### For You (Backend Developer):

1. **Set up database migrations:**
   - Install golang-migrate: `go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest`
   - Create migration files in `backend/migrations/`
   - Run migrations against Supabase

2. **Implement authentication middleware:**
   - JWT validation using Supabase
   - User context injection

3. **Build API endpoints:**
   - Start with `/api/transactions` endpoints
   - See `docs/API.md` for full API spec

4. **Set up GORM models:**
   - Define Go structs for database tables
   - See `docs/DATABASE.md` for schema

### For Claude (Frontend Developer):

1. **Set up Supabase Auth:**
   - Configure Supabase client
   - Create auth context/provider
   - Build login/signup forms

2. **Set up React Query:**
   - Configure QueryClient
   - Create API client functions
   - Build custom hooks for data fetching

3. **Build UI components:**
   - Set up shadcn/ui
   - Create base components (Button, Input, Card, etc.)
   - Build transaction list component

4. **Create pages:**
   - Dashboard
   - Transactions
   - Budgets
   - Settings

## Useful Commands

**Backend:**
```bash
# Run server
go run cmd/api/main.go

# Run tests
go test ./...

# Format code
go fmt ./...

# Run linter
golangci-lint run
```

**Frontend:**
```bash
# Dev server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview

# Lint
npm run lint
```

## Project Structure

```
folda-finances/
├── backend/              # Go REST API
│   ├── cmd/api/         # Application entry point
│   ├── internal/        # Internal packages
│   └── migrations/      # Database migrations (to create)
├── frontend/            # React + TypeScript
│   ├── src/
│   │   ├── components/  # Reusable UI components
│   │   ├── pages/       # Page components
│   │   ├── hooks/       # Custom React hooks
│   │   ├── lib/         # Utilities
│   │   └── api/         # API client
│   └── public/          # Static assets
├── shared/              # Shared types
│   └── types/          # TypeScript definitions
└── docs/               # Documentation
    ├── REQUIREMENTS.md # Full requirements doc
    ├── API.md         # API documentation
    └── DATABASE.md    # Database schema
```

## Documentation

- **[Requirements](docs/REQUIREMENTS.md)** - Complete feature requirements
- **[API Documentation](docs/API.md)** - REST API endpoints
- **[Database Schema](docs/DATABASE.md)** - Database design
- **[Frontend README](frontend/README.md)** - Frontend setup
- **[Backend README](backend/README.md)** - Backend setup

## Troubleshooting

**Port already in use:**
```bash
# Find process using port 8080 (backend)
lsof -i :8080
# or on Windows
netstat -ano | findstr :8080

# Kill the process or change PORT in .env
```

**Database connection issues:**
- Verify DATABASE_URL is correct
- Check Supabase project is running
- Ensure your IP is allowed in Supabase settings

**Frontend can't reach backend:**
- Verify backend is running on port 8080
- Check VITE_API_URL in frontend/.env.local
- Check browser console for CORS errors

## Need Help?

- Check the docs in the `/docs` folder
- Review the inline code comments
- Ask Claude for guidance on frontend tasks
- Refer to official docs:
  - [Go Documentation](https://go.dev/doc/)
  - [React Documentation](https://react.dev/)
  - [Supabase Documentation](https://supabase.com/docs)
  - [Vite Documentation](https://vitejs.dev/)

## What's Next?

See the **Next Steps** section in [docs/REQUIREMENTS.md](docs/REQUIREMENTS.md) for the detailed development roadmap.

**Week 1-2 Focus:**
- Database setup & migrations
- Authentication implementation
- Basic transaction CRUD (backend + frontend)
- Transaction list UI

Good luck! Let's build something great! 🚀
