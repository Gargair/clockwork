import type { Category } from '../types';

export interface CreateCategoryInput {
  name: string;
  description?: string | null;
  parentCategoryId?: string | null;
}

export async function listCategories(projectId: string): Promise<Category[]> {
  throw new Error('Not implemented');
}

export async function createCategory(projectId: string, input: CreateCategoryInput): Promise<Category> {
  throw new Error('Not implemented');
}

export async function getCategory(projectId: string, categoryId: string): Promise<Category> {
  throw new Error('Not implemented');
}

export interface UpdateCategoryInput {
  name: string;
  description?: string | null;
  parentCategoryId?: string | null;
}

export async function updateCategory(
  projectId: string,
  categoryId: string,
  input: UpdateCategoryInput,
): Promise<Category> {
  throw new Error('Not implemented');
}

export async function deleteCategory(projectId: string, categoryId: string): Promise<void> {
  throw new Error('Not implemented');
}


