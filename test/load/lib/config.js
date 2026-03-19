export function getBaseUrl() {
    return __ENV.BASE_URL || 'http://localhost:8080';
}

export function getNumberEnv(name, defaultValue) {
    const value = __ENV[name];
    if (!value) {
        return defaultValue;
    }

    const parsed = Number(value);
    if (Number.isNaN(parsed)) {
        return defaultValue;
    }

    return parsed;
}

export function getDurationEnv(defaultValue) {
    return __ENV.DURATION || defaultValue;
}

export function defaultHeaders() {
    return {
        'Content-Type': 'application/json',
    };
}

export function getNowUnixSeconds() {
    return Math.floor(Date.now() / 1000);
}

export function getMetricsWindow() {
    const now = getNowUnixSeconds();
    const from = getNumberEnv('METRICS_FROM', now - 3600);
    const to = getNumberEnv('METRICS_TO', now + 3600);
    return { from, to };
}
