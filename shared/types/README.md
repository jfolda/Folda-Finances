# Shared Types

This directory contains TypeScript type definitions that are shared between the frontend and backend.

## Usage

### Frontend
The frontend can directly import these types:
```typescript
import { Transaction, CreateTransactionRequest } from '@/types/api';
```

### Backend
For the backend, we can generate Go structs from these TypeScript types using tools like:
- `typescriptify` - Generate Go structs from TS interfaces
- `quicktype` - Convert JSON/TS to Go structs

Or maintain parallel type definitions manually.

## Keeping Types in Sync

1. **Source of truth**: TypeScript definitions in this directory
2. **Frontend**: Import directly
3. **Backend**: Generate or manually sync Go structs

This ensures the API contract is consistent across both applications.
