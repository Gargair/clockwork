import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { startTimer, stopTimer, getActiveTimer, listEntries, type ListEntriesParams } from './time';
import { ApiError } from './http';
import type { TimeEntry } from '../types';

// Mock fetch globally
const fetchMock = vi.fn();
globalThis.fetch = fetchMock;

// Mock API_BASE_URL
vi.mock('./config', () => ({
  API_BASE_URL: 'http://localhost:8080',
}));

describe('time API', { concurrent: false }, () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  const createTimeEntry = (startedAt: string, stoppedAt: string | null = null): TimeEntry => ({
    id: '11111111-1111-4111-8111-111111111111',
    categoryId: '22222222-2222-4222-8222-222222222222',
    startedAt,
    stoppedAt,
    durationSeconds: stoppedAt ? 3600 : null,
    createdAt: startedAt,
    updatedAt: startedAt,
  });

  describe('startTimer', () => {
    it('sends POST request with correct body and validates response', async () => {
      const categoryId = '22222222-2222-4222-8222-222222222222';
      const timeEntry = createTimeEntry('2025-11-02T10:00:00Z', null);

      fetchMock.mockResolvedValueOnce({
        ok: true,
        status: 201,
        text: async () => JSON.stringify(timeEntry),
        headers: new Headers({ 'Content-Type': 'application/json' }),
      } as Response);

      const result = await startTimer(categoryId);

      expect(fetchMock).toHaveBeenCalledWith('http://localhost:8080/api/time/start', {
        method: 'POST',
        headers: {
          Accept: 'application/json',
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ categoryId }),
        signal: undefined,
      });

      expect(result).toEqual(timeEntry);
    });

    it('throws ApiError with code and requestId on 400 invalid_id', async () => {
      const categoryId = 'invalid-id';

      fetchMock.mockResolvedValueOnce({
        ok: false,
        status: 400,
        statusText: 'Bad Request',
        text: async () =>
          JSON.stringify({
            code: 'invalid_id',
            message: 'Invalid category ID',
            requestId: 'req-123',
          }),
        headers: new Headers({ 'Content-Type': 'application/json' }),
      } as Response);

      try {
        await startTimer(categoryId);
        expect.fail('Should have thrown ApiError');
      } catch (err) {
        expect(err).toBeInstanceOf(ApiError);
        const apiError = err as ApiError;
        expect(apiError.status).toBe(400);
        expect(apiError.code).toBe('invalid_id');
        expect(apiError.requestId).toBe('req-123');
      }
    });
  });

  describe('stopTimer', () => {
    it('sends POST request and validates response', async () => {
      const timeEntry = createTimeEntry('2025-11-02T10:00:00Z', '2025-11-02T11:00:00Z');

      fetchMock.mockResolvedValueOnce({
        ok: true,
        status: 200,
        text: async () => JSON.stringify(timeEntry),
        headers: new Headers({ 'Content-Type': 'application/json' }),
      } as Response);

      const result = await stopTimer();

      expect(fetchMock).toHaveBeenCalledWith('http://localhost:8080/api/time/stop', {
        method: 'POST',
        headers: {
          Accept: 'application/json',
        },
        body: undefined,
        signal: undefined,
      });

      expect(result).toEqual(timeEntry);
    });

    it('throws ApiError with no_active_timer code on 409', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: false,
        status: 409,
        statusText: 'Conflict',
        text: async () =>
          JSON.stringify({
            code: 'no_active_timer',
            message: 'No active timer to stop',
            requestId: 'req-456',
          }),
        headers: new Headers({ 'Content-Type': 'application/json' }),
      } as Response);

      try {
        await stopTimer();
        expect.fail('Should have thrown ApiError');
      } catch (err) {
        expect(err).toBeInstanceOf(ApiError);
        const apiError = err as ApiError;
        expect(apiError.status).toBe(409);
        expect(apiError.code).toBe('no_active_timer');
        expect(apiError.requestId).toBe('req-456');
      }
    });
  });

  describe('getActiveTimer', () => {
    it('sends GET request and returns TimeEntry when active', async () => {
      const timeEntry = createTimeEntry('2025-11-02T10:00:00Z', null);

      fetchMock.mockResolvedValueOnce({
        ok: true,
        status: 200,
        text: async () => JSON.stringify(timeEntry),
        headers: new Headers({ 'Content-Type': 'application/json' }),
      } as Response);

      const result = await getActiveTimer();

      expect(fetchMock).toHaveBeenCalledWith('http://localhost:8080/api/time/active', {
        method: 'GET',
        headers: {
          Accept: 'application/json',
        },
        body: undefined,
        signal: undefined,
      });

      expect(result).toEqual(timeEntry);
    });

    it('sends GET request and returns null when no active timer', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: true,
        status: 200,
        text: async () => 'null',
        headers: new Headers({ 'Content-Type': 'application/json' }),
      } as Response);

      const result = await getActiveTimer();

      expect(fetchMock).toHaveBeenCalledWith('http://localhost:8080/api/time/active', {
        method: 'GET',
        headers: {
          Accept: 'application/json',
        },
        body: undefined,
        signal: undefined,
      });

      expect(result).toBeNull();
    });
  });

  describe('listEntries', () => {
    it('sends GET request with categoryId query param', async () => {
      const entries: TimeEntry[] = [
        createTimeEntry('2025-11-02T10:00:00Z', '2025-11-02T11:00:00Z'),
      ];

      fetchMock.mockResolvedValueOnce({
        ok: true,
        status: 200,
        text: async () => JSON.stringify(entries),
        headers: new Headers({ 'Content-Type': 'application/json' }),
      } as Response);

      const params: ListEntriesParams = {
        categoryId: '22222222-2222-4222-8222-222222222222',
      };

      const result = await listEntries(params);

      expect(fetchMock).toHaveBeenCalledWith(
        'http://localhost:8080/api/time/entries?categoryId=22222222-2222-4222-8222-222222222222',
        {
          method: 'GET',
          headers: {
            Accept: 'application/json',
          },
          body: undefined,
          signal: undefined,
        },
      );

      expect(result).toEqual(entries);
    });

    it('sends GET request with categoryId, from, and to query params', async () => {
      const entries: TimeEntry[] = [
        createTimeEntry('2025-11-02T10:00:00Z', '2025-11-02T11:00:00Z'),
      ];

      fetchMock.mockResolvedValueOnce({
        ok: true,
        status: 200,
        text: async () => JSON.stringify(entries),
        headers: new Headers({ 'Content-Type': 'application/json' }),
      } as Response);

      const params: ListEntriesParams = {
        categoryId: '22222222-2222-4222-8222-222222222222',
        from: '2025-11-01T00:00:00Z',
        to: '2025-11-30T23:59:59Z',
      };

      const result = await listEntries(params);

      const expectedUrl =
        'http://localhost:8080/api/time/entries?categoryId=22222222-2222-4222-8222-222222222222&from=2025-11-01T00%3A00%3A00Z&to=2025-11-30T23%3A59%3A59Z';
      expect(fetchMock).toHaveBeenCalledWith(expectedUrl, {
        method: 'GET',
        headers: {
          Accept: 'application/json',
        },
        body: undefined,
        signal: undefined,
      });

      expect(result).toEqual(entries);
    });

    it('throws ApiError with invalid_id code on 400', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: false,
        status: 400,
        statusText: 'Bad Request',
        text: async () =>
          JSON.stringify({
            code: 'invalid_id',
            message: 'Invalid category ID',
            requestId: 'req-789',
          }),
        headers: new Headers({ 'Content-Type': 'application/json' }),
      } as Response);

      try {
        await listEntries({ categoryId: 'invalid-id' });
        expect.fail('Should have thrown ApiError');
      } catch (err) {
        expect(err).toBeInstanceOf(ApiError);
        const apiError = err as ApiError;
        expect(apiError.status).toBe(400);
        expect(apiError.code).toBe('invalid_id');
        expect(apiError.requestId).toBe('req-789');
      }
    });

    it('throws ApiError with invalid_time_range code on 400', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: false,
        status: 400,
        statusText: 'Bad Request',
        text: async () =>
          JSON.stringify({
            code: 'invalid_time_range',
            message: 'Invalid time range',
            requestId: 'req-999',
          }),
        headers: new Headers({ 'Content-Type': 'application/json' }),
      } as Response);

      const params: ListEntriesParams = {
        categoryId: '22222222-2222-4222-8222-222222222222',
        from: '2025-11-30T00:00:00Z',
        to: '2025-11-01T00:00:00Z', // Invalid: from > to
      };

      try {
        await listEntries(params);
        expect.fail('Should have thrown ApiError');
      } catch (err) {
        expect(err).toBeInstanceOf(ApiError);
        const apiError = err as ApiError;
        expect(apiError.status).toBe(400);
        expect(apiError.code).toBe('invalid_time_range');
        expect(apiError.requestId).toBe('req-999');
      }
    });
  });
});

