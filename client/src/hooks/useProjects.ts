import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import type { Project } from '../types';
import {
  listProjects,
  createProject as apiCreateProject,
  updateProject as apiUpdateProject,
  deleteProject as apiDeleteProject,
  type CreateProjectInput,
  type UpdateProjectInput,
} from '../api/projects';
import { ApiError } from '../api/http';

export type UseProjectsStatus = 'idle' | 'loading' | 'success' | 'error';

export interface UseProjectsError {
  message: string;
  code?: string;
  requestId?: string;
}

export interface UseProjectsResult {
  status: UseProjectsStatus;
  projects: Project[];
  error: UseProjectsError | null;
  refresh: () => Promise<void>;
  create: (input: CreateProjectInput) => Promise<void>;
  update: (projectId: string, input: UpdateProjectInput) => Promise<void>;
  remove: (projectId: string) => Promise<void>;
}

export function useProjects(): UseProjectsResult {
  const [status, setStatus] = useState<UseProjectsStatus>('idle');
  const [projects, setProjects] = useState<Project[]>([]);
  const [error, setError] = useState<UseProjectsError | null>(null);
  const mountedRef = useRef<boolean>(true);

  useEffect(() => {
    mountedRef.current = true;
    refresh();
    return () => {
      mountedRef.current = false;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleError = useCallback((err: unknown): UseProjectsError => {
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
      const data = await listProjects();
      if (!mountedRef.current) return;
      setProjects(data);
      setStatus('success');
    } catch (err) {
      if (!mountedRef.current) return;
      setError(handleError(err));
      setStatus('error');
    }
  }, [handleError]);

  const create = useCallback(async (input: CreateProjectInput): Promise<void> => {
    setError(null);
    try {
      await apiCreateProject(input);
      if (!mountedRef.current) return;
      await refresh();
    } catch (err) {
      if (!mountedRef.current) return;
      setError(handleError(err));
      setStatus('error');
    }
  }, [handleError, refresh]);

  const update = useCallback(async (projectId: string, input: UpdateProjectInput): Promise<void> => {
    setError(null);
    try {
      await apiUpdateProject(projectId, input);
      if (!mountedRef.current) return;
      await refresh();
    } catch (err) {
      if (!mountedRef.current) return;
      setError(handleError(err));
      setStatus('error');
    }
  }, [handleError, refresh]);

  const remove = useCallback(async (projectId: string): Promise<void> => {
    setError(null);
    try {
      await apiDeleteProject(projectId);
      if (!mountedRef.current) return;
      await refresh();
    } catch (err) {
      if (!mountedRef.current) return;
      setError(handleError(err));
      setStatus('error');
    }
  }, [handleError, refresh]);

  return useMemo<UseProjectsResult>(() => ({
    status,
    projects,
    error,
    refresh,
    create,
    update,
    remove,
  }), [status, projects, error, refresh, create, update, remove]);
}


