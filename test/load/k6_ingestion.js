import http from 'k6/http';
import { check } from 'k6';

import { defaultHeaders, getBaseUrl, getDurationEnv, getNumberEnv } from './lib/config.js';
import { makeEvent } from './lib/data.js';

const BASE_URL = getBaseUrl();

export const options = {
  vus: getNumberEnv('VUS', 10),
  duration: getDurationEnv('30s'),
  thresholds: {
    http_req_failed: ['rate<0.05'],
    http_req_duration: ['p(95)<1000'],
  },
};

export default function () {
  const payload = JSON.stringify(makeEvent());
  const res = http.post(`${BASE_URL}/events`, payload, {
    headers: defaultHeaders(),
  });

  check(res, {
    'ingestion status is accepted or duplicate': (r) => r.status === 202 || r.status === 409,
    'ingestion response is json-ish': (r) => {
      const body = r.body || '';
      return body.includes('status') || body.includes('message') || body.includes('error');
    },
  });
}
