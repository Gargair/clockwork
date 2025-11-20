import { z } from 'zod';
import { requestJson } from './http';
import { type TimeEntry, TimeEntrySchema, TimeEntryListSchema } from '../types';

export async function startTimer(categoryId: string): Promise<TimeEntry> {
  return requestJson(
    '/api/time/start',
    {
      method: 'POST',
      body: { categoryId },
    },
    TimeEntrySchema,
  );
}

/**
 * Stops the current active timer.
 * @returns The stopped time entry with final `stoppedAt` and `durationSeconds`.
 * @throws {ApiError} Throws `ApiError` with `code: "no_active_timer"` and `status: 409` if there is no active timer.
 */
export async function stopTimer(): Promise<TimeEntry> {
  return requestJson(
    '/api/time/stop',
    {
      method: 'POST',
    },
    TimeEntrySchema,
  );
}

export async function getActiveTimer(): Promise<TimeEntry | null> {
  return requestJson(
    '/api/time/active',
    {
      method: 'GET',
    },
    z.union([TimeEntrySchema, z.null()]),
  );
}

export interface ListEntriesParams {
  categoryId: string;
  from?: string;
  to?: string;
}

export async function listEntries(params: ListEntriesParams): Promise<TimeEntry[]> {
  const searchParams = new URLSearchParams();
  searchParams.set('categoryId', params.categoryId);
  if (params.from !== undefined) {
    searchParams.set('from', params.from);
  }
  if (params.to !== undefined) {
    searchParams.set('to', params.to);
  }

  const path = `/api/time/entries?${searchParams.toString()}`;

  return requestJson(
    path,
    {
      method: 'GET',
    },
    TimeEntryListSchema,
  );
}


