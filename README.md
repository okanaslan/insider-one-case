# Insider One Case - Go Backend

API for an event ingestion and analytics service.

## Development

### Run Locally

```bash
go mod tidy
go run ./cmd/api
```

The service starts on `:8080` by default.

### Run with Docker Compose

```bash
docker compose -f deployments/docker-compose.yml up --build
```

### Run Tests

```bash
go test -v ./...
```

## Endpoints

### API Documentation

The API is documented with OpenAPI and can be accessed at `http://localhost:8080/swagger/index.html` when the server is running.

### `GET /health`

Gets a simple health check response.

```curl
curl -X GET http://localhost:8080/health
```

### `POST /events`

Creates a new event. The request body should be a JSON object with the following structure:

```json
{
  "event_type": "string",
  "user_id": "string",
  "timestamp": "date-time",
  "properties": {
    "key": "value"
  }
}
```

```curl
curl -X POST http://localhost:8080/events \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "user_signup",
    "user_id": "12345",
    "timestamp": "2026-03-18T12:34:56Z",
    "properties": {
      "plan": "pro",
      "referrer": "google"
    }
  }'
```

### `GET /metrics`

Gets aggregated metrics for a given event type and time range. The query parameters are:

- `event_type` (string, required): The type of event to aggregate.
- `start` (string, required, date-time): The start of the time range for the metrics.
- `end` (string, required, date-time): The end of the time range for the metrics.

```curl
curl -X GET http://localhost:8080/metrics?event_type=user_signup&start=2026-03-01T00:00:00Z&end=2026-03-31T23:59:59Z
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
        S1 --> I[Idempotency Store]
        S1 --> Q[In-Memory Queue / Batcher]
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
- Starter routes for health check, event ingestion, and metrics retrieval are defined with placeholders for implementation
- Middleware placeholders for request ID generation, logging, and recovery are included.
- Service and repository layers are stubbed out for event processing and metrics querying.
- ClickHouse and Redis connection bootstraps are included with infra-tolerant startup behavior.
- Worker placeholders in `internal/worker` for ingest loop and batcher are defined.
- Idempotency placeholder store backed by Redis SETNX semantics is included.
- Minimal OpenAPI starter file in `api/openapi.yaml` is included for future API documentation.
- Docker and Docker Compose setup under `deployments` allows for easy local development and testing.
- Developer commands in `Makefile` and starter project documentation in `README.md` provide a clear onboarding path.
- Starter test usage of `testify` in service package demonstrates how to write unit tests with assertions.

## External Packages

- `github.com/gin-gonic/gin`: Used to build clear, lightweight HTTP routing and middleware for API endpoints.
- `github.com/caarlos0/env/v11`: Used to parse environment variables into strongly typed config structs with defaults.
- `github.com/go-playground/validator/v10`: Used to perform request payload validation with simple struct tags.
- `github.com/ClickHouse/clickhouse-go/v2`: Used to connect to ClickHouse for event storage and analytics query execution.
- `github.com/redis/go-redis/v9`: Used to connect to Redis for idempotency and deduplication primitives.
- `github.com/google/uuid`: Used to generate request and event identifiers when they are not supplied.
- `github.com/swaggo/swag`: Used to generate Swagger/OpenAPI docs from Go annotations when enabled.
- `github.com/swaggo/gin-swagger`: Used to expose Swagger UI through a Gin route.
- `github.com/swaggo/files`: Used to serve static Swagger UI assets required by gin-swagger.
- `github.com/stretchr/testify`: Used to make tests more readable with assertions and test helpers.
