import { useCallback, useMemo, useState, type JSX } from 'react';
import type { Project } from '../types';
import { useProjects, type UseProjectsError } from '../hooks/useProjects';
import ProjectForm, { type ProjectFormValues } from '../components/ProjectForm';

export default function Projects(): JSX.Element {
  const { status, projects, error, refresh, create, update, remove } = useProjects();
  const [editingProjectId, setEditingProjectId] = useState<string | null>(null);

  const hasProjects: boolean = projects.length > 0;
  const isLoading: boolean = status === 'loading';

  const handleRefresh = useCallback(async (): Promise<void> => {
    await refresh();
  }, [refresh]);

  const handleCreate = useCallback(
    async (values: ProjectFormValues): Promise<void> => {
      await create(values);
    },
    [create],
  );

  const handleCancelEdit = useCallback((): void => {
    setEditingProjectId(null);
  }, []);

  const startEditById = useMemo<Record<string, () => void>>(() => {
    const entries: Array<[string, () => void]> = projects.map((p: Project) => [
      p.id,
      () => setEditingProjectId(p.id),
    ]);
    return Object.fromEntries(entries);
  }, [projects]);

  const editSubmitById = useMemo<
    Record<string, (values: ProjectFormValues) => Promise<void>>
  >(() => {
    const entries: Array<[string, (values: ProjectFormValues) => Promise<void>]> = projects.map(
      (p: Project) => [
        p.id,
        async (values: ProjectFormValues): Promise<void> => {
          await update(p.id, values);
          setEditingProjectId(null);
        },
      ],
    );
    return Object.fromEntries(entries);
  }, [projects, update]);

  const deleteById = useMemo<Record<string, () => Promise<void>>>(() => {
    const entries: Array<[string, () => Promise<void>]> = projects.map((p: Project) => [
      p.id,
      async (): Promise<void> => {
        await remove(p.id);
        if (editingProjectId === p.id) {
          setEditingProjectId(null);
        }
      },
    ]);
    return Object.fromEntries(entries);
  }, [projects, remove, editingProjectId]);

  return (
    <section>
      <header>
        <h1>Projects</h1>
        <button type="button" onClick={handleRefresh} disabled={isLoading}>
          {isLoading ? 'Refreshing…' : 'Refresh'}
        </button>
      </header>

      {status === 'loading' ? <p>Loading projects…</p> : null}
      {status === 'error' ? renderError(error) : null}

      <div style={{ marginTop: '1rem', marginBottom: '1rem' }}>
        <h2>Create Project</h2>
        <ProjectForm onSubmit={handleCreate} submitLabel="Create" />
      </div>

      <div>
        <h2>Existing Projects</h2>
        {hasProjects ? (
          <table>
            <thead>
              <tr>
                <th>Name</th>
                <th>Description</th>
                <th>Created</th>
                <th>Updated</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {projects.map((p: Project) => {
                const isEditing: boolean = editingProjectId === p.id;
                const startEdit = startEditById[p.id]!;
                const submitEdit = editSubmitById[p.id]!;
                const deleteRow = deleteById[p.id]!;
                return (
                  <tr key={p.id}>
                    <td colSpan={isEditing ? 5 : 1}>
                      {isEditing ? (
                        <div>
                          <ProjectForm
                            initial={{ name: p.name, description: p.description ?? null }}
                            onSubmit={submitEdit}
                            submitLabel="Save"
                          />
                          <div style={{ marginTop: '0.5rem' }}>
                            <button type="button" onClick={handleCancelEdit}>
                              Cancel
                            </button>
                          </div>
                        </div>
                      ) : (
                        p.name
                      )}
                    </td>
                    {isEditing ? null : (
                      <>
                        <td>{p.description ?? ''}</td>
                        <td>{formatDate(p.createdAt)}</td>
                        <td>{formatDate(p.updatedAt)}</td>
                        <td>
                          <button type="button" onClick={startEdit}>
                            Edit
                          </button>
                          <button type="button" onClick={deleteRow}>
                            Delete
                          </button>
                        </td>
                      </>
                    )}
                  </tr>
                );
              })}
            </tbody>
          </table>
        ) : (
          <p>No projects yet.</p>
        )}
      </div>
    </section>
  );
}

function formatDate(value: string): string {
  const d = new Date(value);
  return Number.isNaN(d.getTime()) ? value : d.toLocaleString();
}

function renderError(err: UseProjectsError | null): JSX.Element | null {
  if (!err) return null;
  return (
    <div role="alert">
      <p>Error: {err.message}</p>
      {err.code ? <p>Code: {err.code}</p> : null}
      {err.requestId ? <p>Request ID: {err.requestId}</p> : null}
    </div>
  );
}
