# Changelog

All notable changes to this project will be documented in this file.

The format is based on Keep a Changelog and this project follows Semantic Versioning.

## [0.2.0] - 2026-03-18

### Changed

- Updated event model to match actual contract: replaced `event_id` with composite uniqueness key (`user_id|timestamp|event_name`), added `campaign_id` and `tags` fields, changed `timestamp` to Unix timestamp (`int64`).
- Tightened validation to require all fields except optional `metadata` object and enforce non-empty tags array.

### Removed

- Removed `event_id` from event request model.

## [0.1.1] - 2026-03-18

### Changed

- Updated module Go version to `1.25.0` and aligned the Docker builder image for compatibility.
- Removed obsolete Compose metadata and cleaned up environment template duplication.
- Improved `README.md` with clearer development, endpoint, and dependency documentation.
- Added explicit request parsing and input validation flow for `POST /events` with standardized `400` invalid-request responses.

### Removed

- Removed redundant environment example file.

## [0.1.0] - 2026-03-18

### Added

- Initial Go backend project scaffold with `cmd/api` entrypoint and pragmatic `internal` package layout.
- Environment-driven configuration, structured logging, and graceful HTTP server startup/shutdown flow.
- Stub HTTP routes and middleware for `GET /health`, `POST /events`, and `GET /metrics`.
- Service/repository/worker/idempotency placeholders wired for future ClickHouse and Redis-backed implementation.
- Development assets including Docker/Compose setup, Makefile commands, starter OpenAPI file, and baseline test setup.
