import http from "k6/http";
import { check } from "k6";
import { BASE_URL, headers } from "./config/config.js";


export default function () {
    const res = http.get(`${BASE_URL.replace("/wod/generate", "/wod/list")}?limit=5&offset=0`, { headers });

    check(res, {
        "status is 200": (r) => r.status === 200,
        "has wods": (r) => Array.isArray(r.json("wods")),
        "pagination respected": (r) => r.json("wods").length <= 5,
    });
}