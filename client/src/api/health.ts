import { z } from 'zod';
import { requestJson } from './http';

const HealthSchema = z.object({
  db: z.enum(['up', 'down']),
  time: z.iso.datetime(),
  ok: z.boolean(),
});

export type HealthResponse = z.infer<typeof HealthSchema>;

export async function getHealth(): Promise<HealthResponse> {
  return requestJson('/healthz', undefined, HealthSchema);
}
