import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import type { TimeEntry } from '../types';
import { listEntries, type ListEntriesParams } from '../api/time';
import { ApiError } from '../api/http';

export type UseTimeEntriesStatus = 'idle' | 'loading' | 'success' | 'error';

export interface UseTimeEntriesError {
  message: string;
  code?: string;
  requestId?: string;
}

export interface UseTimeEntriesResult {
  status: UseTimeEntriesStatus;
  entries: TimeEntry[];
  error: UseTimeEntriesError | null;
  refresh: () => Promise<void>;
  load: (params: ListEntriesParams) => Promise<void>;
}

export function useTimeEntries(initialParams?: ListEntriesParams): UseTimeEntriesResult {
  const [status, setStatus] = useState<UseTimeEntriesStatus>('idle');
  const [entries, setEntries] = useState<TimeEntry[]>([]);
  const [error, setError] = useState<UseTimeEntriesError | null>(null);
  const [params, setParams] = useState<ListEntriesParams | undefined>(initialParams);
  const mountedRef = useRef<boolean>(true);

  const handleError = useCallback((err: unknown): UseTimeEntriesError => {
    if (err instanceof ApiError) {
      return { message: err.message, code: err.code, requestId: err.requestId };
    }
    const message: string = err instanceof Error ? err.message : 'Unknown error';
    return { message };
  }, []);

  const performLoad = useCallback(
    async (loadParams: ListEntriesParams): Promise<void> => {
      setStatus('loading');
      setError(null);
      try {
        const data = await listEntries(loadParams);
        if (!mountedRef.current) return;
        setEntries(data);
        setStatus('success');
      } catch (err) {
        if (!mountedRef.current) return;
        setError(handleError(err));
        setStatus('error');
      }
    },
    [handleError],
  );

  const load = useCallback(
    async (loadParams: ListEntriesParams): Promise<void> => {
      setParams(loadParams);
      await performLoad(loadParams);
    },
    [performLoad],
  );

  const refresh = useCallback(async (): Promise<void> => {
    if (params === undefined) {
      // No params set yet, can't refresh
      return;
    }
    await performLoad(params);
  }, [params, performLoad]);

  // Load entries when initial params are provided on mount
  useEffect(() => {
    if (initialParams !== undefined) {
      load(initialParams);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    mountedRef.current = true;
    return () => {
      mountedRef.current = false;
    };
  }, []);

  return useMemo<UseTimeEntriesResult>(
    () => ({
      status,
      entries,
      error,
      refresh,
      load,
    }),
    [status, entries, error, refresh, load],
  );
}

