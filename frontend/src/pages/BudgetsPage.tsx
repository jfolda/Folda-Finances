import { useState, useEffect } from 'react';
import { apiClient } from '../lib/api';
import type { Category, CategoryBudget } from '../../../shared/types/api';

export function BudgetsPage() {
  const [categories, setCategories] = useState<Category[]>([]);
  const [budgets, setBudgets] = useState<CategoryBudget[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editAmount, setEditAmount] = useState('');

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      setLoading(true);
      const [categoriesRes, budgetsRes] = await Promise.all([
        apiClient.getCategories(),
        apiClient.getCategoryBudgets(),
      ]);
      setCategories(categoriesRes.data);
      setBudgets(budgetsRes.data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load data');
    } finally {
      setLoading(false);
    }
  };

  const handleSetBudget = async (categoryId: string, amount: number) => {
    try {
      const existingBudget = budgets.find((b) => b.category_id === categoryId);

      if (existingBudget) {
        // Update existing budget
        await apiClient.updateCategoryBudget(existingBudget.id, { amount });
      } else {
        // Create new budget
        await apiClient.createCategoryBudget({
          category_id: categoryId,
          amount,
          allocation_type: 'pooled',
        });
      }

      await loadData();
      setEditingId(null);
      setEditAmount('');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save budget');
    }
  };

  const handleDelete = async (budgetId: string) => {
    if (!confirm('Are you sure you want to remove this budget?')) return;

    try {
      await apiClient.deleteCategoryBudget(budgetId);
      await loadData();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete budget');
    }
  };

  const startEditing = (categoryId: string) => {
    const budget = budgets.find((b) => b.category_id === categoryId);
    setEditingId(categoryId);
    setEditAmount(budget ? (budget.amount / 100).toFixed(2) : '');
  };

  const cancelEditing = () => {
    setEditingId(null);
    setEditAmount('');
  };

  const saveEditing = (categoryId: string) => {
    const amount = Math.round(parseFloat(editAmount) * 100);
    if (isNaN(amount) || amount < 0) {
      alert('Please enter a valid amount');
      return;
    }
    handleSetBudget(categoryId, amount);
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading budgets...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto px-4 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">Budget Management</h1>
        <p className="mt-2 text-gray-600">
          Set monthly budgets for each category. These amounts will be used to calculate your available spending.
        </p>
      </div>

      {error && (
        <div className="mb-6 rounded-md bg-red-50 p-4">
          <p className="text-sm text-red-800">{error}</p>
        </div>
      )}

      <div className="bg-white shadow rounded-lg overflow-hidden">
        <div className="divide-y divide-gray-200">
          {categories.map((category) => {
            const budget = budgets.find((b) => b.category_id === category.id);
            const isEditing = editingId === category.id;

            return (
              <div
                key={category.id}
                className="p-4 hover:bg-gray-50 transition-colors"
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-3">
                    <div
                      className="w-10 h-10 rounded-full flex items-center justify-center text-xl"
                      style={{ backgroundColor: category.color + '20' }}
                    >
                      {category.icon}
                    </div>
                    <div>
                      <h3 className="text-sm font-medium text-gray-900">
                        {category.name}
                      </h3>
                      {budget && !isEditing && (
                        <p className="text-sm text-gray-500">
                          ${(budget.amount / 100).toFixed(2)}/month
                        </p>
                      )}
                      {!budget && !isEditing && (
                        <p className="text-sm text-gray-400">No budget set</p>
                      )}
                    </div>
                  </div>

                  <div className="flex items-center space-x-2">
                    {isEditing ? (
                      <>
                        <div className="flex items-center space-x-2">
                          <span className="text-gray-500">$</span>
                          <input
                            type="number"
                            step="0.01"
                            min="0"
                            value={editAmount}
                            onChange={(e) => setEditAmount(e.target.value)}
                            onKeyDown={(e) => {
                              if (e.key === 'Enter') {
                                saveEditing(category.id);
                              } else if (e.key === 'Escape') {
                                cancelEditing();
                              }
                            }}
                            className="w-32 rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm px-3 py-2 border"
                            placeholder="0.00"
                            autoFocus
                          />
                        </div>
                        <button
                          onClick={() => saveEditing(category.id)}
                          className="inline-flex items-center px-3 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                        >
                          Save
                        </button>
                        <button
                          onClick={cancelEditing}
                          className="inline-flex items-center px-3 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                        >
                          Cancel
                        </button>
                      </>
                    ) : (
                      <>
                        <button
                          onClick={() => startEditing(category.id)}
                          className="inline-flex items-center px-3 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                        >
                          {budget ? 'Edit' : 'Set Budget'}
                        </button>
                        {budget && (
                          <button
                            onClick={() => handleDelete(budget.id)}
                            className="inline-flex items-center px-3 py-2 border border-gray-300 text-sm font-medium rounded-md text-red-700 bg-white hover:bg-red-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
                          >
                            Remove
                          </button>
                        )}
                      </>
                    )}
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      </div>

      <div className="mt-6 bg-blue-50 rounded-lg p-4">
        <div className="flex">
          <div className="flex-shrink-0">
            <svg
              className="h-5 w-5 text-blue-400"
              viewBox="0 0 20 20"
              fill="currentColor"
            >
              <path
                fillRule="evenodd"
                d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z"
                clipRule="evenodd"
              />
            </svg>
          </div>
          <div className="ml-3 flex-1">
            <h3 className="text-sm font-medium text-blue-800">
              How budgets work
            </h3>
            <div className="mt-2 text-sm text-blue-700">
              <ul className="list-disc list-inside space-y-1">
                <li>Set monthly budget amounts for each category</li>
                <li>
                  Your "What Can I Spend?" page will show how much is available based on your budget period
                </li>
                <li>
                  Budgets are pro-rated based on your period settings (weekly, biweekly, or monthly)
                </li>
                <li>You can update or remove budgets at any time</li>
              </ul>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
