import { useState, useEffect } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { apiClient } from '../lib/api';
import type { ViewPeriod, BudgetInvitation } from '../../../shared/types/api';
import { EnvelopeIcon, CheckIcon, XMarkIcon } from '@heroicons/react/24/outline';

export function SettingsPage() {
  const { user } = useAuth();
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState(false);
  const [error, setError] = useState('');

  const [name, setName] = useState(user?.name || '');
  const [viewPeriod, setViewPeriod] = useState<ViewPeriod>(user?.view_period || 'monthly');
  const [periodStartDate, setPeriodStartDate] = useState(user?.period_start_date || '1');

  // Budget invitation state
  const [invitations, setInvitations] = useState<BudgetInvitation[]>([]);
  const [inviteEmail, setInviteEmail] = useState('');
  const [inviteLoading, setInviteLoading] = useState(false);
  const [inviteError, setInviteError] = useState('');
  const [inviteSuccess, setInviteSuccess] = useState('');

  // Load invitations on mount
  useEffect(() => {
    loadInvitations();
  }, []);

  const loadInvitations = async () => {
    try {
      const response = await apiClient.getBudgetInvitations();
      setInvitations(response.data);
    } catch (err) {
      console.error('Failed to load invitations:', err);
    }
  };

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

  const handleSendInvite = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!user?.budget_id) {
      setInviteError('You must belong to a budget to invite others');
      return;
    }

    setInviteLoading(true);
    setInviteError('');
    setInviteSuccess('');

    try {
      await apiClient.inviteToBudget(user.budget_id, {
        invitee_email: inviteEmail,
        invited_role: 'read_write',
      });
      setInviteSuccess('Invitation sent successfully!');
      setInviteEmail('');
      setTimeout(() => setInviteSuccess(''), 3000);
    } catch (err) {
      setInviteError(err instanceof Error ? err.message : 'Failed to send invitation');
    } finally {
      setInviteLoading(false);
    }
  };

  const handleAcceptInvitation = async (token: string) => {
    try {
      await apiClient.acceptBudgetInvitation(token);
      await loadInvitations();
      window.location.reload(); // Reload to update budget context
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to accept invitation');
    }
  };

  const handleDeclineInvitation = async (token: string) => {
    try {
      await apiClient.declineBudgetInvitation(token);
      await loadInvitations();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to decline invitation');
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

      {/* Budget Invitations Section */}
      {invitations.length > 0 && (
        <div className="mt-8 bg-white shadow rounded-lg p-6">
          <h2 className="text-lg font-medium text-gray-900 mb-4">Budget Invitations</h2>
          <p className="text-sm text-gray-600 mb-4">
            You have been invited to join {invitations.length} budget{invitations.length > 1 ? 's' : ''}.
          </p>

          <div className="space-y-3">
            {invitations.map((invitation) => (
              <div
                key={invitation.id}
                className="flex items-center justify-between p-4 border border-gray-200 rounded-lg bg-blue-50"
              >
                <div className="flex items-center gap-3">
                  <EnvelopeIcon className="h-6 w-6 text-blue-600" />
                  <div>
                    <p className="text-sm font-medium text-gray-900">
                      Budget Invitation
                    </p>
                    <p className="text-xs text-gray-500">
                      Invited {new Date(invitation.created_at).toLocaleDateString()}
                    </p>
                  </div>
                </div>
                <div className="flex gap-2">
                  <button
                    onClick={() => handleAcceptInvitation(invitation.token)}
                    className="inline-flex items-center gap-1 px-3 py-1.5 bg-green-600 text-white rounded-md hover:bg-green-700 text-sm font-medium"
                  >
                    <CheckIcon className="h-4 w-4" />
                    Accept
                  </button>
                  <button
                    onClick={() => handleDeclineInvitation(invitation.token)}
                    className="inline-flex items-center gap-1 px-3 py-1.5 bg-red-600 text-white rounded-md hover:bg-red-700 text-sm font-medium"
                  >
                    <XMarkIcon className="h-4 w-4" />
                    Decline
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Invite to Budget Section */}
      {user?.budget_id && (
        <div className="mt-8 bg-white shadow rounded-lg p-6">
          <h2 className="text-lg font-medium text-gray-900 mb-4">Invite Someone to Your Budget</h2>
          <p className="text-sm text-gray-600 mb-4">
            Share your budget with a partner or family member by sending them an invitation.
          </p>

          <form onSubmit={handleSendInvite} className="space-y-4">
            <div>
              <label htmlFor="inviteEmail" className="block text-sm font-medium text-gray-700 mb-1">
                Email Address
              </label>
              <input
                type="email"
                id="inviteEmail"
                value={inviteEmail}
                onChange={(e) => setInviteEmail(e.target.value)}
                placeholder="partner@example.com"
                className="block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm px-3 py-2 border"
                required
              />
            </div>

            {inviteSuccess && (
              <div className="rounded-md bg-green-50 p-4">
                <p className="text-sm text-green-800">{inviteSuccess}</p>
              </div>
            )}

            {inviteError && (
              <div className="rounded-md bg-red-50 p-4">
                <p className="text-sm text-red-800">{inviteError}</p>
              </div>
            )}

            <button
              type="submit"
              disabled={inviteLoading}
              className="inline-flex justify-center rounded-md border border-transparent bg-blue-600 py-2 px-4 text-sm font-medium text-white shadow-sm hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {inviteLoading ? 'Sending...' : 'Send Invitation'}
            </button>
          </form>
        </div>
      )}
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
