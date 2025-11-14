import { type ZodType } from 'zod';
import { API_BASE_URL } from './config';
import { ErrorResponseSchema, type ErrorResponse } from '../types';

export interface RequestOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';
  headers?: Record<string, string>;
  body?: unknown;
  signal?: AbortSignal;
}

export class ApiError extends Error {
  public readonly status: number;
  public readonly code?: string;
  public readonly requestId?: string;
  public readonly responseBody?: unknown;

  constructor(
    message: string,
    status: number,
    details?: { code?: string; requestId?: string; responseBody?: unknown },
  ) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.code = details?.code;
    this.requestId = details?.requestId;
    this.responseBody = details?.responseBody;
  }
}

function isAbsoluteUrl(path: string): boolean {
  return /^https?:\/\//i.test(path) || path.startsWith('//');
}

function buildUrl(path: string): string {
  if (isAbsoluteUrl(path)) return path;
  const base = API_BASE_URL.replace(/\/+$/, '');
  const suffix = path.startsWith('/') ? path : `/${path}`;
  return `${base}${suffix}`;
}

export async function requestJson<T>(
  path: string,
  options: RequestOptions | undefined,
  schema: ZodType<T, unknown>,
): Promise<T> {
  const url = buildUrl(path);
  const method = options?.method ?? 'GET';

  const headers: Record<string, string> = {
    Accept: 'application/json',
    ...options?.headers,
  };

  let bodyInit: BodyInit | undefined;
  if (options?.body !== undefined) {
    if (!headers['Content-Type']) {
      headers['Content-Type'] = 'application/json';
    }
    bodyInit = JSON.stringify(options.body);
  }

  const response = await fetch(url, {
    method,
    headers,
    body: bodyInit,
    signal: options?.signal,
  });

  // Handle 204 No Content explicitly
  if (response.status === 204) {
    return schema.parse(undefined);
  }

  const text = await response.text();
  const maybeJson: unknown = text.length ? safeJsonParse(text) : null;

  if (!response.ok) {
    // Try to parse structured error
    let parsedError: ErrorResponse | undefined;
    try {
      parsedError = ErrorResponseSchema.parse(maybeJson);
    } catch {
      parsedError = undefined;
    }

    if (parsedError) {
      throw new ApiError(`${parsedError.code}: ${parsedError.message}`, response.status, {
        code: parsedError.code,
        requestId: parsedError.requestId,
        responseBody: maybeJson,
      });
    }

    throw new ApiError(`HTTP ${response.status}: ${response.statusText}`, response.status, {
      responseBody: text.length ? text : undefined,
    });
  }

  // Success: validate with provided schema
  return schema.parse(maybeJson);
}

function safeJsonParse(value: string): unknown {
  try {
    return JSON.parse(value) as unknown;
  } catch {
    return value;
  }
}
