import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { useActiveTimer } from './useActiveTimer';
import { ApiError } from '../api/http';
import type { TimeEntry } from '../types';

// Mock the API layer
const startTimerMock = vi.fn();
const stopTimerMock = vi.fn();
const getActiveTimerMock = vi.fn();

vi.mock('../api/time', () => ({
  startTimer: (...args: unknown[]) => startTimerMock(...args),
  stopTimer: (...args: unknown[]) => stopTimerMock(...args),
  getActiveTimer: (...args: unknown[]) => getActiveTimerMock(...args),
}));

describe('useActiveTimer hook', { concurrent: false }, () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    // Cleanup handled by renderHook
  });

  const createTimeEntry = (startedAt: string, stoppedAt: string | null = null): TimeEntry => ({
    id: '11111111-1111-1111-1111-111111111111',
    categoryId: '22222222-2222-2222-2222-222222222222',
    startedAt,
    stoppedAt,
    durationSeconds: stoppedAt ? 3600 : null,
    createdAt: startedAt,
    updatedAt: startedAt,
  });

  it('loads active timer on mount', async () => {
    const activeTimer = createTimeEntry(new Date().toISOString(), null);
    getActiveTimerMock.mockResolvedValue(activeTimer);

    const { result } = renderHook(() => useActiveTimer());

    expect(result.current.status).toBe('loading');
    expect(getActiveTimerMock).toHaveBeenCalled();

    await waitFor(() => {
      expect(result.current.status).toBe('success');
    });

    expect(result.current.activeTimer).toEqual(activeTimer);
  });

  it('start() calls API and refreshes', async () => {
    const categoryId = '22222222-2222-2222-2222-222222222222';
    const newTimer = createTimeEntry(new Date().toISOString(), null);
    startTimerMock.mockResolvedValue(newTimer);
    getActiveTimerMock.mockResolvedValue(newTimer);

    const { result } = renderHook(() => useActiveTimer());

    await waitFor(() => {
      expect(result.current.status).not.toBe('loading');
    });

    await result.current.start(categoryId);

    expect(startTimerMock).toHaveBeenCalledWith(categoryId);
    expect(getActiveTimerMock).toHaveBeenCalled();
    expect(result.current.activeTimer).toEqual(newTimer);
  });

  it('stop() calls API and refreshes', async () => {
    const activeTimer = createTimeEntry(new Date().toISOString(), null);
    const stoppedTimer = createTimeEntry(
      activeTimer.startedAt,
      new Date().toISOString()
    );
    getActiveTimerMock.mockResolvedValueOnce(activeTimer);
    stopTimerMock.mockResolvedValue(stoppedTimer);
    getActiveTimerMock.mockResolvedValueOnce(null); // After stop, no active timer

    const { result } = renderHook(() => useActiveTimer());

    await waitFor(() => {
      expect(result.current.activeTimer).toEqual(activeTimer);
    });

    await result.current.stop();

    expect(stopTimerMock).toHaveBeenCalled();
    expect(getActiveTimerMock).toHaveBeenCalled();
    await waitFor(() => {
      expect(result.current.activeTimer).toBeNull();
    });
  });

  it('error handling captures ApiError details', async () => {
    const error = new ApiError('no_active_timer: No active timer', 409, {
      code: 'no_active_timer',
      requestId: 'req-123',
    });
    getActiveTimerMock.mockRejectedValue(error);

    const { result } = renderHook(() => useActiveTimer());

    await waitFor(() => {
      expect(result.current.status).toBe('error');
    });

    expect(result.current.error).not.toBeNull();
    expect(result.current.error?.code).toBe('no_active_timer');
    expect(result.current.error?.requestId).toBe('req-123');
    expect(result.current.error?.message).toContain('no_active_timer');
  });

  it('polls at specified interval', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    const activeTimer = createTimeEntry(new Date().toISOString(), null);
    getActiveTimerMock.mockResolvedValue(activeTimer);

    const { result, unmount } = renderHook(() => useActiveTimer({ refreshInterval: 5000 }));

    await waitFor(() => {
      expect(result.current.status).toBe('success');
    });

    const initialCallCount = getActiveTimerMock.mock.calls.length;

    // Advance time by 5 seconds - this should trigger the polling interval
    vi.advanceTimersByTime(5000);

    await waitFor(() => {
      expect(getActiveTimerMock.mock.calls.length).toBeGreaterThan(initialCallCount);
    }, { timeout: 1000 });

    unmount();
    vi.useRealTimers();
  });

  it('refresh() updates active timer', async () => {
    const activeTimer = createTimeEntry(new Date().toISOString(), null);
    getActiveTimerMock.mockResolvedValue(activeTimer);

    const { result } = renderHook(() => useActiveTimer());

    await waitFor(() => {
      expect(result.current.status).toBe('success');
    });

    const initialCallCount = getActiveTimerMock.mock.calls.length;

    await result.current.refresh();

    expect(getActiveTimerMock.mock.calls.length).toBeGreaterThan(initialCallCount);
    expect(result.current.activeTimer).toEqual(activeTimer);
  });
});

