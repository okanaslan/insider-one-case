# Insider One Case - Go Backend

API for an event ingestion and analytics service.

The service accepts event ingestion traffic over HTTP, deduplicates requests with Redis, buffers accepted events in a bounded in-memory queue, persists batches into ClickHouse, and serves aggregated metrics from ClickHouse.

## Development

### Run Locally

```bash
go mod tidy
make migrate
go run ./cmd/api
```

The service starts on `:8080` by default.

If you are using local infrastructure, copy values from `.env.example` into your environment or a local `.env` file before running commands.

### Run with Docker Compose

```bash
docker compose -f deployments/docker-compose.yml up --build
```

Docker Compose starts ClickHouse and Redis, runs the one-shot migration container, and then starts the API.

### Run ClickHouse Migrations

```bash
make migrate
```

Check migration status:

```bash
make migrate-status
```

### Run Tests

```bash
go test -v ./...
```

### Run Load Tests

The project includes k6 scripts under `test/load` for ingestion, metrics, bulk-shape, and mixed scenarios.

Example:

```bash
BASE_URL=http://localhost:8080 k6 run test/load/k6_ingestion.js
```

## Configuration

Main runtime knobs exposed through environment variables:

- `WORKER_BATCH_SIZE`: target batch size before a ClickHouse flush. Current default: `250`.
- `WORKER_FLUSH_INTERVAL_MS`: max wait before flushing a partial batch. Current default: `250`.
- `INGEST_QUEUE_BUFFER_SIZE`: bounded queue capacity for accepted events. Current default: `5000`.
- `INGEST_ENQUEUE_TIMEOUT_MS`: max wait to push into the in-memory queue before returning overload. Current default: `25`.
- `BULK_MAX_EVENTS_PER_REQUEST`: maximum events allowed per bulk ingestion request. Current default: `500`.
- `BULK_PER_REQUEST_TIMEOUT_MS`: per-request timeout for bulk processing logic (not enforced at handler level, for future use). Current default: `1500`.

## Endpoints

### API Documentation

The API is documented with OpenAPI and can be accessed at `http://localhost:8080/swagger/index.html` when the server is running.

### `GET /health`

Gets a simple health check response.

```curl
curl -X GET http://localhost:8080/health
```

### `POST /events`

Accepts an event for asynchronous processing. The request body must be a JSON object with the following structure:

```json
{
  "event_name": "purchase",
  "channel": "mobile",
  "campaign_id": "cmp_123",
  "user_id": "user_456",
  "timestamp": 1710000000,
  "tags": ["promo", "spring"],
  "metadata": {
    "amount": 120,
    "currency": "TRY"
  }
}
```

Notes:

- `timestamp` is a Unix timestamp in seconds.
- `metadata` is optional.
- Deduplication uses the composite key `user_id|timestamp|event_name`.

```curl
curl -X POST http://localhost:8080/events \
  -H "Content-Type: application/json" \
  -d '{
    "event_name": "purchase",
    "channel": "mobile",
    "campaign_id": "cmp_123",
    "user_id": "user_456",
    "timestamp": 1710000000,
    "tags": ["promo", "spring"],
    "metadata": {
      "amount": 120,
      "currency": "TRY"
    }
  }'
```

Typical responses:

- `202 Accepted`: event accepted for asynchronous processing.
- `409 Conflict`: duplicate event detected.
- `429 Too Many Requests`: bounded queue could not accept the event within the enqueue timeout window.
- `400 Bad Request`: malformed or invalid request payload.
- `500 Internal Server Error`: unexpected ingestion failure.

### `POST /events/bulk`

Accepts multiple events for asynchronous processing in a single request. The request body must be a JSON object with an `events` array. Each event follows the same structure as `POST /events`. Returns a partial-success response showing per-item outcomes.

```json
{
  "events": [
    {
      "event_name": "purchase",
      "channel": "mobile",
      "campaign_id": "cmp_123",
      "user_id": "user_456",
      "timestamp": 1710000000,
      "tags": ["promo", "spring"],
      "metadata": { "amount": 120, "currency": "TRY" }
    },
    {
      "event_name": "view_content",
      "channel": "web",
      "campaign_id": "cmp_124",
      "user_id": "user_457",
      "timestamp": 1710000001,
      "tags": ["summer"],
      "metadata": { "page": "/products" }
    }
  ]
}
```

