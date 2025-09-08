// tests/config.js
const prod = 'https://wod-gen.fly.dev'
const local = 'http://localhost:8080'
export const BASE_URL = __ENV.BASE_URL || "http://localhost:8080/api/v1/wod/generate";
export const TOKEN = __ENV.TOKEN;

export const headers = {
    "Authorization": `Bearer ${TOKEN}`,
    "Content-Type": "application/json",
};

export const defaultPayload = JSON.stringify({
    level: "beginner",
    duration_min: 30,
    equipment: ['rower'],
});