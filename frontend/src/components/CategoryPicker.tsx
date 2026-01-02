import { useState } from 'react';
import { PlusIcon } from '@heroicons/react/24/outline';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../lib/api';

interface Category {
  id: string;
  name: string;
  icon: string;
  color: string;
}

interface CategoryPickerProps {
  categories: Category[];
  value: string;
  onChange: (categoryId: string) => void;
  disabled?: boolean;
}

const AVAILABLE_ICONS = [
  '🏠', '🚗', '🍔', '🎬', '💊', '🎓', '👕', '🎮', '✈️', '🎵',
  '📱', '💰', '🛒', '☕', '🏋️', '🐕', '💡', '🔧', '📚', '🎨',
  '🍕', '🚌', '⚡', '💳', '🏥', '🎁', '🌐', '🔑', '🧺', '🚿'
];

const AVAILABLE_COLORS = [
  '#3B82F6', '#EF4444', '#10B981', '#F59E0B', '#8B5CF6',
  '#EC4899', '#14B8A6', '#F97316', '#6366F1', '#84CC16'
];

export function CategoryPicker({ categories, value, onChange, disabled }: CategoryPickerProps) {
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [newCategoryName, setNewCategoryName] = useState('');
  const [selectedIcon, setSelectedIcon] = useState(AVAILABLE_ICONS[0]);
  const [selectedColor, setSelectedColor] = useState(AVAILABLE_COLORS[0]);
  const queryClient = useQueryClient();

  const createMutation = useMutation({
    mutationFn: (data: { name: string; icon: string; color: string }) =>
      apiClient.createCategory(data),
    onSuccess: (response) => {
      queryClient.invalidateQueries({ queryKey: ['categories'] });
      onChange(response.data.id);
      setShowCreateForm(false);
      setNewCategoryName('');
      setSelectedIcon(AVAILABLE_ICONS[0]);
      setSelectedColor(AVAILABLE_COLORS[0]);
    },
  });

  const handleCreateCategory = () => {
    if (!newCategoryName.trim()) return;
    createMutation.mutate({
      name: newCategoryName.trim(),
      icon: selectedIcon,
      color: selectedColor,
    });
  };

  if (showCreateForm) {
    return (
      <div className="border border-gray-200 rounded-lg p-4 bg-gray-50">
        <h3 className="text-sm font-medium text-gray-900 mb-3">Create New Category</h3>

        <div className="space-y-3">
          {/* Category Name */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Category Name
            </label>
            <input
              type="text"
              value={newCategoryName}
              onChange={(e) => setNewCategoryName(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              placeholder="e.g., Pet Care"
              autoFocus
            />
          </div>

          {/* Icon Picker */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Choose Icon
            </label>
            <div className="grid grid-cols-10 gap-2">
              {AVAILABLE_ICONS.map((icon) => (
                <button
                  key={icon}
                  type="button"
                  onClick={() => setSelectedIcon(icon)}
                  className={`w-8 h-8 flex items-center justify-center rounded-lg text-lg transition-all ${
                    selectedIcon === icon
                      ? 'bg-blue-100 ring-2 ring-blue-500'
                      : 'bg-white hover:bg-gray-100 border border-gray-200'
                  }`}
                >
                  {icon}
                </button>
              ))}
            </div>
          </div>

          {/* Color Picker */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Choose Color
            </label>
            <div className="flex gap-2">
              {AVAILABLE_COLORS.map((color) => (
                <button
                  key={color}
                  type="button"
                  onClick={() => setSelectedColor(color)}
                  className={`w-8 h-8 rounded-full transition-all ${
                    selectedColor === color ? 'ring-2 ring-offset-2 ring-gray-400' : ''
                  }`}
                  style={{ backgroundColor: color }}
                />
              ))}
            </div>
          </div>

          {/* Preview */}
          <div className="flex items-center gap-2 p-3 bg-white rounded-lg border border-gray-200">
            <div
              className="w-10 h-10 rounded-full flex items-center justify-center text-xl"
              style={{ backgroundColor: selectedColor + '20' }}
            >
              {selectedIcon}
            </div>
            <span className="text-sm font-medium text-gray-900">
              {newCategoryName || 'Category Preview'}
            </span>
          </div>

          {/* Actions */}
          <div className="flex gap-2 pt-2">
            <button
              type="button"
              onClick={handleCreateCategory}
              disabled={!newCategoryName.trim() || createMutation.isPending}
              className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:bg-gray-300 disabled:cursor-not-allowed font-medium"
            >
              {createMutation.isPending ? 'Creating...' : 'Create Category'}
            </button>
            <button
              type="button"
              onClick={() => setShowCreateForm(false)}
              className="px-4 py-2 text-gray-700 hover:bg-gray-100 rounded-lg font-medium"
            >
              Cancel
            </button>
          </div>

          {createMutation.isError && (
            <p className="text-sm text-red-600">
              Failed to create category. Please try again.
            </p>
          )}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-2">
      <select
        value={value}
        onChange={(e) => onChange(e.target.value)}
        disabled={disabled}
        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 disabled:bg-gray-100"
      >
        <option value="">Select a category</option>
        {categories.map((category) => (
          <option key={category.id} value={category.id}>
            {category.icon} {category.name}
          </option>
        ))}
      </select>

      <button
        type="button"
        onClick={() => setShowCreateForm(true)}
        className="w-full flex items-center justify-center gap-2 px-3 py-2 text-sm text-blue-600 hover:bg-blue-50 border border-blue-200 rounded-lg transition-colors"
      >
        <PlusIcon className="h-4 w-4" />
        <span>Create New Category</span>
      </button>
    </div>
  );
}
