import type { JSX } from 'react';
import { Outlet, Link } from 'react-router-dom';

export default function App(): JSX.Element {
	return (
		<div style={{ fontFamily: 'system-ui, -apple-system, Segoe UI, Roboto, Arial, sans-serif' }}>
			<header style={{ padding: '1rem', borderBottom: '1px solid #e5e7eb' }}>
				<nav style={{ display: 'flex', gap: '1rem', alignItems: 'center' }}>
					<strong>Clockwork</strong>
					<Link to="/">Home</Link>
				</nav>
			</header>
			<main style={{ padding: '1rem' }}>
				<Outlet />
			</main>
			<footer style={{ padding: '1rem', borderTop: '1px solid #e5e7eb', marginTop: '2rem' }}>
				<small>© {new Date().getFullYear()} Clockwork</small>
			</footer>
		</div>
	);
}


