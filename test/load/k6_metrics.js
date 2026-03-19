import http from 'k6/http';
import { check } from 'k6';

import { getBaseUrl, getDurationEnv, getMetricsWindow, getNumberEnv } from './lib/config.js';

const BASE_URL = getBaseUrl();
const EVENT_NAME = __ENV.METRICS_EVENT_NAME || 'purchase';
const GROUP_BY = __ENV.GROUP_BY || '';

export const options = {
    vus: getNumberEnv('VUS', 5),
    duration: getDurationEnv('30s'),
    thresholds: {
        http_req_failed: ['rate<0.05'],
        http_req_duration: ['p(95)<1000'],
    },
};

export default function () {
    const window = getMetricsWindow();
    const params = [`event_name=${encodeURIComponent(EVENT_NAME)}`, `from=${window.from}`, `to=${window.to}`];
    if (GROUP_BY) {
        params.push(`group_by=${encodeURIComponent(GROUP_BY)}`);
    }

    const url = `${BASE_URL}/metrics?${params.join('&')}`;
    const res = http.get(url);

    check(res, {
        'metrics status is 200': (r) => r.status === 200,
        'metrics has total_count': (r) => (r.body || '').includes('total_count'),
        'metrics has unique_users': (r) => (r.body || '').includes('unique_users'),
        'metrics grouped response when group_by set': (r) => !GROUP_BY || (r.body || '').includes('groups'),
    });
}
