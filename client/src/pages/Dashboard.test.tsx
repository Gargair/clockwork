import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { cleanup, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import Dashboard from './Dashboard';
import type { TimeEntry, Project, Category } from '../types';

// Mock the API layers
const listProjectsMock = vi.fn();
const listCategoriesMock = vi.fn();
const startTimerMock = vi.fn();
const stopTimerMock = vi.fn();
const getActiveTimerMock = vi.fn();
const listEntriesMock = vi.fn();

vi.mock('../api/projects', () => ({
  listProjects: (...args: unknown[]) => listProjectsMock(...args),
}));

vi.mock('../api/categories', () => ({
  listCategories: (...args: unknown[]) => listCategoriesMock(...args),
}));

vi.mock('../api/time', () => ({
  startTimer: (...args: unknown[]) => startTimerMock(...args),
  stopTimer: (...args: unknown[]) => stopTimerMock(...args),
  getActiveTimer: (...args: unknown[]) => getActiveTimerMock(...args),
  listEntries: (...args: unknown[]) => listEntriesMock(...args),
}));

function renderWithRouter() {
  return render(
    <MemoryRouter initialEntries={[{ pathname: '/dashboard' }]}>
      <Routes>
        <Route path="/dashboard" element={<Dashboard />} />
      </Routes>
    </MemoryRouter>,
  );
}

describe('Dashboard page', { concurrent: false }, () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    cleanup();
  });

  const createProject = (id: string, name: string): Project => ({
    id,
    name,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  });

  const createCategory = (id: string, name: string, projectId: string): Category => ({
    id,
    projectId,
    parentCategoryId: null,
    name,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  });

  const createTimeEntry = (
    id: string,
    categoryId: string,
    startedAt: string,
    stoppedAt: string | null = null,
  ): TimeEntry => ({
    id,
    categoryId,
    startedAt,
    stoppedAt,
    durationSeconds: stoppedAt ? 3600 : null,
    createdAt: startedAt,
    updatedAt: startedAt,
  });

  it('renders timer controls and entry list', async () => {
    const projects: Project[] = [createProject('proj-1', 'Test Project')];
    const categories: Category[] = [createCategory('cat-1', 'Frontend', 'proj-1')];
    const entries: TimeEntry[] = [
      createTimeEntry('entry-1', 'cat-1', '2025-11-02T10:00:00Z', '2025-11-02T11:00:00Z'),
    ];

    listProjectsMock.mockResolvedValue(projects);
    listCategoriesMock.mockResolvedValue(categories);
    getActiveTimerMock.mockResolvedValue(null);
    listEntriesMock.mockResolvedValue(entries);

    renderWithRouter();

    expect(await screen.findByText('Time Tracking Dashboard')).toBeInTheDocument();
    expect(screen.getByText('Timer Controls')).toBeInTheDocument();
    expect(screen.getByText('Time Entries')).toBeInTheDocument();
  });

  it('starting a timer calls startTimer() and updates UI', async () => {
    const user = userEvent.setup();
    const projects: Project[] = [createProject('proj-1', 'Test Project')];
    const categories: Category[] = [createCategory('cat-1', 'Frontend', 'proj-1')];
    const newTimer = createTimeEntry('timer-1', 'cat-1', new Date().toISOString(), null);

    listProjectsMock.mockResolvedValue(projects);
    listCategoriesMock.mockResolvedValue(categories);
    getActiveTimerMock.mockResolvedValue(null);
    startTimerMock.mockResolvedValue(newTimer);
    getActiveTimerMock.mockResolvedValueOnce(null).mockResolvedValue(newTimer);
    listEntriesMock.mockResolvedValue([]);

    renderWithRouter();

    // Select project
    const projectSelect = await screen.findByLabelText('Project');
    await user.selectOptions(projectSelect, 'proj-1');

    // Wait for categories to load
    await waitFor(() => {
      expect(screen.getByLabelText('Category', { selector: '#dashboard-category' })).toBeInTheDocument();
    });

    // Select category (use ID to avoid conflict with TimerControls category selector)
    const categorySelect = screen.getByLabelText('Category', { selector: '#dashboard-category' });
    await user.selectOptions(categorySelect, 'cat-1');

    // Wait for TimerControls to appear (it only shows when category is selected)
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'Start Timer' })).toBeInTheDocument();
    });

    // Select category in TimerControls (it has its own category selector)
    const timerCategorySelect = screen.getByLabelText('Category', { selector: '#timer-category' });
    await user.selectOptions(timerCategorySelect, 'cat-1');

    // Start timer
    const startButton = screen.getByRole('button', { name: 'Start Timer' });
    await user.click(startButton);

    // Wait for the API call to complete
    await waitFor(() => {
      expect(startTimerMock).toHaveBeenCalledWith('cat-1');
    });
    expect(getActiveTimerMock).toHaveBeenCalled();
  });

  it('stopping a timer calls stopTimer() and clears active state', async () => {
    const user = userEvent.setup();
    const projects: Project[] = [createProject('proj-1', 'Test Project')];
    const categories: Category[] = [createCategory('cat-1', 'Frontend', 'proj-1')];
    const activeTimer = createTimeEntry('timer-1', 'cat-1', new Date().toISOString(), null);
    const stoppedTimer = createTimeEntry(
      'timer-1',
      'cat-1',
      activeTimer.startedAt,
      new Date().toISOString(),
    );

    listProjectsMock.mockResolvedValue(projects);
    listCategoriesMock.mockResolvedValue(categories);
    getActiveTimerMock.mockResolvedValue(activeTimer);
    stopTimerMock.mockResolvedValue(stoppedTimer);
    getActiveTimerMock.mockResolvedValueOnce(activeTimer).mockResolvedValue(null);
    listEntriesMock.mockResolvedValue([]);

    renderWithRouter();

    // Wait for timer to be displayed (need to select project and category first)
    const projectSelect = await screen.findByLabelText('Project');
    await user.selectOptions(projectSelect, 'proj-1');

    await waitFor(() => {
      expect(screen.getByLabelText('Category', { selector: '#dashboard-category' })).toBeInTheDocument();
    });

    // Wait for timer to be displayed
    await waitFor(() => {
      expect(screen.getByText(/Active Timer:/i)).toBeInTheDocument();
    }, { timeout: 3000 });

    // Stop timer
    const stopButton = screen.getByRole('button', { name: 'Stop Timer' });
    await user.click(stopButton);

    expect(stopTimerMock).toHaveBeenCalled();
    expect(getActiveTimerMock).toHaveBeenCalled();
  });

  it('entry list displays entries with formatted durations', async () => {
    const projects: Project[] = [createProject('proj-1', 'Test Project')];
    const categories: Category[] = [createCategory('cat-1', 'Frontend', 'proj-1')];
    const entries: TimeEntry[] = [
      createTimeEntry('entry-1', 'cat-1', '2025-11-02T10:00:00Z', '2025-11-02T11:00:00Z'),
    ];

    listProjectsMock.mockResolvedValue(projects);
    listCategoriesMock.mockResolvedValue(categories);
    getActiveTimerMock.mockResolvedValue(null);
    listEntriesMock.mockResolvedValue(entries);

    renderWithRouter();

    // Select project and category
    const user = userEvent.setup();
    const projectSelect = await screen.findByLabelText('Project');
    await user.selectOptions(projectSelect, 'proj-1');

    await waitFor(() => {
      expect(screen.getByLabelText('Category', { selector: '#dashboard-category' })).toBeInTheDocument();
    });

    const categorySelect = screen.getByLabelText('Category', { selector: '#dashboard-category' });
    await user.selectOptions(categorySelect, 'cat-1');

    // Wait for entries to load - use getAllByText since "Frontend" appears in both category selector and entry list
    await waitFor(() => {
      const frontendElements = screen.getAllByText('Frontend');
      expect(frontendElements.length).toBeGreaterThan(0);
    });

    // Check for duration in entry list (should be in a list item)
    expect(screen.getByText(/1h/)).toBeInTheDocument();
  });

  it('filtering entries by date range calls listEntries() with correct params', async () => {
    const user = userEvent.setup();
    const projects: Project[] = [createProject('proj-1', 'Test Project')];
    const categories: Category[] = [createCategory('cat-1', 'Frontend', 'proj-1')];
    const entries: TimeEntry[] = [];

    listProjectsMock.mockResolvedValue(projects);
    listCategoriesMock.mockResolvedValue(categories);
    getActiveTimerMock.mockResolvedValue(null);
    listEntriesMock.mockResolvedValue(entries);

    renderWithRouter();

    // Select project and category
    const projectSelect = await screen.findByLabelText('Project');
    await user.selectOptions(projectSelect, 'proj-1');

    await waitFor(() => {
      expect(screen.getByLabelText('Category', { selector: '#dashboard-category' })).toBeInTheDocument();
    });

    const categorySelect = screen.getByLabelText('Category', { selector: '#dashboard-category' });
    await user.selectOptions(categorySelect, 'cat-1');

    // Wait for filter inputs
    await waitFor(() => {
      expect(screen.getByLabelText('From Date')).toBeInTheDocument();
    });

    // Wait for entries to load first (category selection triggers initial load)
    await waitFor(() => {
      expect(listEntriesMock).toHaveBeenCalled();
    });

    // Wait for initial load to complete
    await waitFor(() => {
      expect(listEntriesMock).toHaveBeenCalled();
    });

    // Clear previous calls to check only the filter calls
    listEntriesMock.mockClear();

    // Set date filters (date inputs don't need clearing, just type the new value)
    const fromInput = screen.getByLabelText('From Date');
    await user.type(fromInput, '2025-11-01');
    await waitFor(() => {
      expect(listEntriesMock).toHaveBeenCalledWith(
        expect.objectContaining({
          categoryId: 'cat-1',
          from: expect.stringContaining('2025-11-01T00:00:00Z'),
          to: undefined,
        }),
      );
    }, { timeout: 1000 });
    const toInput = screen.getByLabelText('To Date');
    await user.type(toInput, '2025-11-30');
    await waitFor(() => {
      expect(listEntriesMock).toHaveBeenCalledWith(
        expect.objectContaining({
          categoryId: 'cat-1',
          from: expect.stringContaining('2025-11-01T00:00:00Z'),
          to: expect.stringContaining('2025-11-30T23:59:59Z'),
        }),
      );
    }, { timeout: 1000 });
  });
});

