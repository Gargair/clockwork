import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { cleanup, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import EntryList from './EntryList';
import type { TimeEntry } from '../types';

describe('EntryList component', { concurrent: false }, () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    cleanup();
  });

  const createTimeEntry = (
    id: string,
    categoryId: string,
    startedAt: string,
    stoppedAt: string | null = null,
    durationSeconds: number | null = null,
  ): TimeEntry => ({
    id,
    categoryId,
    startedAt,
    stoppedAt,
    durationSeconds,
    createdAt: startedAt,
    updatedAt: startedAt,
  });

  const categories = [
    { id: '22222222-2222-2222-2222-222222222222', name: 'Frontend' },
    { id: '33333333-3333-3333-3333-333333333333', name: 'Backend' },
  ];

  it('renders entries list', () => {
    const entries: TimeEntry[] = [
      createTimeEntry(
        '11111111-1111-1111-1111-111111111111',
        '22222222-2222-2222-2222-222222222222',
        '2025-11-02T10:00:00Z',
        '2025-11-02T11:00:00Z',
        3600,
      ),
      createTimeEntry(
        '44444444-4444-4444-4444-444444444444',
        '33333333-3333-3333-3333-333333333333',
        '2025-11-02T12:00:00Z',
        '2025-11-02T12:30:00Z',
        1800,
      ),
    ];

    render(<EntryList entries={entries} categories={categories} />);

    expect(screen.getByText('Frontend')).toBeInTheDocument();
    expect(screen.getByText('Backend')).toBeInTheDocument();
    expect(screen.getAllByText(/Started:/i).length).toBeGreaterThan(0);
    expect(screen.getAllByText(/Stopped:/i).length).toBeGreaterThan(0);
  });

  it('shows empty state when no entries', () => {
    render(<EntryList entries={[]} categories={categories} />);

    expect(screen.getByText('No entries found.')).toBeInTheDocument();
  });

  it('formats durations correctly', () => {
    const entries: TimeEntry[] = [
      createTimeEntry(
        '11111111-1111-1111-1111-111111111111',
        '22222222-2222-2222-2222-222222222222',
        '2025-11-02T10:00:00Z',
        '2025-11-02T11:00:00Z',
        3600, // 1 hour
      ),
      createTimeEntry(
        '22222222-2222-2222-2222-222222222222',
        '22222222-2222-2222-2222-222222222222',
        '2025-11-02T12:00:00Z',
        '2025-11-02T12:30:00Z',
        1800, // 30 minutes
      ),
      createTimeEntry(
        '33333333-3333-3333-3333-333333333333',
        '22222222-2222-2222-2222-222222222222',
        '2025-11-02T13:00:00Z',
        '2025-11-02T13:00:45Z',
        45, // 45 seconds
      ),
    ];

    render(<EntryList entries={entries} categories={categories} />);

    expect(screen.getByText(/1h/)).toBeInTheDocument();
    expect(screen.getByText(/30m/)).toBeInTheDocument();
    expect(screen.getByText(/45s/)).toBeInTheDocument();
  });

  it('shows active indicator for entries without stoppedAt', () => {
    const entries: TimeEntry[] = [
      createTimeEntry(
        '11111111-1111-1111-1111-111111111111',
        '22222222-2222-2222-2222-222222222222',
        '2025-11-02T10:00:00Z',
        null,
        null,
      ),
    ];

    render(<EntryList entries={entries} categories={categories} />);

    expect(screen.getByText(/Active/i)).toBeInTheDocument();
  });

  it('filter inputs call onFilterChange with correct values', async () => {
    const user = userEvent.setup();
    const onFilterChange = vi.fn();

    render(
      <EntryList
        entries={[]}
        categories={categories}
        onFilterChange={onFilterChange}
      />,
    );

    const fromInput = screen.getByLabelText('From Date');
    const toInput = screen.getByLabelText('To Date');

    await user.type(fromInput, '2025-11-01');
    await user.type(toInput, '2025-11-30');

    // onFilterChange should be called with RFC3339 formatted dates
    expect(onFilterChange).toHaveBeenCalled();
    const calls = onFilterChange.mock.calls;
    expect(calls.length).toBeGreaterThan(0);
    const lastCall = calls.at(-1)?.[0];
    expect(lastCall).toHaveProperty('from');
    expect(lastCall).toHaveProperty('to');
    if (lastCall?.from) {
      expect(lastCall.from).toMatch(/2025-11-01T00:00:00/);
    }
    if (lastCall?.to) {
      expect(lastCall.to).toMatch(/2025-11-30T23:59:59/);
    }
  });

  it('displays error message', () => {
    render(<EntryList entries={[]} categories={categories} error="Failed to load entries" />);

    const alert = screen.getByRole('alert');
    expect(alert).toHaveTextContent('Error: Failed to load entries');
  });

  it('shows loading state', () => {
    render(<EntryList entries={[]} categories={categories} loading={true} />);

    expect(screen.getByText('Loading entries…')).toBeInTheDocument();
  });
});

