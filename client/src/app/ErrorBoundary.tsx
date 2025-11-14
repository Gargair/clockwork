import { Component, type ErrorInfo, type ReactNode } from 'react';
import { ApiError } from '../api/http';

export interface ErrorBoundaryProps {
  children: ReactNode;
}

interface ErrorBoundaryState {
  hasError: boolean;
  error?: unknown;
}

export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  public state: ErrorBoundaryState = { hasError: false };

  public static getDerivedStateFromError(error: unknown): ErrorBoundaryState {
    return { hasError: true, error };
  }

  public componentDidCatch(error: unknown, errorInfo: ErrorInfo): void {
    // Consider wiring this to a logging/telemetry endpoint
    // eslint-disable-next-line no-console
    console.error('ErrorBoundary caught an error', error, errorInfo);
  }

  public render(): ReactNode {
    if (!this.state.hasError) {
      return this.props.children;
    }

    const err = this.state.error;
    const isApiError = err instanceof ApiError;
    const requestId = isApiError ? err.requestId : undefined;
    const code = isApiError ? err.code : undefined;
    const message = extractErrorMessage(err);

    return (
      <div role="alert" style={{ padding: '1rem', color: '#991b1b', background: '#fee2e2' }}>
        <h2 style={{ margin: 0, marginBottom: '0.5rem' }}>An unexpected error occurred</h2>
        <p style={{ margin: 0, marginBottom: '0.5rem' }}>{message}</p>
        {code ? <p style={{ margin: 0, marginBottom: '0.25rem' }}>Code: {code}</p> : null}
        {requestId ? <p style={{ margin: 0 }}>Request ID: {requestId}</p> : null}
      </div>
    );
  }
}

export default ErrorBoundary;

function extractErrorMessage(error: unknown): string {
  if (error instanceof Error) {
    return error.message;
  }
  if (typeof error === 'string') {
    return error;
  }
  return 'Something went wrong.';
}