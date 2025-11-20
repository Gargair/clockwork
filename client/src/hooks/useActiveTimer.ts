import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import type { TimeEntry } from '../types';
import { startTimer, stopTimer, getActiveTimer } from '../api/time';
import { ApiError } from '../api/http';

export type UseActiveTimerStatus = 'idle' | 'loading' | 'success' | 'error';

export interface UseActiveTimerError {
  message: string;
  code?: string;
  requestId?: string;
}

export interface UseActiveTimerResult {
  status: UseActiveTimerStatus;
  activeTimer: TimeEntry | null;
  error: UseActiveTimerError | null;
  refresh: () => Promise<void>;
  start: (categoryId: string) => Promise<void>;
  stop: () => Promise<void>;
}

export interface UseActiveTimerOptions {
  refreshInterval?: number;
}

export function useActiveTimer(options?: UseActiveTimerOptions): UseActiveTimerResult {
  const refreshInterval = options?.refreshInterval ?? 5000;
  const [status, setStatus] = useState<UseActiveTimerStatus>('idle');
  const [activeTimer, setActiveTimer] = useState<TimeEntry | null>(null);
  const [error, setError] = useState<UseActiveTimerError | null>(null);
  const mountedRef = useRef<boolean>(true);
  const pollingIntervalRef = useRef<number | null>(null);

  const handleError = useCallback((err: unknown): UseActiveTimerError => {
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
      const data = await getActiveTimer();
      if (!mountedRef.current) return;
      setActiveTimer(data);
      setStatus('success');
    } catch (err) {
      if (!mountedRef.current) return;
      setError(handleError(err));
      setStatus('error');
    }
  }, [handleError]);

  const start = useCallback(
    async (categoryId: string): Promise<void> => {
      setError(null);
      try {
        await startTimer(categoryId);
        if (!mountedRef.current) return;
        // Refresh to reconcile state (server enforces single active timer)
        await refresh();
      } catch (err) {
        if (!mountedRef.current) return;
        setError(handleError(err));
        setStatus('error');
      }
    },
    [handleError, refresh],
  );

  const stop = useCallback(async (): Promise<void> => {
    setError(null);
    try {
      await stopTimer();
      if (!mountedRef.current) return;
      // Refresh to clear active state
      await refresh();
    } catch (err) {
      if (!mountedRef.current) return;
      setError(handleError(err));
      setStatus('error');
    }
  }, [handleError, refresh]);

  // Load active timer on mount
  useEffect(() => {
    mountedRef.current = true;
    refresh();
    return () => {
      mountedRef.current = false;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Polling: refresh at the specified interval when component is mounted
  // Always poll, regardless of whether there's an active timer
  useEffect(() => {
    // Always poll at the specified interval
    pollingIntervalRef.current = globalThis.setInterval(() => {
      if (mountedRef.current) {
        refresh();
      }
    }, refreshInterval);

    return () => {
      if (pollingIntervalRef.current !== null) {
        clearInterval(pollingIntervalRef.current);
        pollingIntervalRef.current = null;
      }
    };
  }, [refreshInterval, refresh]);

  return useMemo<UseActiveTimerResult>(
    () => ({
      status,
      activeTimer,
      error,
      refresh,
      start,
      stop,
    }),
    [status, activeTimer, error, refresh, start, stop],
  );
}

