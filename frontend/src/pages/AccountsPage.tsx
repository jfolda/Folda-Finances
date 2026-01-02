import { useState, useEffect } from 'react';
import { apiClient } from '../lib/api';
import type { Account, AccountType } from '../../../shared/types/api';
import {
  BanknotesIcon,
  BuildingLibraryIcon,
  CreditCardIcon,
  PlusIcon,
  PencilIcon,
  TrashIcon,
  ChartBarIcon,
} from '@heroicons/react/24/outline';

const ACCOUNT_TYPES: { value: AccountType; label: string; icon: typeof BanknotesIcon }[] = [
  { value: 'checking', label: 'Checking', icon: BuildingLibraryIcon },
  { value: 'savings', label: 'Savings', icon: BuildingLibraryIcon },
  { value: 'credit_card', label: 'Credit Card', icon: CreditCardIcon },
  { value: 'cash', label: 'Cash', icon: BanknotesIcon },
  { value: 'investment', label: 'Investment', icon: ChartBarIcon },
  { value: 'other', label: 'Other', icon: BanknotesIcon },
];

function getAccountTypeInfo(type: AccountType) {
  return ACCOUNT_TYPES.find(t => t.value === type) || ACCOUNT_TYPES[5];
}

export function AccountsPage() {
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showAddForm, setShowAddForm] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);

  // Form state
  const [formData, setFormData] = useState({
    name: '',
    type: 'checking' as AccountType,
    balance: '',
    currency: 'USD',
    notes: '',
  });

  useEffect(() => {
    loadAccounts();
  }, []);

  const loadAccounts = async () => {
    try {
      setLoading(true);
      const res = await apiClient.getAccounts();
      setAccounts(res.data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load accounts');
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    const balanceInCents = Math.round(parseFloat(formData.balance || '0') * 100);

    try {
      if (editingId) {
        await apiClient.updateAccount(editingId, {
          name: formData.name,
          type: formData.type,
          balance: balanceInCents,
          currency: formData.currency,
          notes: formData.notes || '',
        });
      } else {
        await apiClient.createAccount({
          name: formData.name,
          type: formData.type,
          balance: balanceInCents,
          currency: formData.currency,
          notes: formData.notes || '',
        });
      }

      await loadAccounts();
      resetForm();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save account');
    }
  };

  const handleEdit = (account: Account) => {
    setEditingId(account.id);
    setFormData({
      name: account.name,
      type: account.type,
      balance: (account.balance / 100).toFixed(2),
      currency: account.currency,
      notes: account.notes || '',
    });
    setShowAddForm(true);
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this account? Transactions linked to this account will not be deleted.')) {
      return;
    }

    try {
      await apiClient.deleteAccount(id);
      await loadAccounts();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete account');
    }
  };

  const resetForm = () => {
    setFormData({
      name: '',
      type: 'checking',
      balance: '',
      currency: 'USD',
      notes: '',
    });
    setEditingId(null);
    setShowAddForm(false);
  };

  const totalBalance = accounts
    .filter(a => a.is_active)
    .reduce((sum, account) => {
      // For credit cards, negative balance is what you owe
      if (account.type === 'credit_card') {
        return sum - account.balance;
      }
      return sum + account.balance;
    }, 0);

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading accounts...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto px-4 py-8">
      <div className="mb-8">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Accounts</h1>
            <p className="mt-2 text-gray-600">
              Manage your bank accounts, credit cards, and cash.
            </p>
          </div>
          <button
            onClick={() => setShowAddForm(true)}
            className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
          >
            <PlusIcon className="h-5 w-5 mr-2" />
            Add Account
          </button>
        </div>
      </div>

      {error && (
        <div className="mb-6 rounded-md bg-red-50 p-4">
          <p className="text-sm text-red-800">{error}</p>
        </div>
      )}

      {/* Total Net Worth */}
      <div className="mb-6 bg-gradient-to-r from-blue-500 to-blue-600 rounded-lg p-6 text-white shadow-lg">
        <p className="text-sm font-medium opacity-90">Total Net Worth</p>
        <p className="text-4xl font-bold mt-2">
          ${(totalBalance / 100).toFixed(2)}
        </p>
        <p className="text-sm mt-2 opacity-75">
          Across {accounts.filter(a => a.is_active).length} active account{accounts.filter(a => a.is_active).length !== 1 ? 's' : ''}
        </p>
      </div>

      {/* Add/Edit Form */}
      {showAddForm && (
        <div className="mb-6 bg-white shadow rounded-lg p-6">
          <h2 className="text-lg font-medium text-gray-900 mb-4">
            {editingId ? 'Edit Account' : 'Add New Account'}
          </h2>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700">
                Account Name
              </label>
              <input
                type="text"
                required
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm px-3 py-2 border"
                placeholder="e.g., Chase Checking"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700">
                Account Type
              </label>
              <select
                value={formData.type}
                onChange={(e) => setFormData({ ...formData, type: e.target.value as AccountType })}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm px-3 py-2 border"
              >
                {ACCOUNT_TYPES.map((type) => (
                  <option key={type.value} value={type.value}>
                    {type.label}
                  </option>
                ))}
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700">
                Current Balance
              </label>
              <div className="mt-1 relative rounded-md shadow-sm">
                <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                  <span className="text-gray-500 sm:text-sm">$</span>
                </div>
                <input
                  type="number"
                  step="0.01"
                  required
                  value={formData.balance}
                  onChange={(e) => setFormData({ ...formData, balance: e.target.value })}
                  className="block w-full rounded-md border-gray-300 pl-7 focus:border-blue-500 focus:ring-blue-500 sm:text-sm px-3 py-2 border"
                  placeholder="0.00"
                />
              </div>
              {formData.type === 'credit_card' && (
                <p className="mt-1 text-xs text-gray-500">
                  For credit cards, enter your current balance (what you owe). Positive means you owe money.
                </p>
              )}
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700">
                Notes (Optional)
              </label>
              <textarea
                value={formData.notes}
                onChange={(e) => setFormData({ ...formData, notes: e.target.value })}
                rows={2}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm px-3 py-2 border"
                placeholder="Any additional details..."
              />
            </div>

            <div className="flex justify-end space-x-3">
              <button
                type="button"
                onClick={resetForm}
                className="px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
              >
                Cancel
              </button>
              <button
                type="submit"
                className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
              >
                {editingId ? 'Update Account' : 'Add Account'}
              </button>
            </div>
          </form>
        </div>
      )}

      {/* Accounts List */}
      <div className="bg-white shadow rounded-lg overflow-hidden">
        {accounts.length === 0 ? (
          <div className="text-center py-12">
            <BuildingLibraryIcon className="mx-auto h-12 w-12 text-gray-400" />
            <h3 className="mt-2 text-sm font-medium text-gray-900">No accounts</h3>
            <p className="mt-1 text-sm text-gray-500">
              Get started by adding your first account.
            </p>
          </div>
        ) : (
          <div className="divide-y divide-gray-200">
            {accounts.map((account) => {
              const typeInfo = getAccountTypeInfo(account.type);
              const Icon = typeInfo.icon;
              const displayBalance = account.type === 'credit_card'
                ? -account.balance // Show credit card debt as negative
                : account.balance;

              return (
                <div
                  key={account.id}
                  className={`p-4 hover:bg-gray-50 transition-colors ${!account.is_active ? 'opacity-50' : ''}`}
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-4 flex-1">
                      <div className="flex-shrink-0">
                        <div className="w-12 h-12 rounded-full bg-blue-100 flex items-center justify-center">
                          <Icon className="h-6 w-6 text-blue-600" />
                        </div>
                      </div>
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center space-x-2">
                          <h3 className="text-sm font-medium text-gray-900 truncate">
                            {account.name}
                          </h3>
                          {!account.is_active && (
                            <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-800">
                              Inactive
                            </span>
                          )}
                        </div>
                        <p className="text-sm text-gray-500">{typeInfo.label}</p>
                        {account.notes && (
                          <p className="text-xs text-gray-400 mt-1 truncate">
                            {account.notes}
                          </p>
                        )}
                      </div>
                      <div className="text-right">
                        <p className={`text-lg font-semibold ${displayBalance >= 0 ? 'text-gray-900' : 'text-red-600'}`}>
                          ${(Math.abs(displayBalance) / 100).toFixed(2)}
                        </p>
                        <p className="text-xs text-gray-500">{account.currency}</p>
                      </div>
                    </div>
                    <div className="ml-4 flex items-center space-x-2">
                      <button
                        onClick={() => handleEdit(account)}
                        className="p-2 text-gray-400 hover:text-blue-600 focus:outline-none"
                        title="Edit account"
                      >
                        <PencilIcon className="h-5 w-5" />
                      </button>
                      <button
                        onClick={() => handleDelete(account.id)}
                        className="p-2 text-gray-400 hover:text-red-600 focus:outline-none"
                        title="Delete account"
                      >
                        <TrashIcon className="h-5 w-5" />
                      </button>
                    </div>
                  </div>
                </div>
              );
            })}
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
              About Accounts
            </h3>
            <div className="mt-2 text-sm text-blue-700">
              <ul className="list-disc list-inside space-y-1">
                <li>Track balances across all your financial accounts</li>
                <li>Link transactions to specific accounts for better tracking</li>
                <li>Your net worth is calculated automatically</li>
                <li>Inactive accounts are hidden from net worth calculations</li>
              </ul>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
