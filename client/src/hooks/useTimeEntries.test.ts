import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { useTimeEntries } from './useTimeEntries';
import { ApiError } from '../api/http';
import type { TimeEntry } from '../types';

// Mock the API layer
const listEntriesMock = vi.fn();

vi.mock('../api/time', () => ({
  listEntries: (...args: unknown[]) => listEntriesMock(...args),
}));

describe('useTimeEntries hook', { concurrent: false }, () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    // Cleanup handled by renderHook
  });

  const createTimeEntry = (
    id: string,
    categoryId: string,
    startedAt: string,
  ): TimeEntry => ({
    id,
    categoryId,
    startedAt,
    stoppedAt: new Date(Date.parse(startedAt) + 3600000).toISOString(),
    durationSeconds: 3600,
    createdAt: startedAt,
    updatedAt: startedAt,
  });

  it('load() calls API with correct params', async () => {
    const entries: TimeEntry[] = [
      createTimeEntry('11111111-1111-1111-1111-111111111111', 'cat-1', '2025-11-02T10:00:00Z'),
    ];
    listEntriesMock.mockResolvedValue(entries);

    const { result } = renderHook(() => useTimeEntries());

    const params = {
      categoryId: 'cat-1',
      from: '2025-11-01T00:00:00Z',
      to: '2025-11-30T23:59:59Z',
    };

    await result.current.load(params);

    expect(listEntriesMock).toHaveBeenCalledWith(params);
    await waitFor(() => {
      expect(result.current.status).toBe('success');
    });
    expect(result.current.entries).toEqual(entries);
  });

  it('handles query param validation errors', async () => {
    const error = new ApiError('invalid_id: Invalid category ID', 400, {
      code: 'invalid_id',
      requestId: 'req-456',
    });
    listEntriesMock.mockRejectedValue(error);

    const { result } = renderHook(() => useTimeEntries());

    await result.current.load({ categoryId: 'invalid-id' });

    await waitFor(() => {
      expect(result.current.status).toBe('error');
    });

    expect(result.current.error).not.toBeNull();
    expect(result.current.error?.code).toBe('invalid_id');
    expect(result.current.error?.requestId).toBe('req-456');
  });

  it('error handling captures ApiError details', async () => {
    const error = new ApiError('invalid_time_range: Invalid range', 400, {
      code: 'invalid_time_range',
      requestId: 'req-789',
    });
    listEntriesMock.mockRejectedValue(error);

    const { result } = renderHook(() => useTimeEntries());

    await result.current.load({
      categoryId: 'cat-1',
      from: '2025-11-30T00:00:00Z',
      to: '2025-11-01T00:00:00Z', // Invalid: from > to
    });

    await waitFor(() => {
      expect(result.current.status).toBe('error');
    });

    expect(result.current.error).not.toBeNull();
    expect(result.current.error?.code).toBe('invalid_time_range');
    expect(result.current.error?.requestId).toBe('req-789');
    expect(result.current.error?.message).toContain('invalid_time_range');
  });

  it('refresh() reloads with current params', async () => {
    const entries: TimeEntry[] = [
      createTimeEntry('11111111-1111-1111-1111-111111111111', 'cat-1', '2025-11-02T10:00:00Z'),
    ];
    listEntriesMock.mockResolvedValue(entries);

    const { result } = renderHook(() => useTimeEntries());

    const params = { categoryId: 'cat-1' };
    await result.current.load(params);

    await waitFor(() => {
      expect(result.current.status).toBe('success');
    });

    const initialCallCount = listEntriesMock.mock.calls.length;

    await result.current.refresh();

    expect(listEntriesMock.mock.calls.length).toBeGreaterThan(initialCallCount);
    expect(listEntriesMock).toHaveBeenLastCalledWith(params);
  });

  it('loads entries when initial params provided', async () => {
    const entries: TimeEntry[] = [
      createTimeEntry('11111111-1111-1111-1111-111111111111', 'cat-1', '2025-11-02T10:00:00Z'),
    ];
    listEntriesMock.mockResolvedValue(entries);

    const { result } = renderHook(() =>
      useTimeEntries({ categoryId: 'cat-1' }),
    );

    expect(result.current.status).toBe('loading');
    expect(listEntriesMock).toHaveBeenCalledWith({ categoryId: 'cat-1' });

    await waitFor(() => {
      expect(result.current.status).toBe('success');
    });

    expect(result.current.entries).toEqual(entries);
  });

  it('refresh() does nothing when no params set', async () => {
    const { result } = renderHook(() => useTimeEntries());

    await result.current.refresh();

    // Should not call API when no params
    expect(listEntriesMock).not.toHaveBeenCalled();
  });
});

