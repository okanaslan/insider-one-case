import http from 'k6/http';
import { check } from 'k6';

import { defaultHeaders, getBaseUrl, getDurationEnv, getNumberEnv } from './lib/config.js';
import { makeEvent } from './lib/data.js';

const BASE_URL = getBaseUrl();

export const options = {
    stages: [
        { duration: '30s', target: 100 },   // Ramp up
        { duration: '2m', target: 200 },    // Main test at 200 VUs
        { duration: '30s', target: 0 },     // Ramp down
    ],
    
    thresholds: {
        http_req_failed: ['rate<0.01'],          // Stricter: <1% failures
        http_req_duration: ['p(99)<2000'],       // p99 under 2s
        'http_req_duration{status:202}': ['p(95)<100'],  // Happy path <100ms
        'http_req_duration{status:409}': ['p(95)<50'],   // Duplicates fast
        checks: ['rate>0.999'],                  // >99.9% checks pass
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
