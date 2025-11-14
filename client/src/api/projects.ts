import type { Project } from '../types';

export async function listProjects(): Promise<Project[]> {
  throw new Error('Not implemented');
}

export interface CreateProjectInput {
  name: string;
  description?: string | null;
}

export async function createProject(_input: CreateProjectInput): Promise<Project> {
  throw new Error('Not implemented');
}

export async function getProject(_projectId: string): Promise<Project> {
  throw new Error('Not implemented');
}

export interface UpdateProjectInput {
  name: string;
  description?: string | null;
}

export async function updateProject(_projectId: string, _input: UpdateProjectInput): Promise<Project> {
  throw new Error('Not implemented');
}

export async function deleteProject(_projectId: string): Promise<void> {
  throw new Error('Not implemented');
}


