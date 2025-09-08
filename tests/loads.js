import http from 'k6/http';
import { check, sleep } from 'k6';
import { BASE_URL, headers, defaultPayload } from "./config/config.js";

export const options = {
    vus: 20,
    duration: '1m',
};

export default function () {
    const res = http.post(BASE_URL, defaultPayload, { headers });

    check(res, {
        'status is 200': (r) => r.status === 200,
        'has blocks': (r) => Array.isArray(r.json('blocks')) && r.json('blocks').length > 0,
    });

    sleep(1);
}
