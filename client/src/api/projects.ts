import { z } from 'zod';
import { requestJson } from './http';
import { type Project, ProjectSchema, ProjectListSchema } from '../types';

export async function listProjects(): Promise<Project[]> {
  return requestJson('/api/projects', { method: 'GET' }, ProjectListSchema);
}

export interface CreateProjectInput {
  name: string;
  description?: string | null;
}

export async function createProject(input: CreateProjectInput): Promise<Project> {
  return requestJson('/api/projects', { method: 'POST', body: input }, ProjectSchema);
}

export async function getProject(projectId: string): Promise<Project> {
  return requestJson(`/api/projects/${projectId}`, { method: 'GET' }, ProjectSchema);
}

export interface UpdateProjectInput {
  name: string;
  description?: string | null;
}

export async function updateProject(projectId: string, input: UpdateProjectInput): Promise<Project> {
  return requestJson(`/api/projects/${projectId}`, { method: 'PATCH', body: input }, ProjectSchema);
}

export async function deleteProject(projectId: string): Promise<void> {
  return requestJson(`/api/projects/${projectId}`, { method: 'DELETE' }, z.undefined());
}


