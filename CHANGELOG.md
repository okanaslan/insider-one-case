# Changelog

All notable changes to this project will be documented in this file.

## [0.8.2] - 2026-03-19

- Improved ingestion load-test realism by expanding synthetic event/cardinality dimensions (event names, channels, campaigns, and multi-tag combinations).
- Tightened ingestion thresholds to emphasize stricter SLOs (`http_req_failed < 1%`, `p99 < 2s`) and added status-specific latency assertions for `202` and `409` responses.

## [0.8.1] - 2026-03-19

- Removed `ALLOW_START_WITHOUT_INFRA` configuration option to enforce strict infrastructure requirements at startup.
- API now fails immediately if ClickHouse or Redis connections cannot be established, improving operational clarity.

## [0.8.0] - 2026-03-19

- Added central normalization for queue and worker tuning with explicit defaults and min/max bounds.
- Tightened default tuning to `WORKER_BATCH_SIZE=250`, `WORKER_FLUSH_INTERVAL_MS=250`, and `INGEST_QUEUE_BUFFER_SIZE=5000` while keeping enqueue timeout at `25ms`.
- Normalized direct worker configs in the constructor so tests and manual wiring cannot create unbounded or extreme queue settings.
- Added config and worker tests covering defaulting and clamping behavior.

## [0.7.0] - 2026-03-19

- Added bounded enqueue backpressure window via `INGEST_ENQUEUE_TIMEOUT_MS` (default `25ms`) to reduce immediate drops under short queue contention.
- Made worker enqueue context-aware and return timeout/cancel errors explicitly.
- Mapped enqueue pressure to service-level overload sentinel and `POST /events` `429 rate_limited` response.
- Added/updated worker, service, and handler tests for enqueue timeout and overload behavior.

## [0.6.1] - 2026-03-19

- Added event repository tests for nil-connection insert behavior and metadata serialization correctness.
- Improved core-flow test coverage without changing runtime behavior.

## [0.6.0] - 2026-03-19

- Redesigned `GET /metrics` contract to require `event_name`, `from`, and `to`, with optional `group_by=channel` validation and standardized `400`/`500` JSON errors.
- Implemented real ClickHouse-backed metrics queries for total event count, unique user count, and optional grouped aggregation by channel.
- Updated metrics response model to return totals and optional grouped results in case-aligned shape.
- Added tests for metrics handler validation paths, metrics service orchestration, and repository nil-connection behavior.

## [0.5.1] - 2026-03-19

- Improved migration startup reliability by adding ClickHouse connection retry logic in `cmd/migrate`.
- Added ClickHouse healthcheck and updated Compose dependencies so migrations wait for healthy ClickHouse before running.
- Updated Docker image build to include the `migrate` binary for one-shot migration execution in Docker Compose.

## [0.5.0] - 2026-03-19

- Added Goose-based ClickHouse migrations with an embedded `cmd/migrate` command and initial events table migration.
- Moved schema management out of API startup into explicit migration commands (`make migrate`, `make migrate-status`).
- Centralized optional `.env` loading in shared config so both API and migration commands read the same environment settings.

## [0.4.0] - 2026-03-19

- Added ClickHouse event table initialization and batch event persistence.
- Implemented real worker flush path with repository-backed batch writes, timer/size/shutdown flush handling, and queue buffer configuration.
- Added worker tests covering batch-size flush, timer flush, and shutdown flush behavior.

## [0.3.0] - 2026-03-19

- Implemented `POST /events` happy path with Redis deduplication, in-memory queue, and `202 Accepted` response.
- In-process ingestion worker with batch flushing on size threshold, ticker interval, and graceful shutdown drain.
- Service decoupled from worker via `EventEnqueuer` interface; standardized `409`/`500` error response shapes.

## [0.2.0] - 2026-03-18

- Updated event model to match actual contract: replaced `event_id` with composite uniqueness key (`user_id|timestamp|event_name`), added `campaign_id` and `tags` fields, changed `timestamp` to Unix timestamp (`int64`).
- Tightened validation to require all fields except optional `metadata` object and enforce non-empty tags array.
- Removed `event_id` from event request model.

## [0.1.1] - 2026-03-18

- Updated module Go version to `1.25.0` and aligned the Docker builder image for compatibility.
- Removed obsolete Compose metadata and cleaned up environment template duplication.
- Improved `README.md` with clearer development, endpoint, and dependency documentation.
- Added explicit request parsing and input validation flow for `POST /events` with standardized `400` invalid-request responses.
- Removed redundant environment example file.

## [0.1.0] - 2026-03-18

- Initial Go backend project scaffold with `cmd/api` entrypoint and pragmatic `internal` package layout.
- Environment-driven configuration, structured logging, and graceful HTTP server startup/shutdown flow.
- Stub HTTP routes and middleware for `GET /health`, `POST /events`, and `GET /metrics`.
- Service/repository/worker/idempotency placeholders wired for future ClickHouse and Redis-backed implementation.
- Development assets including Docker/Compose setup, Makefile commands, starter OpenAPI file, and baseline test setup.
