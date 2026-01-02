// API client for Folda Finances backend
import type {
  User,
  UpdateUserSettingsRequest,
  Account,
  CreateAccountRequest,
  UpdateAccountRequest,
  Transaction,
  CreateTransactionRequest,
  UpdateTransactionRequest,
  Category,
  CategoryBudget,
  CreateCategoryBudgetRequest,
  UpdateCategoryBudgetRequest,
  ExpectedIncome,
  CreateExpectedIncomeRequest,
  UpdateExpectedIncomeRequest,
  SpendingAvailableResponse,
  Budget,
  CreateBudgetRequest,
  BudgetMember,
  BudgetInvitation,
  CreateBudgetInvitationRequest,
  ApiResponse,
  ApiError,
  PaginatedResponse,
} from '../../../shared/types/api';

const API_BASE_URL = import.meta.env.VITE_API_URL || '/api';

class ApiClient {
  private baseUrl: string;
  private token: string | null = null;

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl;
  }

  setToken(token: string | null) {
    this.token = token;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    };

    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }

    const response = await fetch(`${this.baseUrl}${endpoint}`, {
      ...options,
      headers,
    });

    if (!response.ok) {
      const error: ApiError = await response.json();
      throw new Error(error.message || 'An error occurred');
    }

    return response.json();
  }

  // Auth endpoints
  async getCurrentUser(): Promise<ApiResponse<User>> {
    return this.request<ApiResponse<User>>('/auth/me');
  }

  async updateUserSettings(
    data: UpdateUserSettingsRequest
  ): Promise<ApiResponse<User>> {
    return this.request<ApiResponse<User>>('/auth/me', {
      method: 'PATCH',
      body: JSON.stringify(data),
    });
  }

  // "What Can I Spend?" endpoints
  async getSpendingAvailable(): Promise<ApiResponse<SpendingAvailableResponse>> {
    return this.request<ApiResponse<SpendingAvailableResponse>>(
      '/spending/available'
    );
  }

  // Transaction endpoints
  async getTransactions(params?: {
    category_id?: string;
    user_id?: string;
    start_date?: string;
    end_date?: string;
    page?: number;
    per_page?: number;
  }): Promise<ApiResponse<PaginatedResponse<Transaction>>> {
    const queryParams = new URLSearchParams();
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined) {
          queryParams.append(key, String(value));
        }
      });
    }
    const query = queryParams.toString();
    return this.request<ApiResponse<PaginatedResponse<Transaction>>>(
      `/transactions${query ? `?${query}` : ''}`
    );
  }

  async getTransaction(id: string): Promise<ApiResponse<Transaction>> {
    return this.request<ApiResponse<Transaction>>(`/transactions/${id}`);
  }

  async createTransaction(
    data: CreateTransactionRequest
  ): Promise<ApiResponse<Transaction>> {
    return this.request<ApiResponse<Transaction>>('/transactions', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async updateTransaction(
    id: string,
    data: UpdateTransactionRequest
  ): Promise<ApiResponse<Transaction>> {
    return this.request<ApiResponse<Transaction>>(`/transactions/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async deleteTransaction(id: string): Promise<ApiResponse<void>> {
    return this.request<ApiResponse<void>>(`/transactions/${id}`, {
      method: 'DELETE',
    });
  }

  // Account endpoints
  async getAccounts(): Promise<ApiResponse<Account[]>> {
    return this.request<ApiResponse<Account[]>>('/accounts');
  }

  async createAccount(
    data: CreateAccountRequest
  ): Promise<ApiResponse<Account>> {
    return this.request<ApiResponse<Account>>('/accounts', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async getAccount(id: string): Promise<ApiResponse<Account>> {
    return this.request<ApiResponse<Account>>(`/accounts/${id}`);
  }

  async updateAccount(
    id: string,
    data: UpdateAccountRequest
  ): Promise<ApiResponse<Account>> {
    return this.request<ApiResponse<Account>>(`/accounts/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async deleteAccount(id: string): Promise<ApiResponse<void>> {
    return this.request<ApiResponse<void>>(`/accounts/${id}`, {
      method: 'DELETE',
    });
  }

  // Category endpoints
  async getCategories(): Promise<ApiResponse<Category[]>> {
    return this.request<ApiResponse<Category[]>>('/categories');
  }

  // Budget endpoints
  async getBudgets(): Promise<ApiResponse<Budget[]>> {
    return this.request<ApiResponse<Budget[]>>('/budgets');
  }

  async createBudget(
    data: CreateBudgetRequest
  ): Promise<ApiResponse<Budget>> {
    return this.request<ApiResponse<Budget>>('/budgets', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async getBudgetMembers(budgetId: string): Promise<ApiResponse<BudgetMember[]>> {
    return this.request<ApiResponse<BudgetMember[]>>(
      `/budgets/${budgetId}/members`
    );
  }

  async inviteToBudget(
    budgetId: string,
    data: CreateBudgetInvitationRequest
  ): Promise<ApiResponse<BudgetInvitation>> {
    return this.request<ApiResponse<BudgetInvitation>>(
      `/budgets/${budgetId}/invite`,
      {
        method: 'POST',
        body: JSON.stringify(data),
      }
    );
  }

  async removeBudgetMember(
    budgetId: string,
    userId: string
  ): Promise<ApiResponse<void>> {
    return this.request<ApiResponse<void>>(
      `/budgets/${budgetId}/members/${userId}`,
      {
        method: 'DELETE',
      }
    );
  }

  // Budget invitation endpoints
  async getBudgetInvitations(): Promise<ApiResponse<BudgetInvitation[]>> {
    return this.request<ApiResponse<BudgetInvitation[]>>('/budget-invitations');
  }

  async acceptBudgetInvitation(
    token: string
  ): Promise<ApiResponse<void>> {
    return this.request<ApiResponse<void>>(
      `/budget-invitations/${token}/accept`,
      {
        method: 'POST',
      }
    );
  }

  async declineBudgetInvitation(
    token: string
  ): Promise<ApiResponse<void>> {
    return this.request<ApiResponse<void>>(
      `/budget-invitations/${token}/decline`,
      {
        method: 'POST',
      }
    );
  }

  // Category budget endpoints
  async getCategoryBudgets(): Promise<ApiResponse<CategoryBudget[]>> {
    return this.request<ApiResponse<CategoryBudget[]>>('/category-budgets');
  }

  async createCategoryBudget(
    data: CreateCategoryBudgetRequest
  ): Promise<ApiResponse<CategoryBudget>> {
    return this.request<ApiResponse<CategoryBudget>>('/category-budgets', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async updateCategoryBudget(
    id: string,
    data: UpdateCategoryBudgetRequest
  ): Promise<ApiResponse<CategoryBudget>> {
    return this.request<ApiResponse<CategoryBudget>>(`/category-budgets/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async deleteCategoryBudget(id: string): Promise<ApiResponse<void>> {
    return this.request<ApiResponse<void>>(`/category-budgets/${id}`, {
      method: 'DELETE',
    });
  }

  // Expected income endpoints
  async getExpectedIncome(): Promise<ApiResponse<ExpectedIncome[]>> {
    return this.request<ApiResponse<ExpectedIncome[]>>('/expected-income');
  }

  async createExpectedIncome(
    data: CreateExpectedIncomeRequest
  ): Promise<ApiResponse<ExpectedIncome>> {
    return this.request<ApiResponse<ExpectedIncome>>('/expected-income', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async updateExpectedIncome(
    id: string,
    data: UpdateExpectedIncomeRequest
  ): Promise<ApiResponse<ExpectedIncome>> {
    return this.request<ApiResponse<ExpectedIncome>>(`/expected-income/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async deleteExpectedIncome(id: string): Promise<ApiResponse<void>> {
    return this.request<ApiResponse<void>>(`/expected-income/${id}`, {
      method: 'DELETE',
    });
  }
}

export const apiClient = new ApiClient(API_BASE_URL);
