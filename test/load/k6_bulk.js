import http from 'k6/http';
import { check } from 'k6';

import { defaultHeaders, getBaseUrl, getDurationEnv, getNumberEnv } from './lib/config.js';
import { makeBulkEvents } from './lib/data.js';

const BASE_URL = getBaseUrl();
const BULK_SIZE = getNumberEnv('BULK_SIZE', 50);

export const options = {
    vus: getNumberEnv('VUS', 5),
    duration: getDurationEnv('30s'),
    thresholds: {
        http_req_failed: ['rate<0.20'],
        http_req_duration: ['p(95)<1500'],
    },
};

export default function () {
    const body = JSON.stringify({ events: makeBulkEvents(BULK_SIZE) });

    const res = http.post(`${BASE_URL}/events/bulk`, body, {
        headers: defaultHeaders(),
    });

    check(res, {
        'bulk status is 200/202 (or 404 if endpoint not implemented yet)': (r) =>
            r.status === 200 || r.status === 202 || r.status === 404,
        'bulk response body is non-empty': (r) => (r.body || '').length > 0,
    });
}
