# Clockwork Documentation

Clockwork is an application to track time for projects and their categories. Documentation is organized into focused pages below.

## Table of contents
- [Overview and MVP](#overview-and-mvp)
- [Architecture](architecture.md)
- [Domain model](domain-model.md)
- [API](api.md)
- [Development](development.md)
- [Testing](testing.md)
- [Deployment](deployment.md)
- [Security](security.md)
- [Roadmap](roadmap.md)

## Overview and MVP
- Client–server application
  - Server written in Go
  - Client written in TypeScript (Vite + React SPA)
  - In production, the Go server serves the client as static files
  - Intended to run in a Kubernetes cluster
  - Data persisted in a PostgreSQL database
- Server development follows test-driven development (TDD)

MVP scope (single user):
- Create, edit, and delete projects
- Create, edit, and delete categories
- Categories are hierarchical and belong to a project
- A category is not moved to a different project once created
- Track time via Start/Stop controls
  - Time is tracked on categories only
  - Only one category can be actively tracked at any given moment

See also: `../LICENSE`.

