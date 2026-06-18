# signalFlow

Real-time energy event processing system. Ingests meter readings from distributed
energy assets (solar farms, wind turbines, battery storage), processes them
asynchronously via NATS JetStream, persists to ClickHouse, and automatically
detects imbalance risk when actual production deviates more than 20% below expected.

## How it works

Two separate processes run at the same time:

```
Meter Device
     │
     │  POST /readings  (with JWT token)
     ▼
┌──────────────────────────────────┐
│  API Process (:8080)             │
│                                  │
│  POST /token    → issues JWT     │
│  POST /assets   → PostgreSQL     │
│  POST /readings → NATS JetStream │  ← responds immediately, does not wait
│  GET  /alerts   ← PostgreSQL     │
│  GET  /health                    │
│  GET  /metrics  (Prometheus)     │
│  GET  /swagger/                  │
└──────────────┬───────────────────┘
               │ publishes event to NATS JetStream
               ▼
┌──────────────────────────────────┐
│  Consumer Process                │
│                                  │
│  1. Pull message from NATS       │
│  2. Validate asset exists        │
│  3. Write reading → ClickHouse   │
│  4. ACK message (at-least-once)  │
│  5. If actual < 80% of expected: │
│       Write alert → PostgreSQL   │
└──────────────────────────────────┘
               │
     ┌─────────┴──────────┐
     ▼                    ▼
 PostgreSQL           ClickHouse
 assets, alerts       readings (time series)
```

The API never blocks — it accepts a reading and returns `202` in milliseconds.
The consumer processes it asynchronously. If the consumer restarts, NATS
re-delivers any unacknowledged messages so no readings are lost.

## Tech Stack

| Layer | Technology | Why |
|---|---|---|
| Language | Go 1.25 | Concurrency, performance |
| HTTP Router | chi | Lightweight, middleware-friendly |
| Message Broker | NATS JetStream | Durable queue, at-least-once delivery |
| Operational DB | PostgreSQL | Assets and alerts (transactional) |
| Analytics DB | ClickHouse | Readings time series (high volume) |
| Auth | JWT (HS256) | Stateless token validation |
| Logging | zerolog | Structured JSON logs |
| Metrics | Prometheus | HTTP counters, durations |
| Migrations | Atlas | SQL schema migrations |
| Infrastructure | Docker Compose | Local dev environment |

## Prerequisites

