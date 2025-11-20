import { useCallback, useEffect, useMemo, useRef, useState, type JSX } from 'react';
import type { TimeEntry } from '../types';

export interface TimerControlsProps {
  activeTimer: TimeEntry | null;
  categories: Array<{ id: string; name: string }>;
  onStart: (categoryId: string) => void | Promise<void>;
  onStop: () => void | Promise<void>;
  loading?: boolean;
  error?: string | null;
}

export default function TimerControls(props: Readonly<TimerControlsProps>): JSX.Element {
  const { activeTimer, categories, onStart, onStop, loading = false, error = null } = props;
  const [selectedCategoryId, setSelectedCategoryId] = useState<string>('');
  const [elapsedTime, setElapsedTime] = useState<number>(0);
  const intervalRef = useRef<number | null>(null);

  // Calculate elapsed time from startedAt to now if stoppedAt is null
  useEffect(() => {
    if (activeTimer?.stoppedAt !== null) {
      setElapsedTime(0);
      return;
    }

    // Calculate initial elapsed time
    const startedAt = new Date(activeTimer.startedAt).getTime();
    const updateElapsed = (): void => {
      const now = Date.now();
      const elapsed = Math.floor((now - startedAt) / 1000);
      setElapsedTime(elapsed);
    };

    updateElapsed();

    // Update every second
    intervalRef.current = globalThis.setInterval(updateElapsed, 1000);

    return () => {
      if (intervalRef.current !== null) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
    };
  }, [activeTimer]);

  const handleCategoryChange = useCallback((e: React.ChangeEvent<HTMLSelectElement>): void => {
    setSelectedCategoryId(e.target.value);
  }, []);

  const handleStart = useCallback(
    async (e: React.FormEvent<HTMLFormElement>): Promise<void> => {
      e.preventDefault();
      if (selectedCategoryId === '') {
        return;
      }
      await onStart(selectedCategoryId);
    },
    [selectedCategoryId, onStart],
  );

  const handleStop = useCallback(async (): Promise<void> => {
    await onStop();
  }, [onStop]);

  const formattedElapsed = useMemo<string>(() => {
    const hours = Math.floor(elapsedTime / 3600);
    const minutes = Math.floor((elapsedTime % 3600) / 60);
    const seconds = elapsedTime % 60;

    if (hours > 0) {
      return `${hours}h ${minutes}m ${seconds}s`;
    }
    if (minutes > 0) {
      return `${minutes}m ${seconds}s`;
    }
    return `${seconds}s`;
  }, [elapsedTime]);

  const categoryName = useMemo<string | null>(() => {
    if (activeTimer === null) {
      return null;
    }
    const category = categories.find((c) => c.id === activeTimer.categoryId);
    return category?.name ?? null;
  }, [activeTimer, categories]);

  return (
    <div>
      {error ? (
        <div role="alert" style={{ color: 'var(--color-danger, red)', marginBottom: '1rem' }}>
          <p>Error: {error}</p>
        </div>
      ) : null}

      {activeTimer === null ? (
        <form onSubmit={handleStart}>
          <div style={{ marginBottom: '0.75rem' }}>
            <label htmlFor="timer-category">Category</label>
            <select
              id="timer-category"
              name="category"
              value={selectedCategoryId}
              onChange={handleCategoryChange}
              disabled={loading}
              required
            >
              <option value="">Select a category</option>
              {categories.map((category) => (
                <option key={category.id} value={category.id}>
                  {category.name}
                </option>
              ))}
            </select>
          </div>
          <div>
            <button type="submit" disabled={loading || selectedCategoryId === ''}>
              {loading ? 'Starting…' : 'Start Timer'}
            </button>
          </div>
        </form>
      ) : (
        <div>
          <div style={{ marginBottom: '1rem' }}>
            <p>
              <strong>Active Timer:</strong> {categoryName ?? 'Unknown Category'}
            </p>
            <p>
              <strong>Elapsed Time:</strong> {formattedElapsed}
            </p>
            <p style={{ fontSize: '0.875rem', color: 'var(--color-muted, gray)' }}>
              Started: {new Date(activeTimer.startedAt).toLocaleString()}
            </p>
          </div>
          <div>
            <button type="button" onClick={handleStop} disabled={loading}>
              {loading ? 'Stopping…' : 'Stop Timer'}
            </button>
          </div>
        </div>
      )}
    </div>
  );
}

