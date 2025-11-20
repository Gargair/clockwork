import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import {
  listProjects,
  createProject,
  getProject,
  updateProject,
  deleteProject,
  type CreateProjectInput,
  type UpdateProjectInput,
} from './projects';
import { ApiError } from './http';
import type { Project } from '../types';

// Mock fetch globally
const fetchMock = vi.fn();
globalThis.fetch = fetchMock;

// Mock API_BASE_URL
vi.mock('./config', () => ({
  API_BASE_URL: 'http://localhost:8080',
}));

describe('projects API', { concurrent: false }, () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  const createProjectData = (id: string, name: string, description?: string): Project => ({
    id,
    name,
    description,
    createdAt: '2025-11-02T10:00:00Z',
    updatedAt: '2025-11-02T10:00:00Z',
  });

  const validUuid1 = '11111111-1111-4111-8111-111111111111';
  const validUuid2 = '22222222-2222-4222-8222-222222222222';
  const validUuid3 = '33333333-3333-4333-8333-333333333333';
  const validUuid4 = '44444444-4444-4444-8444-444444444444';
  const validUuid5 = '55555555-5555-4555-8555-555555555555';
  const validUuid6 = '66666666-6666-4666-8666-666666666666';
  const validUuid7 = '77777777-7777-4777-8777-777777777777';

  describe('listProjects', () => {
    it('sends GET request and validates response array', async () => {
      const projects: Project[] = [
        createProjectData(validUuid1, 'Project 1', 'Description 1'),
        createProjectData(validUuid2, 'Project 2'),
      ];

      fetchMock.mockResolvedValueOnce({
        ok: true,
        status: 200,
        text: async () => JSON.stringify(projects),
        headers: new Headers({ 'Content-Type': 'application/json' }),
      } as Response);

      const result = await listProjects();

      expect(fetchMock).toHaveBeenCalledWith('http://localhost:8080/api/projects', {
        method: 'GET',
        headers: {
          Accept: 'application/json',
        },
        body: undefined,
        signal: undefined,
      });

      expect(result).toEqual(projects);
    });

    it('throws ApiError on error response', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: false,
        status: 500,
        statusText: 'Internal Server Error',
        text: async () =>
          JSON.stringify({
            code: 'internal',
            message: 'Internal server error',
            requestId: 'req-500',
          }),
        headers: new Headers({ 'Content-Type': 'application/json' }),
      } as Response);

      try {
        await listProjects();
        expect.fail('Should have thrown ApiError');
      } catch (err) {
        expect(err).toBeInstanceOf(ApiError);
        const apiError = err as ApiError;
        expect(apiError.status).toBe(500);
        expect(apiError.code).toBe('internal');
        expect(apiError.requestId).toBe('req-500');
      }
    });
  });

  describe('createProject', () => {
    it('sends POST request with correct body and validates response', async () => {
      const input: CreateProjectInput = {
        name: 'New Project',
        description: 'Project description',
      };
      const project = createProjectData(validUuid3, input.name, input.description ?? undefined);

      fetchMock.mockResolvedValueOnce({
        ok: true,
        status: 201,
        text: async () => JSON.stringify(project),
        headers: new Headers({ 'Content-Type': 'application/json' }),
      } as Response);

      const result = await createProject(input);

      expect(fetchMock).toHaveBeenCalledWith('http://localhost:8080/api/projects', {
        method: 'POST',
        headers: {
          Accept: 'application/json',
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(input),
        signal: undefined,
      });

      expect(result).toEqual(project);
    });

    it('sends POST request with null description', async () => {
      const input: CreateProjectInput = {
        name: 'New Project',
        description: null,
      };
      const project = createProjectData(validUuid4, input.name);

      fetchMock.mockResolvedValueOnce({
        ok: true,
        status: 201,
        text: async () => JSON.stringify(project),
        headers: new Headers({ 'Content-Type': 'application/json' }),
      } as Response);

      const result = await createProject(input);

      expect(fetchMock).toHaveBeenCalledWith('http://localhost:8080/api/projects', {
        method: 'POST',
        headers: {
          Accept: 'application/json',
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(input),
        signal: undefined,
      });

      expect(result).toEqual(project);
    });

    it('throws ApiError with invalid_project_name code on 400', async () => {
      const input: CreateProjectInput = { name: '' };

      fetchMock.mockResolvedValueOnce({
        ok: false,
        status: 400,
        statusText: 'Bad Request',
        text: async () =>
          JSON.stringify({
            code: 'invalid_project_name',
            message: 'Project name is required',
            requestId: 'req-invalid',
          }),
        headers: new Headers({ 'Content-Type': 'application/json' }),
      } as Response);

      try {
        await createProject(input);
        expect.fail('Should have thrown ApiError');
      } catch (err) {
        expect(err).toBeInstanceOf(ApiError);
        const apiError = err as ApiError;
        expect(apiError.status).toBe(400);
        expect(apiError.code).toBe('invalid_project_name');
      }
    });
  });

  describe('getProject', () => {
    it('sends GET request with projectId in path and validates response', async () => {
      const projectId = validUuid5;
      const project = createProjectData(projectId, 'Test Project');

      fetchMock.mockResolvedValueOnce({
        ok: true,
        status: 200,
        text: async () => JSON.stringify(project),
        headers: new Headers({ 'Content-Type': 'application/json' }),
      } as Response);

      const result = await getProject(projectId);

      expect(fetchMock).toHaveBeenCalledWith(`http://localhost:8080/api/projects/${projectId}`, {
        method: 'GET',
        headers: {
          Accept: 'application/json',
        },
        body: undefined,
        signal: undefined,
      });

      expect(result).toEqual(project);
    });

    it('throws ApiError with not_found code on 404', async () => {
      const projectId = '00000000-0000-0000-0000-000000000000';

      fetchMock.mockResolvedValueOnce({
        ok: false,
        status: 404,
        statusText: 'Not Found',
        text: async () =>
          JSON.stringify({
            code: 'not_found',
            message: 'Project not found',
            requestId: 'req-404',
          }),
        headers: new Headers({ 'Content-Type': 'application/json' }),
      } as Response);

      try {
        await getProject(projectId);
        expect.fail('Should have thrown ApiError');
      } catch (err) {
        expect(err).toBeInstanceOf(ApiError);
        const apiError = err as ApiError;
        expect(apiError.status).toBe(404);
        expect(apiError.code).toBe('not_found');
      }
    });
  });

  describe('updateProject', () => {
    it('sends PATCH request with correct body and validates response', async () => {
      const projectId = validUuid6;
      const input: UpdateProjectInput = {
        name: 'Updated Project',
        description: 'Updated description',
      };
      const project = createProjectData(projectId, input.name, input.description ?? undefined);

      fetchMock.mockResolvedValueOnce({
        ok: true,
        status: 200,
        text: async () => JSON.stringify(project),
        headers: new Headers({ 'Content-Type': 'application/json' }),
      } as Response);

      const result = await updateProject(projectId, input);

      expect(fetchMock).toHaveBeenCalledWith(`http://localhost:8080/api/projects/${projectId}`, {
        method: 'PATCH',
        headers: {
          Accept: 'application/json',
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(input),
        signal: undefined,
      });

      expect(result).toEqual(project);
    });
  });

  describe('deleteProject', () => {
    it('sends DELETE request and handles 204 No Content', async () => {
      const projectId = validUuid7;

      fetchMock.mockResolvedValueOnce({
        ok: true,
        status: 204,
        text: async () => '',
        headers: new Headers(),
      } as Response);

      await deleteProject(projectId);

      expect(fetchMock).toHaveBeenCalledWith(`http://localhost:8080/api/projects/${projectId}`, {
        method: 'DELETE',
        headers: {
          Accept: 'application/json',
        },
        body: undefined,
        signal: undefined,
      });
    });

    it('throws ApiError with not_found code on 404', async () => {
      const projectId = '00000000-0000-0000-0000-000000000000';

      fetchMock.mockResolvedValueOnce({
        ok: false,
        status: 404,
        statusText: 'Not Found',
        text: async () =>
          JSON.stringify({
            code: 'not_found',
            message: 'Project not found',
            requestId: 'req-del-404',
          }),
        headers: new Headers({ 'Content-Type': 'application/json' }),
      } as Response);

      await expect(deleteProject(projectId)).rejects.toThrow(ApiError);
    });
  });
});

