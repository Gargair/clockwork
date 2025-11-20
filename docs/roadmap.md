# Roadmap

- Multi-user support with authentication
- Reporting and analytics (daily/weekly summaries, per-category totals)
- Data export/import (CSV/JSON)
- Integrations (issue trackers, calendars)
- Offline-friendly client
- Role-based access control (post multi-user)

## Post-initial release polish

- Adopt Tailwind CSS for styling to standardize design tokens, speed up UI development, and reduce custom CSS surface area. Gradually migrate existing styles, starting with layout and common components.
- Replace native `Date` usage with a time library (e.g., `date-fns`, `dayjs`, or `luxon`) for better date handling, timezone support, and formatting in client components (`TimerControls.tsx`, `EntryList.tsx`, and related hooks)
- Replace state handling through hooks with mobX and proper observables