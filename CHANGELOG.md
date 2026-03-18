# Changelog

All notable changes to this project will be documented in this file.

The format is based on Keep a Changelog and this project follows Semantic Versioning.

## [0.1.1] - 2026-03-18

### Changed

- Updated module Go version to `1.25.0` and aligned the Docker builder image for compatibility.
- Removed obsolete Compose metadata and cleaned up environment template duplication.
- Improved `README.md` with clearer development, endpoint, and dependency documentation.

### Removed

- Removed redundant environment example duplication.

## [0.1.0] - 2026-03-18

### Added

- Initial Go backend project scaffold with `cmd/api` entrypoint and pragmatic `internal` package layout.
- Environment-driven configuration, structured logging, and graceful HTTP server startup/shutdown flow.
- Stub HTTP routes and middleware for `GET /health`, `POST /events`, and `GET /metrics`.
- Service/repository/worker/idempotency placeholders wired for future ClickHouse and Redis-backed implementation.
- Development assets including Docker/Compose setup, Makefile commands, starter OpenAPI file, and baseline test setup.
