Got it 👍 Here’s the English version of the README I drafted for you:

````markdown
# WOD Generator API

A **Go** API that generates reproducible Workouts of the Day (WODs) based on user level, duration, available equipment, and an optional generation seed.

## 🚀 Features

- Generates structured WODs as blocks (movement + parameters).
- Supports 3 levels: `beginner`, `intermediate`, `advanced`.
- Configurable duration between **15 and 120 minutes**.
- Takes available equipment into account (falls back to bodyweight moves if none).
- Deterministic results with a `seed` (re-run the same WOD).
- Secured API: **Bearer token authentication** + **rate limiting**.
- Healthchecks available (`/healthz`, `/readyz`).

## 📦 Installation & Run

```bash
git clone https://github.com/youruser/wod-gen.git
cd wod-gen

# run lint & checks
make lint

# run locally (default port: 8080)
go run ./cmd/wod-gen
````

Main environment variables (see `internal/config`):

* `HTTP_PORT`: HTTP port (e.g. `8080`)
* `AUTH_API_KEYS`: list of API keys (hashed)
* `RATE_LIMIT_STRATEGY`: rate limiting strategy

## 🔌 Endpoints

### `POST /api/v1/wod/generate`

Generate a WOD.

**Example request:**

```bash
curl -X POST http://localhost:8080/api/v1/wod/generate \
  -H "Authorization: Bearer <API_KEY>" \
  -H "Content-Type: application/json" \
  -d '{
    "level": "intermediate",
    "duration_min": 45,
    "equipment": ["rower", "dumbbell"],
    "seed": "123c10ba-9ab6-45b9-acc9-56a16b488482"
  }'
```

**Example response:**

```json
{
  "id": "1e89b9ed-b4a7-4cee-9b13-89a88e0a3642",
  "seed": "123c10ba-9ab6-45b9-acc9-56a16b488482",
  "created_at": "2025-09-07T00:38:00Z",
  "duration_min": 45,
  "level": "intermediate",
  "generator_version": "v1",
  "equipment": ["rower", "dumbbell"],
  "blocks": [
    {"name": "Run", "params": {"meters": 900}},
    {"name": "Push-ups", "params": {"reps": 15}},
    {"name": "Burpees Broad Jump", "params": {"meters": 10}}
  ]
}
```

## ⚙️ Development

- **Language & Framework**
    - Go (1.21+)
    - HTTP framework: [Gin](https://gin-gonic.com/)

- **API & Contracts**
    - OpenAPI spec in `openapi.yml`
    - Code generated via [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen)

- **Configuration**
    - Environment variables parsed with [caarlos0/env](https://github.com/caarlos0/env)
    - Struct validation with [go-playground/validator](https://github.com/go-playground/validator)

- **Core logic**
    - Movement catalog defined in YAML (`internal/core/catalog.yml`)
    - Randomized but reproducible workout generation (seeded RNG)
    - Equipment-aware filtering with fallback to bodyweight-only moves
    - Duration-based block allocation (`blocksForDuration`)

- **Security**
    - Bearer token authentication (API keys, hashed)
    - Rate limiting middleware with configurable strategy
    - Security headers (HSTS, X-Frame-Options, etc.)

- **Observability**
    - Structured logging via Go `slog`
    - Request IDs injected into logs
    - Panic recovery middleware

- **Resilience**
    - Graceful shutdown on `SIGINT`/`SIGTERM`
    - Request timeouts
    - Body size limits
    - Health (`/healthz`) and readiness (`/readyz`) endpoints

- **Quality & Tooling**
    - Linting with `golangci-lint`
    - Security scanning with `govulncheck`
    - Static error checking with `errcheck`
    - Tests with race detector and coverage reporting

✨ A project to automatically generate varied and balanced workouts!

