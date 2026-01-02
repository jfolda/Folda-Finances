import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../lib/api';
import { PencilIcon, TrashIcon } from '@heroicons/react/24/outline';
import { CategoryPicker } from '../components/CategoryPicker';

export function TransactionDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const [isEditing, setIsEditing] = useState(false);
  const [description, setDescription] = useState('');
  const [amount, setAmount] = useState('');
  const [categoryId, setCategoryId] = useState('');
  const [accountId, setAccountId] = useState('');
  const [date, setDate] = useState('');
  const [error, setError] = useState('');

  const { data: transactionResponse, isLoading: loadingTransaction } = useQuery({
    queryKey: ['transaction', id],
    queryFn: () => apiClient.getTransaction(id!),
    enabled: !!id,
  });

  // Initialize form when transaction loads
  useEffect(() => {
    if (transactionResponse?.data && !isEditing && !description) {
      const tx = transactionResponse.data;
      setDescription(tx.description);
      setAmount((Math.abs(tx.amount) / 100).toFixed(2));
      setCategoryId(tx.category_id);
      setAccountId(tx.account_id || '');
      setDate(tx.date.split('T')[0]);
    }
  }, [transactionResponse, isEditing, description]);

  const { data: categoriesResponse } = useQuery({
    queryKey: ['categories'],
    queryFn: () => apiClient.getCategories(),
  });

  const { data: accountsResponse } = useQuery({
    queryKey: ['accounts'],
    queryFn: () => apiClient.getAccounts(),
  });

  const updateMutation = useMutation({
    mutationFn: (data: any) => apiClient.updateTransaction(id!, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['transaction', id] });
      queryClient.invalidateQueries({ queryKey: ['transactions'] });
      queryClient.invalidateQueries({ queryKey: ['spending', 'available'] });
      setIsEditing(false);
      setError('');
    },
    onError: (err) => {
      setError(err instanceof Error ? err.message : 'Failed to update transaction');
    },
  });

  const deleteMutation = useMutation({
    mutationFn: () => apiClient.deleteTransaction(id!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['transactions'] });
      queryClient.invalidateQueries({ queryKey: ['spending', 'available'] });
      navigate('/transactions');
    },
    onError: (err) => {
      setError(err instanceof Error ? err.message : 'Failed to delete transaction');
    },
  });

  const handleUpdate = () => {
    setError('');

    if (!description.trim()) {
      setError('Please enter a description');
      return;
    }

    if (!amount || parseFloat(amount) === 0) {
      setError('Please enter a valid amount');
      return;
    }

    if (!categoryId) {
      setError('Please select a category');
      return;
    }

    const amountInCents = Math.round(parseFloat(amount) * 100);

    updateMutation.mutate({
      description: description.trim(),
      amount: -Math.abs(amountInCents),
      category_id: categoryId,
      date,
      account_id: accountId || null,
    });
  };

  const handleDelete = () => {
    if (confirm('Are you sure you want to delete this transaction? This cannot be undone.')) {
      deleteMutation.mutate();
    }
  };

  const categories = categoriesResponse?.data || [];
  const accounts = accountsResponse?.data.filter(a => a.is_active) || [];
  const transaction = transactionResponse?.data;

  if (loadingTransaction) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading transaction...</p>
        </div>
      </div>
    );
  }

  if (!transaction) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <p className="text-gray-600">Transaction not found</p>
          <button
            onClick={() => navigate('/transactions')}
            className="mt-4 text-blue-600 hover:text-blue-500"
          >
            Back to Transactions
          </button>
        </div>
      </div>
    );
  }

  const category = categories.find(c => c.id === transaction.category_id);
  const account = accounts.find(a => a.id === transaction.account_id);

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-2xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        <div className="mb-6">
          <button
            onClick={() => navigate('/transactions')}
            className="text-sm text-blue-600 hover:text-blue-500"
          >
            ← Back to Transactions
          </button>
        </div>

        <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
          {/* Header */}
          <div className="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
            <h1 className="text-2xl font-bold text-gray-900">
              {isEditing ? 'Edit Transaction' : 'Transaction Details'}
            </h1>
            {!isEditing && (
              <div className="flex gap-2">
                <button
                  onClick={() => setIsEditing(true)}
                  className="p-2 text-gray-400 hover:text-blue-600"
                  title="Edit transaction"
                >
                  <PencilIcon className="h-5 w-5" />
                </button>
                <button
                  onClick={handleDelete}
                  className="p-2 text-gray-400 hover:text-red-600"
                  title="Delete transaction"
                  disabled={deleteMutation.isPending}
                >
                  <TrashIcon className="h-5 w-5" />
                </button>
              </div>
            )}
          </div>

          {error && (
            <div className="mx-6 mt-4 rounded-md bg-red-50 p-4">
              <p className="text-sm text-red-800">{error}</p>
            </div>
          )}

          {/* Content */}
          <div className="p-6">
            {isEditing ? (
              <div className="space-y-6">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Description *
                  </label>
                  <input
                    type="text"
                    value={description}
                    onChange={(e) => setDescription(e.target.value)}
                    className="block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm px-3 py-2 border"
                    placeholder="e.g., Grocery shopping"
                    required
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Amount *
                  </label>
                  <div className="relative">
                    <div className="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
                      <span className="text-gray-500 sm:text-sm">$</span>
                    </div>
                    <input
                      type="number"
                      step="0.01"
                      value={amount}
                      onChange={(e) => setAmount(e.target.value)}
                      className="block w-full rounded-md border-gray-300 pl-7 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm px-3 py-2 border"
                      placeholder="0.00"
                      required
                    />
                  </div>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Category *
                  </label>
                  <CategoryPicker
                    categories={categories}
                    value={categoryId}
                    onChange={setCategoryId}
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Account (Optional)
                  </label>
                  <select
                    value={accountId}
                    onChange={(e) => setAccountId(e.target.value)}
                    className="block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm px-3 py-2 border"
                  >
                    <option value="">No account selected</option>
                    {accounts.map((acc) => (
                      <option key={acc.id} value={acc.id}>
                        {acc.name} (${(acc.balance / 100).toFixed(2)})
                      </option>
                    ))}
                  </select>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Date *
                  </label>
                  <input
                    type="date"
                    value={date}
                    onChange={(e) => setDate(e.target.value)}
                    className="block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm px-3 py-2 border"
                    required
                  />
                </div>

                <div className="flex gap-3 pt-4 border-t border-gray-200">
                  <button
                    onClick={() => {
                      setIsEditing(false);
                      setError('');
                      // Reset to original values
                      const tx = transaction;
                      setDescription(tx.description);
                      setAmount((Math.abs(tx.amount) / 100).toFixed(2));
                      setCategoryId(tx.category_id);
                      setAccountId(tx.account_id || '');
                      setDate(tx.date.split('T')[0]);
                    }}
                    className="flex-1 px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
                  >
                    Cancel
                  </button>
                  <button
                    onClick={handleUpdate}
                    disabled={updateMutation.isPending}
                    className="flex-1 px-4 py-2 text-sm font-medium text-white bg-blue-600 border border-transparent rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {updateMutation.isPending ? 'Saving...' : 'Save Changes'}
                  </button>
                </div>
              </div>
            ) : (
              <div className="space-y-6">
                {/* Amount */}
                <div>
                  <p className="text-sm text-gray-500">Amount</p>
                  <p className="text-3xl font-bold text-red-600 mt-1">
                    −${(Math.abs(transaction.amount) / 100).toFixed(2)}
                  </p>
                </div>

                {/* Description */}
                <div>
                  <p className="text-sm text-gray-500">Description</p>
                  <p className="text-lg text-gray-900 mt-1">{transaction.description}</p>
                </div>

                {/* Category */}
                <div>
                  <p className="text-sm text-gray-500">Category</p>
                  <div className="flex items-center gap-2 mt-1">
                    {category && (
                      <>
                        <div
                          className="w-8 h-8 rounded-full flex items-center justify-center text-base"
                          style={{ backgroundColor: category.color + '20' }}
                        >
                          {category.icon}
                        </div>
                        <span className="text-lg text-gray-900">{category.name}</span>
                      </>
                    )}
                  </div>
                </div>

                {/* Account */}
                {account && (
                  <div>
                    <p className="text-sm text-gray-500">Account</p>
                    <p className="text-lg text-gray-900 mt-1">{account.name}</p>
                  </div>
                )}

                {/* Date */}
                <div>
                  <p className="text-sm text-gray-500">Date</p>
                  <p className="text-lg text-gray-900 mt-1">
                    {new Date(transaction.date).toLocaleDateString('en-US', {
                      weekday: 'long',
                      year: 'numeric',
                      month: 'long',
                      day: 'numeric',
                    })}
                  </p>
                </div>

                {/* Metadata */}
                <div className="pt-4 border-t border-gray-200">
                  <p className="text-xs text-gray-400">
                    Created {new Date(transaction.created_at).toLocaleString()}
                  </p>
                  {transaction.updated_at !== transaction.created_at && (
                    <p className="text-xs text-gray-400 mt-1">
                      Last updated {new Date(transaction.updated_at).toLocaleString()}
                    </p>
                  )}
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
