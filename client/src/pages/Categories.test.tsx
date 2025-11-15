import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { cleanup, render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import type { Category } from '../types';
import Categories from './Categories';
import { ApiError } from '../api/http';

// Mock the API layer used by the hook
const listCategoriesMock = vi.fn();
const createCategoryMock = vi.fn();
const updateCategoryMock = vi.fn();
const deleteCategoryMock = vi.fn();

vi.mock('../api/categories', () => ({
  listCategories: (...args: unknown[]) => listCategoriesMock(...args),
  createCategory: (...args: unknown[]) => createCategoryMock(...args),
  updateCategory: (...args: unknown[]) => updateCategoryMock(...args),
  deleteCategory: (...args: unknown[]) => deleteCategoryMock(...args),
}));

const projectId = '11111111-1111-1111-1111-111111111111';

function renderWithRouter() {
  return render(
    <MemoryRouter
      basename=""
      initialEntries={[{ pathname: `/projects/${projectId}/categories` }]}
      initialIndex={0}
    >
      <Routes>
        <Route path="/projects/:projectId/categories" element={<Categories />} />
      </Routes>
    </MemoryRouter>,
  );
}

describe('Categories page', { concurrent: false }, () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    cleanup();
  });

  it('renders list as a tree', async () => {
    const categories: Category[] = [
      {
        id: 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
        projectId,
        parentCategoryId: null,
        name: 'Frontend',
        description: 'Frontend work',
        createdAt: new Date('2025-11-02T10:00:00Z').toISOString(),
        updatedAt: new Date('2025-11-02T10:00:00Z').toISOString(),
      },
      {
        id: 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
        projectId,
        parentCategoryId: 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
        name: 'React',
        createdAt: new Date('2025-11-02T11:00:00Z').toISOString(),
        updatedAt: new Date('2025-11-02T11:00:00Z').toISOString(),
      },
    ];
    listCategoriesMock.mockResolvedValue(categories);

    renderWithRouter();

    // Category names appear
    expect(await screen.findByText('Frontend')).toBeInTheDocument();
    expect(screen.queryByText('React')).not.toBeInTheDocument();
  });

  it('creates a top-level category', async () => {
    const current: Category[] = [];
    listCategoriesMock.mockImplementation(async () => current.slice());
    createCategoryMock.mockImplementation(
      async (
        _projectId: string,
        {
          name,
          description,
          parentCategoryId,
        }: {
          name: string;
          description?: string | null;
          parentCategoryId?: string | null;
        },
      ) => {
        const now = new Date().toISOString();
        const newCategory: Category = {
          id: 'cccccccc-cccc-cccc-cccc-cccccccccccc',
          projectId,
          parentCategoryId: parentCategoryId ?? null,
          name,
          description: description === null ? undefined : description,
          createdAt: now,
          updatedAt: now,
        };
        current.push(newCategory);
        return newCategory;
      },
    );

    const user = userEvent.setup();
    renderWithRouter();

    // Click create button
    await user.click(await screen.findByRole('button', { name: 'Create Category' }));

    // Fill form
    const nameInput = await screen.findByLabelText('Name');
    await user.type(nameInput, 'Backend');
    await user.click(screen.getByRole('button', { name: 'Create' }));

    // New category appears
    expect(await screen.findByText('Backend')).toBeInTheDocument();
  });

  it('creates a child category', async () => {
    const now = new Date().toISOString();
    const parent: Category = {
      id: 'dddddddd-dddd-dddd-dddd-dddddddddddd',
      projectId,
      parentCategoryId: null,
      name: 'Frontend',
      createdAt: now,
      updatedAt: now,
    };
    const current: Category[] = [parent];
    listCategoriesMock.mockImplementation(async () => current.slice());
    createCategoryMock.mockImplementation(
      async (
        _projectId: string,
        { name, parentCategoryId }: { name: string; parentCategoryId?: string | null },
      ) => {
        const newCategory: Category = {
          id: 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee',
          projectId,
          parentCategoryId: parentCategoryId ?? null,
          name,
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
        };
        current.push(newCategory);
        return newCategory;
      },
    );

    const user = userEvent.setup();
    renderWithRouter();

    // Wait for parent to appear
    await screen.findByText('Frontend');

    // Click "Add Child" on parent
    const addChildButtons = screen.getAllByRole('button', { name: /Add subcategory to Frontend/i });
    await user.click(addChildButtons[0]!);

    // Fill form (parent should be pre-selected)
    const nameInput = await screen.findByLabelText('Name');
    await user.type(nameInput, 'React');
    await user.click(screen.getByRole('button', { name: 'Create' }));

    const expandButton = screen.queryByRole('button', { name: /Expand/i });
    
    if(expandButton) {
      // Click to expand (if collapsed)
      await user.click(expandButton);
    }

    // Child category appears
    expect(await screen.findByText('React')).toBeInTheDocument();
  });

  it('updates category name and parent (re-parenting)', async () => {
    const now = new Date().toISOString();
    const parent1: Category = {
      id: 'ffffffff-ffff-ffff-ffff-ffffffffffff',
      projectId,
      parentCategoryId: null,
      name: 'Frontend',
      createdAt: now,
      updatedAt: now,
    };
    const parent2: Category = {
      id: 'gggggggg-gggg-gggg-gggg-gggggggggggg',
      projectId,
      parentCategoryId: null,
      name: 'Backend',
      createdAt: now,
      updatedAt: now,
    };
    const child: Category = {
      id: 'hhhhhhhh-hhhh-hhhh-hhhh-hhhhhhhhhhhh',
      projectId,
      parentCategoryId: 'ffffffff-ffff-ffff-ffff-ffffffffffff',
      name: 'React',
      createdAt: now,
      updatedAt: now,
    };
    const current: Category[] = [parent1, parent2, child];
    listCategoriesMock.mockImplementation(async () => current.slice());
    updateCategoryMock.mockImplementation(
      async (
        _projectId: string,
        categoryId: string,
        { name, parentCategoryId }: { name: string; parentCategoryId?: string | null },
      ) => {
        const idx = current.findIndex((c) => c.id === categoryId);
        if (idx >= 0) {
          current[idx] = {
            ...current[idx]!,
            name,
            parentCategoryId: parentCategoryId ?? null,
            updatedAt: new Date().toISOString(),
          };
        }
        return current[idx];
      },
    );

    const user = userEvent.setup();
    renderWithRouter();

    const frontendNameNode = await screen.findByText('Frontend');
    
    const frontendExpandButton = within(frontendNameNode.parentElement!).queryByRole('button', { name: /Expand/i });
    
    if(frontendExpandButton) {
      // Click to expand (if collapsed)
      await user.click(frontendExpandButton);
    }

    // Wait for categories to appear
    const reactNameNode = await screen.findByText('React');

    // Click Edit on child
    const editButton = within(reactNameNode.parentElement!).getByRole('button', { name: /Edit/i });
    await user.click(editButton);

    // Update name and change parent
    const nameInput = await screen.findByDisplayValue('React');
    await user.clear(nameInput);
    await user.type(nameInput, 'React v2');

    const parentSelect = screen.getByLabelText('Parent Category');
    await user.selectOptions(parentSelect, 'gggggggg-gggg-gggg-gggg-gggggggggggg');

    await user.click(screen.getByRole('button', { name: 'Save' }));

    const backendExpandNode = await screen.findByText('Backend');
    
    const backendExpandButton = within(backendExpandNode.parentElement!).queryByRole('button', { name: /Expand/i });
    
    if(backendExpandButton) {
      // Click to expand (if collapsed)
      await user.click(backendExpandButton);
    }

    // Updated category appears
    expect(await screen.findByText('React v2')).toBeInTheDocument();
    expect(updateCategoryMock).toHaveBeenCalledWith(
      projectId,
      'hhhhhhhh-hhhh-hhhh-hhhh-hhhhhhhhhhhh',
      {
        name: 'React v2',
        description: null,
        parentCategoryId: 'gggggggg-gggg-gggg-gggg-gggggggggggg',
      },
    );
  });

  it('deletes a category', async () => {
    const now = new Date().toISOString();
    const category: Category = {
      id: 'iiiiiiii-iiii-iiii-iiii-iiiiiiiiiiii',
      projectId,
      parentCategoryId: null,
      name: 'Remove this category',
      createdAt: now,
      updatedAt: now,
    };
    const current: Category[] = [category];
    listCategoriesMock.mockImplementation(async () => current.slice());
    deleteCategoryMock.mockImplementation(async (_projectId: string, id: string) => {
      const idx = current.findIndex((c) => c.id === id);
      if (idx >= 0) current.splice(idx, 1);
    });

    const user = userEvent.setup();
    renderWithRouter();

    // Wait for category to appear
    await screen.findByText('Remove this category');

    // Delete
    const deleteButtons = screen.getAllByRole('button', { name: /Delete/i });
    const deleteButton = Array.from(deleteButtons).find((btn) =>
      btn.getAttribute('aria-label')?.includes('Remove this category'),
    );
    await user.click(deleteButton!);

    // Category should be removed
    expect(await screen.findByText('No categories yet.')).toBeInTheDocument();
  });

  it('shows error details with requestId when available', async () => {
    const err = new ApiError('internal: oops', 500, { code: 'internal', requestId: 'req-xyz' });
    listCategoriesMock.mockRejectedValue(err);

    renderWithRouter();

    const alert = await screen.findByRole('alert');
    expect(within(alert).getByText(/Error:/i)).toBeInTheDocument();
    expect(within(alert).getByText(/Code:/i)).toHaveTextContent('Code: internal');
    expect(within(alert).getByText(/Request ID:/i)).toHaveTextContent('Request ID: req-xyz');
  });

  it('shows user-friendly message for category_cycle error', async () => {
    const err = new ApiError('category_cycle: cycle detected', 409, {
      code: 'category_cycle',
      requestId: 'req-cycle',
    });
    createCategoryMock.mockRejectedValue(err);

    const user = userEvent.setup();
    renderWithRouter();

    // Try to create a category that would cause a cycle
    await user.click(await screen.findByRole('button', { name: 'Create Category' }));
    const nameInput = await screen.findByLabelText('Name');
    await user.type(nameInput, 'Test');
    await user.click(screen.getByRole('button', { name: 'Create' }));

    // Error message should be user-friendly
    const alert = await screen.findByRole('alert');
    expect(within(alert).getByText(/Cannot create a cycle/i)).toBeInTheDocument();
    expect(within(alert).getByText(/Request ID:/i)).toHaveTextContent('Request ID: req-cycle');
  });
});
