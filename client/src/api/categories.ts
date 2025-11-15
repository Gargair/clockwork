import { z } from 'zod';
import { requestJson } from './http';
import { type Category, CategorySchema, CategoryListSchema } from '../types';

export interface CreateCategoryInput {
  name: string;
  description?: string | null;
  parentCategoryId?: string | null;
}

export async function listCategories(projectId: string): Promise<Category[]> {
  return requestJson(
    `/api/projects/${projectId}/categories`,
    { method: 'GET' },
    CategoryListSchema,
  );
}

export async function createCategory(
  projectId: string,
  input: CreateCategoryInput,
): Promise<Category> {
  return requestJson(
    `/api/projects/${projectId}/categories`,
    { method: 'POST', body: input },
    CategorySchema,
  );
}

export async function getCategory(projectId: string, categoryId: string): Promise<Category> {
  return requestJson(
    `/api/projects/${projectId}/categories/${categoryId}`,
    { method: 'GET' },
    CategorySchema,
  );
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
  return requestJson(
    `/api/projects/${projectId}/categories/${categoryId}`,
    { method: 'PATCH', body: input },
    CategorySchema,
  );
}

export async function deleteCategory(projectId: string, categoryId: string): Promise<void> {
  return requestJson(
    `/api/projects/${projectId}/categories/${categoryId}`,
    { method: 'DELETE' },
    z.undefined(),
  );
}
