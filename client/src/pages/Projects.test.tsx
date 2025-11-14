import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { cleanup, render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import type { Project } from '../types';
import Projects from './Projects';
import { ApiError } from '../api/http';

// Mock the API layer used by the hook
const listProjectsMock = vi.fn();
const createProjectMock = vi.fn();
const updateProjectMock = vi.fn();
const deleteProjectMock = vi.fn();

vi.mock('../api/projects', () => ({
  listProjects: (...args: unknown[]) => listProjectsMock(...args),
  createProject: (...args: unknown[]) => createProjectMock(...args),
  updateProject: (...args: unknown[]) => updateProjectMock(...args),
  deleteProject: (...args: unknown[]) => deleteProjectMock(...args),
}));

describe('Projects page', { concurrent: false }, () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    cleanup();
  });

  it('renders list from listProjects()', async () => {
    const projects: Project[] = [
      {
        id: '11111111-1111-1111-1111-111111111111',
        name: 'Alpha',
        description: 'First',
        createdAt: new Date('2025-11-02T10:00:00Z').toISOString(),
        updatedAt: new Date('2025-11-02T10:00:00Z').toISOString(),
      },
      {
        id: '22222222-2222-2222-2222-222222222222',
        name: 'Beta',
        createdAt: new Date('2025-11-03T10:00:00Z').toISOString(),
        updatedAt: new Date('2025-11-03T10:00:00Z').toISOString(),
      },
    ];
    listProjectsMock.mockResolvedValue(projects);

    render(<Projects />);

    // Names appear
    expect(await screen.findByText('Alpha')).toBeInTheDocument();
    expect(screen.getByText('Beta')).toBeInTheDocument();
  });

  it('creates a project and refreshes the list', async () => {
    // Keep a mutable list backing listProjects
    const current: Project[] = [];
    listProjectsMock.mockImplementation(async () => current.slice());
    createProjectMock.mockImplementation(
      async ({ name, description }: { name: string; description?: string | null }) => {
        const now = new Date().toISOString();
        const newProject: Project = {
          id: '33333333-3333-3333-3333-333333333333',
          name,
          description: description === null ? undefined : description,
          createdAt: now,
          updatedAt: now,
        };
        current.push(newProject);
        return newProject;
      },
    );

    render(<Projects />);

    const user = userEvent.setup();
    // Create form (only one at start)
    const nameInput = await screen.findByLabelText('Name');
    const descInput = screen.getByLabelText('Description (optional)');
    await user.type(nameInput, 'New Project');
    await user.type(descInput, 'Optional');
    await user.click(screen.getByRole('button', { name: 'Create' }));

    // After create, list should show new project
    expect(await screen.findByText('New Project')).toBeInTheDocument();
  });

  it('updates an existing project via inline edit', async () => {
    const now = new Date().toISOString();
    const current: Project[] = [
      {
        id: '44444444-4444-4444-4444-444444444444',
        name: 'Gamma',
        description: 'G',
        createdAt: now,
        updatedAt: now,
      },
    ];
    listProjectsMock.mockImplementation(async () => current.slice());
    updateProjectMock.mockImplementation(
      async (id: string, { name, description }: { name: string; description?: string | null }) => {
        const idx = current.findIndex((p) => p.id === id);
        if (idx >= 0) {
          current[idx] = {
            ...current[idx]!,
            name,
            description: description === null ? undefined : description,
            updatedAt: new Date().toISOString(),
          };
        }
        return current[idx];
      },
    );

    const user = userEvent.setup();
    render(<Projects />);
    
    // Start edit
    await user.click(await screen.findByRole('button', { name: 'Edit' }));
    const projectTable = await screen.findByRole('table');
    const editNameInput = within(projectTable).getByDisplayValue('Gamma');
    await user.clear(editNameInput);
    await user.type(editNameInput, 'Gamma v2');
    await user.click(within(projectTable).getByRole('button', { name: 'Save' }));

    // Updated name visible
    expect(within(projectTable).getByText('Gamma v2')).toBeInTheDocument();
    expect(updateProjectMock).toHaveBeenCalledWith('44444444-4444-4444-4444-444444444444', {
      name: 'Gamma v2',
      description: 'G',
    });
  });

  it('deletes a project', async () => {
    const now = new Date().toISOString();
    const current: Project[] = [
      {
        id: '55555555-5555-5555-5555-555555555555',
        name: 'Delta',
        createdAt: now,
        updatedAt: now,
      },
    ];
    listProjectsMock.mockImplementation(async () => current.slice());
    deleteProjectMock.mockImplementation(async (id: string) => {
      const idx = current.findIndex((p) => p.id === id);
      if (idx >= 0) current.splice(idx, 1);
    });

    const user = userEvent.setup();
    render(<Projects />);

    // Delete
    await user.click(await screen.findByRole('button', { name: 'Delete' }));
    // Row should no longer be present
    expect(await screen.findByText('No projects yet.')).toBeInTheDocument();
  });

  it('shows error details when list fails', async () => {
    const err = new ApiError('internal: oops', 500, { code: 'internal', requestId: 'req-abc' });
    listProjectsMock.mockRejectedValue(err);

    render(<Projects />);

    const alert = await screen.findByRole('alert');
    expect(within(alert).getByText(/Error:/i)).toBeInTheDocument();
    expect(within(alert).getByText(/Code:/i)).toHaveTextContent('Code: internal');
    expect(within(alert).getByText(/Request ID:/i)).toHaveTextContent('Request ID: req-abc');
  });
});
