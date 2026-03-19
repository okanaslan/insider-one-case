import http from 'k6/http';
import { check } from 'k6';

import { defaultHeaders, getBaseUrl, getDurationEnv, getNumberEnv } from './lib/config.js';
import { makeBulkEvents } from './lib/data.js';

const BASE_URL = getBaseUrl();
const BULK_SIZE = getNumberEnv('BULK_SIZE', 100);

export const options = {
    stages: [
        { duration: '30s', target: 50 },    // Ramp up
        { duration: '2m', target: 100 },    // Main test at 100 VUs
        { duration: '30s', target: 0 },     // Ramp down
    ],
    
    thresholds: {
        http_req_failed: ['rate<0.05'],          // <5% failures for bulk
        http_req_duration: ['p(99)<2000'],       // p99 under 2s
        'http_req_duration{status:202}': ['p(95)<500'],  // Happy path <500ms
        checks: ['rate>0.99'],                   // >99% checks pass
    },
};

export default function () {
    const payload = JSON.stringify({ events: makeBulkEvents(BULK_SIZE) });
    const res = http.post(`${BASE_URL}/events/bulk`, payload, {
        headers: defaultHeaders(),
    });

    check(res, {
        'bulk status is 202': (r) => r.status === 202,
        'bulk response has summary': (r) => (r.body || '').includes('summary'),
        'bulk response has accepted count': (r) => (r.body || '').includes('accepted'),
    });
}
