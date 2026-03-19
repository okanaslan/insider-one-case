# Load Tests (k6)

This package contains practical k6 scripts for load testing the current backend API.

## Prerequisites

- k6 installed locally
- Backend running at `http://localhost:8080` (or set `BASE_URL`)
- ClickHouse and Redis running when testing full ingest + metrics behavior

## API Assumptions

These scripts target the current repository contract:

- `POST /events`
- `GET /metrics`
- `GET /health`
- `POST /events/bulk` (prepared for future use; may currently return `404`)

Event payload shape used by scripts:

```json
{
  "event_name": "purchase",
  "channel": "mobile",
  "campaign_id": "cmp_1",
  "user_id": "user_1",
  "timestamp": 1710000000,
  "tags": ["promo"],
  "metadata": {
    "amount": 120
  }
}
```

## Scripts

- `test/load/k6_ingestion.js`: single event ingestion load (`POST /events`)
- `test/load/k6_bulk.js`: bulk ingestion load (`POST /events/bulk`)
- `test/load/k6_metrics.js`: metrics query load (`GET /metrics`)
- `test/load/k6_mixed.js`: mixed workload (single ingestion + bulk + metrics)

## Run Commands

```bash
k6 run test/load/k6_ingestion.js
k6 run test/load/k6_bulk.js
k6 run test/load/k6_metrics.js
k6 run test/load/k6_mixed.js
```

## Environment Variables

- `BASE_URL` default: `http://localhost:8080`
- `VUS` default: script-specific
- `DURATION` default: `30s`
- `BULK_SIZE` default: `50` (or `20` in mixed scenario)
- `METRICS_EVENT_NAME` default: `purchase`
- `GROUP_BY` optional for metrics script (`channel`)
- `METRICS_FROM`, `METRICS_TO` optional unix-second bounds for metrics

Examples:

```bash
BASE_URL=http://localhost:8080 VUS=20 DURATION=1m k6 run test/load/k6_ingestion.js
BASE_URL=http://localhost:8080 BULK_SIZE=100 k6 run test/load/k6_bulk.js
BASE_URL=http://localhost:8080 METRICS_EVENT_NAME=purchase GROUP_BY=channel k6 run test/load/k6_metrics.js
```

## Notes

- `k6_bulk.js` and the bulk scenario in `k6_mixed.js` are intentionally ready for future `/events/bulk` support.
- If `/events/bulk` is not implemented yet, those scripts may receive `404`; this is currently treated as expected for compatibility.
- Update scripts if endpoint contracts change (paths, payload fields, status codes, or metrics response shape).