Notes:

- Minimum 1 event, maximum `BULK_MAX_EVENTS_PER_REQUEST` events per request (default 500).
- Each event is validated and processed independently.
- Response includes per-item status: `accepted`, `duplicate`, `invalid`, `overloaded`, or `error`.

Example response (partial success):

```json
{
  "status": "accepted_partial",
  "summary": {
    "total": 2,
    "accepted": 1,
    "duplicate": 1,
    "invalid": 0,
    "overloaded": 0,
    "error": 0
  },
}
```

Typical responses:

- `202 Accepted`: valid request.
- `400 Bad Request`: invalid (malformed JSON, empty array, oversized array, >1 invalid field per item).
- `500 Internal Server Error`: fatal handler/service failure.

### `GET /metrics`

Returns aggregated metrics for a given event name and time range. Query parameters:

- `event_name` (string, required): event type to query.
- `from` (int64, required): inclusive Unix timestamp lower bound.
- `to` (int64, required): exclusive Unix timestamp upper bound.
- `group_by` (string, optional): currently only `channel` is supported.

```curl
curl -X GET "http://localhost:8080/metrics?event_name=purchase&from=1710000000&to=1710086400&group_by=channel"
```

Example response:

```json
{
  "event_name": "purchase",
  "from": 1710000000,
  "to": 1710086400,
  "total_count": 1280,
  "unique_users": 911,
  "group_by": "channel",
  "groups": [
    {
      "key": "mobile",
      "count": 900,
      "unique_users": 650
    },
    {
      "key": "web",
      "count": 380,
      "unique_users": 261
    }
  ]
}
```

## System Architecture

``` mermaid
flowchart TD
    C[Client / Producer] --> A[Go API Service<br/>Gin HTTP Server]

    subgraph API[Application Layer]
        A --> H[HTTP Handlers]
        H --> S1[Event Service]
        H --> S2[Metrics Service]
        S1 --> V[Validator]
        S1 --> I[Redis Idempotency Store]
        S1 --> Q[Bounded In-Memory Queue / Batcher]
        S2 --> R2[Metrics Repository]
        subgraph Endpoints[Endpoints]
            A --> HEALTH[GET /health]
            A --> EVENTS[POST /events]
            A --> METRICS[GET /metrics]
        end
    end

    subgraph BG[Background Processing]
        Q --> W[Ingest Worker]
        W --> R1[Event Repository]
    end

    subgraph DATA[Data Layer]
        I --> REDIS[(Redis)]
        R1 --> CH[(ClickHouse)]
        R2 --> CH
    end


    EVENTS --> H
    METRICS --> H

    REDIS -. dedup / reserve event .-> I
    CH -. raw events + aggregations .-> R1
    CH -. query metrics .-> R2
```

## System Design Highlights

- Structure follows `golang-standards/project-layout` ideas pragmatically.
- Config loading is centralized in `internal/config` with environment variable support and validation.
- HTTP server setup includes graceful shutdown and structured logging.
- `POST /events` validates the case-specific payload, reserves a dedupe key in Redis, and hands accepted events to an async worker.
- The worker keeps a bounded in-memory queue and flushes to ClickHouse on batch-size threshold, timer, and shutdown drain.
- `GET /metrics` runs ClickHouse-backed totals and optional channel grouping queries.
- Schema changes are handled through an explicit migration command and embedded migration files rather than implicit startup DDL.
- Docker Compose includes a dedicated migration service so local environments converge before the API starts.
- k6 scripts under `test/load` cover ingestion, metrics, and mixed-traffic scenarios for load validation.

## Architectural Decisions

These are the main design decisions we took while developing the service and why they matter.

### 1. Asynchronous ingestion instead of synchronous ClickHouse writes

`POST /events` does not write directly to ClickHouse on the request path. Accepted events are enqueued and persisted by the worker in batches.

Why:

- keeps request latency low under normal load,
- reduces per-request write overhead,
- makes ClickHouse inserts more efficient through batching.

Tradeoff:

- ingestion becomes eventually consistent rather than immediately durable at response time.

### 2. Bounded queue with short enqueue backpressure window

The in-memory queue is intentionally bounded (default 5000 events) and enqueue waits only for a short configured timeout (default 25ms) before returning overload.

Why:

- avoids unbounded memory growth,
- protects the process during traffic spikes,
- turns saturation into an explicit `429` instead of deeper internal failure,
- reflects actual queue contention directly to clients for transparent backpressure.

