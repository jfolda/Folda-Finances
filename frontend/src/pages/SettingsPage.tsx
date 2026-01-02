import { useState } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { apiClient } from '../lib/api';
import type { ViewPeriod } from '../../../shared/types/api';

export function SettingsPage() {
  const { user } = useAuth();
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState(false);
  const [error, setError] = useState('');

  const [name, setName] = useState(user?.name || '');
  const [viewPeriod, setViewPeriod] = useState<ViewPeriod>(user?.view_period || 'monthly');
  const [periodStartDate, setPeriodStartDate] = useState(user?.period_start_date || '1');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    setSuccess(false);

    try {
      await apiClient.updateUserSettings({
        name,
        view_period: viewPeriod,
        period_start_date: periodStartDate,
      });
      setSuccess(true);
      setTimeout(() => setSuccess(false), 3000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update settings');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-3xl mx-auto px-4 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">Settings</h1>
        <p className="mt-2 text-gray-600">
          Manage your account settings and preferences
        </p>
      </div>

      <div className="bg-white shadow rounded-lg">
        <form onSubmit={handleSubmit} className="p-6 space-y-6">
          {/* Profile Section */}
          <div>
            <h2 className="text-lg font-medium text-gray-900 mb-4">Profile</h2>
            <div>
              <label
                htmlFor="name"
                className="block text-sm font-medium text-gray-700"
              >
                Name
              </label>
              <input
                type="text"
                id="name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm px-3 py-2 border"
              />
            </div>
          </div>

          {/* Budget Period Settings */}
          <div className="border-t pt-6">
            <h2 className="text-lg font-medium text-gray-900 mb-4">
              Budget Period Settings
            </h2>
            <p className="text-sm text-gray-600 mb-4">
              These settings control how "What Can I Spend?" calculates your available budget.
            </p>

            <div className="space-y-4">
              {/* View Period */}
              <div>
                <label
                  htmlFor="viewPeriod"
                  className="block text-sm font-medium text-gray-700 mb-2"
                >
                  Budget Period
                </label>
                <select
                  id="viewPeriod"
                  value={viewPeriod}
                  onChange={(e) => setViewPeriod(e.target.value as ViewPeriod)}
                  className="block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm px-3 py-2 border"
                >
                  <option value="weekly">Weekly</option>
                  <option value="biweekly">Biweekly (every 2 weeks)</option>
                  <option value="monthly">Monthly</option>
                </select>
                <p className="mt-2 text-sm text-gray-500">
                  Your budget calculations will be based on this time period.
                </p>
              </div>

              {/* Period Start Date */}
              <div>
                <label
                  htmlFor="periodStartDate"
                  className="block text-sm font-medium text-gray-700 mb-2"
                >
                  Period Start {viewPeriod === 'monthly' ? 'Day' : 'Day of Week'}
                </label>

                {viewPeriod === 'monthly' ? (
                  <select
                    id="periodStartDate"
                    value={periodStartDate}
                    onChange={(e) => setPeriodStartDate(e.target.value)}
                    className="block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm px-3 py-2 border"
                  >
                    {Array.from({ length: 28 }, (_, i) => i + 1).map((day) => (
                      <option key={day} value={day}>
                        {day}{getOrdinalSuffix(day)} of each month
                      </option>
                    ))}
                  </select>
                ) : (
                  <select
                    id="periodStartDate"
                    value={periodStartDate}
                    onChange={(e) => setPeriodStartDate(e.target.value)}
                    className="block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm px-3 py-2 border"
                  >
                    <option value="0">Sunday</option>
                    <option value="1">Monday</option>
                    <option value="2">Tuesday</option>
                    <option value="3">Wednesday</option>
                    <option value="4">Thursday</option>
                    <option value="5">Friday</option>
                    <option value="6">Saturday</option>
                  </select>
                )}

                <p className="mt-2 text-sm text-gray-500">
                  {viewPeriod === 'monthly'
                    ? 'The day of the month when your budget period starts (e.g., payday).'
                    : 'The day of the week when your budget period starts.'}
                </p>
              </div>
            </div>
          </div>

          {/* Success/Error Messages */}
          {success && (
            <div className="rounded-md bg-green-50 p-4">
              <p className="text-sm text-green-800">Settings updated successfully!</p>
            </div>
          )}

          {error && (
            <div className="rounded-md bg-red-50 p-4">
              <p className="text-sm text-red-800">{error}</p>
            </div>
          )}

          {/* Submit Button */}
          <div className="flex justify-end border-t pt-6">
            <button
              type="submit"
              disabled={loading}
              className="inline-flex justify-center rounded-md border border-transparent bg-blue-600 py-2 px-4 text-sm font-medium text-white shadow-sm hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {loading ? 'Saving...' : 'Save Changes'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

function getOrdinalSuffix(day: number): string {
  if (day > 3 && day < 21) return 'th';
  switch (day % 10) {
    case 1: return 'st';
    case 2: return 'nd';
    case 3: return 'rd';
    default: return 'th';
  }
}
