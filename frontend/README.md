# Folda Finances - Frontend

React + TypeScript + Vite frontend for Folda Finances budgeting application.

## Features Implemented

### Core Features
- ✅ **"What Can I Spend?" Dashboard** - Real-time per-category spending view (main feature)
- ✅ **Authentication** - Login, Signup pages using Supabase Auth
- ✅ **Transaction Management** - Add, view, and filter transactions
- ⏳ **Budget Management** - Create and manage monthly budgets (coming next)
- ⏳ **Multi-User Budgets** - Invite users and manage budget members (coming next)
- ⏳ **Expected Income** - Track expected income sources (coming next)
- ⏳ **User Settings** - Configure view period and preferences (coming next)

## Tech Stack

- **React 18** - UI framework
- **TypeScript** - Type safety
- **Vite** - Build tool and dev server
- **TanStack Query** (React Query) - Data fetching and caching
- **React Router** - Client-side routing
- **Tailwind CSS** - Styling
- **Heroicons** - Icon library
- **Supabase** - Authentication and database

## Getting Started

### 1. Install Dependencies

```bash
cd frontend
npm install
```

### 2. Environment Setup

Create a `.env` file:

```bash
cp .env.example .env
```

Add your Supabase credentials to `.env`.

### 3. Run Development Server

```bash
npm run dev
```

Visit `http://localhost:3000`

## Scripts

- `npm run dev` - Start development server
- `npm run build` - Build for production
- `npm run preview` - Preview production build
- `npm run lint` - Run ESLint
