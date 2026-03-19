import http from 'k6/http';
import { check } from 'k6';

import {
  defaultHeaders,
  getBaseUrl,
  getDurationEnv,
  getMetricsWindow,
  getNumberEnv,
} from './lib/config.js';
import { makeBulkEvents, makeEvent } from './lib/data.js';

const BASE_URL = getBaseUrl();
const BULK_SIZE = getNumberEnv('BULK_SIZE', 20);
const METRICS_EVENT_NAME = __ENV.METRICS_EVENT_NAME || 'purchase';

export const options = {
  scenarios: {
    ingestion: {
      executor: 'constant-vus',
      vus: getNumberEnv('INGESTION_VUS', 7),
      duration: getDurationEnv('30s'),
      exec: 'runIngestion',
    },
    bulk: {
      executor: 'constant-vus',
      vus: getNumberEnv('BULK_VUS', 2),
      duration: getDurationEnv('30s'),
      exec: 'runBulk',
    },
    metrics: {
      executor: 'constant-vus',
      vus: getNumberEnv('METRICS_VUS', 1),
      duration: getDurationEnv('30s'),
      exec: 'runMetrics',
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.10'],
    http_req_duration: ['p(95)<1500'],
  },
};

export function runIngestion() {
  const payload = JSON.stringify(makeEvent());
  const res = http.post(`${BASE_URL}/events`, payload, {
    headers: defaultHeaders(),
  });

  check(res, {
    'mixed ingestion status is accepted or duplicate': (r) => r.status === 202 || r.status === 409,
  });
}

export function runBulk() {
  const body = JSON.stringify({ events: makeBulkEvents(BULK_SIZE) });
  const res = http.post(`${BASE_URL}/events/bulk`, body, {
    headers: defaultHeaders(),
  });

  check(res, {
    'mixed bulk status is 200/202 (or 404 if endpoint not implemented)': (r) =>
      r.status === 200 || r.status === 202 || r.status === 404,
  });
}

export function runMetrics() {
  const window = getMetricsWindow();
  const url = `${BASE_URL}/metrics?event_name=${encodeURIComponent(
    METRICS_EVENT_NAME,
  )}&from=${window.from}&to=${window.to}&group_by=channel`;

  const res = http.get(url);

  check(res, {
    'mixed metrics status is 200': (r) => r.status === 200,
    'mixed metrics has total_count': (r) => (r.body || '').includes('total_count'),
  });
}
