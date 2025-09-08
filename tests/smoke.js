import http from 'k6/http';
import { check } from 'k6';
import { BASE_URL } from "./config/config.js";

export default function () {
    const res = http.get(`${BASE_URL}/healthz`);
    check(res, { 'status is 200': (r) => r.status === 200 });
}
