import { z } from 'zod';
import {
	ProjectSchema,
	CategorySchema,
	TimeEntrySchema,
	ErrorResponseSchema,
} from './schemas';

export { ProjectSchema, ProjectListSchema } from './schemas';
export { CategorySchema, CategoryListSchema } from './schemas';
export { TimeEntrySchema, TimeEntryListSchema } from './schemas';
export { ErrorResponseSchema } from './schemas';

export type Project = z.infer<typeof ProjectSchema>;
export type Category = z.infer<typeof CategorySchema>;
export type TimeEntry = z.infer<typeof TimeEntrySchema>;
export type ErrorResponse = z.infer<typeof ErrorResponseSchema>;