- Go 1.25+
- Docker + Docker Compose
- `atlas` CLI — [install](https://atlasgo.io/getting-started)
- `golangci-lint` — [install](https://golangci-lint.run/usage/install/) (optional, for linting)

## Quick Start

```bash
# 1. Clone and enter the project
git clone <repo-url> signalFlow && cd signalFlow

# 2. Copy environment config
cp .env.example .env
# Edit .env and set JWT_SECRET and API_KEY (see Environment Variables below)

# 3. Start infrastructure (PostgreSQL, ClickHouse, NATS)
make docker-up

# 4. Run database migrations
export $(cat .env | xargs)
make migrate

# 5. Start API process (terminal 1)
make run-api

# 6. Start consumer process (terminal 2)
make run-consumer

# 7. Verify everything works
./scripts/test-flow.sh
```

## Environment Variables

Copy `.env.example` to `.env`. The defaults work for local development except
`JWT_SECRET` and `API_KEY` — set those to any strings you choose.

| Variable | Default | Description |
|---|---|---|
| `API_PORT` | `8080` | API server port |
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/signalflow?sslmode=disable` | PostgreSQL DSN |
| `CLICKHOUSE_URL` | `clickhouse://localhost:9000/signalflow` | ClickHouse DSN |
| `NATS_URL` | `nats://localhost:4222` | NATS server address |
| `NATS_STREAM` | `READINGS` | JetStream stream name |
| `NATS_SUBJECT` | `readings.ingest` | JetStream subject |
| `IMBALANCE_THRESHOLD` | `0.20` | Alert when actual output is this % below expected |
| `JWT_SECRET` | `dev-secret-change-in-production` | Signs and verifies JWT tokens — change in production |
| `API_KEY` | `dev-api-key` | Pre-shared key used to obtain a JWT token |

## API Endpoints

### Public (no token required)

| Method | Path | Description |
|---|---|---|
| `POST` | `/token` | Exchange an API key for a JWT token |
| `GET` | `/health` | Health check — PostgreSQL and NATS connectivity |
| `GET` | `/metrics` | Prometheus metrics |
| `GET` | `/swagger/` | Interactive API docs |

### Protected (JWT required)

All protected endpoints require:
```
Authorization: Bearer <your-token>
```

| Method | Path | Description |
|---|---|---|
| `POST` | `/assets` | Register a new energy asset |
| `POST` | `/readings` | Submit a meter reading (returns `202` immediately) |
| `GET` | `/alerts` | List imbalance alerts, optionally filtered by `?asset_id=` |

## Full Usage Walkthrough

### Step 1 — Get a token

```bash
curl -s -X POST localhost:8080/token \
  -H "Content-Type: application/json" \
  -d '{"api_key":"dev-api-key"}' | jq
```

Response:
```json
{ "token": "eyJhbGci..." }
```

Save it:
```bash
TOKEN="eyJhbGci..."
```

Tokens expire after 24 hours. Request a new one when needed.

### Step 2 — Create an asset

```bash
curl -s -X POST localhost:8080/assets \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"Solar Farm A","expected_output":100}' | jq
```

Response:
```json
{
  "id": "a1b2c3d4-...",
  "name": "Solar Farm A",
  "expected_output": 100
}
```

Save the `id`:
```bash
ASSET_ID="a1b2c3d4-..."
```

### Step 3 — Submit a healthy reading (no alert)

Actual output is 90kW — above 80% of the 100kW expected output.

```bash
curl -s -X POST localhost:8080/readings \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"asset_id\": \"$ASSET_ID\",
    \"actual_output\": 90,
    \"recorded_at\": \"$(date -u +%FT%TZ)\"
  }"
```

Returns `202 Accepted` immediately. The consumer processes it in the background.

### Step 4 — Submit an imbalance reading (triggers alert)

Actual output is 50kW — below 80% of the 100kW expected output.

```bash
curl -s -X POST localhost:8080/readings \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"asset_id\": \"$ASSET_ID\",
    \"actual_output\": 50,
    \"recorded_at\": \"$(date -u +%FT%TZ)\"
  }"
```

### Step 5 — Check alerts

Wait 1–2 seconds for the consumer to process, then:

```bash
curl -s "localhost:8080/alerts?asset_id=$ASSET_ID" \
  -H "Authorization: Bearer $TOKEN" | jq
```

You should see one alert with `deviation_pct` showing how far below threshold the reading was.

### What happens without a token

```bash
curl -s localhost:8080/alerts
# → 401 Unauthorized
```

## Running Tests

```bash
make test         # all unit tests (skips DB/NATS tests if infra is down)
make test-race    # same, with Go race detector enabled
make lint         # golangci-lint
```

End-to-end test (requires running stack):
```bash
./scripts/test-flow.sh
```

## Make Targets

```bash
make docker-up    # start PostgreSQL, ClickHouse, NATS
make docker-down  # stop and remove containers
make run-api      # start API process
make run-consumer # start consumer process
make migrate      # apply database migrations
make build        # compile both binaries to bin/
make test         # run all tests
make swagger      # regenerate Swagger docs from annotations
```

## Project Structure

```
signalFlow/
├── cmd/
│   ├── api/main.go           # API process entrypoint
│   └── consumer/main.go      # Consumer process entrypoint
├── config/
│   └── config.go             # All env var loading in one place
├── internal/
│   ├── api/
│   │   ├── handlers/         # One file per endpoint
│   │   ├── middleware/        # Auth, metrics, request ID
│   │   └── router.go         # Route registration
│   ├── consumer/
│   │   └── consumer.go       # NATS subscriber + imbalance logic
│   ├── db/
│   │   ├── queries/          # Raw SQL query functions
│   │   ├── postgres.go       # PostgreSQL connection
│   │   └── clickhouse.go     # ClickHouse connection
│   └── models/
│       └── models.go         # Shared structs (Asset, Reading, Alert)
├── docker-compose.yml        # Local infrastructure
├── Makefile                  # Common tasks
└── .env.example              # Environment variable template
```
