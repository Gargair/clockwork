import { z } from 'zod';
import { requestJson } from './http';

const HealthSchema = z.object({
  status: z.string(),
});

export type HealthResponse = z.infer<typeof HealthSchema>;

export async function getHealth(): Promise<HealthResponse> {
  return requestJson('/healthz', undefined, HealthSchema);
}
