import type { JSX } from 'react';
import { Outlet, Link } from 'react-router-dom';

export default function App(): JSX.Element {
  return (
    <div className="app">
      <header>
        <nav>
          <strong>Clockwork</strong>
          <Link to="/">Home</Link>
        </nav>
      </header>
      <main>
        <Outlet />
      </main>
      <footer>
        <small>© {new Date().getFullYear()} Clockwork</small>
      </footer>
    </div>
  );
}
