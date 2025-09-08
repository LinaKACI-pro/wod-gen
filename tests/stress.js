import http from 'k6/http';
import { check } from 'k6';
import { BASE_URL, headers, defaultPayload } from "./config/config.js";

export const options = {
    stages: [
        { duration: '30s', target: 10 },
        { duration: '30s', target: 50 },
        { duration: '30s', target: 100 },
        { duration: '30s', target: 0 },
    ],
};

export default function () {
    const res = http.post(BASE_URL, defaultPayload, { headers });

    check(res, {
        'status is 200': (r) => r.status === 200,
    });
}
