import { z } from 'zod';

export const ProjectSchema = z.object({
  id: z.uuid(),
  name: z.string(),
  description: z.string().optional(),
  createdAt: z.iso.datetime(),
  updatedAt: z.iso.datetime(),
});

export const ProjectListSchema = z.array(ProjectSchema);

export const CategorySchema = z.object({
  id: z.uuid(),
  projectId: z.uuid(),
  parentCategoryId: z.uuid().nullable(),
  name: z.string(),
  description: z.string().optional(),
  createdAt: z.iso.datetime(),
  updatedAt: z.iso.datetime(),
});

export const CategoryListSchema = z.array(CategorySchema);

export const TimeEntrySchema = z.object({
  id: z.uuid(),
  categoryId: z.uuid(),
  startedAt: z.iso.datetime(),
  stoppedAt: z.iso.datetime().nullable(),
  durationSeconds: z.number().int().nonnegative().nullable(),
  createdAt: z.iso.datetime(),
  updatedAt: z.iso.datetime(),
});

export const TimeEntryListSchema = z.array(TimeEntrySchema);

export const ErrorResponseSchema = z.object({
  code: z.string(),
  message: z.string(),
  requestId: z.string(),
});
