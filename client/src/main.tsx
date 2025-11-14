import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import App from './app/App.tsx';
import ErrorBoundary from './app/ErrorBoundary.tsx';
import Home from './pages/Home.tsx';
import './style.css';

const root = document.getElementById('app');
if (!root) {
  throw new Error('Root element #app not found');
}

ReactDOM.createRoot(root).render(
  <React.StrictMode>
    <ErrorBoundary>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<App />}>
            <Route index element={<Home />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </ErrorBoundary>
  </React.StrictMode>,
);
