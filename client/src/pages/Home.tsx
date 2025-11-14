import { useEffect, useState, type JSX } from 'react';
import { getHealth, type HealthResponse } from '../api/health';
import { ApiError } from '../api/http';

type LoadStatus = 'idle' | 'loading' | 'success' | 'error';

interface ErrorState {
  message: string;
  code?: string;
  requestId?: string;
}

export default function Home(): JSX.Element {
  const [status, setStatus] = useState<LoadStatus>('idle');
  const [health, setHealth] = useState<HealthResponse | null>(null);
  const [error, setError] = useState<ErrorState | null>(null);

  useEffect(() => {
    let isMounted = true;
    setStatus('loading');
    getHealth()
      .then((res) => {
        if (!isMounted) return;
        setHealth(res);
        setError(null);
        setStatus('success');
      })
      .catch((err: unknown) => {
        if (!isMounted) return;
        const base: ErrorState =
          err instanceof Error ? { message: err.message } : { message: 'Unknown error' };
        if (err instanceof ApiError) {
          base.code = err.code;
          base.requestId = err.requestId;
        }
        setError(base);
        setHealth(null);
        setStatus('error');
      });
    return () => {
      isMounted = false;
    };
  }, []);

  return (
    <section>
      <h1>Welcome to Clockwork</h1>
      <p>Scaffold complete. Start building the app.</p>

      <div style={{ marginTop: '1rem' }}>
        <h2>API Health</h2>
        {status === 'loading' ? <p>Checking health…</p> : null}
        {status === 'success' && health ? <p>Status: {health.ok ? 'OK' : 'Error'}</p> : null}
        {status === 'error' && error ? (
          <div role="alert">
            <p>Error: {error.message}</p>
            {error.code ? <p>Code: {error.code}</p> : null}
            {error.requestId ? <p>Request ID: {error.requestId}</p> : null}
          </div>
        ) : null}
      </div>
    </section>
  );
}
