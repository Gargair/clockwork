import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import {
  listCategories,
  createCategory,
  getCategory,
  updateCategory,
  deleteCategory,
  type CreateCategoryInput,
  type UpdateCategoryInput,
} from './categories';
import { ApiError } from './http';
import type { Category } from '../types';

// Mock fetch globally
const fetchMock = vi.fn();
globalThis.fetch = fetchMock;

// Mock API_BASE_URL
vi.mock('./config', () => ({
  API_BASE_URL: 'http://localhost:8080',
}));

describe('categories API', { concurrent: false }, () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  const projectId = '11111111-1111-4111-8111-111111111111';

  const validUuid2 = '22222222-2222-4222-8222-222222222222';
  const validUuid3 = '33333333-3333-4333-8333-333333333333';
  const validUuid4 = '44444444-4444-4444-8444-444444444444';
  const validUuid5 = '55555555-5555-4555-8555-555555555555';
  const validUuid6 = '66666666-6666-4666-8666-666666666666';
  const validUuid8 = '88888888-8888-4888-8888-888888888888';
  const validUuid9 = '99999999-9999-4999-8999-999999999999';
  const validUuidA = 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa';

  const createCategoryData = (
    id: string,
    name: string,
    parentCategoryId: string | null = null,
    description?: string,
  ): Category => ({
    id,
    projectId,
    parentCategoryId,
    name,
    description,
    createdAt: '2025-11-02T10:00:00Z',
    updatedAt: '2025-11-02T10:00:00Z',
  });

  describe('listCategories', () => {
    it('sends GET request with projectId in path and validates response', async () => {
      const categories: Category[] = [
        createCategoryData(validUuid2, 'Frontend'),
        createCategoryData(validUuid3, 'Backend'),
      ];

      fetchMock.mockResolvedValueOnce({
        ok: true,
        status: 200,
        text: async () => JSON.stringify(categories),
        headers: new Headers({ 'Content-Type': 'application/json' }),
      } as Response);

      const result = await listCategories(projectId);

      expect(fetchMock).toHaveBeenCalledWith(
        `http://localhost:8080/api/projects/${projectId}/categories`,
        {
          method: 'GET',
          headers: {
            Accept: 'application/json',
          },
          body: undefined,
          signal: undefined,
        },
      );

      expect(result).toEqual(categories);
    });

    it('throws ApiError with invalid_id code on 400', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: false,
        status: 400,
        statusText: 'Bad Request',
        text: async () =>
          JSON.stringify({
            code: 'invalid_id',
            message: 'Invalid project ID',
            requestId: 'req-cat-400',
          }),
        headers: new Headers({ 'Content-Type': 'application/json' }),
      } as Response);

      try {
        await listCategories('invalid-id');
        expect.fail('Should have thrown ApiError');
      } catch (err) {
        expect(err).toBeInstanceOf(ApiError);
        const apiError = err as ApiError;
        expect(apiError.status).toBe(400);
        expect(apiError.code).toBe('invalid_id');
      }
    });
  });

  describe('createCategory', () => {
    it('sends POST request with correct body and validates response', async () => {
      const input: CreateCategoryInput = {
        name: 'New Category',
        description: 'Category description',
        parentCategoryId: null,
      };
      const category = createCategoryData(validUuid4, input.name, null, input.description ?? undefined);

      fetchMock.mockResolvedValueOnce({
        ok: true,
        status: 201,
        text: async () => JSON.stringify(category),
        headers: new Headers({ 'Content-Type': 'application/json' }),
      } as Response);

      const result = await createCategory(projectId, input);

      expect(fetchMock).toHaveBeenCalledWith(
        `http://localhost:8080/api/projects/${projectId}/categories`,
        {
          method: 'POST',
          headers: {
            Accept: 'application/json',
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(input),
          signal: undefined,
        },
      );

      expect(result).toEqual(category);
    });

    it('sends POST request with parentCategoryId', async () => {
      const parentId = validUuid5;
      const input: CreateCategoryInput = {
        name: 'Child Category',
        parentCategoryId: parentId,
      };
      const category = createCategoryData(validUuid6, input.name, parentId);

      fetchMock.mockResolvedValueOnce({
        ok: true,
        status: 201,
        text: async () => JSON.stringify(category),
        headers: new Headers({ 'Content-Type': 'application/json' }),
      } as Response);

      const result = await createCategory(projectId, input);

      expect(fetchMock).toHaveBeenCalledWith(
        `http://localhost:8080/api/projects/${projectId}/categories`,
        {
          method: 'POST',
          headers: {
            Accept: 'application/json',
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(input),
          signal: undefined,
        },
      );

      expect(result).toEqual(category);
    });

    it('throws ApiError with category_cycle code on 409', async () => {
      const input: CreateCategoryInput = {
        name: 'Category',
        parentCategoryId: '77777777-7777-7777-7777-777777777777',
      };

      fetchMock.mockResolvedValueOnce({
        ok: false,
        status: 409,
        statusText: 'Conflict',
        text: async () =>
          JSON.stringify({
            code: 'category_cycle',
            message: 'Category cycle detected',
            requestId: 'req-cycle',
          }),
        headers: new Headers({ 'Content-Type': 'application/json' }),
      } as Response);

      try {
        await createCategory(projectId, input);
        expect.fail('Should have thrown ApiError');
      } catch (err) {
        expect(err).toBeInstanceOf(ApiError);
        const apiError = err as ApiError;
        expect(apiError.status).toBe(409);
        expect(apiError.code).toBe('category_cycle');
      }
    });
  });

  describe('getCategory', () => {
    it('sends GET request with projectId and categoryId in path', async () => {
      const categoryId = validUuid8;
      const category = createCategoryData(categoryId, 'Test Category');

      fetchMock.mockResolvedValueOnce({
        ok: true,
        status: 200,
        text: async () => JSON.stringify(category),
        headers: new Headers({ 'Content-Type': 'application/json' }),
      } as Response);

      const result = await getCategory(projectId, categoryId);

      expect(fetchMock).toHaveBeenCalledWith(
        `http://localhost:8080/api/projects/${projectId}/categories/${categoryId}`,
        {
          method: 'GET',
          headers: {
            Accept: 'application/json',
          },
          body: undefined,
          signal: undefined,
        },
      );

      expect(result).toEqual(category);
    });
  });

  describe('updateCategory', () => {
    it('sends PATCH request with correct body and validates response', async () => {
      const categoryId = validUuid9;
      const input: UpdateCategoryInput = {
        name: 'Updated Category',
        description: 'Updated description',
        parentCategoryId: null,
      };
      const category = createCategoryData(categoryId, input.name, null, input.description ?? undefined);

      fetchMock.mockResolvedValueOnce({
        ok: true,
        status: 200,
        text: async () => JSON.stringify(category),
        headers: new Headers({ 'Content-Type': 'application/json' }),
      } as Response);

      const result = await updateCategory(projectId, categoryId, input);

      expect(fetchMock).toHaveBeenCalledWith(
        `http://localhost:8080/api/projects/${projectId}/categories/${categoryId}`,
        {
          method: 'PATCH',
          headers: {
            Accept: 'application/json',
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(input),
          signal: undefined,
        },
      );

      expect(result).toEqual(category);
    });
  });

  describe('deleteCategory', () => {
    it('sends DELETE request and handles 204 No Content', async () => {
      const categoryId = validUuidA;

      fetchMock.mockResolvedValueOnce({
        ok: true,
        status: 204,
        text: async () => '',
        headers: new Headers(),
      } as Response);

      await deleteCategory(projectId, categoryId);

      expect(fetchMock).toHaveBeenCalledWith(
        `http://localhost:8080/api/projects/${projectId}/categories/${categoryId}`,
        {
          method: 'DELETE',
          headers: {
            Accept: 'application/json',
          },
          body: undefined,
          signal: undefined,
        },
      );
    });
  });
});

