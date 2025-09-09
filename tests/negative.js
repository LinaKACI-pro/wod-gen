import http from 'k6/http';
import { check } from 'k6';
import { BASE_URL, headers, defaultPayload } from "./config/config.js";

export default function () {
    // Known 400 errors
    // invalid level
    let res = http.post(BASE_URL, JSON.stringify({ level: "invalid", duration_min: 30 }), { headers });
    check(res, { 'invalid level → 400': (r) => r.status === 400 });

    // duration too short
    res = http.post(BASE_URL, JSON.stringify({ level: "beginner", duration_min: 5 }), { headers });
    check(res, { 'invalid duration → 400': (r) => r.status === 400 });

    // Auth tests
    res = http.post(BASE_URL, defaultPayload, {
        headers: { 'Content-Type': 'application/json' },
    });
    check(res, { 'no token → 401': (r) => r.status === 401 });

    res = http.post(BASE_URL, defaultPayload, {
        headers: { 'Authorization': 'Bearer wrongtoken', 'Content-Type': 'application/json' },
    });
    check(res, { 'bad token → 401': (r) => r.status === 401 });
}
