import type { TimeEntry } from '../types';

export async function startTimer(categoryId: string): Promise<TimeEntry> {
  throw new Error('Not implemented');
}

export async function stopTimer(): Promise<TimeEntry> {
  throw new Error('Not implemented');
}

export async function getActiveTimer(): Promise<TimeEntry | null> {
  throw new Error('Not implemented');
}

export interface ListEntriesParams {
  categoryId: string;
  from?: string;
  to?: string;
}

export async function listEntries(params: ListEntriesParams): Promise<TimeEntry[]> {
  throw new Error('Not implemented');
}


