import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import type { Category } from '../types';
import {
  listCategories,
  createCategory as apiCreateCategory,
  updateCategory as apiUpdateCategory,
  deleteCategory as apiDeleteCategory,
  type CreateCategoryInput,
  type UpdateCategoryInput,
} from '../api/categories';
import { ApiError } from '../api/http';

export type UseCategoriesStatus = 'idle' | 'loading' | 'success' | 'error';

export interface UseCategoriesError {
  message: string;
  code?: string;
  requestId?: string;
}

export interface CategoryNode extends Category {
  children: CategoryNode[];
}

export interface UseCategoriesResult {
  status: UseCategoriesStatus;
  categories: Category[];
  error: UseCategoriesError | null;
  refresh: () => Promise<void>;
  create: (input: CreateCategoryInput) => Promise<void>;
  update: (categoryId: string, input: UpdateCategoryInput) => Promise<void>;
  remove: (categoryId: string) => Promise<void>;
}

export function useCategories(projectId: string): UseCategoriesResult {
  const [status, setStatus] = useState<UseCategoriesStatus>('idle');
  const [categories, setCategories] = useState<Category[]>([]);
  const [error, setError] = useState<UseCategoriesError | null>(null);
  const mountedRef = useRef<boolean>(true);

  useEffect(() => {
    mountedRef.current = true;
    refresh();
    return () => {
      mountedRef.current = false;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [projectId]);

  const handleError = useCallback((err: unknown): UseCategoriesError => {
    if (err instanceof ApiError) {
      return { message: err.message, code: err.code, requestId: err.requestId };
    }
    const message: string = err instanceof Error ? err.message : 'Unknown error';
    return { message };
  }, []);

  const refresh = useCallback(async (): Promise<void> => {
    setStatus('loading');
    setError(null);
    try {
      const data = await listCategories(projectId);
      if (!mountedRef.current) return;
      setCategories(data);
      setStatus('success');
    } catch (err) {
      if (!mountedRef.current) return;
      setError(handleError(err));
      setStatus('error');
    }
  }, [projectId, handleError]);

  const create = useCallback(
    async (input: CreateCategoryInput): Promise<void> => {
      setError(null);
      try {
        await apiCreateCategory(projectId, input);
        if (!mountedRef.current) return;
        await refresh();
      } catch (err) {
        if (!mountedRef.current) return;
        setError(handleError(err));
        setStatus('error');
      }
    },
    [projectId, handleError, refresh],
  );

  const update = useCallback(
    async (categoryId: string, input: UpdateCategoryInput): Promise<void> => {
      setError(null);
      try {
        await apiUpdateCategory(projectId, categoryId, input);
        if (!mountedRef.current) return;
        await refresh();
      } catch (err) {
        if (!mountedRef.current) return;
        setError(handleError(err));
        setStatus('error');
      }
    },
    [projectId, handleError, refresh],
  );

  const remove = useCallback(
    async (categoryId: string): Promise<void> => {
      setError(null);
      try {
        await apiDeleteCategory(projectId, categoryId);
        if (!mountedRef.current) return;
        await refresh();
      } catch (err) {
        if (!mountedRef.current) return;
        setError(handleError(err));
        setStatus('error');
      }
    },
    [projectId, handleError, refresh],
  );

  return useMemo<UseCategoriesResult>(
    () => ({
      status,
      categories,
      error,
      refresh,
      create,
      update,
      remove,
    }),
    [status, categories, error, refresh, create, update, remove],
  );
}

/**
 * Converts a flat list of categories into a tree structure.
 * Categories are sorted by name for stable ordering.
 * Orphaned categories (with invalid parentCategoryId) are handled gracefully
 * by placing them at the root level.
 */
export function buildCategoryTree(categories: Category[]): CategoryNode[] {
  // Create a map for quick lookup
  const categoryMap = new Map<string, CategoryNode>();
  const rootNodes: CategoryNode[] = [];

  // First pass: create all nodes
  const sortedCategories = [...categories].sort((a, b) => a.name.localeCompare(b.name));
  for (const category of sortedCategories) {
    categoryMap.set(category.id, {
      ...category,
      children: [],
    });
  }

  // Second pass: build the tree structure
  for (const category of sortedCategories) {
    const node = categoryMap.get(category.id)!;
    const parentId = category.parentCategoryId;

    if (parentId === null) {
      // Root level category
      rootNodes.push(node);
    } else {
      // Child category
      const parent = categoryMap.get(parentId);
      if (parent) {
        parent.children.push(node);
        // Sort children by name for stable ordering
        parent.children.sort((a, b) => a.name.localeCompare(b.name));
      } else {
        // Orphaned category (parent doesn't exist) - place at root level
        rootNodes.push(node);
      }
    }
  }

  return rootNodes;
}

