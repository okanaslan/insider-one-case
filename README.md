# Insider One Case - Go Backend Scaffold

API for an event ingestion and analytics service.

## Run Locally

```bash
go mod tidy
go run ./cmd/api
```

The service starts on `:8080` by default.

## Run with Docker Compose

```bash
docker compose -f deployments/docker-compose.yml up --build
```

## Endpoints (Stub)

- `GET /health`
- `POST /events`
- `GET /metrics`

## System Architectrue

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

## Notes

- Structure follows `golang-standards/project-layout` ideas pragmatically.
- Business logic, SQL, worker ingestion pipeline, and metrics computation are intentionally stubbed.
- ClickHouse and Redis startup checks are present.
- If infra is unavailable and `ALLOW_START_WITHOUT_INFRA=true`, app logs warnings and continues.

## Next Implementation Areas

- Event ingestion queueing and batching flow.
- ClickHouse insert and metrics query SQL.
- Rich validation and idempotency strategy.
- Swagger generation into `docs` via Makefile.