Operational behavior:

- When the queue is full, the service returns `429 Too Many Requests` to the caller.
- Clients should retry with exponential backoff; `429` indicates transient queue saturation, not a permanent failure.
- Load testing has validated ~23k req/s sustained throughput with p95 latency <15ms under this model.
- Adjust `INGEST_QUEUE_BUFFER_SIZE`, `WORKER_BATCH_SIZE`, and `WORKER_FLUSH_INTERVAL_MS` based on your throughput targets.

Tradeoff:

- under sustained overload, some otherwise valid requests are rejected and must be retried by clients.

### 3. Redis-backed idempotency before enqueue

The service uses Redis to reserve a composite uniqueness key (`user_id|timestamp|event_name`) before the event enters the async pipeline. The reservation carries a 24-hour TTL.

Why:

- stops duplicate events from creating duplicated downstream writes,
- keeps deduplication logic off the ClickHouse query path,
- gives a clear `409 Conflict` contract to callers for duplicate detection.

Operational behavior:

- When an event arrives, the service attempts to reserve its uniqueness key in Redis.
- If the key already exists, the event is rejected with a `409 Conflict` response.
- If reservation succeeds, the event proceeds to the in-memory queue.
- If Redis is unavailable, the service proceeds with a warning log; eventual duplicates may write to ClickHouse but will be detected on replay.

This design prioritizes availability and forward progress over strict single-write guarantees.

Tradeoff:

- deduplication depends on Redis availability and TTL-based reservation semantics; Redis outages allow potential duplicates to pass through.

### 4. ClickHouse as both raw event store and analytics backend

Raw events are persisted in ClickHouse and metrics are queried from the same store.

Why:

- keeps the architecture small,
- fits append-heavy event workloads well,
- allows aggregate queries without introducing a second analytics system.

Tradeoff:

- application code needs to be explicit about batching, schema evolution, and query shape.

### 5. Explicit migrations instead of schema creation at API startup

Schema changes are applied through `cmd/migrate` and Goose migrations, not hidden inside API boot.

Why:

- startup behavior is more predictable,
- schema changes become versioned and reviewable,
- local and containerized environments follow the same migration flow.

Tradeoff:

- there is an extra operational step outside the API process itself.

### 6. Prefer backpressure over fixed application-level rate limiting

We considered adding in-process request rate limiting but intentionally kept overload protection centered on queue backpressure and bounded batching.

Why:

- it targets the actual bottleneck directly,
- it avoids rejecting legitimate bursts prematurely,
- it adds almost no extra request-path policy complexity.

Tradeoff:

- protection is reactive to internal pressure rather than enforcing a fixed front-door ceiling.

### 7. No-retry ingestion worker for simplicity

The ingestion worker flushes batches to ClickHouse on three triggers: batch-size threshold, timer interval, or graceful shutdown. **If a flush fails, the batch is dropped and not retried** so events in that batch are lost.

Why:

- eliminates retry complexity, dead-letter queue management, and distributed consensus for retried events,
- guarantees forward progress; the worker does not block,
- keeps operational logic simple for a case study.

Operational behavior:

- Worker logs the error with batch count and error details when a flush fails.
- Subsequent batches continue processing normally after a failed flush.

Tradeoff:

- transient ClickHouse failures or network issues can cause event loss.
- In production, consider adding a dead-letter queue, retry logic with exponential backoff, or dual-write to a persistent queue before deleting from memory.

### 8. Eventually consistent metrics without caching

Metrics are queried directly from ClickHouse without application-level caching. There is a lag between event acceptance (in-memory queue) and metrics visibility (ClickHouse persistence).

Why:

- avoids cache coherency complexity and stale-data expiration logic,
- keeps metrics semantically accurate to the ClickHouse state,
- simplifies the ingestion path by avoiding metrics updates on the critical path.

Operational behavior:

- Under normal load, metrics lag is <1 second.
- Under backlog conditions, lag can be several seconds depending on worker batch size, flush interval, and ClickHouse write throughput.
- Adjust `WORKER_BATCH_SIZE` and `WORKER_FLUSH_INTERVAL_MS` to tune freshness vs. batching efficiency.

Tradeoff:

- metrics visibility is delayed by the async pipeline; they are eventually rather than immediately consistent with ingestion.
