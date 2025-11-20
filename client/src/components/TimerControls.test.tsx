import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { cleanup, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import TimerControls from './TimerControls';
import type { TimeEntry } from '../types';

describe('TimerControls component', { concurrent: false }, () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    cleanup();
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

  const categories = [
    { id: '22222222-2222-2222-2222-222222222222', name: 'Frontend' },
    { id: '33333333-3333-3333-3333-333333333333', name: 'Backend' },
  ];

  it('renders start controls when no active timer', () => {
    const onStart = vi.fn();
    const onStop = vi.fn();

    render(
      <TimerControls
        activeTimer={null}
        categories={categories}
        onStart={onStart}
        onStop={onStop}
      />,
    );

    expect(screen.getByLabelText('Category')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Start Timer' })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Stop Timer' })).not.toBeInTheDocument();
  });

  it('renders stop controls with elapsed time when active', async () => {
    const onStart = vi.fn();
    const onStop = vi.fn();
    const startedAt = new Date(Date.now() - 5000).toISOString(); // 5 seconds ago
    const activeTimer = createTimeEntry(startedAt, null);

    render(
      <TimerControls
        activeTimer={activeTimer}
        categories={categories}
        onStart={onStart}
        onStop={onStop}
      />,
    );

    expect(screen.getByText(/Active Timer:/i)).toBeInTheDocument();
    expect(screen.getByText('Frontend')).toBeInTheDocument();
    expect(screen.getByText(/Elapsed Time:/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Stop Timer' })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Start Timer' })).not.toBeInTheDocument();

    // Elapsed time should be displayed (component calculates from startedAt to now)
    // The exact value may vary slightly, so just check that elapsed time is shown
    await waitFor(
      () => {
        const elapsedText = screen.getByText(/Elapsed Time:/i).parentElement?.textContent;
        expect(elapsedText).toMatch(/\d+[hms]/);
      },
      { timeout: 2000 },
    );
  });

  it('start action invokes callback correctly', async () => {
    const user = userEvent.setup();
    const onStart = vi.fn().mockResolvedValue(undefined);
    const onStop = vi.fn();

    render(
      <TimerControls
        activeTimer={null}
        categories={categories}
        onStart={onStart}
        onStop={onStop}
      />,
    );

    const categorySelect = screen.getByLabelText('Category');
    await user.selectOptions(categorySelect, '22222222-2222-2222-2222-222222222222');
    await user.click(screen.getByRole('button', { name: 'Start Timer' }));

    expect(onStart).toHaveBeenCalledWith('22222222-2222-2222-2222-222222222222');
    expect(onStop).not.toHaveBeenCalled();
  });

  it('stop action invokes callback correctly', async () => {
    const user = userEvent.setup();
    const onStart = vi.fn();
    const onStop = vi.fn().mockResolvedValue(undefined);
    const activeTimer = createTimeEntry(new Date().toISOString(), null);

    render(
      <TimerControls
        activeTimer={activeTimer}
        categories={categories}
        onStart={onStart}
        onStop={onStop}
      />,
    );

    await user.click(screen.getByRole('button', { name: 'Stop Timer' }));

    expect(onStop).toHaveBeenCalled();
    expect(onStart).not.toHaveBeenCalled();
  });

  it('loading state disables controls', () => {
    const onStart = vi.fn();
    const onStop = vi.fn();

    render(
      <TimerControls
        activeTimer={null}
        categories={categories}
        onStart={onStart}
        onStop={onStop}
        loading={true}
      />,
    );

    const categorySelect = screen.getByLabelText('Category');
    const startButton = screen.getByRole('button', { name: 'Starting…' });

    expect(categorySelect).toBeDisabled();
    expect(startButton).toBeDisabled();
  });

  it('displays error message', () => {
    const onStart = vi.fn();
    const onStop = vi.fn();

    render(
      <TimerControls
        activeTimer={null}
        categories={categories}
        onStart={onStart}
        onStop={onStop}
        error="Something went wrong"
      />,
    );

    const alert = screen.getByRole('alert');
    expect(alert).toHaveTextContent('Error: Something went wrong');
  });

  it('updates elapsed time every second', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    const onStart = vi.fn();
    const onStop = vi.fn();
    const now = Date.now();
    vi.setSystemTime(now);
    const startedAt = new Date(now - 2000).toISOString(); // 2 seconds ago
    const activeTimer = createTimeEntry(startedAt, null);

    const { unmount } = render(
      <TimerControls
        activeTimer={activeTimer}
        categories={categories}
        onStart={onStart}
        onStop={onStop}
      />,
    );

    // Initially should show around 2s (component calculates from startedAt to now)
    await waitFor(
      () => {
        const elapsedText = screen.getByText(/Elapsed Time:/i).parentElement?.textContent;
        expect(elapsedText).toMatch(/2s/);
      },
      { timeout: 2000 },
    );

    // Advance time by 1 second
    vi.advanceTimersByTime(1000);
    vi.setSystemTime(now + 1000);

    // Should now show 3s
    await waitFor(
      () => {
        const elapsedText = screen.getByText(/Elapsed Time:/i).parentElement?.textContent;
        expect(elapsedText).toMatch(/3s/);
      },
      { timeout: 2000 },
    );

    unmount();
    vi.useRealTimers();
  });
});

