import { describe, it, expect } from 'vitest';
import { buildCategoryTree } from './useCategories';
import type { Category } from '../types';

describe('buildCategoryTree', () => {
  const projectId = '11111111-1111-1111-1111-111111111111';

  const createCategory = (
    id: string,
    name: string,
    parentId: string | null = null,
    description?: string,
  ): Category => ({
    id,
    projectId,
    parentCategoryId: parentId,
    name,
    description,
    createdAt: new Date('2025-11-02T10:00:00Z').toISOString(),
    updatedAt: new Date('2025-11-02T10:00:00Z').toISOString(),
  });

  it('builds acyclic nesting correctly', () => {
    const categories: Category[] = [
      createCategory('1', 'Frontend', null),
      createCategory('2', 'Backend', null),
      createCategory('3', 'React', '1'),
      createCategory('4', 'Vue', '1'),
      createCategory('5', 'Components', '3'),
    ];

    const tree = buildCategoryTree(categories);

    // Should have 2 root nodes
    expect(tree).toHaveLength(2);
    expect(tree[0]?.name).toBe('Backend'); // Sorted by name
    expect(tree[1]?.name).toBe('Frontend');

    // Frontend should have 2 children
    const frontend = tree[1]!;
    expect(frontend.children).toHaveLength(2);
    expect(frontend.children[0]?.name).toBe('React'); // Sorted by name
    expect(frontend.children[1]?.name).toBe('Vue');

    // React should have 1 child
    const react = frontend.children[0]!;
    expect(react.children).toHaveLength(1);
    expect(react.children[0]?.name).toBe('Components');
  });

  it('handles single root category', () => {
    const categories: Category[] = [createCategory('1', 'Root', null)];

    const tree = buildCategoryTree(categories);

    expect(tree).toHaveLength(1);
    expect(tree[0]?.name).toBe('Root');
    expect(tree[0]?.children).toHaveLength(0);
  });

  it('handles multiple levels of nesting', () => {
    const categories: Category[] = [
      createCategory('1', 'Level1', null),
      createCategory('2', 'Level2', '1'),
      createCategory('3', 'Level3', '2'),
      createCategory('4', 'Level4', '3'),
    ];

    const tree = buildCategoryTree(categories);

    expect(tree).toHaveLength(1);
    const level1 = tree[0]!;
    expect(level1.children).toHaveLength(1);
    const level2 = level1.children[0]!;
    expect(level2.children).toHaveLength(1);
    const level3 = level2.children[0]!;
    expect(level3.children).toHaveLength(1);
    expect(level3.children[0]?.name).toBe('Level4');
  });

  it('sorts categories by name for stable ordering', () => {
    const categories: Category[] = [
      createCategory('3', 'Zebra', null),
      createCategory('1', 'Alpha', null),
      createCategory('2', 'Beta', null),
    ];

    const tree = buildCategoryTree(categories);

    expect(tree).toHaveLength(3);
    expect(tree[0]?.name).toBe('Alpha');
    expect(tree[1]?.name).toBe('Beta');
    expect(tree[2]?.name).toBe('Zebra');
  });

  it('sorts children by name for stable ordering', () => {
    const categories: Category[] = [
      createCategory('1', 'Parent', null),
      createCategory('3', 'Zebra Child', '1'),
      createCategory('2', 'Alpha Child', '1'),
      createCategory('4', 'Beta Child', '1'),
    ];

    const tree = buildCategoryTree(categories);

    const parent = tree[0]!;
    expect(parent.children).toHaveLength(3);
    expect(parent.children[0]?.name).toBe('Alpha Child');
    expect(parent.children[1]?.name).toBe('Beta Child');
    expect(parent.children[2]?.name).toBe('Zebra Child');
  });

  it('handles orphaned categories defensively by placing at root', () => {
    // Category with parentCategoryId that doesn't exist
    const categories: Category[] = [
      createCategory('1', 'Valid Root', null),
      createCategory('2', 'Orphan', '99999999-9999-9999-9999-999999999999'), // Non-existent parent
    ];

    const tree = buildCategoryTree(categories);

    // Should have 2 root nodes (valid root + orphan)
    expect(tree).toHaveLength(2);
    expect(tree[0]?.name).toBe('Orphan'); // Sorted by name
    expect(tree[1]?.name).toBe('Valid Root');
    expect(tree[0]?.children).toHaveLength(0);
  });

  it('handles empty array', () => {
    const categories: Category[] = [];
    const tree = buildCategoryTree(categories);
    expect(tree).toHaveLength(0);
  });

  it('preserves all category properties', () => {
    const categories: Category[] = [
      createCategory('1', 'Test', null, 'Test description'),
    ];

    const tree = buildCategoryTree(categories);

    expect(tree[0]?.id).toBe('1');
    expect(tree[0]?.projectId).toBe(projectId);
    expect(tree[0]?.parentCategoryId).toBeNull();
    expect(tree[0]?.name).toBe('Test');
    expect(tree[0]?.description).toBe('Test description');
    expect(tree[0]?.createdAt).toBeDefined();
    expect(tree[0]?.updatedAt).toBeDefined();
  });
});

