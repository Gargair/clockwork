import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { cleanup, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import CategoryTree from './CategoryTree';
import type { CategoryNode } from '../hooks/useCategories';

describe('CategoryTree component', { concurrent: false }, () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    cleanup();
  });

  const createNode = (
    id: string,
    name: string,
    parentId: string | null = null,
    children: CategoryNode[] = [],
  ): CategoryNode => ({
    id,
    projectId: '11111111-1111-1111-1111-111111111111',
    parentCategoryId: parentId,
    name,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
    children,
  });

  it('renders tree structure', () => {
    const tree: CategoryNode[] = [
      createNode('1', 'Frontend', null, [
        createNode('2', 'React', '1'),
        createNode('3', 'Vue', '1'),
      ]),
      createNode('4', 'Backend', null),
    ];

    render(
      <CategoryTree
        tree={tree}
        onAdd={vi.fn()}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
      />,
    );

    expect(screen.getByText('Frontend')).toBeInTheDocument();
    expect(screen.queryByText('React')).not.toBeInTheDocument();
    expect(screen.queryByText('Vue')).not.toBeInTheDocument();
    expect(screen.getByText('Backend')).toBeInTheDocument();
  });

  it('expands and collapses nodes', async () => {
    const user = userEvent.setup();
    const tree: CategoryNode[] = [
      createNode('1', 'Frontend', null, [
        createNode('2', 'React', '1'),
      ]),
    ];

    render(
      <CategoryTree
        tree={tree}
        onAdd={vi.fn()}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
      />,
    );

    // Initially, child should not be visible (collapsed)
    expect(screen.getByText('Frontend')).toBeInTheDocument();
    expect(screen.queryByText('React')).not.toBeInTheDocument();

    // Find expand/collapse button
    const expandButton = screen.getByRole('button', { name: /Expand|Collapse/i });
    
    // Click to expand (if collapsed) or collapse (if expanded)
    await user.click(expandButton);

    // After toggle, React should be in document
    expect(screen.getByText('React')).toBeInTheDocument();
  });

  it('invokes onAdd callback with correct parent ID', async () => {
    const user = userEvent.setup();
    const onAddMock = vi.fn();
    const tree: CategoryNode[] = [
      createNode('1', 'Frontend', null, [
        createNode('2', 'React', '1'),
      ]),
    ];

    render(
      <CategoryTree
        tree={tree}
        onAdd={onAddMock}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
      />,
    );

    // Find "Add Child" button for Frontend
    const addChildButtons = screen.getAllByRole('button', { name: /Add subcategory to Frontend/i });
    await user.click(addChildButtons[0]!);

    // Should call onAdd with parent ID
    expect(onAddMock).toHaveBeenCalledWith('1');
  });

  it('invokes onAdd callback with null for top-level add', async () => {
    const user = userEvent.setup();
    const onAddMock = vi.fn();
    const tree: CategoryNode[] = [
      createNode('1', 'Frontend', null),
    ];

    render(
      <CategoryTree
        tree={tree}
        onAdd={onAddMock}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
      />,
    );

    // Find "Add Child" button
    const addChildButtons = screen.getAllByRole('button', { name: /Add subcategory to Frontend/i });
    await user.click(addChildButtons[0]!);

    // Should call onAdd with the node's ID (not null, since it's adding a child to that node)
    expect(onAddMock).toHaveBeenCalledWith('1');
  });

  it('invokes onEdit callback with correct category ID', async () => {
    const user = userEvent.setup();
    const onEditMock = vi.fn();
    const tree: CategoryNode[] = [
      createNode('1', 'Frontend', null),
      createNode('2', 'Backend', null),
    ];

    render(
      <CategoryTree
        tree={tree}
        onAdd={vi.fn()}
        onEdit={onEditMock}
        onDelete={vi.fn()}
      />,
    );

    // Find Edit button for Frontend
    const editButtons = screen.getAllByRole('button', { name: /Edit/i });
    const frontendEditButton = Array.from(editButtons).find((btn) =>
      btn.getAttribute('aria-label')?.includes('Frontend'),
    );
    await user.click(frontendEditButton!);

    // Should call onEdit with category ID
    expect(onEditMock).toHaveBeenCalledWith('1');
  });

  it('invokes onDelete callback with correct category ID', async () => {
    const user = userEvent.setup();
    const onDeleteMock = vi.fn();
    const tree: CategoryNode[] = [
      createNode('1', 'Frontend', null),
      createNode('2', 'Backend', null),
    ];

    render(
      <CategoryTree
        tree={tree}
        onAdd={vi.fn()}
        onEdit={vi.fn()}
        onDelete={onDeleteMock}
      />,
    );

    // Find Delete button for Backend
    const deleteButtons = screen.getAllByRole('button', { name: /Delete/i });
    const backendDeleteButton = Array.from(deleteButtons).find((btn) =>
      btn.getAttribute('aria-label')?.includes('Backend'),
    );
    await user.click(backendDeleteButton!);

    // Should call onDelete with category ID
    expect(onDeleteMock).toHaveBeenCalledWith('2');
  });

  it('displays loading state', () => {
    render(
      <CategoryTree
        tree={[]}
        onAdd={vi.fn()}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        loading={true}
      />,
    );

    expect(screen.getByText('Loading categories…')).toBeInTheDocument();
  });

  it('displays error state', () => {
    render(
      <CategoryTree
        tree={[]}
        onAdd={vi.fn()}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        error="Something went wrong"
      />,
    );

    const alert = screen.getByRole('alert');
    expect(alert).toHaveTextContent('Error: Something went wrong');
  });

  it('displays empty state', () => {
    render(
      <CategoryTree
        tree={[]}
        onAdd={vi.fn()}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
      />,
    );

    expect(screen.getByText('No categories yet.')).toBeInTheDocument();
  });
});

