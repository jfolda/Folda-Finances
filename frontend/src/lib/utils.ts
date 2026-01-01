// Utility functions
import { type ClassValue, clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

// Format cents to dollar string
export function formatCurrency(cents: number): string {
  const dollars = cents / 100;
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
  }).format(dollars);
}

// Parse dollar string to cents
export function parseCurrency(dollarString: string): number {
  const cleaned = dollarString.replace(/[$,]/g, '');
  const dollars = parseFloat(cleaned);
  return Math.round(dollars * 100);
}

// Format date to YYYY-MM-DD
export function formatDate(date: Date): string {
  return date.toISOString().split('T')[0];
}

// Calculate pro-rated budget based on view period
export function calculateProratedBudget(
  monthlyAmount: number,
  viewPeriod: 'weekly' | 'biweekly' | 'monthly'
): number {
  const DAYS_PER_MONTH = 30.44; // Average days per month

  switch (viewPeriod) {
    case 'weekly':
      return Math.round(monthlyAmount * (7 / DAYS_PER_MONTH));
    case 'biweekly':
      return Math.round(monthlyAmount * (14 / DAYS_PER_MONTH));
    case 'monthly':
      return monthlyAmount;
    default:
      return monthlyAmount;
  }
}

// Get status color based on percentage used
export function getStatusColor(percentageUsed: number): string {
  if (percentageUsed > 100) return 'text-red-600';
  if (percentageUsed >= 75) return 'text-yellow-600';
  return 'text-green-600';
}

// Get background color based on percentage used
export function getStatusBgColor(percentageUsed: number): string {
  if (percentageUsed > 100) return 'bg-red-50';
  if (percentageUsed >= 75) return 'bg-yellow-50';
  return 'bg-green-50';
}

// Calculate days remaining in period
export function calculateDaysRemaining(endDate: string): number {
  const end = new Date(endDate);
  const now = new Date();
  const diffTime = end.getTime() - now.getTime();
  const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
  return Math.max(0, diffDays);
}
