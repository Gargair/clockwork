import { z } from 'zod';

export const ProjectSchema = z.object({
  id: z.string().uuid(),
  name: z.string(),
  description: z.string().optional(),
  createdAt: z.string().datetime(),
  updatedAt: z.string().datetime(),
});

export const ProjectListSchema = z.array(ProjectSchema);

export const CategorySchema = z.object({
  id: z.string().uuid(),
  projectId: z.string().uuid(),
  parentCategoryId: z.string().uuid().nullable(),
  name: z.string(),
  description: z.string().optional(),
  createdAt: z.string().datetime(),
  updatedAt: z.string().datetime(),
});

export const CategoryListSchema = z.array(CategorySchema);

export const TimeEntrySchema = z.object({
  id: z.string().uuid(),
  categoryId: z.string().uuid(),
  startedAt: z.string().datetime(),
  stoppedAt: z.string().datetime().nullable(),
  durationSeconds: z.number().int().nonnegative().nullable(),
  createdAt: z.string().datetime(),
  updatedAt: z.string().datetime(),
});

export const TimeEntryListSchema = z.array(TimeEntrySchema);

export const ErrorResponseSchema = z.object({
  code: z.string(),
  message: z.string(),
  requestId: z.string(),
});
