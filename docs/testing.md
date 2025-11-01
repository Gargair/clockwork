# Testing

## Principles
- TDD for server development: write failing tests first, then implement and refactor
- Favor fast, deterministic tests with clear assertions

## Server tests
- Unit tests for domain logic and invariants (single active timer, category hierarchy)
- Integration tests for HTTP handlers, persistence, and edge cases
- Command: `go test ./...`

## Client tests
- Component tests for critical UI flows (start/stop, listing entries)
- Prefer headless DOM testing where possible
- Command: `npm test` (to be defined in client scripts)

