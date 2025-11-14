import type { Category } from '../types';

export interface CreateCategoryInput {
  name: string;
  description?: string | null;
  parentCategoryId?: string | null;
}

export async function listCategories(_projectId: string): Promise<Category[]> {
  throw new Error('Not implemented');
}

export async function createCategory(_projectId: string, _input: CreateCategoryInput): Promise<Category> {
  throw new Error('Not implemented');
}

export async function getCategory(_projectId: string, _categoryId: string): Promise<Category> {
  throw new Error('Not implemented');
}

export interface UpdateCategoryInput {
  name: string;
  description?: string | null;
  parentCategoryId?: string | null;
}

export async function updateCategory(
  _projectId: string,
  _categoryId: string,
  _input: UpdateCategoryInput,
): Promise<Category> {
  throw new Error('Not implemented');
}

export async function deleteCategory(_projectId: string, _categoryId: string): Promise<void> {
  throw new Error('Not implemented');
}


