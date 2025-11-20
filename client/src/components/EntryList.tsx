import { useCallback, useState, type JSX } from 'react';
import type { TimeEntry } from '../types';

export interface EntryListProps {
  entries: TimeEntry[];
  categories: Array<{ id: string; name: string }>;
  loading?: boolean;
  error?: string | null;
  onFilterChange?: (params: { from?: string; to?: string }) => void;
}

export default function EntryList(props: Readonly<EntryListProps>): JSX.Element {
  const { entries, categories, loading = false, error = null, onFilterChange } = props;
  const [fromDate, setFromDate] = useState<string>('');
  const [toDate, setToDate] = useState<string>('');

  const formatDuration = useCallback((durationSeconds: number | null): string => {
    if (durationSeconds === null) {
      return 'N/A';
    }
    const hours = Math.floor(durationSeconds / 3600);
    const minutes = Math.floor((durationSeconds % 3600) / 60);
    const seconds = durationSeconds % 60;

    if (hours > 0) {
      return `${hours}h ${minutes}m`;
    }
    if (minutes > 0) {
      return `${minutes}m`;
    }
    return `${seconds}s`;
  }, []);

  const getCategoryName = useCallback(
    (categoryId: string): string => {
      const category = categories.find((c) => c.id === categoryId);
      return category?.name ?? 'Unknown Category';
    },
    [categories],
  );

  const handleFromChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>): void => {
      const value = e.target.value;
      setFromDate(value);
      if (onFilterChange) {
        const from = value ? `${value}T00:00:00Z` : undefined;
        onFilterChange({ from, to: toDate ? `${toDate}T23:59:59Z` : undefined });
      }
    },
    [toDate, onFilterChange],
  );

  const handleToChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>): void => {
      const value = e.target.value;
      setToDate(value);
      if (onFilterChange) {
        onFilterChange({
          from: fromDate ? `${fromDate}T00:00:00Z` : undefined,
          to: value ? `${value}T23:59:59Z` : undefined,
        });
      }
    },
    [fromDate, onFilterChange],
  );

  if (loading) {
    return <p>Loading entries…</p>;
  }

  if (error) {
    return (
      <div role="alert">
        <p>Error: {error}</p>
      </div>
    );
  }

  return (
    <div>
      {onFilterChange ? (
        <div style={{ marginBottom: '1rem', display: 'flex', gap: '1rem', alignItems: 'flex-end' }}>
          <div>
            <label htmlFor="filter-from">From Date</label>
            <input
              id="filter-from"
              type="date"
              value={fromDate}
              onChange={handleFromChange}
              style={{ display: 'block', marginTop: '0.25rem' }}
            />
          </div>
          <div>
            <label htmlFor="filter-to">To Date</label>
            <input
              id="filter-to"
              type="date"
              value={toDate}
              onChange={handleToChange}
              style={{ display: 'block', marginTop: '0.25rem' }}
            />
          </div>
        </div>
      ) : null}

      {entries.length === 0 ? (
        <p>No entries found.</p>
      ) : (
        <ul style={{ listStyle: 'none', paddingLeft: 0 }}>
          {entries.map((entry) => {
            const startDate = new Date(entry.startedAt);
            const stopDate = entry.stoppedAt ? new Date(entry.stoppedAt) : null;
            const duration = entry.durationSeconds;

            return (
              <li
                key={entry.id}
                style={{
                  padding: '0.75rem',
                  marginBottom: '0.5rem',
                  border: '1px solid var(--color-border, #ddd)',
                  borderRadius: '0.25rem',
                }}
              >
                <div style={{ marginBottom: '0.5rem' }}>
                  <strong>{getCategoryName(entry.categoryId)}</strong>
                </div>
                <div style={{ fontSize: '0.875rem', color: 'var(--color-muted, gray)' }}>
                  <div>
                    <strong>Started:</strong> {startDate.toLocaleString()}
                  </div>
                  {stopDate ? (
                    <div>
                      <strong>Stopped:</strong> {stopDate.toLocaleString()}
                    </div>
                  ) : (
                    <div style={{ color: 'var(--color-warning, orange)' }}>
                      <strong>Active</strong>
                    </div>
                  )}
                  <div>
                    <strong>Duration:</strong> {formatDuration(duration)}
                  </div>
                </div>
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}

