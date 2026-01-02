import { useState, useEffect } from 'react';
import { apiClient } from '../lib/api';
import type { Category, CategoryBudget } from '../../../shared/types/api';
import { PencilIcon, TrashIcon, PlusIcon } from '@heroicons/react/24/outline';

export function BudgetsPage() {
  const [categories, setCategories] = useState<Category[]>([]);
  const [budgets, setBudgets] = useState<CategoryBudget[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editAmount, setEditAmount] = useState('');
  const [showAddForm, setShowAddForm] = useState(false);
  const [selectedCategoryId, setSelectedCategoryId] = useState('');

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
    setShowAddForm(false);
  };

  const startAdding = () => {
    setShowAddForm(true);
    setEditingId(null);
    setSelectedCategoryId('');
    setEditAmount('');
  };

  const cancelEditing = () => {
    setEditingId(null);
    setEditAmount('');
    setShowAddForm(false);
    setSelectedCategoryId('');
  };

  const saveEditing = (categoryId: string) => {
    const amount = Math.round(parseFloat(editAmount) * 100);
    if (isNaN(amount) || amount < 0) {
      alert('Please enter a valid amount');
      return;
    }
    handleSetBudget(categoryId, amount);
  };

  const saveNewBudget = () => {
    if (!selectedCategoryId) {
      alert('Please select a category');
      return;
    }
    const amount = Math.round(parseFloat(editAmount) * 100);
    if (isNaN(amount) || amount < 0) {
      alert('Please enter a valid amount');
      return;
    }
    handleSetBudget(selectedCategoryId, amount);
    setShowAddForm(false);
  };

  // Get categories that already have budgets
  const budgetedCategories = categories.filter((cat) =>
    budgets.some((b) => b.category_id === cat.id)
  );

  // Get categories without budgets for the add form
  const unbudgetedCategories = categories.filter((cat) =>
    !budgets.some((b) => b.category_id === cat.id)
  );

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

      {/* Budgeted Categories */}
      <div className="bg-white shadow rounded-lg overflow-hidden">
        {budgetedCategories.length === 0 && !showAddForm ? (
          <div className="p-12 text-center">
            <p className="text-gray-500 mb-4">No budgets set yet.</p>
            <button
              onClick={startAdding}
              className="inline-flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
            >
              <PlusIcon className="h-5 w-5" />
              Add Your First Budget
            </button>
          </div>
        ) : (
          <div className="divide-y divide-gray-200">
            {budgetedCategories.map((category) => {
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
                            className="inline-flex items-center px-3 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700"
                          >
                            Save
                          </button>
                          <button
                            onClick={cancelEditing}
                            className="inline-flex items-center px-3 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50"
                          >
                            Cancel
                          </button>
                        </>
                      ) : (
                        <>
                          <button
                            onClick={() => startEditing(category.id)}
                            className="p-2 text-gray-400 hover:text-blue-600"
                            title="Edit budget"
                          >
                            <PencilIcon className="h-5 w-5" />
                          </button>
                          <button
                            onClick={() => budget && handleDelete(budget.id)}
                            className="p-2 text-gray-400 hover:text-red-600"
                            title="Remove budget"
                          >
                            <TrashIcon className="h-5 w-5" />
                          </button>
                        </>
                      )}
                    </div>
                  </div>
                </div>
              );
            })}

            {/* Add New Budget Form */}
            {showAddForm && unbudgetedCategories.length > 0 && (
              <div className="p-4 bg-blue-50 border-t-2 border-blue-200">
                <h3 className="text-sm font-medium text-gray-900 mb-3">
                  Add New Budget
                </h3>
                <div className="flex items-end gap-3">
                  <div className="flex-1">
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Category
                    </label>
                    <select
                      value={selectedCategoryId}
                      onChange={(e) => setSelectedCategoryId(e.target.value)}
                      className="block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm px-3 py-2 border"
                    >
                      <option value="">Select a category</option>
                      {unbudgetedCategories.map((cat) => (
                        <option key={cat.id} value={cat.id}>
                          {cat.icon} {cat.name}
                        </option>
                      ))}
                    </select>
                  </div>
                  <div className="flex-1">
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Monthly Budget
                    </label>
                    <div className="flex items-center">
                      <span className="text-gray-500 mr-2">$</span>
                      <input
                        type="number"
                        step="0.01"
                        min="0"
                        value={editAmount}
                        onChange={(e) => setEditAmount(e.target.value)}
                        onKeyDown={(e) => {
                          if (e.key === 'Enter') {
                            saveNewBudget();
                          } else if (e.key === 'Escape') {
                            cancelEditing();
                          }
                        }}
                        className="block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm px-3 py-2 border"
                        placeholder="0.00"
                      />
                    </div>
                  </div>
                  <div className="flex gap-2">
                    <button
                      onClick={saveNewBudget}
                      className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 text-sm font-medium"
                    >
                      Add
                    </button>
                    <button
                      onClick={cancelEditing}
                      className="px-4 py-2 border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50 text-sm font-medium"
                    >
                      Cancel
                    </button>
                  </div>
                </div>
              </div>
            )}

            {/* Large Add Button */}
            {!showAddForm && unbudgetedCategories.length > 0 && (
              <button
                onClick={startAdding}
                className="w-full p-6 flex items-center justify-center gap-2 text-blue-600 hover:bg-blue-50 transition-colors border-t-2 border-dashed border-gray-200 hover:border-blue-300"
              >
                <PlusIcon className="h-6 w-6" />
                <span className="font-medium">Add Budget</span>
              </button>
            )}
          </div>
        )}
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
