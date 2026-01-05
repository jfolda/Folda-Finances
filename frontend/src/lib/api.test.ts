import { describe, it, expect, beforeEach, vi } from 'vitest';
import { apiClient } from './api';

// Mock fetch globally
global.fetch = vi.fn();

describe('ApiClient - Budget Splitting', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('getCurrentBudgetMembers', () => {
    it('should fetch budget members successfully', async () => {
      const mockMembers = [
        {
          id: 'user-1',
          email: 'user1@example.com',
          name: 'User 1',
          budget_id: 'budget-1',
        },
        {
          id: 'user-2',
          email: 'user2@example.com',
          name: 'User 2',
          budget_id: 'budget-1',
        },
      ];

      (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ data: mockMembers }),
      });

      const result = await apiClient.getCurrentBudgetMembers();

      expect(result.data).toEqual(mockMembers);
      expect(global.fetch).toHaveBeenCalledWith(
        '/api/budget/members',
        expect.objectContaining({
          headers: expect.objectContaining({
            'Content-Type': 'application/json',
          }),
        })
      );
    });
  });

  describe('getCategoryBudgetSplits', () => {
    it('should fetch splits for a category budget', async () => {
      const budgetId = 'budget-123';
      const mockSplits = [
        {
          id: 'split-1',
          category_budget_id: budgetId,
          user_id: 'user-1',
          allocation_amount: 60000,
          allocation_percentage: null,
        },
        {
          id: 'split-2',
          category_budget_id: budgetId,
          user_id: 'user-2',
          allocation_amount: 40000,
          allocation_percentage: null,
        },
      ];

      (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ data: mockSplits }),
      });

      const result = await apiClient.getCategoryBudgetSplits(budgetId);

      expect(result.data).toEqual(mockSplits);
      expect(global.fetch).toHaveBeenCalledWith(
        `/api/category-budgets/${budgetId}/splits`,
        expect.objectContaining({
          headers: expect.objectContaining({
            'Content-Type': 'application/json',
          }),
        })
      );
    });
  });

  describe('updateCategoryBudgetSplits', () => {
    it('should update splits with fixed amounts', async () => {
      const budgetId = 'budget-123';
      const splits = [
        {
          user_id: 'user-1',
          allocation_amount: 60000,
        },
        {
          user_id: 'user-2',
          allocation_amount: 40000,
        },
      ];

      const mockResponse = [
        {
          id: 'split-1',
          category_budget_id: budgetId,
          ...splits[0],
        },
        {
          id: 'split-2',
          category_budget_id: budgetId,
          ...splits[1],
        },
      ];

      (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ data: mockResponse }),
      });

      const result = await apiClient.updateCategoryBudgetSplits(budgetId, splits);

      expect(result.data).toEqual(mockResponse);
      expect(global.fetch).toHaveBeenCalledWith(
        `/api/category-budgets/${budgetId}/splits`,
        expect.objectContaining({
          method: 'PUT',
          body: JSON.stringify({ splits }),
        })
      );
    });

    it('should update splits with percentages', async () => {
      const budgetId = 'budget-123';
      const splits = [
        {
          user_id: 'user-1',
          allocation_percentage: 60.0,
        },
        {
          user_id: 'user-2',
          allocation_percentage: 40.0,
        },
      ];

      (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ data: [] }),
      });

      await apiClient.updateCategoryBudgetSplits(budgetId, splits);

      expect(global.fetch).toHaveBeenCalledWith(
        `/api/category-budgets/${budgetId}/splits`,
        expect.objectContaining({
          method: 'PUT',
          body: JSON.stringify({ splits }),
        })
      );
    });
  });

  describe('createCategoryBudget', () => {
    it('should create a pooled budget', async () => {
      const budgetData = {
        category_id: 'cat-1',
        amount: 100000,
        allocation_type: 'pooled' as const,
      };

      const mockResponse = {
        id: 'budget-1',
        budget_id: 'budget-123',
        ...budgetData,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      };

      (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ data: mockResponse }),
      });

      const result = await apiClient.createCategoryBudget(budgetData);

      expect(result.data).toEqual(mockResponse);
      expect(result.data.allocation_type).toBe('pooled');
    });

    it('should create a split budget', async () => {
      const budgetData = {
        category_id: 'cat-1',
        amount: 100000,
        allocation_type: 'split' as const,
      };

      const mockResponse = {
        id: 'budget-1',
        budget_id: 'budget-123',
        ...budgetData,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      };

      (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ data: mockResponse }),
      });

      const result = await apiClient.createCategoryBudget(budgetData);

      expect(result.data.allocation_type).toBe('split');
    });
  });

  describe('Error Handling', () => {
    it('should throw error when splits request fails', async () => {
      (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
        ok: false,
        status: 400,
        json: async () => ({ error: 'all users must belong to the same budget' }),
      });

      await expect(
        apiClient.updateCategoryBudgetSplits('budget-123', [
          { user_id: 'user-1', allocation_amount: 50000 },
        ])
      ).rejects.toThrow();
    });

    it('should throw error when getting members fails', async () => {
      (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
        ok: false,
        status: 401,
        json: async () => ({ error: 'unauthorized' }),
      });

      await expect(apiClient.getCurrentBudgetMembers()).rejects.toThrow();
    });
  });
});

describe('ApiClient - Token Management', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should set token correctly', () => {
    const token = 'test-token-123';
    apiClient.setToken(token);

    // Token should be included in subsequent requests
    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: true,
      json: async () => ({ data: [] }),
    });

    apiClient.getCurrentBudgetMembers();

    expect(global.fetch).toHaveBeenCalledWith(
      expect.any(String),
      expect.objectContaining({
        headers: expect.objectContaining({
          Authorization: `Bearer ${token}`,
        }),
      })
    );
  });

  it('should clear token when set to null', () => {
    apiClient.setToken('test-token');
    apiClient.setToken(null);

    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: true,
      json: async () => ({ data: [] }),
    });

    apiClient.getCurrentBudgetMembers();

    const headers = (global.fetch as ReturnType<typeof vi.fn>).mock.calls[0][1]
      ?.headers as Record<string, string>;

    expect(headers.Authorization).toBeUndefined();
  });
});
