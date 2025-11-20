import { useCallback, useEffect, useMemo, useState, type JSX } from 'react';
import { useActiveTimer, type UseActiveTimerError } from '../hooks/useActiveTimer';
import { useTimeEntries, type UseTimeEntriesError } from '../hooks/useTimeEntries';
import { useCategories } from '../hooks/useCategories';
import { useProjects } from '../hooks/useProjects';
import TimerControls from '../components/TimerControls';
import EntryList from '../components/EntryList';

export default function Dashboard(): JSX.Element {
  const { projects } = useProjects();
  const [selectedProjectId, setSelectedProjectId] = useState<string>('');
  const [selectedCategoryId, setSelectedCategoryId] = useState<string>('');

  const {
    status: timerStatus,
    activeTimer,
    error: timerError,
    start: startTimer,
    stop: stopTimer,
  } = useActiveTimer();

  const {
    status: categoriesStatus,
    categories,
    error: categoriesError,
  } = useCategories(selectedProjectId || '');

  const {
    status: entriesStatus,
    entries,
    error: entriesError,
    load: loadEntries,
    refresh: refreshEntries,
  } = useTimeEntries();

  // Load entries when category is selected
  useEffect(() => {
    if (selectedCategoryId) {
      loadEntries({ categoryId: selectedCategoryId });
    }
  }, [loadEntries, selectedCategoryId]);

  const handleProjectChange = useCallback((e: React.ChangeEvent<HTMLSelectElement>): void => {
    setSelectedProjectId(e.target.value);
    setSelectedCategoryId(''); // Reset category when project changes
  }, []);

  const handleCategoryChange = useCallback((e: React.ChangeEvent<HTMLSelectElement>): void => {
    setSelectedCategoryId(e.target.value);
  }, []);

  const handleStart = useCallback(
    async (categoryId: string): Promise<void> => {
      await startTimer(categoryId);
      await refreshEntries();
    },
    [startTimer, refreshEntries],
  );

  const handleStop = useCallback(async (): Promise<void> => {
    await stopTimer();
    await refreshEntries();
  }, [stopTimer, refreshEntries]);

  const handleFilterChange = useCallback(
    (params: { from?: string; to?: string }): void => {
      if (selectedCategoryId) {
        loadEntries({ categoryId: selectedCategoryId, ...params });
      }
    },
    [selectedCategoryId, loadEntries],
  );

  const categoryOptions = useMemo<Array<{ id: string; name: string }>>(() => {
    return categories.map((cat) => ({ id: cat.id, name: cat.name }));
  }, [categories]);

  const timerLoading = timerStatus === 'loading';
  const categoriesLoading = categoriesStatus === 'loading';
  const entriesLoading = entriesStatus === 'loading';

  const timerErrorMessage = useMemo<string | null>(() => {
    if (!timerError) return null;
    return getTimerErrorMessage(timerError);
  }, [timerError]);

  const entriesErrorMessage = useMemo<string | null>(() => {
    if (!entriesError) return null;
    return getEntriesErrorMessage(entriesError);
  }, [entriesError]);

  const renderCategorySelector = useCallback((): JSX.Element => {
    if (categoriesLoading) {
      return <p>Loading categories…</p>;
    }
    if (categoriesError) {
      return (
        <div role="alert">
          <p>Error loading categories: {categoriesError.message}</p>
        </div>
      );
    }
    return (
      <select
        id="dashboard-category"
        value={selectedCategoryId}
        onChange={handleCategoryChange}
        style={{ display: 'block', marginTop: '0.25rem', minWidth: '200px' }}
      >
        <option value="">Select a category</option>
        {categories.map((category) => (
          <option key={category.id} value={category.id}>
            {category.name}
          </option>
        ))}
      </select>
    );
  }, [categoriesLoading, categoriesError, categories, selectedCategoryId, handleCategoryChange]);

  return (
    <section>
      <h1>Time Tracking Dashboard</h1>

      <div style={{ marginBottom: '2rem' }}>
        <h2>Project & Category Selection</h2>
        <div style={{ marginBottom: '0.75rem' }}>
          <label htmlFor="dashboard-project">Project</label>
          <select
            id="dashboard-project"
            value={selectedProjectId}
            onChange={handleProjectChange}
            style={{ display: 'block', marginTop: '0.25rem', minWidth: '200px' }}
          >
            <option value="">Select a project</option>
            {projects.map((project) => (
              <option key={project.id} value={project.id}>
                {project.name}
              </option>
            ))}
          </select>
        </div>

        {selectedProjectId ? (
          <div style={{ marginBottom: '0.75rem' }}>
            <label htmlFor="dashboard-category">Category</label>
            {renderCategorySelector()}
          </div>
        ) : null}
      </div>

      <div style={{ marginBottom: '2rem' }}>
        <h2>Timer Controls</h2>
        {selectedProjectId && categoryOptions.length > 0 ? (
          <TimerControls
            activeTimer={activeTimer}
            categories={categoryOptions}
            onStart={handleStart}
            onStop={handleStop}
            loading={timerLoading}
            error={timerErrorMessage}
          />
        ) : (
          <p>Please select a project and category to start a timer.</p>
        )}
      </div>

      <div style={{ marginBottom: '2rem' }}>
        <h2>Time Entries</h2>
        {selectedCategoryId ? (
          <EntryList
            entries={entries}
            categories={categoryOptions}
            loading={entriesLoading}
            error={entriesErrorMessage}
            onFilterChange={handleFilterChange}
          />
        ) : (
          <p>Please select a category to view time entries.</p>
        )}
      </div>
    </section>
  );
}

function getTimerErrorMessage(err: UseActiveTimerError): string {
  if (err.code === 'no_active_timer') {
    return 'No active timer to stop.';
  }
  return err.message;
}

function getEntriesErrorMessage(err: UseTimeEntriesError): string {
  switch (err.code) {
    case 'invalid_id':
      return 'Invalid category ID. Please select a valid category.';
    case 'invalid_time':
      return 'Invalid date format. Please check your date filters.';
    case 'invalid_time_range':
      return 'Invalid date range. The "from" date must be before the "to" date.';
    default:
      return err.message;
  }
}
