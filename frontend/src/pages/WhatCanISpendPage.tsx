// "What Can I Spend?" page - CORE FEATURE
import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../lib/api';
import { formatCurrency, getStatusColor, calculateDaysRemaining } from '../lib/utils';
import { Link } from 'react-router-dom';
import { PlusIcon, ChevronDownIcon, ChevronUpIcon } from '@heroicons/react/24/outline';

export function WhatCanISpendPage() {
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['spending', 'available'],
    queryFn: () => apiClient.getSpendingAvailable(),
    staleTime: 30000, // 30 seconds
    refetchOnWindowFocus: true,
  });

  const handleRefresh = () => {
    refetch();
  };

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading your budget...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center max-w-md mx-auto px-4">
          <div className="rounded-md bg-red-50 p-4">
            <p className="text-sm text-red-800">
              {error instanceof Error ? error.message : 'Failed to load spending data'}
            </p>
          </div>
          <button
            onClick={handleRefresh}
            className="mt-4 text-sm text-blue-600 hover:text-blue-500"
          >
            Try again
          </button>
        </div>
      </div>
    );
  }

  const spendingData = data?.data;
  if (!spendingData) return null;

  const { period, summary, categories } = spendingData;
  const daysRemaining = calculateDaysRemaining(period.end_date);

  return (
    <div className="min-h-screen bg-gray-50 pb-20">
      {/* Header */}
      <div className="bg-white shadow-sm border-b border-gray-200">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold text-gray-900">
                What Can I Spend?
              </h1>
              <p className="mt-1 text-sm text-gray-500">
                Your real-time spending guide
              </p>
            </div>
            <button
              onClick={handleRefresh}
              className="text-sm text-blue-600 hover:text-blue-500 font-medium"
            >
              Refresh
            </button>
          </div>
        </div>
      </div>

      {/* Period Info */}
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-blue-900">Current Period</p>
              <p className="text-lg font-semibold text-blue-900">
                {new Date(period.start_date).toLocaleDateString()} -{' '}
                {new Date(period.end_date).toLocaleDateString()}
              </p>
            </div>
            <div className="text-right">
              <p className="text-sm font-medium text-blue-900">
                {daysRemaining} {daysRemaining === 1 ? 'day' : 'days'} left
              </p>
              <p className="text-sm text-blue-700 capitalize">{period.type} period</p>
            </div>
          </div>

          {/* Progress bar */}
          <div className="mt-3">
            <div className="h-2 bg-blue-200 rounded-full overflow-hidden">
              <div
                className="h-full bg-blue-600 transition-all duration-300"
                style={{
                  width: `${Math.min(
                    100,
                    ((period.days_remaining / getDaysInPeriod(period.type)) * 100)
                  )}%`,
                }}
              />
            </div>
          </div>
        </div>

        {/* Summary Card */}
        <div className="mt-6 bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Period Summary</h2>
          <div className="grid grid-cols-3 gap-4">
            <div>
              <p className="text-sm text-gray-500">Budgeted</p>
              <p className="text-2xl font-bold text-gray-900">
                {formatCurrency(summary.total_budgeted)}
              </p>
            </div>
            <div>
              <p className="text-sm text-gray-500">Spent</p>
              <p className="text-2xl font-bold text-gray-900">
                {formatCurrency(summary.total_spent)}
              </p>
            </div>
            <div>
              <p className="text-sm text-gray-500">Available</p>
              <p className={`text-2xl font-bold ${getStatusColor(
                (summary.total_spent / summary.total_budgeted) * 100
              )}`}>
                {formatCurrency(summary.total_available)}
              </p>
            </div>
          </div>
        </div>

        {/* Category Breakdown */}
        <div className="mt-6 space-y-4">
          <h2 className="text-lg font-semibold text-gray-900">By Category</h2>

          {categories.length === 0 ? (
            <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-8 text-center">
              <p className="text-gray-500">No budgets set up yet.</p>
              <Link
                to="/budgets"
                className="mt-4 inline-block text-blue-600 hover:text-blue-500 font-medium"
              >
                Create your first budget
              </Link>
            </div>
          ) : (
            categories.map((category) => (
              <CategoryCard key={category.category_id} category={category} />
            ))
          )}
        </div>
      </div>

      {/* Floating Add Button */}
      <Link
        to="/transactions/new"
        className="fixed bottom-6 right-6 bg-blue-600 text-white rounded-full p-4 shadow-lg hover:bg-blue-700 transition-all hover:scale-110"
      >
        <PlusIcon className="h-6 w-6" />
      </Link>
    </div>
  );
}

function CategoryCard({ category }: { category: any }) {
  const [isExpanded, setIsExpanded] = useState(false);
  const percentageUsed = category.percentage_used;
  const statusColor = getStatusColor(percentageUsed);

  return (
    <div className="bg-white rounded-lg border border-gray-200 hover:border-gray-300 transition-colors">
      {/* Compact Header - Always Visible */}
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className="w-full px-4 py-3 flex items-center justify-between text-left"
      >
        <div className="flex items-center gap-3 flex-1 min-w-0">
          <div
            className="w-8 h-8 rounded-full flex items-center justify-center text-base flex-shrink-0"
            style={{ backgroundColor: category.category_color + '20' }}
          >
            {category.category_icon}
          </div>
          <div className="flex-1 min-w-0">
            <h3 className="text-sm font-medium text-gray-900 truncate">
              {category.category_name}
            </h3>
            {category.is_split && (
              <p className="text-xs text-gray-500">Your share</p>
            )}
          </div>
          <div className="text-right flex-shrink-0">
            <p className={`text-lg font-bold ${statusColor}`}>
              {category.available >= 0
                ? formatCurrency(category.available)
                : `−${formatCurrency(Math.abs(category.available))}`}
            </p>
          </div>
          <div className="flex-shrink-0 ml-2">
            {isExpanded ? (
              <ChevronUpIcon className="h-5 w-5 text-gray-400" />
            ) : (
              <ChevronDownIcon className="h-5 w-5 text-gray-400" />
            )}
          </div>
        </div>
      </button>

      {/* Expanded Details */}
      {isExpanded && (
        <div className="px-4 pb-4 border-t border-gray-100">
          <div className="pt-3 space-y-3">
            {/* Budget Details */}
            <div className="grid grid-cols-2 gap-4 text-sm">
              <div>
                <p className="text-gray-500">Budgeted</p>
                <p className="font-semibold text-gray-900">
                  {formatCurrency(category.budgeted)}
                </p>
              </div>
              <div>
                <p className="text-gray-500">Spent</p>
                <p className="font-semibold text-gray-900">
                  {formatCurrency(category.spent)}
                </p>
              </div>
            </div>

            {/* Progress Bar */}
            <div>
              <div className="h-2 bg-gray-100 rounded-full overflow-hidden">
                <div
                  className={`h-full transition-all duration-300 ${
                    percentageUsed > 100
                      ? 'bg-red-600'
                      : percentageUsed >= 75
                      ? 'bg-yellow-500'
                      : 'bg-green-500'
                  }`}
                  style={{ width: `${Math.min(100, percentageUsed)}%` }}
                />
              </div>
              <p className="text-xs text-gray-500 mt-1 text-right">
                {percentageUsed.toFixed(0)}% used
              </p>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

function getDaysInPeriod(type: string): number {
  switch (type) {
    case 'weekly':
      return 7;
    case 'biweekly':
      return 14;
    case 'monthly':
      return 30;
    default:
      return 30;
  }
}
